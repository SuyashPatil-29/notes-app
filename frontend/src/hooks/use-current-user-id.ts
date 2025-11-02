import { useUser } from '@clerk/clerk-react'

/**
 * Hook to get the current user's unique ID from Clerk
 * Returns a consistent user ID for use in realtime features
 */
export const useCurrentUserId = () => {
  const { user, isLoaded } = useUser()

  if (!isLoaded || !user) {
    return null
  }

  // Clerk's user.id is a stable, unique identifier
  return user.id
}

