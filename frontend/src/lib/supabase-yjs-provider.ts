import * as Y from 'yjs'
import { Awareness, encodeAwarenessUpdate, applyAwarenessUpdate } from 'y-protocols/awareness'
import { createClient } from '@/lib/supabase/client'
import { RealtimeChannel, REALTIME_SUBSCRIBE_STATES } from '@supabase/supabase-js'
import axios from 'axios'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'

/**
 * Custom Yjs provider that uses Supabase Realtime for synchronization
 * Implements conflict-free real-time collaboration using Yjs CRDT
 */
export class SupabaseYjsProvider {
  public doc: Y.Doc
  public awareness: Awareness
  private noteId: string
  private channel: RealtimeChannel | null = null
  private supabase = createClient()
  private synced = false
  private getAuthToken: () => Promise<string | null>
  
  // Update buffering for backend saves
  private updateBuffer: Uint8Array[] = []
  private saveTimer: NodeJS.Timeout | null = null
  private readonly SAVE_DEBOUNCE_MS = 10000 // Save every 10 seconds
  
  // Retry logic for failed saves
  private failedUpdates: Uint8Array[] = []
  private retrying = false
  
  // Store initial JSON content for editor initialization
  public initialContent: string | null = null
  
  // Awareness heartbeat to let others know we're active
  private awarenessUpdateInterval: NodeJS.Timeout | null = null
  private readonly AWARENESS_HEARTBEAT_MS = 30000 // 30 seconds
  
  // Event callbacks
  private onSyncedCallback?: () => void
  private onStatusCallback?: (status: { status: string }) => void
  private onErrorCallback?: (error: Error) => void

  constructor(noteId: string, doc: Y.Doc, user: any, getAuthToken: () => Promise<string | null>) {
    this.noteId = noteId
    this.doc = doc
    this.getAuthToken = getAuthToken
    
    // Initialize awareness for cursor/selection tracking
    this.awareness = new Awareness(this.doc)
    this.awareness.setLocalState({
      user: {
        name: user?.fullName || 'Anonymous',
        clerkId: user?.id,
        color: this.generateUserColor(user?.id || 'default'),
        image: user?.imageUrl,
      },
    })
    
    // Initialize connection
    this.initialize()
  }

  private async initialize() {
    try {
      // Fetch initial document state
      await this.fetchInitialState()
      
      // Setup Supabase Realtime channel
      this.setupRealtimeChannel()
      
      // Listen for local changes
      this.setupLocalChangeListener()
      
    } catch (error) {
      console.error('[SupabaseYjsProvider] Initialization error:', error)
      this.onErrorCallback?.(error as Error)
    }
  }

  private async fetchInitialState() {
    try {
      const token = await this.getAuthToken()
      
      console.log(`[SupabaseYjsProvider] Fetching initial state for note: ${this.noteId}`)
      
      const response = await axios.get(
        `${API_BASE_URL}/note/${this.noteId}/yjs-state`,
        {
          responseType: 'arraybuffer',
          headers: token ? { Authorization: `Bearer ${token}` } : {},
          withCredentials: true,
        }
      )
      
      // Check if response indicates initialization is needed
      if (response.headers['content-type']?.includes('application/json')) {
        // Parse JSON response
        const jsonData = JSON.parse(new TextDecoder().decode(response.data))
        
        if (jsonData.requiresInit) {
          await this.initializeFromJSON(jsonData.noteContent)
          return
        }
      }
      
      // Apply binary Yjs state
      if (response.data && response.data.byteLength > 0) {
        const update = new Uint8Array(response.data)
        Y.applyUpdate(this.doc, update)
      }
      
      this.synced = true
      this.onSyncedCallback?.()
      
    } catch (error: any) {
      console.error(`[SupabaseYjsProvider] Failed to fetch initial state for note ${this.noteId}:`, error)
      console.error(`[SupabaseYjsProvider] Error status: ${error.response?.status}, Message: ${error.message}`)
      
      // If 403, it's likely a permission issue - retry after delay
      if (error.response?.status === 403) {
        console.warn(`[SupabaseYjsProvider] Access denied (403) for note ${this.noteId}, retrying in 2 seconds...`)
        await new Promise(resolve => setTimeout(resolve, 2000))
        try {
          // Retry once
          await this.fetchInitialState()
          return
        } catch (retryError) {
          console.error(`[SupabaseYjsProvider] Retry failed for note ${this.noteId}:`, retryError)
        }
      }
      
      // If 404 or note doesn't exist, initialize empty
      if (error.response?.status === 404) {
        console.log(`[SupabaseYjsProvider] Note ${this.noteId} not found (404), initializing empty`)
        this.synced = true
        this.onSyncedCallback?.()
      } else {
        throw error
      }
    }
  }

