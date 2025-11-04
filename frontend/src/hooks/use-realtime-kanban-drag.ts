import { createClient } from '@/lib/supabase/client'
import { RealtimeChannel, REALTIME_SUBSCRIBE_STATES } from '@supabase/supabase-js'
import { useCallback, useEffect, useRef, useState } from 'react'

const supabase = createClient()

/**
 * Throttle a callback to a certain delay
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

/**
 * Generate a stable hash from a string (Clerk user ID)
 */
const hashStringToNumber = (str: string): number => {
  let hash = 0
  for (let i = 0; i < str.length; i++) {
    const char = str.charCodeAt(i)
    hash = (hash << 5) - hash + char
    hash = hash & hash
  }
  return Math.abs(hash)
}

type DragState = {
  taskId: string
  position: { x: number; y: number }
  user: {
    id: string
    name: string
    clerkId?: string
  }
  color: string
  timestamp: number
  column?: string | null
}

type HoverState = {
  taskId: string | null
  user: {
    id: string
    name: string
    clerkId?: string
  }
  color: string
  timestamp: number
}

export const useRealtimeKanbanDrag = ({
  boardId,
  username,
  clerkUserId,
  enabled = true,
}: {
  boardId: string
  username: string
  clerkUserId?: string | null
  enabled?: boolean
}) => {
  const [userId] = useState(() =>
    clerkUserId ? hashStringToNumber(clerkUserId) : Math.floor(Math.random() * 100000)
  )
  const [color] = useState(() =>
    clerkUserId ? `hsl(${hashStringToNumber(clerkUserId + 'color') % 360}, 100%, 70%)` : 'hsl(200, 100%, 70%)'
  )
  
  const [activeDrags, setActiveDrags] = useState<Record<string, DragState>>({})
  const [cardHovers, setCardHovers] = useState<Record<string, HoverState>>({})
  const channelRef = useRef<RealtimeChannel | null>(null)
  const currentDragRef = useRef<DragState | null>(null)

  // Cleanup stale drag states (older than 5 seconds)
  useEffect(() => {
    const interval = setInterval(() => {
      const now = Date.now()
      setActiveDrags((prev) => {
        const cleaned = { ...prev }
        let hasChanges = false
        
        Object.keys(cleaned).forEach((key) => {
          if (now - cleaned[key].timestamp > 5000) {
            delete cleaned[key]
            hasChanges = true
          }
        })
        
        return hasChanges ? cleaned : prev
      })

      setCardHovers((prev) => {
        const cleaned = { ...prev }
        let hasChanges = false
        
        Object.keys(cleaned).forEach((key) => {
          if (now - cleaned[key].timestamp > 3000) {
            delete cleaned[key]
            hasChanges = true
          }
        })
        
        return hasChanges ? cleaned : prev
      })
    }, 1000)

    return () => clearInterval(interval)
  }, [])

  // Setup channel and listeners
  useEffect(() => {
    if (!enabled || !username || !clerkUserId || !boardId) {
      return
    }

    const channelName = `kanban-board-${boardId}-drag`
    const channel = supabase.channel(channelName, {
      config: {
        presence: {
          key: clerkUserId,
        },
      },
    })

    channel
      .on('presence', { event: 'leave' }, ({ leftPresences }) => {
        leftPresences.forEach((element) => {
          const leftClerkId = element.clerkId
          if (leftClerkId) {
            setActiveDrags((prev) => {
              const newDrags = { ...prev }
              Object.keys(newDrags).forEach((key) => {
                if (newDrags[key].user.clerkId === leftClerkId) {
                  delete newDrags[key]
                }
              })
              return newDrags
            })

            setCardHovers((prev) => {
              const newHovers = { ...prev }
              Object.keys(newHovers).forEach((key) => {
                if (newHovers[key].user.clerkId === leftClerkId) {
                  delete newHovers[key]
                }
              })
              return newHovers
            })
          }
        })
      })
      .on('broadcast', { event: 'kanban-drag-start' }, (data: { payload: DragState }) => {
        const { payload } = data
        
        // Don't render own drag
        if (payload.user.clerkId === clerkUserId) {
          return
        }

        setActiveDrags((prev) => ({
          ...prev,
          [payload.user.id]: payload,
        }))
      })
      .on('broadcast', { event: 'kanban-drag-move' }, (data: { payload: DragState }) => {
        const { payload } = data
        
        // Don't render own drag
        if (payload.user.clerkId === clerkUserId) {
          return
        }

        setActiveDrags((prev) => ({
          ...prev,
          [payload.user.id]: payload,
        }))
      })
      .on('broadcast', { event: 'kanban-drag-end' }, (data: { payload: { userId: string; clerkId?: string } }) => {
        const { payload } = data
        
        // Don't process own events
        if (payload.clerkId === clerkUserId) {
          return
        }

        setActiveDrags((prev) => {
          const newDrags = { ...prev }
          delete newDrags[payload.userId]
          return newDrags
        })
      })
      .on('broadcast', { event: 'kanban-card-hover' }, (data: { payload: HoverState }) => {
        const { payload } = data
        
        // Don't render own hover
        if (payload.user.clerkId === clerkUserId) {
          return
        }

        if (payload.taskId === null) {
          // Remove hover
          setCardHovers((prev) => {
            const newHovers = { ...prev }
            delete newHovers[payload.user.id]
            return newHovers
          })
        } else {
          // Add/update hover
          setCardHovers((prev) => ({
            ...prev,
            [payload.user.id]: payload,
          }))
        }
      })
      .subscribe(async (status) => {
        if (status === REALTIME_SUBSCRIBE_STATES.SUBSCRIBED) {
          await channel.track({
            clerkId: clerkUserId,
            name: username,
            userId: userId.toString(),
          })
          channelRef.current = channel
        } else if (
          status === REALTIME_SUBSCRIBE_STATES.CHANNEL_ERROR ||
          status === REALTIME_SUBSCRIBE_STATES.TIMED_OUT ||
          status === REALTIME_SUBSCRIBE_STATES.CLOSED
        ) {
          setActiveDrags({})
          setCardHovers({})
          channelRef.current = null
        }
      })

    return () => {
      channel.unsubscribe()
      channelRef.current = null
      setActiveDrags({})
      setCardHovers({})
    }
  }, [boardId, username, clerkUserId, userId, color, enabled])

  // Broadcast drag start
  const broadcastDragStart = useCallback((taskId: string, position: { x: number; y: number }) => {
    if (!channelRef.current) return

    const payload: DragState = {
      taskId,
      position,
      user: {
        id: userId.toString(),
        name: username,
        clerkId: clerkUserId || undefined,
      },
      color,
      timestamp: Date.now(),
    }

    currentDragRef.current = payload

    channelRef.current.send({
      type: 'broadcast',
      event: 'kanban-drag-start',
      payload,
    })
  }, [userId, username, clerkUserId, color])

  // Broadcast drag move (throttled)
  const broadcastDragMoveInternal = useCallback((position: { x: number; y: number }, column?: string | null) => {
    if (!channelRef.current || !currentDragRef.current) return

    const payload: DragState = {
      ...currentDragRef.current,
      position,
      column,
      timestamp: Date.now(),
    }

    currentDragRef.current = payload

    channelRef.current.send({
      type: 'broadcast',
      event: 'kanban-drag-move',
      payload,
    })
  }, [])

  const broadcastDragMove = useThrottleCallback(broadcastDragMoveInternal, 50)

  // Broadcast drag end
  const broadcastDragEnd = useCallback(() => {
    if (!channelRef.current) return

    channelRef.current.send({
      type: 'broadcast',
      event: 'kanban-drag-end',
      payload: {
        userId: userId.toString(),
        clerkId: clerkUserId || undefined,
      },
    })

    currentDragRef.current = null
  }, [userId, clerkUserId])

  // Broadcast card hover
  const broadcastCardHoverInternal = useCallback((taskId: string | null) => {
    if (!channelRef.current) return

    const payload: HoverState = {
      taskId,
      user: {
        id: userId.toString(),
        name: username,
        clerkId: clerkUserId || undefined,
      },
      color,
      timestamp: Date.now(),
    }

    channelRef.current.send({
      type: 'broadcast',
      event: 'kanban-card-hover',
      payload,
    })
  }, [userId, username, clerkUserId, color])

  const broadcastCardHover = useThrottleCallback(broadcastCardHoverInternal, 100)

  return {
    activeDrags,
    cardHovers,
    broadcastDragStart,
    broadcastDragMove,
    broadcastDragEnd,
    broadcastCardHover,
    userColor: color,
  }
}

