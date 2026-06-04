import { useState } from "react";
import { useThemeGenerator } from "../hooks/useThemeGenerator";

export function ThemeGenerator() {
  const [input, setInput] = useState("");
  const { theme, createTheme, loading } = useThemeGenerator();

  return (
    <div style={{ padding: 20 }}>
      <h2>AI Theme Generator</h2>

      <input
        value={input}
        placeholder="e.g. Christmas, Formula 1, Harry Potter..."
        onChange={(e) => setInput(e.target.value)}
      />

      <button
        onClick={() => createTheme(input)}
        disabled={!input || loading}
      >
        {loading ? "Generating..." : "Generate Theme"}
      </button>

      {theme && (
        <div style={{ marginTop: 20 }}>
          <h3>
            {theme.icon} {theme.name}
          </h3>

          <p>{theme.music.genre} - {theme.music.mood}</p>

          <p>{theme.music.description}</p>

          <div>
            <strong>Voice:</strong> {theme.voice_recommendation}
          </div>

          <ul>
            {theme.phrases.map((p, i) => (
              <li key={i}>{p}</li>
            ))}
          </ul>
        </div>
      )}
    </div>
  );
}
