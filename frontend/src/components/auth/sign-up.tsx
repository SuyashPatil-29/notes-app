import { useState } from 'react'
import { useSignUp } from '@clerk/clerk-react'
import { useNavigate } from 'react-router-dom'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { toast } from 'sonner'
import { Mail, Lock, User, Github, Chrome, Loader2, Check } from 'lucide-react'
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card'
import { Separator } from '@/components/ui/separator'

export function SignUp() {
  const { isLoaded, signUp, setActive } = useSignUp()
  const navigate = useNavigate()
  
  const [firstName, setFirstName] = useState('')
  const [lastName, setLastName] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [verifying, setVerifying] = useState(false)
  const [code, setCode] = useState('')
  const [isLoading, setIsLoading] = useState(false)
  const [isOAuthLoading, setIsOAuthLoading] = useState<string | null>(null)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    
    if (!isLoaded) return
    
    setIsLoading(true)
    
    try {
      await signUp.create({
        firstName,
        lastName,
        emailAddress: email,
        password,
      })

      // Send email verification code
      await signUp.prepareEmailAddressVerification({ strategy: 'email_code' })
      
      setVerifying(true)
      toast.success('Verification code sent to your email')
    } catch (err: any) {
      console.error('Error signing up:', err)
      toast.error(err.errors?.[0]?.message || 'Failed to sign up. Please try again.')
    } finally {
      setIsLoading(false)
    }
  }

  const handleVerify = async (e: React.FormEvent) => {
    e.preventDefault()
    
    if (!isLoaded) return
    
    setIsLoading(true)
    
    try {
      const completeSignUp = await signUp.attemptEmailAddressVerification({
        code,
      })

      if (completeSignUp.status === 'complete') {
        await setActive({ session: completeSignUp.createdSessionId })
        navigate('/')
      } else {
        console.log('Sign up result:', completeSignUp)
      }
    } catch (err: any) {
      console.error('Error verifying:', err)
      toast.error(err.errors?.[0]?.message || 'Invalid verification code. Please try again.')
    } finally {
      setIsLoading(false)
    }
  }

  const handleOAuthSignUp = async (strategy: 'oauth_google' | 'oauth_github') => {
    if (!isLoaded) return
    
    setIsOAuthLoading(strategy)
    
    try {
      await signUp.authenticateWithRedirect({
        strategy,
        redirectUrl: '/sso-callback',
        redirectUrlComplete: '/',
      })
    } catch (err: any) {
      console.error('OAuth error:', err)
      toast.error(err.errors?.[0]?.message || 'Failed to sign up with OAuth')
      setIsOAuthLoading(null)
    }
  }

  if (verifying) {
    return (
      <Card className="w-full max-w-md">
        <CardHeader className="space-y-1">
          <CardTitle className="text-2xl font-bold text-center">Verify your email</CardTitle>
          <CardDescription className="text-center">
            We sent a verification code to <span className="font-medium">{email}</span>
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleVerify} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="code">Verification Code</Label>
              <Input
                id="code"
                type="text"
                placeholder="Enter 6-digit code"
                value={code}
                onChange={(e) => setCode(e.target.value)}
                disabled={isLoading || !isLoaded}
                required
                maxLength={6}
                className="text-center text-2xl tracking-widest"
              />
            </div>

            <Button
              type="submit"
              className="w-full"
              disabled={isLoading || !isLoaded || code.length !== 6}
            >
              {isLoading ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Verifying...
                </>
              ) : (
                <>
                  <Check className="mr-2 h-4 w-4" />
                  Verify Email
                </>
              )}
            </Button>
          </form>
        </CardContent>
        <CardFooter className="flex flex-col space-y-2">
          <div className="text-sm text-muted-foreground text-center">
            Didn't receive the code?{' '}
            <button
              type="button"
              className="text-primary hover:underline font-medium"
              onClick={async () => {
                try {
                  await signUp?.prepareEmailAddressVerification({ strategy: 'email_code' })
                  toast.success('Verification code resent')
                } catch (err: any) {
                  toast.error('Failed to resend code')
                }
              }}
            >
              Resend
            </button>
          </div>
          <div className="text-sm text-muted-foreground text-center">
            <button
              type="button"
              className="text-primary hover:underline"
              onClick={() => setVerifying(false)}
            >
              Use a different email
            </button>
          </div>
        </CardFooter>
      </Card>
    )
  }

  return (
    <Card className="w-full max-w-md">
      <CardHeader className="space-y-1">
        <CardTitle className="text-2xl font-bold text-center">Create an account</CardTitle>
        <CardDescription className="text-center">
          Enter your information to get started
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {/* OAuth Buttons */}
        <div className="grid grid-cols-2 gap-4">
          <Button
            variant="outline"
            type="button"
            disabled={!isLoaded || isLoading || isOAuthLoading !== null}
            onClick={() => handleOAuthSignUp('oauth_google')}
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
            onClick={() => handleOAuthSignUp('oauth_github')}
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

        {/* Sign Up Form */}
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="firstName">First Name</Label>
              <div className="relative">
                <User className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
                <Input
                  id="firstName"
                  type="text"
                  placeholder="John"
                  value={firstName}
                  onChange={(e) => setFirstName(e.target.value)}
                  disabled={isLoading || !isLoaded}
                  required
                  className="pl-10"
                />
              </div>
            </div>

            <div className="space-y-2">
              <Label htmlFor="lastName">Last Name</Label>
              <Input
                id="lastName"
                type="text"
                placeholder="Doe"
                value={lastName}
                onChange={(e) => setLastName(e.target.value)}
                disabled={isLoading || !isLoaded}
                required
              />
            </div>
          </div>

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
                minLength={8}
                className="pl-10"
              />
            </div>
            <p className="text-xs text-muted-foreground">
              Must be at least 8 characters
            </p>
          </div>

          <Button
            type="submit"
            className="w-full"
            disabled={isLoading || !isLoaded || isOAuthLoading !== null}
          >
            {isLoading ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Creating account...
              </>
            ) : (
              'Create Account'
            )}
          </Button>
        </form>
      </CardContent>
      <CardFooter>
        <div className="text-sm text-muted-foreground text-center w-full">
          Already have an account?{' '}
          <a
            href="/sign-in"
            className="text-primary hover:underline font-medium"
            onClick={(e) => {
              e.preventDefault()
              navigate('/sign-in')
            }}
          >
            Sign in
          </a>
        </div>
      </CardFooter>
    </Card>
  )
}

