import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from '@/components/ui/breadcrumb'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Separator } from '@/components/ui/separator'
import { LeftSidebarTrigger } from '@/components/ui/left-sidebar'
import { RightSidebarTrigger } from '@/components/ui/right-sidebar'
import { ThemeSelector } from '@/components/ThemeSelector'
import { Button } from '@/components/ui/button'
import { LogOut, User, ArrowLeft, RotateCcw, Settings } from 'lucide-react'
import { Link } from 'react-router-dom'
import { handleGoogleLogout } from '@/utils/auth'
import type { AuthenticatedUser } from '@/types/backend'
import api from '@/utils/api'
import { toast } from 'sonner'
import { ModeToggle } from './ModeToggle'

export interface HeaderBreadcrumbItem {
  label: string
  href?: string
}

interface HeaderProps {
  user: AuthenticatedUser | null
  breadcrumbs?: HeaderBreadcrumbItem[]
  onOnboardingReset?: () => void
}

export function Header({ user, breadcrumbs = [{ label: 'Dashboard' }], onOnboardingReset }: HeaderProps) {
  const handleResetOnboarding = async () => {
    if (!confirm("Reset onboarding? This will require you to complete onboarding again.")) {
      return;
    }

    try {
      await api.delete("/onboarding");
      toast.success("Onboarding reset. Please refresh the page.");
      if (onOnboardingReset) {
        onOnboardingReset();
      } else {
        // Fallback: reload the page
        window.location.reload();
      }
    } catch (error: any) {
      console.error("Reset onboarding error:", error);
      toast.error("Failed to reset onboarding");
    }
  };
  return (
    <header className="flex h-16 shrink-0 items-center border-b bg-background sticky top-0 z-10">
      <div className="flex items-center gap-2 px-4 w-full">
          <Button
            variant="ghost"
            size="icon"
          onClick={() => window.history.back()}
            className="hover:bg-accent"
          title="Go back"
          >
          <ArrowLeft className="h-5 w-5" />
          </Button>
        <LeftSidebarTrigger />
        <Separator orientation="vertical" className="h-4" />

        {/* Breadcrumbs */}
        <Breadcrumb>
          <BreadcrumbList>
            {breadcrumbs.map((item, index) => (
              <div key={index} className="flex items-center gap-2">
                {index > 0 && <BreadcrumbSeparator />}
                <BreadcrumbItem>
                  {index === breadcrumbs.length - 1 ? (
                    <BreadcrumbPage>{item.label}</BreadcrumbPage>
                  ) : item.href ? (
                    <BreadcrumbLink asChild>
                      <Link to={item.href}>{item.label}</Link>
                    </BreadcrumbLink>
                  ) : (
                    <span>{item.label}</span>
                  )}
                </BreadcrumbItem>
              </div>
            ))}
          </BreadcrumbList>
        </Breadcrumb>

        {/* Right Side Actions */}
        <div className="ml-auto flex items-center gap-2">
          <ModeToggle />
          <ThemeSelector />
          <RightSidebarTrigger />
          {user && (
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <button className="flex items-center gap-2 rounded-full hover:opacity-80 transition-opacity focus:outline-none focus:ring-2 focus:ring-ring">
                  {user.imageUrl ? (
                    <img
                      src={user.imageUrl}
                      alt={user.name}
                      className="w-9 h-9 rounded-full ring-2 ring-border"
                    />
                  ) : (
                    <div className="w-9 h-9 rounded-full bg-primary flex items-center justify-center ring-2 ring-border">
                      <User className="w-5 h-5 text-primary-foreground" />
                    </div>
                  )}
                </button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end" className="w-56">
                <DropdownMenuLabel className="font-normal">
                  <div className="flex flex-col space-y-1">
                    <p className="text-sm font-medium leading-none">{user.name}</p>
                    <p className="text-xs leading-none text-muted-foreground">
                      {user.email}
                    </p>
                  </div>
                </DropdownMenuLabel>
                <DropdownMenuSeparator />
                <DropdownMenuItem asChild className="cursor-pointer">
                  <Link to="/profile">
                    <Settings className="mr-2 h-4 w-4" />
                    <span>Profile & Settings</span>
                  </Link>
                </DropdownMenuItem>
                {import.meta.env.DEV && (
                  <>
                    <DropdownMenuSeparator />
                    <DropdownMenuItem onClick={handleResetOnboarding} className="cursor-pointer text-yellow-600 dark:text-yellow-400">
                      <RotateCcw className="mr-2 h-4 w-4" />
                      <span>Reset Onboarding (Dev)</span>
                    </DropdownMenuItem>
                  </>
                )}
                <DropdownMenuSeparator />
                <DropdownMenuItem onClick={handleGoogleLogout} className="cursor-pointer">
                  <LogOut className="mr-2 h-4 w-4" />
                  <span>Log out</span>
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          )}
        </div>
      </div>
    </header>
  )
}

