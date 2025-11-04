import { useParams, useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { TaskBoard } from '@/components/TaskBoard'
import { Header } from '@/components/Header'
import { useUser } from '@/hooks/auth'
import { getTaskBoard } from '@/utils/tasks'

export function TaskBoardPage() {
  const { boardId } = useParams<{ boardId: string }>()
  const navigate = useNavigate()
  const { user } = useUser()

  // Fetch task board for breadcrumbs
  const { data: taskBoard, isLoading } = useQuery({
    queryKey: ['taskBoard', boardId],
    queryFn: () => getTaskBoard(boardId!),
    enabled: !!boardId,
    refetchOnWindowFocus: false,
  })

  if (!boardId) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="text-center space-y-4">
          <p className="text-lg text-destructive">Invalid task board ID</p>
          <button
            onClick={() => navigate('/dashboard')}
            className="text-sm text-muted-foreground hover:text-foreground"
          >
            ← Back to Dashboard
          </button>
        </div>
      </div>
    )
  }

  // Build breadcrumbs
  const breadcrumbs = [
    { label: 'Dashboard', href: '/dashboard' },
    { label: 'Task Boards', href: '/dashboard' },
    { label: isLoading ? 'Loading...' : taskBoard?.name || 'Task Board' },
  ]

  return (
    <div className="flex flex-col h-full">
      <Header user={user} breadcrumbs={breadcrumbs} />
      <TaskBoard
        boardId={boardId}
        onNavigateBack={() => navigate('/dashboard')}
        onBoardDeleted={() => navigate('/dashboard')}
        className="flex-1"
      />
    </div>
  )
}
