import { useState } from 'react'
import { useSignIn } from '@clerk/clerk-react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { toast } from 'sonner'
import { Mail, Lock, Github, Chrome, Loader2 } from 'lucide-react'
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card'
import { Separator } from '@/components/ui/separator'

export function SignIn() {
  const { isLoaded, signIn, setActive } = useSignIn()
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [isLoading, setIsLoading] = useState(false)
  const [isOAuthLoading, setIsOAuthLoading] = useState<string | null>(null)
  
  // Get redirect URL from query params
  const redirectUrl = searchParams.get('redirect_url') || '/'

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    
    if (!isLoaded) return
    
    setIsLoading(true)
    
    try {
      const result = await signIn.create({
        identifier: email,
        password,
      })

      if (result.status === 'complete') {
        await setActive({ session: result.createdSessionId })
        navigate(redirectUrl)
      } else {
        // Handle other statuses if needed (e.g., 2FA)
        console.log('Sign in result:', result)
      }
    } catch (err: any) {
      console.error('Error signing in:', err)
      toast.error(err.errors?.[0]?.message || 'Failed to sign in. Please check your credentials.')
    } finally {
      setIsLoading(false)
    }
  }

  const handleOAuthSignIn = async (strategy: 'oauth_google' | 'oauth_github') => {
    if (!isLoaded) return
    
    setIsOAuthLoading(strategy)
    
    try {
      await signIn.authenticateWithRedirect({
        strategy,
        redirectUrl: '/sso-callback',
        redirectUrlComplete: redirectUrl,
      })
    } catch (err: any) {
      console.error('OAuth error:', err)
      toast.error(err.errors?.[0]?.message || 'Failed to sign in with OAuth')
      setIsOAuthLoading(null)
    }
  }

  return (
    <Card className="w-full max-w-md">
      <CardHeader className="space-y-1">
        <CardTitle className="text-2xl font-bold text-center">Welcome back</CardTitle>
        <CardDescription className="text-center">
          Sign in to your account to continue
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {/* OAuth Buttons */}
        <div className="grid grid-cols-2 gap-4">
          <Button
            variant="outline"
            type="button"
            disabled={!isLoaded || isLoading || isOAuthLoading !== null}
            onClick={() => handleOAuthSignIn('oauth_google')}
          >
            {isOAuthLoading === 'oauth_google' ? (
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
            ) : (
              <Chrome className="mr-2 h-4 w-4" />
            )}
            Google
          </Button>
          <Button
            variant="outline"
            type="button"
            disabled={!isLoaded || isLoading || isOAuthLoading !== null}
            onClick={() => handleOAuthSignIn('oauth_github')}
          >
            {isOAuthLoading === 'oauth_github' ? (
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
            ) : (
              <Github className="mr-2 h-4 w-4" />
            )}
            GitHub
          </Button>
        </div>

        <div className="relative">
          <div className="absolute inset-0 flex items-center">
            <Separator />
          </div>
          <div className="relative flex justify-center text-xs uppercase">
            <span className="bg-background px-2 text-muted-foreground">Or continue with</span>
          </div>
        </div>

        {/* Email/Password Form */}
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="email">Email</Label>
            <div className="relative">
              <Mail className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
              <Input
                id="email"
                type="email"
                placeholder="name@example.com"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                disabled={isLoading || !isLoaded}
                required
                className="pl-10"
              />
            </div>
          </div>

          <div className="space-y-2">
            <Label htmlFor="password">Password</Label>
            <div className="relative">
              <Lock className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
              <Input
                id="password"
                type="password"
                placeholder="••••••••"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                disabled={isLoading || !isLoaded}
                required
                className="pl-10"
              />
            </div>
          </div>

          <Button
            type="submit"
            className="w-full"
            disabled={isLoading || !isLoaded || isOAuthLoading !== null}
          >
            {isLoading ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Signing in...
              </>
            ) : (
              'Sign In'
            )}
          </Button>
        </form>
      </CardContent>
      <CardFooter className="flex flex-col space-y-2">
        <div className="text-sm text-muted-foreground text-center">
          Don't have an account?{' '}
          <a
            href="/sign-up"
            className="text-primary hover:underline font-medium"
            onClick={(e) => {
              e.preventDefault()
              navigate('/sign-up')
            }}
          >
            Sign up
          </a>
        </div>
        <div className="text-sm text-muted-foreground text-center">
          <a
            href="/forgot-password"
            className="text-primary hover:underline"
            onClick={(e) => {
              e.preventDefault()
              toast.info('Password reset functionality coming soon!')
            }}
          >
            Forgot your password?
          </a>
        </div>
      </CardFooter>
    </Card>
  )
}

