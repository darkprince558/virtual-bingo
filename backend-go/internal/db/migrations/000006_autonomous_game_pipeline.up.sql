ALTER TABLE players
  ADD COLUMN player_icon text,
  ADD COLUMN player_avatar_color text,
  ADD COLUMN player_avatar_label text;

CREATE TABLE game_call_deck (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  game_run_id uuid NOT NULL REFERENCES game_runs(id) ON DELETE CASCADE,
  word_set_word_id uuid REFERENCES word_set_words(id) ON DELETE SET NULL,
  word text NOT NULL,
  sequence integer NOT NULL,
  shuffle_seed text NOT NULL,
  shuffle_version text NOT NULL,
  locked_at timestamptz NOT NULL DEFAULT now(),
  called_word_id uuid REFERENCES called_words(id) ON DELETE SET NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT game_call_deck_word_not_blank CHECK (btrim(word) <> ''),
  CONSTRAINT game_call_deck_sequence_positive CHECK (sequence > 0),
  CONSTRAINT game_call_deck_sequence_unique UNIQUE (game_run_id, sequence),
  CONSTRAINT game_call_deck_word_unique UNIQUE (game_run_id, word)
);

CREATE INDEX game_call_deck_game_run_called_idx ON game_call_deck (game_run_id, called_word_id, sequence);

CREATE TABLE caller_assets (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  game_run_id uuid NOT NULL REFERENCES game_runs(id) ON DELETE CASCADE,
  call_deck_item_id uuid NOT NULL REFERENCES game_call_deck(id) ON DELETE CASCADE,
  word text NOT NULL,
  sequence integer NOT NULL,
  line text NOT NULL,
  audio_url text,
  storage_key text,
  voice_name text,
  provider text NOT NULL DEFAULT 'unknown',
  status text NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'ready', 'failed', 'fallback')),
  error_reason text,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT caller_assets_deck_item_unique UNIQUE (call_deck_item_id),
  CONSTRAINT caller_assets_word_not_blank CHECK (btrim(word) <> ''),
  CONSTRAINT caller_assets_line_not_blank CHECK (btrim(line) <> ''),
  CONSTRAINT caller_assets_sequence_positive CHECK (sequence > 0)
);

CREATE INDEX caller_assets_game_run_sequence_idx ON caller_assets (game_run_id, sequence);
CREATE INDEX caller_assets_status_idx ON caller_assets (status);

CREATE TABLE delivery_batches (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  game_run_id uuid NOT NULL REFERENCES game_runs(id) ON DELETE CASCADE,
  channel text NOT NULL CHECK (channel IN ('email', 'teams')),
  purpose text NOT NULL CHECK (purpose IN ('host_review', 'player_invite', 'reminder', 'summary')),
  status text NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'sent', 'failed', 'skipped')),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX delivery_batches_game_run_idx ON delivery_batches (game_run_id, created_at DESC);

CREATE TABLE delivery_attempts (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  batch_id uuid NOT NULL REFERENCES delivery_batches(id) ON DELETE CASCADE,
  game_run_id uuid NOT NULL REFERENCES game_runs(id) ON DELETE CASCADE,
  channel text NOT NULL CHECK (channel IN ('email', 'teams')),
  purpose text NOT NULL CHECK (purpose IN ('host_review', 'player_invite', 'reminder', 'summary')),
  recipient_email text NOT NULL,
  recipient_user_id uuid REFERENCES users(id) ON DELETE SET NULL,
  subject text NOT NULL,
  template_key text NOT NULL,
  body_preview text NOT NULL,
  link_url text NOT NULL,
  game_code text NOT NULL,
  status text NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'sent', 'failed', 'skipped')),
  error_reason text,
  sent_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT delivery_attempts_recipient_email_not_blank CHECK (btrim(recipient_email) <> '')
);

CREATE INDEX delivery_attempts_game_run_idx ON delivery_attempts (game_run_id, created_at DESC);
CREATE INDEX delivery_attempts_status_idx ON delivery_attempts (status);

CREATE TABLE theme_generation_jobs (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  game_run_id uuid REFERENCES game_runs(id) ON DELETE SET NULL,
  status text NOT NULL DEFAULT 'running' CHECK (status IN ('running', 'succeeded', 'failed')),
  provider text NOT NULL DEFAULT 'unknown',
  prompt text NOT NULL,
  error_message text,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT theme_generation_jobs_prompt_not_blank CHECK (btrim(prompt) <> '')
);

CREATE INDEX theme_generation_jobs_game_run_idx ON theme_generation_jobs (game_run_id, created_at DESC);

CREATE TABLE themes (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  game_run_id uuid REFERENCES game_runs(id) ON DELETE SET NULL,
  generation_job_id uuid REFERENCES theme_generation_jobs(id) ON DELETE SET NULL,
  name text NOT NULL,
  summary text NOT NULL,
  tokens jsonb NOT NULL,
  status text NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'approved', 'rejected', 'archived')),
  provider text NOT NULL DEFAULT 'unknown',
  created_by_user_id uuid REFERENCES users(id) ON DELETE SET NULL,
  approved_by_user_id uuid REFERENCES users(id) ON DELETE SET NULL,
  approved_at timestamptz,
  rejected_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT themes_name_not_blank CHECK (btrim(name) <> ''),
  CONSTRAINT themes_summary_not_blank CHECK (btrim(summary) <> ''),
  CONSTRAINT themes_tokens_object CHECK (jsonb_typeof(tokens) = 'object')
);

CREATE INDEX themes_status_idx ON themes (status);
CREATE INDEX themes_game_run_idx ON themes (game_run_id, created_at DESC);

CREATE TABLE theme_approvals (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  theme_id uuid NOT NULL REFERENCES themes(id) ON DELETE CASCADE,
  game_run_id uuid REFERENCES game_runs(id) ON DELETE SET NULL,
  actor_user_id uuid REFERENCES users(id) ON DELETE SET NULL,
  status text NOT NULL CHECK (status IN ('approved', 'rejected')),
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX theme_approvals_theme_idx ON theme_approvals (theme_id, created_at DESC);
