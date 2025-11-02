import { AvatarStack } from '@/components/avatar-stack'
import { useRealtimePresenceRoom } from '@/hooks/use-realtime-presence-room'
import { useClerkUserCached } from '@/hooks/use-clerk-user-cached'
import { useMemo } from 'react'

export const RealtimeAvatarStack = ({ roomName }: { roomName: string }) => {
  const { isSignedIn, isLoaded } = useClerkUserCached()
  const { users: usersMap } = useRealtimePresenceRoom(roomName)

  const avatars = useMemo(() => {
    return Object.values(usersMap).map((user) => ({
      name: user.name,
      image: user.image,
    }))
  }, [usersMap])

  // Don't render if user is not loaded or not signed in
  if (!isLoaded || !isSignedIn) {
    return null
  }

  return <AvatarStack avatars={avatars} />
}
