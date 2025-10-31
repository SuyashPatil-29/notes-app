export const themeColors = {
  claude: {
    primary: "hsl(15.1111, 55.5556%, 52.3529%)",
    secondary: "hsl(46.1538, 22.8070%, 88.8235%)",
    accent: "hsl(50, 7.5000%, 84.3137%)",
    background: "hsl(46.1538, 22.8070%, 88.8235%)",
  },
  tech: {
    primary: "hsl(21.7450, 65.6388%, 55.4902%)",
    secondary: "hsl(180, 17.5879%, 39.0196%)",
    accent: "hsl(0, 0%, 93.3333%)",
    background: "hsl(220, 13.0435%, 90.9804%)",
  },
  minimal: {
    primary: "hsl(0, 0%, 94.1176%)",
    secondary: "hsl(0, 0%, 20%)",
    accent: "hsl(0, 0%, 37.6471%)",
    background: "hsl(0, 0%, 81.5686%)",
  },
  gruvbox: {
    primary: "hsl(23.7, 87.7193%, 44.7059%)",
    secondary: "hsl(43.1579, 58.7629%, 80.9804%)",
    accent: "hsl(48.4615, 86.6667%, 88.2353%)",
    background: "hsl(41.9704, 95.3052%, 58.2353%)",
  },
  notebook: {
    primary: "hsl(0, 0%, 97.6471%)",
    secondary: "hsl(0, 0%, 22.7451%)",
    accent: "hsl(47.4419, 64.1791%, 86.8627%)",
    background: "hsl(0, 0%, 45.0980%)",
  },
  supabase: {
    primary: "hsl(151.3274, 66.8639%, 66.8627%)",
    secondary: "hsl(0, 0%, 99.2157%)",
    accent: "hsl(0, 0%, 98.8235%)",
    background: "hsl(0, 0%, 92.9412%)",
  },
  pink: {
    primary: "hsl(333.2673, 42.9787%, 46.0784%)",
    secondary: "hsl(314.6667, 61.6438%, 85.6863%)",
    accent: "hsl(314.6667, 61.6438%, 85.6863%)",
    background: "hsl(304.8000, 60.9756%, 83.9216%)",
  },
  orange: {
    primary: "hsl(15.1111, 55.5556%, 52.3529%)",
    secondary: "hsl(46.1538, 22.8070%, 88.8235%)",
    accent: "hsl(46.1538, 22.8070%, 88.8235%)",
    background: "hsl(50, 7.5000%, 84.3137%)",
  },
};

export type ThemeName = keyof typeof themeColors;

export function getThemeColors(theme: string) {
  return themeColors[theme as ThemeName] || themeColors.notebook;
}