  private async initializeFromJSON(jsonContent: string) {
    try {
      // The content will be set by the editor when onCreate is called
      // We just store it in initialContent for now
      // Tiptap's Collaboration extension will handle converting it to Yjs format
      
      // Store the JSON content for editor to use
      this.initialContent = jsonContent
      
      // Mark as synced
      this.synced = true
      this.onSyncedCallback?.()
      
    } catch (error) {
      console.error('[SupabaseYjsProvider] Failed to initialize from JSON:', error)
      throw error
    }
  }
  
  // Method to be called by the editor after it has loaded the initial content
  // Note: The editor now waits for Yjs to sync before calling this, so we just validate and send
  public async markInitialized() {
    try {
      const token = await this.getAuthToken()
      const initialState = Y.encodeStateAsUpdate(this.doc)
      
      // Final safety check that we have actual content (not just empty doc)
      // Empty Yjs doc is typically 2 bytes: [0, 0]
      if (!initialState || initialState.length <= 2) {
        console.warn(`[SupabaseYjsProvider] Cannot initialize - Yjs document is empty (${initialState?.length || 0} bytes)`)
        console.warn(`[SupabaseYjsProvider] Editor should have waited for content. Check NoteEditor initialization.`)
        return
      }
      
      console.log(`[SupabaseYjsProvider] Marking document as initialized (${initialState.length} bytes)`)
      
      await axios.post(
        `${API_BASE_URL}/note/${this.noteId}/yjs-init`,
        initialState,
        {
          headers: {
            'Content-Type': 'application/octet-stream',
            ...(token ? { Authorization: `Bearer ${token}` } : {}),
          },
          withCredentials: true,
        }
      )
      
      console.log(`[SupabaseYjsProvider] Successfully marked document as initialized`)
      
    } catch (error) {
      console.error('[SupabaseYjsProvider] Failed to mark document as initialized:', error)
      throw error
    }
  }

  private setupRealtimeChannel() {
    const channelName = `yjs-sync-${this.noteId}`
    
    this.channel = this.supabase.channel(channelName, {
      config: {
        broadcast: {
          self: false, // Don't receive our own broadcasts
        },
      },
    })
    
    // Listen for updates from other clients
    this.channel
      .on('broadcast', { event: 'yjs-update' }, (payload) => {
        this.handleRemoteUpdate(payload.payload)
      })
      .on('broadcast', { event: 'awareness-update' }, (payload) => {
        this.handleRemoteAwarenessUpdate(payload.payload)
      })
      .subscribe((status) => {
        
        if (status === REALTIME_SUBSCRIBE_STATES.SUBSCRIBED) {
          this.onStatusCallback?.({ status: 'connected' })
          
          // Start awareness heartbeat
          this.startAwarenessHeartbeat()
        } else if (status === REALTIME_SUBSCRIBE_STATES.CHANNEL_ERROR) {
          console.error('[SupabaseYjsProvider] Channel error')
          this.onStatusCallback?.({ status: 'error' })
        } else if (status === REALTIME_SUBSCRIBE_STATES.TIMED_OUT) {
          console.error('[SupabaseYjsProvider] Channel timed out')
          this.onStatusCallback?.({ status: 'disconnected' })
        }
      })
  }

  private startAwarenessHeartbeat() {
    // Clear any existing interval
    if (this.awarenessUpdateInterval) {
      clearInterval(this.awarenessUpdateInterval)
    }
    
    // Send periodic awareness updates to let others know we're active
    this.awarenessUpdateInterval = setInterval(() => {
      const localClientId = this.awareness.clientID
      this.broadcastAwarenessUpdate([localClientId])
    }, this.AWARENESS_HEARTBEAT_MS)
    
    // Send initial awareness update
    const localClientId = this.awareness.clientID
    this.broadcastAwarenessUpdate([localClientId])
  }

