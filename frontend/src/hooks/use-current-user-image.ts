import { useClerkUserCached } from '@/hooks/use-clerk-user-cached'

/**
 * Hook to get the current user's profile image from Clerk
 * Returns the user's image URL or null if not available
 * Uses cached version for better performance
 */
export const useCurrentUserImage = () => {
  const { user, isLoaded } = useClerkUserCached()

  if (!isLoaded) {
    return null
  }

  // Clerk provides imageUrl directly
  return user?.imageUrl || null
}
