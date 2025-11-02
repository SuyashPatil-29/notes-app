import { useClerkUserCached } from '@/hooks/use-clerk-user-cached'

/**
 * Hook to get the current user's name from Clerk
 * Returns the user's full name, first name, or '?' as fallback
 * Uses cached version for better performance
 */
export const useCurrentUserName = () => {
  const { user, isLoaded } = useClerkUserCached()

  if (!isLoaded) {
    return '?'
  }

  // Priority: fullName > firstName + lastName > firstName > email username > '?'
  return (
    user?.fullName ||
    (user?.firstName && user?.lastName ? `${user.firstName} ${user.lastName}` : null) ||
    user?.firstName ||
    user?.primaryEmailAddress?.split('@')[0] ||
    '?'
  )
}