  private setupLocalChangeListener() {
    // Listen for local document updates
    this.doc.on('update', (update: Uint8Array, origin: any) => {
      // Ignore updates from remote sources
      if (origin === this) return
      
      // Broadcast to other clients via Supabase
      this.broadcastUpdate(update)
      
      // Buffer update for backend save
      this.bufferUpdateForBackend(update)
    })
    
    // Listen for local awareness changes (cursor position, selection, etc.)
    this.awareness.on('update', ({ added, updated, removed }: any) => {
      const changedClients = added.concat(updated).concat(removed)
      this.broadcastAwarenessUpdate(changedClients)
    })
  }

  private broadcastUpdate(update: Uint8Array) {
    if (!this.channel) return
    
    // Convert Uint8Array to base64 for JSON transport
    const base64Update = this.uint8ArrayToBase64(update)
    
    this.channel.send({
      type: 'broadcast',
      event: 'yjs-update',
      payload: { update: base64Update },
    })
  }

  private handleRemoteUpdate(payload: any) {
    try {
      // Convert base64 back to Uint8Array
      const update = this.base64ToUint8Array(payload.update)
      
      // Apply the update with 'this' as origin to prevent echo
      Y.applyUpdate(this.doc, update, this)
      
    } catch (error) {
      console.error('[SupabaseYjsProvider] Error applying remote update:', error)
    }
  }

  private broadcastAwarenessUpdate(changedClients: number[]) {
    if (!this.channel) return
    
    try {
      // Encode awareness update using y-protocols
      const update = encodeAwarenessUpdate(this.awareness, changedClients)
      const base64Update = this.uint8ArrayToBase64(update)
      
      // Broadcast awareness to other clients
      this.channel.send({
        type: 'broadcast',
        event: 'awareness-update',
        payload: { 
          update: base64Update,
          clientId: this.awareness.clientID 
        },
      })
    } catch (error) {
      console.error('[SupabaseYjsProvider] Error broadcasting awareness:', error)
    }
  }

  private handleRemoteAwarenessUpdate(payload: any) {
    try {
      const { update: base64Update, clientId } = payload
      
      // Don't apply our own awareness updates
      if (clientId === this.awareness.clientID) return
      
      // Decode and apply awareness update
      const update = this.base64ToUint8Array(base64Update)
      applyAwarenessUpdate(this.awareness, update, this)
      
    } catch (error) {
      console.error('[SupabaseYjsProvider] Error applying remote awareness:', error)
    }
  }

  private bufferUpdateForBackend(update: Uint8Array) {
    this.updateBuffer.push(update)
    
    // Reset save timer
    if (this.saveTimer) {
      clearTimeout(this.saveTimer)
    }
    
    this.saveTimer = setTimeout(() => {
      this.flushUpdatesToBackend()
    }, this.SAVE_DEBOUNCE_MS)
  }

  private async flushUpdatesToBackend() {
    if (this.updateBuffer.length === 0) return
    
    try {
      // Merge all buffered updates (we'll send complete state instead)
      Y.mergeUpdates(this.updateBuffer)
      this.updateBuffer = []
      
      // Get the complete state after applying updates
      const completeState = Y.encodeStateAsUpdate(this.doc)
      
      // Get auth token
      const token = await this.getAuthToken()
      
      console.log(`[SupabaseYjsProvider] Flushing updates to backend for note: ${this.noteId}`)
      
      // Send to backend
      await axios.post(
        `${API_BASE_URL}/note/${this.noteId}/yjs-update`,
        completeState,
        {
          headers: {
            'Content-Type': 'application/octet-stream',
            ...(token ? { Authorization: `Bearer ${token}` } : {}),
          },
          withCredentials: true,
        }
      )
      
      console.log(`[SupabaseYjsProvider] Successfully flushed updates for note: ${this.noteId}`)
      
      // Retry any failed updates
      if (this.failedUpdates.length > 0 && !this.retrying) {
        this.retryFailedUpdates()
      }
      
    } catch (error: any) {
      console.error(`[SupabaseYjsProvider] Failed to flush updates for note ${this.noteId}:`, error)
      console.error(`[SupabaseYjsProvider] Error status: ${error.response?.status}, Message: ${error.message}`)
      
      // If 403, log detailed info
      if (error.response?.status === 403) {
        console.error(`[SupabaseYjsProvider] Permission denied (403) when trying to update note ${this.noteId}`)
        console.error('[SupabaseYjsProvider] This may indicate the note was created but access permissions are not set correctly')
        // Notify the error callback
        this.onErrorCallback?.(new Error(`Permission denied when updating note ${this.noteId}`))
      }
      
      // Queue for retry
      const mergedUpdate = Y.mergeUpdates(this.updateBuffer)
      this.failedUpdates.push(mergedUpdate)
      this.updateBuffer = []
    }
  }

