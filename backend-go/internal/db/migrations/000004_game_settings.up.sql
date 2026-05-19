CREATE TABLE game_run_settings (
  game_run_id uuid PRIMARY KEY REFERENCES game_runs(id) ON DELETE CASCADE,
  marking_mode text NOT NULL DEFAULT 'manual' CHECK (marking_mode IN ('manual', 'assist', 'auto_mark')),
  allow_player_marking_mode_choice boolean NOT NULL DEFAULT false,
  show_claim_readiness boolean NOT NULL DEFAULT true,
  voice_claim_mode text NOT NULL DEFAULT 'off' CHECK (voice_claim_mode IN ('off', 'optional', 'required')),
  voice_claim_autoplay boolean NOT NULL DEFAULT false,
  caller_mode text NOT NULL DEFAULT 'off' CHECK (caller_mode IN ('off', 'text_only', 'tts')),
  theme_mode text NOT NULL DEFAULT 'default' CHECK (theme_mode IN ('default', 'manual', 'ai_generated')),
  theme_id uuid,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE player_preferences (
  player_id uuid PRIMARY KEY REFERENCES players(id) ON DELETE CASCADE,
  marking_mode text CHECK (marking_mode IS NULL OR marking_mode IN ('manual', 'assist', 'auto_mark')),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);
