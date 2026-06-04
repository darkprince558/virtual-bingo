export interface ThemePalette {
  primary: string;
  secondary: string;
  accent: string;
}

export interface ThemeMusic {
  genre: string;
  mood: string;
  description: string;
}

export interface Theme {
  name: string;
  icon: string;
  palette: ThemePalette;
  music: ThemeMusic;
  voice_recommendation: string;
  phrases: string[];
}
