import { createClient } from '@/lib/supabase/client'
import { RealtimeChannel, REALTIME_SUBSCRIBE_STATES } from '@supabase/supabase-js'
import { useEffect, useRef, useCallback } from 'react'
import type { Task } from '@/types/backend'

const supabase = createClient()

type TaskUpdatePayload = {
  taskId: string
  changes: Partial<Task>
  userId: string
  clerkId?: string
  timestamp: number
}

type TaskCreatedPayload = {
  task: Task
  userId: string
  clerkId?: string
  timestamp: number
}

type TaskDeletedPayload = {
  taskId: string
  userId: string
  clerkId?: string
  timestamp: number
}

export const useRealtimeKanbanUpdates = ({
  boardId,
  clerkUserId,
  onTaskCreated,
  onTaskUpdated,
  onTaskDeleted,
  enabled = true,
}: {
  boardId: string
  clerkUserId?: string | null
  onTaskCreated?: (task: Task) => void
  onTaskUpdated?: (taskId: string, changes: Partial<Task>) => void
  onTaskDeleted?: (taskId: string) => void
  enabled?: boolean
}) => {
  const channelRef = useRef<RealtimeChannel | null>(null)

  // Setup channel and listeners
  useEffect(() => {
    if (!enabled || !clerkUserId || !boardId) {
      return
    }

    const channelName = `kanban-board-${boardId}-updates`
    const channel = supabase.channel(channelName, {
      config: {
        presence: {
          key: clerkUserId,
        },
      },
    })

    channel
      .on('broadcast', { event: 'kanban-task-created' }, (data: { payload: TaskCreatedPayload }) => {
        const { payload } = data
        
        // Don't process own events
        if (payload.clerkId === clerkUserId) {
          return
        }

        onTaskCreated?.(payload.task)
      })
      .on('broadcast', { event: 'kanban-task-updated' }, (data: { payload: TaskUpdatePayload }) => {
        const { payload } = data
        
        // Don't process own events
        if (payload.clerkId === clerkUserId) {
          return
        }

        onTaskUpdated?.(payload.taskId, payload.changes)
      })
      .on('broadcast', { event: 'kanban-task-deleted' }, (data: { payload: TaskDeletedPayload }) => {
        const { payload } = data
        
        // Don't process own events
        if (payload.clerkId === clerkUserId) {
          return
        }

        onTaskDeleted?.(payload.taskId)
      })
      .subscribe((status) => {
        if (status === REALTIME_SUBSCRIBE_STATES.SUBSCRIBED) {
          channelRef.current = channel
        } else if (
          status === REALTIME_SUBSCRIBE_STATES.CHANNEL_ERROR ||
          status === REALTIME_SUBSCRIBE_STATES.TIMED_OUT ||
          status === REALTIME_SUBSCRIBE_STATES.CLOSED
        ) {
          channelRef.current = null
        }
      })

    return () => {
      channel.unsubscribe()
      channelRef.current = null
    }
  }, [boardId, clerkUserId, onTaskCreated, onTaskUpdated, onTaskDeleted, enabled])

  // Broadcast task created
  const broadcastTaskCreated = useCallback((task: Task) => {
    if (!channelRef.current) return

    const payload: TaskCreatedPayload = {
      task,
      userId: clerkUserId || 'unknown',
      clerkId: clerkUserId || undefined,
      timestamp: Date.now(),
    }

    channelRef.current.send({
      type: 'broadcast',
      event: 'kanban-task-created',
      payload,
    })
  }, [clerkUserId])

  // Broadcast task updated
  const broadcastTaskUpdated = useCallback((taskId: string, changes: Partial<Task>) => {
    if (!channelRef.current) return

    const payload: TaskUpdatePayload = {
      taskId,
      changes,
      userId: clerkUserId || 'unknown',
      clerkId: clerkUserId || undefined,
      timestamp: Date.now(),
    }

    channelRef.current.send({
      type: 'broadcast',
      event: 'kanban-task-updated',
      payload,
    })
  }, [clerkUserId])

  // Broadcast task deleted
  const broadcastTaskDeleted = useCallback((taskId: string) => {
    if (!channelRef.current) return

    const payload: TaskDeletedPayload = {
      taskId,
      userId: clerkUserId || 'unknown',
      clerkId: clerkUserId || undefined,
      timestamp: Date.now(),
    }

    channelRef.current.send({
      type: 'broadcast',
      event: 'kanban-task-deleted',
      payload,
    })
  }, [clerkUserId])

  return {
    broadcastTaskCreated,
    broadcastTaskUpdated,
    broadcastTaskDeleted,
  }
}

