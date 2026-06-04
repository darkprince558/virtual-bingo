import { useVoice } from "../hooks/useVoice";

const VOICES = [
  {
    id: "us_neutral",
    name: "US English (Default)"
  },
  {
    id: "us_female_warm",
    name: "Warm Female (US)"
  },
  {
    id: "us_male_broadcaster",
    name: "Broadcaster (US)"
  },
  {
    id: "uk_posh",
    name: "UK Posh"
  },
  {
    id: "spooky",
    name: "Spooky"
  }
];

export function VoiceSelector() {
  const { voice, setVoice } = useVoice();

  return (
    <div>
      <label>Voice</label>

      <select
        value={voice.id}
        onChange={(e) => {
          const selected = VOICES.find(v => v.id === e.target.value);
          if (selected) setVoice(selected as any);
        }}
      >
        {VOICES.map(v => (
          <option key={v.id} value={v.id}>
            {v.name}
          </option>
        ))}
      </select>

      <p>Current: {voice.name}</p>
    </div>
  );
}
