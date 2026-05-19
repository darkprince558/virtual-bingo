CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE users (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  external_subject text UNIQUE,
  display_name text NOT NULL,
  email text NOT NULL,
  role text NOT NULL DEFAULT 'player' CHECK (role IN ('admin', 'host', 'player', 'viewer')),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT users_email_not_blank CHECK (btrim(email) <> '')
);

CREATE UNIQUE INDEX users_email_lower_idx ON users (lower(email));

CREATE TABLE host_privilege_requests (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  status text NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected', 'cancelled')),
  reason text,
  reviewed_by_user_id uuid REFERENCES users(id) ON DELETE SET NULL,
  reviewed_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX host_privilege_requests_user_id_idx ON host_privilege_requests (user_id);
CREATE INDEX host_privilege_requests_status_idx ON host_privilege_requests (status);

CREATE TABLE word_sets (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name text NOT NULL,
  status text NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'approved', 'archived')),
  source text NOT NULL DEFAULT 'manual' CHECK (source IN ('manual', 'seed', 'ai_generated')),
  created_by_user_id uuid REFERENCES users(id) ON DELETE SET NULL,
  approved_by_user_id uuid REFERENCES users(id) ON DELETE SET NULL,
  approved_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT word_sets_name_not_blank CHECK (btrim(name) <> '')
);

CREATE TABLE word_set_words (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  word_set_id uuid NOT NULL REFERENCES word_sets(id) ON DELETE CASCADE,
  word text NOT NULL,
  sort_order integer NOT NULL,
  is_active boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT word_set_words_word_not_blank CHECK (btrim(word) <> ''),
  CONSTRAINT word_set_words_sort_order_positive CHECK (sort_order > 0),
  CONSTRAINT word_set_words_sort_order_unique UNIQUE (word_set_id, sort_order),
  CONSTRAINT word_set_words_word_unique UNIQUE (word_set_id, word)
);

CREATE TABLE game_templates (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  host_user_id uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  default_word_set_id uuid REFERENCES word_sets(id) ON DELETE SET NULL,
  name text NOT NULL,
  status text NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'active', 'paused', 'archived')),
  recurrence_rule text,
  time_zone text NOT NULL DEFAULT 'America/Toronto',
  min_players integer NOT NULL DEFAULT 6,
  winning_pattern_mode text NOT NULL DEFAULT 'random' CHECK (winning_pattern_mode IN ('random', 'manual')),
  ai_content_mode text NOT NULL DEFAULT 'manual' CHECK (ai_content_mode IN ('manual', 'draft_review', 'autopilot')),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT game_templates_name_not_blank CHECK (btrim(name) <> ''),
  CONSTRAINT game_templates_min_players_positive CHECK (min_players > 0)
);

CREATE INDEX game_templates_host_user_id_idx ON game_templates (host_user_id);

CREATE TABLE game_runs (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  template_id uuid REFERENCES game_templates(id) ON DELETE SET NULL,
  host_user_id uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  word_set_id uuid REFERENCES word_sets(id) ON DELETE SET NULL,
  code text NOT NULL,
  name text NOT NULL,
  status text NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'content_generating', 'content_review', 'scheduled', 'invites_sent', 'lobby_open', 'live', 'paused', 'finished', 'reward_fulfillment_pending', 'complete', 'cancelled', 'failed')),
  scheduled_start_at timestamptz,
  started_at timestamptz,
  ended_at timestamptz,
  current_called_word_id uuid,
  winning_pattern text,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT game_runs_code_not_blank CHECK (btrim(code) <> ''),
  CONSTRAINT game_runs_name_not_blank CHECK (btrim(name) <> ''),
  CONSTRAINT game_runs_code_unique UNIQUE (code)
);

CREATE INDEX game_runs_template_id_idx ON game_runs (template_id);
CREATE INDEX game_runs_host_user_id_idx ON game_runs (host_user_id);
CREATE INDEX game_runs_status_idx ON game_runs (status);

CREATE TABLE allowed_players (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  game_run_id uuid NOT NULL REFERENCES game_runs(id) ON DELETE CASCADE,
  email text NOT NULL,
  display_name text NOT NULL,
  source text NOT NULL DEFAULT 'manual' CHECK (source IN ('manual', 'seed', 'graph_group', 'csv')),
  created_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT allowed_players_email_not_blank CHECK (btrim(email) <> ''),
  CONSTRAINT allowed_players_display_name_not_blank CHECK (btrim(display_name) <> '')
);

CREATE UNIQUE INDEX allowed_players_game_run_email_idx ON allowed_players (game_run_id, lower(email));

CREATE TABLE players (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  game_run_id uuid NOT NULL REFERENCES game_runs(id) ON DELETE CASCADE,
  user_id uuid REFERENCES users(id) ON DELETE SET NULL,
  email text NOT NULL,
  display_name text NOT NULL,
  connection_state text NOT NULL DEFAULT 'offline' CHECK (connection_state IN ('online', 'offline', 'disconnected')),
  state text NOT NULL DEFAULT 'joined' CHECK (state IN ('joined', 'waiting', 'playing', 'claimed_bingo', 'confirmed_winner', 'rejected_claim', 'disconnected')),
  joined_at timestamptz NOT NULL DEFAULT now(),
  last_seen_at timestamptz NOT NULL DEFAULT now(),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT players_email_not_blank CHECK (btrim(email) <> ''),
  CONSTRAINT players_display_name_not_blank CHECK (btrim(display_name) <> '')
);

