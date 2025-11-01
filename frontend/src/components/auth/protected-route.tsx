import { useAuth } from '@clerk/clerk-react'
import { Navigate } from 'react-router-dom'

interface PublicOnlyRouteProps {
  children: React.ReactNode
}

/**
 * Wrapper for routes that should only be accessible to unauthenticated users
 * Redirects authenticated users to /dashboard
 */
export function PublicOnlyRoute({ children }: PublicOnlyRouteProps) {
  const { isSignedIn, isLoaded } = useAuth()

  // Wait for auth state to load
  if (!isLoaded) {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center">
        <div className="text-muted-foreground text-lg">Loading...</div>
      </div>
    )
  }

  // If user is signed in, redirect to dashboard
  if (isSignedIn) {
    return <Navigate to="/dashboard" replace />
  }

  // Otherwise, show the public route
  return <>{children}</>
}

