import {
  CircleCheckIcon,
  InfoIcon,
  Loader2Icon,
  OctagonXIcon,
  TriangleAlertIcon,
} from "lucide-react"
import { useTheme } from "next-themes"
import { Toaster as Sonner, type ToasterProps } from "sonner"

const Toaster = ({ ...props }: ToasterProps) => {
  const { theme = "system" } = useTheme()

  return (
    <Sonner
      theme={theme as ToasterProps["theme"]}
      className="toaster group"
      position="bottom-right"
      duration={3500}
      gap={8}
      toastOptions={{
        style: {
          background: "hsl(var(--primary) / 0.1)",
          border: "1px solid hsl(var(--primary) / 0.2)",
          color: "hsl(var(--foreground))",
        },
        classNames: {
          toast: "group toast",
          description: "group-[.toast]:text-muted-foreground",
          actionButton:
            "group-[.toast]:bg-primary group-[.toast]:text-primary-foreground",
          cancelButton:
            "group-[.toast]:bg-muted group-[.toast]:text-muted-foreground",
          success: "!bg-primary/25 !text-primary !border-primary/30",
          error: "!bg-destructive/25 !text-destructive !border-destructive/30",
          warning: "!bg-amber-500/25 !text-amber-700 dark:!text-amber-400 !border-amber-500/30",
          info: "!bg-blue-500/25 !text-blue-700 dark:!text-blue-400 !border-blue-500/30",
        },
      }}
      icons={{
        success: <CircleCheckIcon className="h-5 w-5" strokeWidth={2.5} />,
        info: <InfoIcon className="h-5 w-5" strokeWidth={2.5} />,
        warning: <TriangleAlertIcon className="h-5 w-5" strokeWidth={2.5} />,
        error: <OctagonXIcon className="h-5 w-5" strokeWidth={2.5} />,
        loading: <Loader2Icon className="h-5 w-5 animate-spin" strokeWidth={2.5} />,
      }}
      {...props}
    />
  )
}

export { Toaster }