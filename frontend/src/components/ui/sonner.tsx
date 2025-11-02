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
        classNames: {
          toast:
            "group toast group-[.toaster]:bg-primary/15 group-[.toaster]:text-foreground group-[.toaster]:border group-[.toaster]:border-primary/30 group-[.toaster]:shadow-xl group-[.toaster]:backdrop-blur-md group-[.toaster]:rounded-[calc(var(--radius)-2px)]",
          description: "group-[.toast]:text-muted-foreground group-[.toast]:text-sm",
          actionButton:
            "group-[.toast]:bg-primary group-[.toast]:text-primary-foreground group-[.toast]:hover:bg-primary/90 group-[.toast]:rounded-[calc(var(--radius)-6px)] group-[.toast]:font-medium",
          cancelButton:
            "group-[.toast]:bg-secondary group-[.toast]:text-secondary-foreground group-[.toast]:hover:bg-secondary/80 group-[.toast]:rounded-[calc(var(--radius)-6px)]",
          success:
            "!group-[.toaster]:bg-primary/20 !group-[.toaster]:text-primary !group-[.toaster]:border-primary/40",
          error:
            "!group-[.toaster]:bg-destructive/20 !group-[.toaster]:text-destructive !group-[.toaster]:border-destructive/40",
          warning:
            "!group-[.toaster]:bg-accent/70 !group-[.toaster]:text-accent-foreground !group-[.toaster]:border-accent",
          info:
            "!group-[.toaster]:bg-muted/80 !group-[.toaster]:text-foreground !group-[.toaster]:border-border",
        },
      }}
      icons={{
        success: <CircleCheckIcon className="size-[1.125rem]" strokeWidth={2.5} />,
        info: <InfoIcon className="size-[1.125rem]" strokeWidth={2.5} />,
        warning: <TriangleAlertIcon className="size-[1.125rem]" strokeWidth={2.5} />,
        error: <OctagonXIcon className="size-[1.125rem]" strokeWidth={2.5} />,
        loading: <Loader2Icon className="size-[1.125rem] animate-spin" strokeWidth={2.5} />,
      }}
      {...props}
    />
  )
}

export { Toaster }
