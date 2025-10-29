import { useTheme } from "@/components/theme-provider";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Button } from "@/components/ui/button";
import { Palette } from "lucide-react";
import { useEffect, useState } from "react";

const themes = [
  {
    name: "Claude",
    value: "orange",
    colors: [
      "hsl(15.1111, 55.5556%, 52.3529%)",
      "hsl(46.1538, 22.8070%, 88.8235%)",
      "hsl(46.1538, 22.8070%, 88.8235%)",
      "hsl(50, 7.5000%, 84.3137%)",
    ],
  },
  {
    name: "Dark Matter",
    value: "tech",
    colors: [
      "hsl(21.7450, 65.6388%, 55.4902%)",
      "hsl(180, 17.5879%, 39.0196%)",
      "hsl(0, 0%, 93.3333%)",
      "hsl(220, 13.0435%, 90.9804%)",
    ],
  },
  {
    name: "Graphite",
    value: "minimal",
    colors: [
      "hsl(0, 0%, 94.1176%)",
      "hsl(0, 0%, 20%)",
      "hsl(0, 0%, 37.6471%)",
      "hsl(0, 0%, 81.5686%)",
    ],
  },
  {
    name: "Gruvbox",
    value: "gruvbox",
    colors: [
      "hsl(23.7, 87.7193%, 44.7059%)",
      "hsl(43.1579, 58.7629%, 80.9804%)",
      "hsl(48.4615, 86.6667%, 88.2353%)",
      "hsl(41.9704, 95.3052%, 58.2353%)",
    ],
  },
  {
    name: "Notebook",
    value: "notebook",
    colors: [
      "hsl(0, 0%, 97.6471%)",
      "hsl(0, 0%, 22.7451%)",
      "hsl(47.4419, 64.1791%, 86.8627%)",
      "hsl(0, 0%, 45.0980%)",
    ],
  },
  {
    name: "Supabase",
    value: "supabase",
    colors: [
      "hsl(151.3274, 66.8639%, 66.8627%)",
      "hsl(0, 0%, 99.2157%)",
      "hsl(0, 0%, 98.8235%)",
      "hsl(0, 0%, 92.9412%)",
    ],
  },
  {
    name: "T3 Chat",
    value: "pink",
    colors: [
      "hsl(333.2673, 42.9787%, 46.0784%)",
      "hsl(314.6667, 61.6438%, 85.6863%)",
      "hsl(314.6667, 61.6438%, 85.6863%)",
      "hsl(304.8000, 60.9756%, 83.9216%)",
    ],
  },
];

export function ThemeSelector() {
  const { theme, setTheme } = useTheme();
  const [mounted, setMounted] = useState(false);

  useEffect(() => setMounted(true), []);

  if (!mounted) return null;

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" size="icon" className="relative">
          <Palette className="h-5 w-5" />
          <span className="sr-only">Select theme</span>
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="w-56">
        <DropdownMenuRadioGroup 
          value={theme} 
          onValueChange={(value) => setTheme(value as "notebook" | "tech" | "minimal" | "orange" | "pink" | "supabase" | "gruvbox")}
        >
        {themes.map((themeOption) => (
            <DropdownMenuRadioItem
              key={themeOption.value}
              value={themeOption.value}
              className="cursor-pointer"
            >
              <div className="flex items-center gap-3 w-full">
                <span className="min-w-[80px] text-sm">{themeOption.name}</span>
                <div className="flex gap-0.5 ml-auto">
                {themeOption.colors.map((color, index) => (
                  <div
                    key={index}
                    style={{ backgroundColor: color }}
                      className="w-4 h-4 rounded-sm border border-border"
                  />
                ))}
                </div>
              </div>
            </DropdownMenuRadioItem>
        ))}
        </DropdownMenuRadioGroup>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
