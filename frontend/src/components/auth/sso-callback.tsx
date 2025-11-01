import { useEffect } from 'react'
import { useClerk } from '@clerk/clerk-react'
import { useNavigate } from 'react-router-dom'
import { Loader2 } from 'lucide-react'

export function SSOCallback() {
  const { handleRedirectCallback } = useClerk()
  const navigate = useNavigate()

  useEffect(() => {
    const handleCallback = async () => {
      try {
        await handleRedirectCallback({
          redirectUrl: '/',
        })
        navigate('/')
      } catch (error) {
        console.error('Error handling OAuth callback:', error)
        navigate('/sign-in')
      }
    }

    handleCallback()
  }, [handleRedirectCallback, navigate])

  return (
    <div className="min-h-screen bg-background flex items-center justify-center">
      <div className="text-center space-y-4">
        <Loader2 className="h-12 w-12 animate-spin text-primary mx-auto" />
        <h2 className="text-xl font-semibold">Completing sign in...</h2>
        <p className="text-muted-foreground">Please wait while we verify your account.</p>
      </div>
    </div>
  )
}

