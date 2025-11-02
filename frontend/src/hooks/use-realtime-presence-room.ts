'use client'

import { useCurrentUserImage } from '@/hooks/use-current-user-image'
import { useCurrentUserName } from '@/hooks/use-current-user-name'
import { useCurrentUserId } from '@/hooks/use-current-user-id'
import { createClient } from '@/lib/supabase/client'
import { REALTIME_SUBSCRIBE_STATES, RealtimeChannel } from '@supabase/supabase-js'
import { useEffect, useState, useRef } from 'react'

const supabase = createClient()

export type RealtimeUser = {
  id: string
  name: string
  image: string
}

export const useRealtimePresenceRoom = (roomName: string) => {
  const currentUserImage = useCurrentUserImage()
  const currentUserName = useCurrentUserName()
  const currentUserId = useCurrentUserId()

  const [users, setUsers] = useState<Record<string, RealtimeUser>>({})
  const channelRef = useRef<RealtimeChannel | null>(null)
  const isTrackingRef = useRef(false)

  useEffect(() => {
    // Don't proceed if we don't have user info yet
    if (!currentUserId || !currentUserName) {
      return
    }

    const room = supabase.channel(roomName, {
      config: {
        presence: {
          key: currentUserId, // Use Clerk user ID as the presence key
        },
      },
    })

    channelRef.current = room

    room
      .on('presence', { event: 'sync' }, () => {
        const newState = room.presenceState<{ 
          id: string
          name: string
          image: string | null 
        }>()

        const newUsers = Object.fromEntries(
          Object.entries(newState).map(([key, values]) => {
            const userData = values[0]
            return [
              key,
              { 
                id: userData.id,
                name: userData.name, 
                image: userData.image || '' 
              },
            ]
          })
        ) as Record<string, RealtimeUser>
        setUsers(newUsers)
      })
      .on('presence', { event: 'join' }, ({ key, newPresences }) => {
        console.log('User joined:', key, newPresences)
      })
      .on('presence', { event: 'leave' }, ({ key, leftPresences }) => {
        console.log('User left:', key, leftPresences)
      })
      .subscribe(async (status) => {
        if (status === REALTIME_SUBSCRIBE_STATES.SUBSCRIBED) {
          // Only track once
          if (!isTrackingRef.current) {
            isTrackingRef.current = true
            await room.track({
              id: currentUserId,
              name: currentUserName,
              image: currentUserImage,
            })
          }
        } else if (status === REALTIME_SUBSCRIBE_STATES.CHANNEL_ERROR) {
          console.error('Realtime channel error')
          isTrackingRef.current = false
        } else if (status === REALTIME_SUBSCRIBE_STATES.TIMED_OUT) {
          console.error('Realtime channel timed out')
          isTrackingRef.current = false
        } else if (status === REALTIME_SUBSCRIBE_STATES.CLOSED) {
          console.log('Realtime channel closed')
          isTrackingRef.current = false
        }
      })

    return () => {
      isTrackingRef.current = false
      if (channelRef.current) {
        channelRef.current.unsubscribe()
        channelRef.current = null
      }
    }
  }, [roomName]) // Only re-subscribe when roomName changes

  // Update presence when user info changes (but don't re-subscribe)
  useEffect(() => {
    if (channelRef.current && isTrackingRef.current && currentUserId) {
      channelRef.current.track({
        id: currentUserId,
        name: currentUserName,
        image: currentUserImage,
      })
    }
  }, [currentUserName, currentUserImage, currentUserId])

  return { users }
}