CREATE UNIQUE INDEX players_game_run_email_idx ON players (game_run_id, lower(email));
CREATE INDEX players_game_run_id_idx ON players (game_run_id);

CREATE TABLE bingo_cards (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  game_run_id uuid NOT NULL REFERENCES game_runs(id) ON DELETE CASCADE,
  player_id uuid NOT NULL REFERENCES players(id) ON DELETE CASCADE,
  seed text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT bingo_cards_seed_not_blank CHECK (btrim(seed) <> ''),
  CONSTRAINT bingo_cards_player_unique UNIQUE (player_id)
);

CREATE INDEX bingo_cards_game_run_id_idx ON bingo_cards (game_run_id);

CREATE TABLE bingo_card_cells (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  card_id uuid NOT NULL REFERENCES bingo_cards(id) ON DELETE CASCADE,
  row_index integer NOT NULL,
  col_index integer NOT NULL,
  word text NOT NULL,
  is_free_space boolean NOT NULL DEFAULT false,
  marked_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT bingo_card_cells_row_range CHECK (row_index BETWEEN 0 AND 4),
  CONSTRAINT bingo_card_cells_col_range CHECK (col_index BETWEEN 0 AND 4),
  CONSTRAINT bingo_card_cells_word_not_blank CHECK (btrim(word) <> ''),
  CONSTRAINT bingo_card_cells_position_unique UNIQUE (card_id, row_index, col_index)
);

CREATE TABLE called_words (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  game_run_id uuid NOT NULL REFERENCES game_runs(id) ON DELETE CASCADE,
  word_set_word_id uuid REFERENCES word_set_words(id) ON DELETE SET NULL,
  word text NOT NULL,
  called_by_user_id uuid REFERENCES users(id) ON DELETE SET NULL,
  sequence integer NOT NULL,
  called_at timestamptz NOT NULL DEFAULT now(),
  created_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT called_words_word_not_blank CHECK (btrim(word) <> ''),
  CONSTRAINT called_words_sequence_positive CHECK (sequence > 0),
  CONSTRAINT called_words_sequence_unique UNIQUE (game_run_id, sequence),
  CONSTRAINT called_words_word_unique UNIQUE (game_run_id, word)
);

ALTER TABLE game_runs
  ADD CONSTRAINT game_runs_current_called_word_fk
  FOREIGN KEY (current_called_word_id) REFERENCES called_words(id) ON DELETE SET NULL;

CREATE TABLE bingo_claims (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  game_run_id uuid NOT NULL REFERENCES game_runs(id) ON DELETE CASCADE,
  player_id uuid NOT NULL REFERENCES players(id) ON DELETE CASCADE,
  pattern text NOT NULL,
  status text NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'valid', 'invalid', 'confirmed', 'rejected')),
  validation_result jsonb NOT NULL DEFAULT '{}'::jsonb,
  claimed_at timestamptz NOT NULL DEFAULT now(),
  reviewed_by_user_id uuid REFERENCES users(id) ON DELETE SET NULL,
  reviewed_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT bingo_claims_pattern_not_blank CHECK (btrim(pattern) <> '')
);

CREATE INDEX bingo_claims_game_run_id_idx ON bingo_claims (game_run_id);
CREATE INDEX bingo_claims_player_id_idx ON bingo_claims (player_id);
CREATE INDEX bingo_claims_status_idx ON bingo_claims (status);

CREATE TABLE winners (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  game_run_id uuid NOT NULL REFERENCES game_runs(id) ON DELETE CASCADE,
  player_id uuid NOT NULL REFERENCES players(id) ON DELETE CASCADE,
  claim_id uuid REFERENCES bingo_claims(id) ON DELETE SET NULL,
  placement integer NOT NULL,
  pattern text NOT NULL,
  confirmed_at timestamptz NOT NULL DEFAULT now(),
  created_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT winners_placement_range CHECK (placement BETWEEN 1 AND 3),
  CONSTRAINT winners_pattern_not_blank CHECK (btrim(pattern) <> ''),
  CONSTRAINT winners_placement_unique UNIQUE (game_run_id, placement)
);

CREATE INDEX winners_game_run_id_idx ON winners (game_run_id);
CREATE INDEX winners_player_id_idx ON winners (player_id);

CREATE TABLE audit_events (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  game_run_id uuid REFERENCES game_runs(id) ON DELETE SET NULL,
  actor_user_id uuid REFERENCES users(id) ON DELETE SET NULL,
  event_type text NOT NULL,
  entity_type text NOT NULL,
  entity_id uuid,
  payload jsonb NOT NULL DEFAULT '{}'::jsonb,
  created_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT audit_events_event_type_not_blank CHECK (btrim(event_type) <> ''),
  CONSTRAINT audit_events_entity_type_not_blank CHECK (btrim(entity_type) <> '')
);

CREATE INDEX audit_events_game_run_id_idx ON audit_events (game_run_id);
CREATE INDEX audit_events_actor_user_id_idx ON audit_events (actor_user_id);
CREATE INDEX audit_events_event_type_idx ON audit_events (event_type);
