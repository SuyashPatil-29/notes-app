import { createContext, useContext, useEffect, useState } from "react"

type ColorScheme = "dark" | "light" | "system"
type ThemeVariant = "notebook" | "grayscale" | "tech" | "minimal" | "orange" | "pink"

type ThemeProviderProps = {
  children: React.ReactNode
  defaultTheme?: ThemeVariant
  defaultColorScheme?: ColorScheme
  storageKey?: string
  themeStorageKey?: string
}

type ThemeProviderState = {
  theme: ThemeVariant
  colorScheme: ColorScheme
  setTheme: (theme: ThemeVariant) => void
  setColorScheme: (scheme: ColorScheme) => void
}

const initialState: ThemeProviderState = {
  theme: "notebook",
  colorScheme: "system",
  setTheme: () => null,
  setColorScheme: () => null,
}

const ThemeProviderContext = createContext<ThemeProviderState>(initialState)

export function ThemeProvider({
  children,
  defaultTheme = "notebook",
  defaultColorScheme = "system",
  storageKey = "vite-ui-color-scheme",
  themeStorageKey = "vite-ui-theme",
  ...props
}: ThemeProviderProps) {
  const [colorScheme, setColorSchemeState] = useState<ColorScheme>(
    () => (localStorage.getItem(storageKey) as ColorScheme) || defaultColorScheme
  )
  const [theme, setThemeState] = useState<ThemeVariant>(
    () => (localStorage.getItem(themeStorageKey) as ThemeVariant) || defaultTheme
  )

  useEffect(() => {
    const root = window.document.documentElement

    // Set data-theme attribute
    root.setAttribute("data-theme", theme)

    // Handle color scheme (dark/light)
    root.classList.remove("light", "dark")

    if (colorScheme === "system") {
      const systemTheme = window.matchMedia("(prefers-color-scheme: dark)")
        .matches
        ? "dark"
        : "light"

      root.classList.add(systemTheme)
    } else {
      root.classList.add(colorScheme)
    }
  }, [theme, colorScheme])

  const value = {
    theme,
    colorScheme,
    setTheme: (newTheme: ThemeVariant) => {
      localStorage.setItem(themeStorageKey, newTheme)
      setThemeState(newTheme)
    },
    setColorScheme: (scheme: ColorScheme) => {
      localStorage.setItem(storageKey, scheme)
      setColorSchemeState(scheme)
    },
  }

  return (
    <ThemeProviderContext.Provider {...props} value={value}>
      {children}
    </ThemeProviderContext.Provider>
  )
}

export const useTheme = () => {
  const context = useContext(ThemeProviderContext)

  if (context === undefined)
    throw new Error("useTheme must be used within a ThemeProvider")

  return context
}