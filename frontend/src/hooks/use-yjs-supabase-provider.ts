'use client'

import { useEffect, useState, useRef, useCallback } from 'react'
import { createClient } from '@/lib/supabase/client'
import * as Y from 'yjs'
import { RealtimeChannel } from '@supabase/supabase-js'
import { Awareness, encodeAwarenessUpdate, applyAwarenessUpdate } from 'y-protocols/awareness'

const supabase = createClient()

export type ProviderStatus = 'connecting' | 'connected' | 'disconnected' | 'error'

// Custom Y.js provider using Supabase Realtime
class SupabaseYjsProvider {
  public awareness: Awareness
  public doc: Y.Doc
  public document: Y.Doc // Alias for Tiptap compatibility
  public configuration: { doc: Y.Doc; awareness: Awareness } // For CollaborationCursor compatibility
  public channelName: string
  public onStatusChange: (status: ProviderStatus) => void
  public onSyncedChange: (synced: boolean) => void
  private channel: RealtimeChannel | null = null
  private connected = false

  constructor(
    doc: Y.Doc,
    channelName: string,
    onStatusChange: (status: ProviderStatus) => void,
    onSyncedChange: (synced: boolean) => void
  ) {
    this.doc = doc
    this.document = doc // Set both for compatibility
    this.awareness = new Awareness(doc)
    this.configuration = { doc: this.doc, awareness: this.awareness } // Set configuration
    this.channelName = channelName
    this.onStatusChange = onStatusChange
    this.onSyncedChange = onSyncedChange
    this.setupProvider()
  }

  private setupProvider() {
    console.log('[Y.js Provider] Setting up channel:', this.channelName)
    this.onStatusChange('connecting')

    // Create Supabase Realtime channel
    this.channel = supabase.channel(this.channelName, {
      config: {
        broadcast: { ack: false, self: true },
      },
    })

    // Listen for Y.js updates from other clients
    this.channel.on('broadcast', { event: 'yjs-update' }, ({ payload }) => {
      console.log('[Y.js Provider] Received yjs-update from broadcast:', payload)
      if (payload.update) {
        const update = new Uint8Array(payload.update)
        console.log('[Y.js Provider] Applying remote update, size:', update.length)
        Y.applyUpdate(this.doc, update, 'remote')
      }
    })

    // Listen for awareness updates (cursor positions, user info)
    this.channel.on('broadcast', { event: 'yjs-awareness' }, ({ payload }) => {
      console.log('[Y.js Provider] Received yjs-awareness from broadcast:', payload)
      if (payload.update) {
        const update = new Uint8Array(payload.update)
        console.log('[Y.js Provider] Applying awareness update')
        applyAwarenessUpdate(this.awareness, update, 'remote')
      }
    })

    // Handle doc updates - broadcast to other clients
    this.doc.on('update', (update: Uint8Array, origin: any) => {
      console.log('[Y.js Provider] Document update:', { 
        origin, 
        connected: this.connected,
        updateSize: update.length 
      })
      
      // Don't broadcast updates that came from remote
      if (origin !== 'remote' && this.connected) {
        console.log('[Y.js Provider] Broadcasting update to other clients')
        this.channel?.send({
          type: 'broadcast',
          event: 'yjs-update',
          payload: { update: Array.from(update) },
        })
      }
    })

    // Handle awareness updates - broadcast cursor/user info
    this.awareness.on('update', ({ added, updated, removed }: any, origin: any) => {
      console.log('[Y.js Provider] Awareness update:', { 
        added, 
        updated, 
        removed, 
        origin,
        connected: this.connected 
      })
      
      if (origin !== 'remote' && this.connected) {
        const changedClients = added.concat(updated).concat(removed)
        const update = encodeAwarenessUpdate(this.awareness, changedClients)
        console.log('[Y.js Provider] Broadcasting awareness update')
        this.channel?.send({
          type: 'broadcast',
          event: 'yjs-awareness',
          payload: { update: Array.from(update) },
        })
      }
    })

    // Subscribe to channel
    this.channel.subscribe((status) => {
      console.log('[Y.js Provider] Channel status:', status)
      
      if (status === 'SUBSCRIBED') {
        this.connected = true
        this.onStatusChange('connected')
        
        // Send initial sync
        setTimeout(() => {
          this.onSyncedChange(true)
        }, 100)
      } else if (status === 'CHANNEL_ERROR') {
        this.connected = false
        this.onStatusChange('error')
      } else if (status === 'TIMED_OUT') {
        this.connected = false
        this.onStatusChange('error')
      } else if (status === 'CLOSED') {
        this.connected = false
        this.onStatusChange('disconnected')
      }
    })
  }

  destroy() {
    console.log('[Y.js Provider] Destroying provider')
    if (this.channel) {
      this.channel.unsubscribe()
      this.channel = null
    }
    this.connected = false
    this.awareness.destroy()
  }
}

export const useYjsSupabaseProvider = (noteId: string, enabled: boolean = true) => {
  const [status, setStatus] = useState<ProviderStatus>('disconnected')
  const [isSynced, setIsSynced] = useState(false)
  const ydocRef = useRef<Y.Doc | null>(null)
  const providerRef = useRef<SupabaseYjsProvider | null>(null)

  // Initialize Y.Doc once
  if (!ydocRef.current) {
    ydocRef.current = new Y.Doc()
  }

  const handleStatusChange = useCallback((newStatus: ProviderStatus) => {
    setStatus(newStatus)
  }, [])

  const handleSyncedChange = useCallback((synced: boolean) => {
    setIsSynced(synced)
  }, [])

  useEffect(() => {
    if (!noteId || !enabled) {
      return
    }

    const ydoc = ydocRef.current!
    
    console.log('[Y.js Provider] Initializing for note:', noteId)

    // Create custom Supabase Y.js provider
    const provider = new SupabaseYjsProvider(
      ydoc,
      `note-collab-${noteId}`,
      handleStatusChange,
      handleSyncedChange
    )

    providerRef.current = provider

    // Cleanup
    return () => {
      console.log('[Y.js Provider] Cleaning up')
      if (providerRef.current) {
        providerRef.current.destroy()
        providerRef.current = null
      }
    }
  }, [noteId, enabled, handleStatusChange, handleSyncedChange])

  return {
    ydoc: ydocRef.current,
    provider: providerRef.current,
    awareness: providerRef.current?.awareness,
    status,
    isSynced,
  }
}

