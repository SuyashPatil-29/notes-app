import { useClerkUserCached } from '@/hooks/use-clerk-user-cached'

/**
 * Hook to get the current user's unique ID from Clerk
 * Returns a consistent user ID for use in realtime features
 * Uses cached version for better performance
 */
export const useCurrentUserId = () => {
  const { user, isLoaded } = useClerkUserCached()

  if (!isLoaded || !user) {
    return null
  }

  // Clerk's user.id is a stable, unique identifier
  return user.id
}