  private async retryFailedUpdates() {
    if (this.failedUpdates.length === 0 || this.retrying) return
    
    this.retrying = true
    
    while (this.failedUpdates.length > 0) {
      const update = this.failedUpdates[0]
      
      try {
        const token = await this.getAuthToken()
        
        await axios.post(
          `${API_BASE_URL}/note/${this.noteId}/yjs-update`,
          update,
          {
            headers: {
              'Content-Type': 'application/octet-stream',
              ...(token ? { Authorization: `Bearer ${token}` } : {}),
            },
            withCredentials: true,
          }
        )
        
        // Success, remove from failed queue
        this.failedUpdates.shift()
        
      } catch (error) {
        console.error('[SupabaseYjsProvider] Retry failed:', error)
        // Wait before next retry
        await new Promise(resolve => setTimeout(resolve, 5000))
        break
      }
    }
    
    this.retrying = false
  }

  // Utility functions for base64 conversion
  private uint8ArrayToBase64(array: Uint8Array): string {
    return btoa(String.fromCharCode(...array))
  }

  private base64ToUint8Array(base64: string): Uint8Array {
    const binaryString = atob(base64)
    const len = binaryString.length
    const bytes = new Uint8Array(len)
    for (let i = 0; i < len; i++) {
      bytes[i] = binaryString.charCodeAt(i)
    }
    return bytes
  }

  // Generate a consistent color for a user
  private generateUserColor(userId: string): string {
    const colors = [
      '#FF6B6B', '#4ECDC4', '#45B7D1', '#FFA07A',
      '#98D8C8', '#F7DC6F', '#BB8FCE', '#85C1E2',
      '#F8B739', '#52B788', '#E07A5F', '#81B29A',
    ]
    
    // Use user ID to generate consistent color
    let hash = 0
    for (let i = 0; i < userId.length; i++) {
      hash = userId.charCodeAt(i) + ((hash << 5) - hash)
    }
    
    return colors[Math.abs(hash) % colors.length]
  }

  // Event handlers
  public on(event: 'synced', callback: () => void): void
  public on(event: 'status', callback: (status: { status: string }) => void): void
  public on(event: 'error', callback: (error: Error) => void): void
  public on(event: string, callback: any): void {
    switch (event) {
      case 'synced':
        this.onSyncedCallback = callback
        if (this.synced) callback()
        break
      case 'status':
        this.onStatusCallback = callback
        break
      case 'error':
        this.onErrorCallback = callback
        break
    }
  }

  // Cleanup
  public destroy() {
    // Flush any pending updates
    if (this.updateBuffer.length > 0) {
      this.flushUpdatesToBackend()
    }
    
    // Clear timers
    if (this.saveTimer) {
      clearTimeout(this.saveTimer)
    }
    
    if (this.awarenessUpdateInterval) {
      clearInterval(this.awarenessUpdateInterval)
      this.awarenessUpdateInterval = null
    }
    
    // Set local awareness state to null to signal we're leaving
    this.awareness.setLocalState(null)
    
    // Broadcast final awareness update
    const localClientId = this.awareness.clientID
    this.broadcastAwarenessUpdate([localClientId])
    
    // Unsubscribe from channel
    if (this.channel) {
      this.channel.unsubscribe()
      this.channel = null
    }
    
    // Destroy awareness
    this.awareness.destroy()
  }

  // Manual sync trigger
  public async sync() {
    await this.flushUpdatesToBackend()
  }

  // Get connection status
  public get connected(): boolean {
    return this.channel !== null && this.synced
  }
}

