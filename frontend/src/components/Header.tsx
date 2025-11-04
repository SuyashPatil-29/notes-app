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
import { KbdKey } from './ui/kbd'
import { TooltipProvider, Tooltip, TooltipTrigger, TooltipContent } from './ui/tooltip'

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
      <div className="flex items-center gap-1 sm:gap-2 px-2 sm:px-4 w-full overflow-hidden">
        {/* Back Button - Hidden on mobile */}
        <Button
          variant="ghost"
          size="icon"
          onClick={() => window.history.back()}
          className="hover:bg-accent hidden sm:flex shrink-0"
          title="Go back"
        >
          <ArrowLeft className="h-5 w-5" />
        </Button>

        <LeftSidebarTrigger className="shrink-0" />
        <Separator orientation="vertical" className="h-4 hidden sm:block shrink-0" />

        {/* Breadcrumbs - Responsive with truncation */}
        <div className="flex-1 min-w-0 overflow-hidden">
          <Breadcrumb>
            <BreadcrumbList className="flex-nowrap">
              {breadcrumbs.map((item, index) => (
                <div key={index} className="flex items-center gap-1 sm:gap-2 min-w-0">
                  {index > 0 && <BreadcrumbSeparator className="shrink-0" />}
                  <BreadcrumbItem className="min-w-0">
                    {index === breadcrumbs.length - 1 ? (
                      <BreadcrumbPage className="truncate max-w-[120px] sm:max-w-[200px] md:max-w-none">
                        {item.label}
                      </BreadcrumbPage>
                    ) : item.href ? (
                      <BreadcrumbLink asChild>
                        <Link to={item.href} className="truncate max-w-[80px] sm:max-w-[150px] md:max-w-none block">
                          {item.label}
                        </Link>
                      </BreadcrumbLink>
                    ) : (
                      <span className="truncate max-w-[80px] sm:max-w-[150px] md:max-w-none block">
                        {item.label}
                      </span>
                    )}
                  </BreadcrumbItem>
                </div>
              ))}
            </BreadcrumbList>
          </Breadcrumb>
        </div>

        {/* Custom Actions - Responsive */}
        {actions && (
          <div className="flex items-center gap-1 sm:gap-2 shrink-0">
            {actions}
          </div>
        )}

        <TooltipProvider>
          <Tooltip>
            <TooltipTrigger asChild>
              <div className={`shrink-0 ${actions ? "ml-1 sm:ml-2" : ""}`}>
                <ClerkLoaded>
                  <button
                    type="button"
                    className="flex items-center gap-1 px-2 rounded bg-muted hover:bg-accent transition border text-muted-foreground text-[11px] font-mono font-medium h-9"
                    title="Search (Ctrl+K or Cmd+K)"
                    aria-label="Open search"
                    onClick={() => {
                      window.dispatchEvent(new Event('openCommandMenu'));
                    }}
                  >

                    <span className="flex items-center gap-0.5">
                      <KbdKey
                        className="shrink-0 text-[9px] px-1 py-0 leading-none"
                        aria-label={navigator.platform.includes('Mac') ? "⌘" : "Ctrl"}
                      >
                        {navigator.platform.includes('Mac') ? "⌘" : "Ctrl"}
                      </KbdKey>
                      <KbdKey
                        className="shrink-0 text-[9px] px-1 py-0 leading-none"
                        aria-label="K"
                      >
                        K
                      </KbdKey>
                    </span>
                  </button>
                </ClerkLoaded>
              </div>
            </TooltipTrigger>
            <TooltipContent side="top" className="max-w-xs">
              <p>Search notes, notebooks, and more</p>
            </TooltipContent>
          </Tooltip>
        </TooltipProvider>

        {/* Organization Switcher - Responsive */}
        <div className={`shrink-0 ${actions ? "ml-1 sm:ml-2" : ""}`}>
          <ClerkLoaded>
            <OrganizationSwitcher />
          </ClerkLoaded>
        </div>

        {/* Right Side Actions - Responsive: Hide less critical items on mobile */}
        <div className="flex items-center gap-0.5 sm:gap-1 md:gap-2 shrink-0 ml-auto">
          <div className="shrink-0">
            <ModeToggle />
          </div>
          {/* Hide ThemeSelector on small screens */}
          <div className="hidden lg:block shrink-0">
            <ThemeSelector />
          </div>
          <RightSidebarTrigger className="shrink-0" />
          <ClerkLoading>
            <Skeleton className="h-8 w-8 sm:h-9 sm:w-9 rounded-full shrink-0" />
          </ClerkLoading>
          <ClerkLoaded>
            {user ? (
              <div className="shrink-0">
                <UserButton
                  appearance={{
                    elements: {
                      avatarBox: "w-8 h-8 sm:w-9 sm:h-9 ring-2 ring-border"
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
              </div>
            ) : (
              <Skeleton className="h-8 w-8 sm:h-9 sm:w-9 rounded-full shrink-0" />
            )}
          </ClerkLoaded>
        </div>
      </div>
    </header>
  )
}

