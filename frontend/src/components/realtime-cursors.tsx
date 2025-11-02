import { Cursor } from '@/components/cursor'
import { useRealtimeCursors } from '@/hooks/use-realtime-cursors'
import { useCurrentUserName } from '@/hooks/use-current-user-name'
import { useCurrentUserId } from '@/hooks/use-current-user-id'
import { useUser } from '@clerk/clerk-react'

const THROTTLE_MS = 50

export const RealtimeCursors = ({
  roomName,
  username,
}: {
  roomName: string
  username?: string
}) => {
  // Get user data from Clerk
  const { isSignedIn, isLoaded } = useUser()
  const clerkUserId = useCurrentUserId()
  const defaultUsername = useCurrentUserName()

  // Use provided username or fall back to Clerk username
  const finalUsername = username || defaultUsername

  const { cursors } = useRealtimeCursors({
    roomName,
    username: finalUsername,
    throttleMs: THROTTLE_MS,
    clerkUserId,
  })

  // Don't render if user is not loaded or not signed in
  if (!isLoaded || !isSignedIn) {
    return null
  }

  return (
    <div>
      {Object.keys(cursors).map((id) => (
        <Cursor
          key={id}
          className="fixed transition-transform ease-in-out z-50"
          style={{
            transitionDuration: '20ms',
            top: 0,
            left: 0,
            transform: `translate(${cursors[id].position.x}px, ${cursors[id].position.y}px)`,
          }}
          color={cursors[id].color}
          name={cursors[id].user.name}
        />
      ))}
    </div>
  )
}
