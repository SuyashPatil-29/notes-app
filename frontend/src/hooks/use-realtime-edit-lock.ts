'use client'

import { useCurrentUserImage } from '@/hooks/use-current-user-image'
import { useCurrentUserName } from '@/hooks/use-current-user-name'
import { useCurrentUserId } from '@/hooks/use-current-user-id'
import { createClient } from '@/lib/supabase/client'
import { REALTIME_SUBSCRIBE_STATES, RealtimeChannel } from '@supabase/supabase-js'
import { useEffect, useState, useRef, useCallback } from 'react'

const supabase = createClient()

const HEARTBEAT_INTERVAL = 5000 // 5 seconds
const STALE_LOCK_TIMEOUT = 30000 // 30 seconds

export type EditorInfo = {
  clerkId: string
  name: string
  image: string
  editingStartedAt: number
  lastHeartbeat: number
}

export type CurrentEditor = {
  clerkId: string
  name: string
  image: string
} | null

export const useRealtimeEditLock = (roomName: string) => {
  const currentUserImage = useCurrentUserImage()
  const currentUserName = useCurrentUserName()
  const currentUserId = useCurrentUserId()

  const [currentEditor, setCurrentEditor] = useState<CurrentEditor>(null)
  const [isEditing, setIsEditing] = useState(false)
  const [canEdit, setCanEdit] = useState(true)
  
  const channelRef = useRef<RealtimeChannel | null>(null)
  const heartbeatIntervalRef = useRef<NodeJS.Timeout | null>(null)

  // Clean up stale locks
  const checkStaleLocks = useCallback((presenceState: Record<string, EditorInfo[]>) => {
    const now = Date.now()
    const editors = Object.values(presenceState).flat()
    
    console.log('[EditLock] Checking presence state:', editors)
    
    // Filter out stale locks (no heartbeat for 30+ seconds)
    const activeEditors = editors.filter(
      (editor) => now - editor.lastHeartbeat < STALE_LOCK_TIMEOUT
    )
    
    console.log('[EditLock] Active editors:', activeEditors)
    
    if (activeEditors.length === 0) {
      console.log('[EditLock] No active editors')
      setCurrentEditor(null)
      setCanEdit(true)
    } else {
      const editor = activeEditors[0]
      console.log('[EditLock] Current editor:', editor.name, 'Current user:', currentUserId)
      setCurrentEditor({
        clerkId: editor.clerkId,
        name: editor.name,
        image: editor.image,
      })
      // Can edit if no editor or if you are the editor
      setCanEdit(!editor || editor.clerkId === currentUserId)
    }
  }, [currentUserId])

  // Request edit lock
  const requestEdit = useCallback(async () => {
    if (!currentUserId || !currentUserName || !channelRef.current) {
      console.log('[EditLock] Cannot request edit - missing user info or channel')
      return false
    }

    // Check actual presence state instead of local state to avoid race conditions
    const presenceState = channelRef.current.presenceState<EditorInfo>()
    const now = Date.now()
    const editors = Object.values(presenceState).flat()
    
    // Filter out stale locks
    const activeEditors = editors.filter(
      (editor) => now - editor.lastHeartbeat < STALE_LOCK_TIMEOUT
    )
    
    // Check if someone else is actively editing
    const otherEditor = activeEditors.find(editor => editor.clerkId !== currentUserId)
    if (otherEditor) {
      console.log('[EditLock] Cannot edit - someone else is editing:', otherEditor.name)
      return false
    }

    const editorInfo: EditorInfo = {
      clerkId: currentUserId,
      name: currentUserName,
      image: currentUserImage || '',
      editingStartedAt: now,
      lastHeartbeat: now,
    }

    try {
      console.log('[EditLock] Claiming edit lock:', currentUserName)
      await channelRef.current.track(editorInfo)
      setIsEditing(true)
      
      // Start heartbeat
      if (heartbeatIntervalRef.current) {
        clearInterval(heartbeatIntervalRef.current)
      }
      
      heartbeatIntervalRef.current = setInterval(() => {
        if (channelRef.current) {
          channelRef.current.track({
            ...editorInfo,
            lastHeartbeat: Date.now(),
          })
        }
      }, HEARTBEAT_INTERVAL)
      
      return true
    } catch (error) {
      console.error('[EditLock] Failed to request edit lock:', error)
      return false
    }
  }, [currentUserId, currentUserName, currentUserImage])

  // Release edit lock
  const releaseEdit = useCallback(async () => {
    if (!channelRef.current) {
      return
    }

    try {
      console.log('[EditLock] Releasing edit lock')
      // Stop heartbeat
      if (heartbeatIntervalRef.current) {
        clearInterval(heartbeatIntervalRef.current)
        heartbeatIntervalRef.current = null
      }

      // Untrack presence
      await channelRef.current.untrack()
      setIsEditing(false)
      setCurrentEditor(null)
      setCanEdit(true)
    } catch (error) {
      console.error('[EditLock] Failed to release edit lock:', error)
    }
  }, [])

  useEffect(() => {
    // Don't proceed if we don't have user info yet
    if (!currentUserId || !currentUserName) {
      return
    }

    const lockChannelName = `${roomName}-edit-lock`
    const channel = supabase.channel(lockChannelName, {
      config: {
        presence: {
          key: currentUserId,
        },
      },
    })

    channelRef.current = channel

    channel
      .on('presence', { event: 'sync' }, () => {
        console.log('[EditLock] Presence sync event')
        const presenceState = channel.presenceState<EditorInfo>()
        checkStaleLocks(presenceState)
      })
      .on('presence', { event: 'join' }, ({ newPresences }) => {
        console.log('[EditLock] Someone joined:', newPresences)
        // Someone joined - recheck locks
        const presenceState = channel.presenceState<EditorInfo>()
        checkStaleLocks(presenceState)
      })
      .on('presence', { event: 'leave' }, ({ leftPresences }) => {
        console.log('[EditLock] Someone left:', leftPresences)
        // Someone left - recheck locks
        const presenceState = channel.presenceState<EditorInfo>()
        checkStaleLocks(presenceState)
      })
      .subscribe(async (status) => {
        console.log('[EditLock] Channel subscription status:', status)
        if (status === REALTIME_SUBSCRIBE_STATES.SUBSCRIBED) {
          console.log('[EditLock] Edit lock channel subscribed successfully')
        } else if (status === REALTIME_SUBSCRIBE_STATES.CHANNEL_ERROR) {
          console.error('[EditLock] Edit lock channel error')
        } else if (status === REALTIME_SUBSCRIBE_STATES.TIMED_OUT) {
          console.error('[EditLock] Edit lock channel timed out')
        } else if (status === REALTIME_SUBSCRIBE_STATES.CLOSED) {
          console.log('[EditLock] Edit lock channel closed')
        }
      })

    return () => {
      console.log('[EditLock] Cleaning up channel')
      // Stop heartbeat on unmount
      if (heartbeatIntervalRef.current) {
        clearInterval(heartbeatIntervalRef.current)
        heartbeatIntervalRef.current = null
      }
      
      // Untrack presence (if we were tracking)
      channel.untrack()
      
      // Unsubscribe channel
      channel.unsubscribe()
      channelRef.current = null
    }
  }, [roomName, currentUserId, currentUserName, checkStaleLocks])

  return {
    currentEditor,
    requestEdit,
    releaseEdit,
    isEditing,
    canEdit,
  }
}

