import { useState } from "react";
import type { Voice } from "../types/voice";

const DEFAULT_VOICE: Voice = {
  id: "us_neutral",
  name: "US English (Default)",
  locale: "en-US",
  style: "neutral"
};

export function useVoice() {
  const [voice, setVoice] = useState<Voice>(DEFAULT_VOICE);

  return {
    voice,
    setVoice
  };
}
