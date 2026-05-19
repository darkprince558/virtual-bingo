CREATE TABLE game_event_outbox (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  game_run_id uuid NOT NULL REFERENCES game_runs(id) ON DELETE CASCADE,
  type text NOT NULL,
  entity_id uuid,
  payload jsonb NOT NULL DEFAULT '{}'::jsonb,
  sequence bigint NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT game_event_outbox_type_not_blank CHECK (btrim(type) <> ''),
  CONSTRAINT game_event_outbox_sequence_positive CHECK (sequence > 0),
  CONSTRAINT game_event_outbox_sequence_unique UNIQUE (game_run_id, sequence)
);

CREATE INDEX game_event_outbox_game_run_sequence_idx ON game_event_outbox (game_run_id, sequence);
CREATE INDEX game_event_outbox_type_idx ON game_event_outbox (type);
