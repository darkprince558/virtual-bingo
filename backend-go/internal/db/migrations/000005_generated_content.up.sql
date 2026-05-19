CREATE TABLE content_generation_jobs (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  game_run_id uuid NOT NULL REFERENCES game_runs(id) ON DELETE CASCADE,
  job_type text NOT NULL CHECK (job_type IN ('game_prep')),
  status text NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'running', 'succeeded', 'failed')),
  provider text NOT NULL DEFAULT 'unknown',
  error_message text,
  retry_count integer NOT NULL DEFAULT 0,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX content_generation_jobs_game_run_id_idx ON content_generation_jobs (game_run_id, created_at DESC);
CREATE INDEX content_generation_jobs_status_idx ON content_generation_jobs (status);

CREATE TABLE generated_game_content (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  game_run_id uuid NOT NULL REFERENCES game_runs(id) ON DELETE CASCADE,
  generation_job_id uuid REFERENCES content_generation_jobs(id) ON DELETE SET NULL,
  status text NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'generated', 'edited', 'locked', 'failed')),
  topic text NOT NULL,
  summary text NOT NULL,
  generated_words jsonb NOT NULL DEFAULT '[]'::jsonb,
  current_words jsonb NOT NULL DEFAULT '[]'::jsonb,
  caller_style text,
  theme_prompt text,
  review_window_opens_at timestamptz,
  review_window_closes_at timestamptz,
  locked_at timestamptz,
  locked_word_set_id uuid REFERENCES word_sets(id) ON DELETE SET NULL,
  generation_provider text NOT NULL DEFAULT 'unknown',
  generation_error text,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT generated_game_content_game_run_unique UNIQUE (game_run_id),
  CONSTRAINT generated_game_content_topic_not_blank CHECK (btrim(topic) <> ''),
  CONSTRAINT generated_game_content_summary_not_blank CHECK (btrim(summary) <> ''),
  CONSTRAINT generated_game_content_words_arrays CHECK (jsonb_typeof(generated_words) = 'array' AND jsonb_typeof(current_words) = 'array')
);

CREATE INDEX generated_game_content_status_idx ON generated_game_content (status);
CREATE INDEX generated_game_content_locked_at_idx ON generated_game_content (locked_at);

CREATE TABLE game_run_content_reviews (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  game_run_id uuid NOT NULL REFERENCES game_runs(id) ON DELETE CASCADE,
  content_id uuid NOT NULL REFERENCES generated_game_content(id) ON DELETE CASCADE,
  actor_user_id uuid REFERENCES users(id) ON DELETE SET NULL,
  edited_topic text,
  edited_summary text,
  edited_words jsonb NOT NULL DEFAULT '[]'::jsonb,
  caller_style text,
  created_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT game_run_content_reviews_words_array CHECK (jsonb_typeof(edited_words) = 'array')
);

CREATE INDEX game_run_content_reviews_game_run_id_idx ON game_run_content_reviews (game_run_id, created_at DESC);
