import type { Theme } from "../types/theme";

export async function generateTheme(topic: string): Promise<Theme> {
  const res = await fetch("/api/theme/generate", {
    method: "POST",
    headers: {
      "Content-Type": "application/json"
    },
    body: JSON.stringify({ topic })
  });

  if (!res.ok) {
    throw new Error("Theme generation failed");
  }

  return res.json();
}
