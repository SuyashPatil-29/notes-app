import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from '@/components/ui/breadcrumb'
import { Separator } from '@/components/ui/separator'
import { LeftSidebarTrigger } from '@/components/ui/left-sidebar'
import { RightSidebarTrigger } from '@/components/ui/right-sidebar'
import { ThemeSelector } from '@/components/ThemeSelector'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import { ArrowLeft, RotateCcw, Settings } from 'lucide-react'
import { Link } from 'react-router-dom'
import { ClerkLoaded, ClerkLoading, UserButton } from '@clerk/clerk-react'
import type { AuthenticatedUser } from '@/types/backend'
import api from '@/utils/api'
import { toast } from 'sonner'
import { ModeToggle } from './ModeToggle'
import { OrganizationSwitcher } from './organizations/OrganizationSwitcher'

export interface HeaderBreadcrumbItem {
  label: string
  href?: string
}

interface HeaderProps {
  user: AuthenticatedUser | null
  breadcrumbs?: HeaderBreadcrumbItem[]
  onOnboardingReset?: () => void
  actions?: React.ReactNode
}

export function Header({ user, breadcrumbs = [{ label: 'Dashboard' }], onOnboardingReset, actions }: HeaderProps) {
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

        {/* Custom Actions */}
        {actions && (
          <>
            <div className="ml-auto flex items-center gap-2">
              {actions}
            </div>
          </>
        )}

        {/* Organization Switcher */}
        <div className={actions ? "ml-4" : "ml-auto"}>
          <ClerkLoaded>
            <OrganizationSwitcher />
          </ClerkLoaded>
        </div>

        {/* Right Side Actions */}
        <div className="flex items-center gap-2">
          <ModeToggle />
          <ThemeSelector />
          <RightSidebarTrigger />
          <ClerkLoading>
            <Skeleton className="h-9 w-9 rounded-full" />
          </ClerkLoading>
          <ClerkLoaded>
            {user ? (
              <UserButton
                appearance={{
                  elements: {
                    avatarBox: "w-9 h-9 ring-2 ring-border"
                  }
                }}
              >
                <UserButton.MenuItems>
                  <UserButton.Link
                    label="Profile & Settings"
                    labelIcon={<Settings className="w-4 h-4" />}
                    href="/profile"
                  />
                  {import.meta.env.DEV && (
                    <UserButton.Action
                      label="Reset Onboarding (Dev)"
                      labelIcon={<RotateCcw className="w-4 h-4" />}
                      onClick={handleResetOnboarding}
                    />
                  )}
                </UserButton.MenuItems>
              </UserButton>
            ) : (
              <Skeleton className="h-9 w-9 rounded-full" />
            )}
          </ClerkLoaded>
        </div>
      </div>
    </header>
  )
}

