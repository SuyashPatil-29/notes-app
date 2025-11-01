import { useEffect } from 'react'
import { useTheme } from './theme-provider'

/**
 * This component ensures Clerk components properly sync with the app's theme.
 * Clerk automatically detects dark/light mode from the root element's class,
 * but we use this wrapper to handle system theme changes in real-time.
 */
export function ClerkThemeWrapper({ children }: { children: React.ReactNode }) {
  const { colorScheme } = useTheme()

  useEffect(() => {
    // Listen for system theme changes when using 'system' color scheme
    if (colorScheme === 'system') {
      const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
      
      const handleChange = (e: MediaQueryListEvent) => {
        const root = document.documentElement
        root.classList.remove('light', 'dark')
        root.classList.add(e.matches ? 'dark' : 'light')
      }

      mediaQuery.addEventListener('change', handleChange)
      return () => mediaQuery.removeEventListener('change', handleChange)
    }
  }, [colorScheme])

  return <>{children}</>
}

