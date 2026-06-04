import { useState } from "react";
import type { Theme } from "../types/theme";
import { generateTheme } from "../api/themeApi";

export function useThemeGenerator() {
  const [theme, setTheme] = useState<Theme | null>(null);
  const [loading, setLoading] = useState(false);

  async function createTheme(topic: string) {
    setLoading(true);

    try {
      const result = await generateTheme(topic);
      setTheme(result);
    } finally {
      setLoading(false);
    }
  }

  return {
    theme,
    loading,
    createTheme
  };
}
