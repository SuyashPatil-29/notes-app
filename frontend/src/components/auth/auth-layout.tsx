import type { ReactNode } from 'react'
import { BookOpen } from 'lucide-react'
import { ModeToggle } from '@/components/ModeToggle'
import { ThemeSelector } from '@/components/ThemeSelector'

interface AuthLayoutProps {
  children: ReactNode
}

export function AuthLayout({ children }: AuthLayoutProps) {
  return (
    <div className="min-h-screen grid lg:grid-cols-2">
      {/* Header with theme options */}
      <header className="fixed top-0 left-0 right-0 z-50 border-b bg-background/95 backdrop-blur supports-backdrop-filter:bg-background/60">
        <div className="flex h-14 items-center justify-between px-6">
          <div className="flex items-center gap-2">
            <BookOpen className="h-6 w-6 text-primary" />
            <span className="text-xl font-bold">Notes App</span>
          </div>
          <div className="flex items-center gap-2">
            <ModeToggle />
            <ThemeSelector />
          </div>
        </div>
      </header>

      {/* Left side - Auth form */}
      <div className="flex items-center justify-center p-8 pt-24">
        <div className="w-full max-w-md space-y-8">
          {/* Auth component */}
          {children}
        </div>
      </div>

      {/* Right side - Decorative background */}
      <div className="hidden lg:block relative bg-linear-to-br from-primary/10 via-primary/5 to-background pt-14">
        <div className="absolute inset-0 flex items-center justify-center p-12 pt-24">
          <div className="max-w-2xl space-y-8 text-center">
            <h1 className="text-4xl font-bold tracking-tight">
              Organize Your Thoughts,<br />Amplify Your Ideas
            </h1>
            <p className="text-xl text-muted-foreground">
              Create beautiful notes with AI assistance, organize them in notebooks,
              and share your knowledge with the world.
            </p>
          </div>
        </div>
        
        {/* Decorative elements */}
        <div className="absolute top-0 right-0 w-64 h-64 bg-primary/10 rounded-full blur-3xl" />
        <div className="absolute bottom-0 left-0 w-96 h-96 bg-primary/5 rounded-full blur-3xl" />
      </div>
    </div>
  )
}

