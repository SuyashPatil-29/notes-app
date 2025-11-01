import { cn } from "@/lib/utils"

function Skeleton({ className, ...props }: React.ComponentProps<"div">) {
  return (
    <div
      data-slot="skeleton"
      className={cn(
        // Base skeleton with improved dark mode contrast
        "bg-muted/80 dark:bg-muted/60 animate-pulse rounded-md",
        // Slightly more visible in dark mode
        "dark:opacity-80",
        className
      )}
      {...props}
    />
  )
}

export { Skeleton }
