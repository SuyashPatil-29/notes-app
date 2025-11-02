import { useUser } from '@clerk/clerk-react'

/**
 * Hook to get the current user's name from Clerk
 * Returns the user's full name, first name, or '?' as fallback
 */
export const useCurrentUserName = () => {
  const { user, isLoaded } = useUser()

  if (!isLoaded) {
    return '?'
  }

  // Priority: fullName > firstName + lastName > firstName > email username > '?'
  return (
    user?.fullName ||
    (user?.firstName && user?.lastName ? `${user.firstName} ${user.lastName}` : null) ||
    user?.firstName ||
    user?.primaryEmailAddress?.emailAddress?.split('@')[0] ||
    '?'
  )
}
