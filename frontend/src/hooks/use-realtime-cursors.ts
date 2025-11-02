import { createClient } from '@/lib/supabase/client'
import { RealtimeChannel, REALTIME_SUBSCRIBE_STATES } from '@supabase/supabase-js'
import { useCallback, useEffect, useRef, useState } from 'react'

/**
 * Throttle a callback to a certain delay, It will only call the callback if the delay has passed, with the arguments
 * from the last call
 */
const useThrottleCallback = <Params extends unknown[], Return>(
  callback: (...args: Params) => Return,
  delay: number
) => {
  const lastCall = useRef(0)
  const timeout = useRef<NodeJS.Timeout | null>(null)

  return useCallback(
    (...args: Params) => {
      const now = Date.now()
      const remainingTime = delay - (now - lastCall.current)

      if (remainingTime <= 0) {
        if (timeout.current) {
          clearTimeout(timeout.current)
          timeout.current = null
        }
        lastCall.current = now
        callback(...args)
      } else if (!timeout.current) {
        timeout.current = setTimeout(() => {
          lastCall.current = Date.now()
          timeout.current = null
          callback(...args)
        }, remainingTime)
      }
    },
    [callback, delay]
  )
}

const supabase = createClient()

const generateRandomColor = () => `hsl(${Math.floor(Math.random() * 360)}, 100%, 70%)`

/**
 * Generate a stable hash from a string (Clerk user ID)
 * This ensures the same user always gets the same numeric ID
 */
const hashStringToNumber = (str: string): number => {
  let hash = 0
  for (let i = 0; i < str.length; i++) {
    const char = str.charCodeAt(i)
    hash = (hash << 5) - hash + char
    hash = hash & hash // Convert to 32bit integer
  }
  return Math.abs(hash)
}

const generateRandomNumber = () => Math.floor(Math.random() * 100)

const EVENT_NAME = 'realtime-cursor-move'

type CursorEventPayload = {
  position: {
    x: number
    y: number
  }
  user: {
    id: number
    name: string
    clerkId?: string // Add Clerk ID for better user tracking
  }
  color: string
  timestamp: number
}

export const useRealtimeCursors = ({
  roomName,
  username,
  throttleMs,
  clerkUserId,
}: {
  roomName: string
  username: string
  throttleMs: number
  clerkUserId?: string | null // Optional Clerk user ID for stable identification
}) => {
  // Use Clerk user ID to generate a stable numeric ID and color
  const [userId] = useState(() =>
    clerkUserId ? hashStringToNumber(clerkUserId) : generateRandomNumber()
  )
  const [color] = useState(() =>
    clerkUserId ? `hsl(${hashStringToNumber(clerkUserId + 'color') % 360}, 100%, 70%)` : generateRandomColor()
  )
  const [cursors, setCursors] = useState<Record<string, CursorEventPayload>>({})
  const cursorPayload = useRef<CursorEventPayload | null>(null)

  const channelRef = useRef<RealtimeChannel | null>(null)

  const callback = useCallback(
    (event: MouseEvent) => {
      const { clientX, clientY } = event

      const payload: CursorEventPayload = {
        position: {
          x: clientX,
          y: clientY,
        },
        user: {
          id: userId,
          name: username,
          clerkId: clerkUserId || undefined,
        },
        color: color,
        timestamp: new Date().getTime(),
      }

      cursorPayload.current = payload

      channelRef.current?.send({
        type: 'broadcast',
        event: EVENT_NAME,
        payload: payload,
      })
    },
    [color, userId, username, clerkUserId]
  )

  const handleMouseMove = useThrottleCallback(callback, throttleMs)

  useEffect(() => {
    // Don't proceed if we don't have required user info
    if (!username || !clerkUserId) {
      return
    }

    // Use a unique channel name to avoid conflicts with presence channel
    const cursorChannelName = `${roomName}-cursors`
    const channel = supabase.channel(cursorChannelName, {
      config: {
        presence: {
          key: clerkUserId, // Use Clerk user ID as the presence key
        },
      },
    })

    channel
      .on('presence', { event: 'leave' }, ({ leftPresences }) => {
        leftPresences.forEach(function (element) {
          // Remove cursor when user leaves by Clerk ID
          const leftClerkId = element.clerkId
          if (leftClerkId) {
          setCursors((prev) => {
              // Find and remove cursor by clerkId
              const newCursors = { ...prev }
              Object.keys(newCursors).forEach((key) => {
                if (newCursors[key].user.clerkId === leftClerkId) {
                  delete newCursors[key]
            }
              })
              return newCursors
          })
          }
        })
      })
      .on('presence', { event: 'join' }, () => {
        if (!cursorPayload.current) return

        // All cursors broadcast their position when a new cursor joins
        channelRef.current?.send({
          type: 'broadcast',
          event: EVENT_NAME,
          payload: cursorPayload.current,
        })
      })
      .on('broadcast', { event: EVENT_NAME }, (data: { payload: CursorEventPayload }) => {
        const { user } = data.payload
        
        // Don't render your own cursor (check by Clerk ID)
        if (user.clerkId === clerkUserId) {
          return
        }

        setCursors((prev) => {
          // Remove own cursor if it exists
          const newCursors = { ...prev }
          if (newCursors[userId]) {
            delete newCursors[userId]
          }

          newCursors[user.id] = data.payload
          return newCursors
        })
      })
      .subscribe(async (status) => {
        if (status === REALTIME_SUBSCRIBE_STATES.SUBSCRIBED) {
          // Track presence with Clerk user ID and additional metadata
          await channel.track({
            clerkId: clerkUserId,
            name: username,
            userId: userId,
          })
          channelRef.current = channel
        } else if (status === REALTIME_SUBSCRIBE_STATES.CHANNEL_ERROR) {
          setCursors({})
          channelRef.current = null
        } else if (status === REALTIME_SUBSCRIBE_STATES.TIMED_OUT) {
          setCursors({})
          channelRef.current = null
        } else if (status === REALTIME_SUBSCRIBE_STATES.CLOSED) {
          setCursors({})
          channelRef.current = null
        }
      })

    return () => {
      channel.unsubscribe()
      channelRef.current = null
      setCursors({})
    }
  }, [roomName, username, clerkUserId, userId, color])

  useEffect(() => {
    // Add event listener for mousemove
    window.addEventListener('mousemove', handleMouseMove)

    // Cleanup on unmount
    return () => {
      window.removeEventListener('mousemove', handleMouseMove)
    }
  }, [handleMouseMove])

  return { cursors }
}
