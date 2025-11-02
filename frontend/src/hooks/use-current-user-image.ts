import { useUser } from '@clerk/clerk-react'

/**
 * Hook to get the current user's profile image from Clerk
 * Returns the user's image URL or null if not available
 */
export const useCurrentUserImage = () => {
  const { user, isLoaded } = useUser()

  if (!isLoaded) {
    return null
  }

  // Clerk provides imageUrl directly
  return user?.imageUrl || null
}
