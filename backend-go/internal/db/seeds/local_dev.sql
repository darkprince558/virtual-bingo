BEGIN;

CREATE TEMP TABLE local_locked_word_sets AS
SELECT locked_word_set_id AS id
FROM generated_game_content
WHERE game_run_id = '00000000-0000-0000-0000-000000000401'
  AND locked_word_set_id IS NOT NULL;

UPDATE game_runs
SET current_called_word_id = NULL,
    word_set_id = '00000000-0000-0000-0000-000000000201',
    status = 'lobby_open',
    started_at = NULL,
    ended_at = NULL,
    updated_at = now()
WHERE id = '00000000-0000-0000-0000-000000000401';

DELETE FROM player_preferences
WHERE player_id IN (
  SELECT id FROM players WHERE game_run_id = '00000000-0000-0000-0000-000000000401'
);

DELETE FROM winners WHERE game_run_id = '00000000-0000-0000-0000-000000000401';
DELETE FROM bingo_claims WHERE game_run_id = '00000000-0000-0000-0000-000000000401';
DELETE FROM bingo_card_cells
WHERE card_id IN (
  SELECT id FROM bingo_cards WHERE game_run_id = '00000000-0000-0000-0000-000000000401'
);
DELETE FROM bingo_cards WHERE game_run_id = '00000000-0000-0000-0000-000000000401';
DELETE FROM players WHERE game_run_id = '00000000-0000-0000-0000-000000000401';
DELETE FROM called_words WHERE game_run_id = '00000000-0000-0000-0000-000000000401';
DELETE FROM caller_assets WHERE game_run_id = '00000000-0000-0000-0000-000000000401';
DELETE FROM game_call_deck WHERE game_run_id = '00000000-0000-0000-0000-000000000401';
DELETE FROM delivery_attempts WHERE game_run_id = '00000000-0000-0000-0000-000000000401';
DELETE FROM delivery_batches WHERE game_run_id = '00000000-0000-0000-0000-000000000401';
DELETE FROM theme_approvals WHERE game_run_id = '00000000-0000-0000-0000-000000000401';
DELETE FROM themes WHERE game_run_id = '00000000-0000-0000-0000-000000000401';
DELETE FROM theme_generation_jobs WHERE game_run_id = '00000000-0000-0000-0000-000000000401';
DELETE FROM game_run_content_reviews WHERE game_run_id = '00000000-0000-0000-0000-000000000401';
DELETE FROM generated_game_content WHERE game_run_id = '00000000-0000-0000-0000-000000000401';
DELETE FROM content_generation_jobs WHERE game_run_id = '00000000-0000-0000-0000-000000000401';
DELETE FROM game_run_settings WHERE game_run_id = '00000000-0000-0000-0000-000000000401';
DELETE FROM game_event_outbox WHERE game_run_id = '00000000-0000-0000-0000-000000000401';
DELETE FROM audit_events WHERE game_run_id = '00000000-0000-0000-0000-000000000401';
DELETE FROM word_sets WHERE id IN (SELECT id FROM local_locked_word_sets);

INSERT INTO users (id, external_subject, display_name, email, role)
VALUES
  ('00000000-0000-0000-0000-000000000101', 'local-host', 'Local Development Host', 'host@example.local', 'host')
ON CONFLICT (id) DO UPDATE
SET display_name = EXCLUDED.display_name,
    email = EXCLUDED.email,
    role = EXCLUDED.role,
    updated_at = now();

INSERT INTO word_sets (id, name, status, source, created_by_user_id, approved_by_user_id, approved_at)
VALUES
  ('00000000-0000-0000-0000-000000000201', 'Local Development Workplace Bingo', 'approved', 'seed', '00000000-0000-0000-0000-000000000101', '00000000-0000-0000-0000-000000000101', now())
ON CONFLICT (id) DO UPDATE
SET name = EXCLUDED.name,
    status = EXCLUDED.status,
    updated_at = now();

INSERT INTO word_set_words (word_set_id, word, sort_order)
VALUES
  ('00000000-0000-0000-0000-000000000201', 'Standup', 1),
  ('00000000-0000-0000-0000-000000000201', 'Sprint Planning', 2),
  ('00000000-0000-0000-0000-000000000201', 'Code Review', 3),
  ('00000000-0000-0000-0000-000000000201', 'Deployment', 4),
  ('00000000-0000-0000-0000-000000000201', 'Retrospective', 5),
  ('00000000-0000-0000-0000-000000000201', 'Client Review', 6),
  ('00000000-0000-0000-0000-000000000201', 'Documentation', 7),
  ('00000000-0000-0000-0000-000000000201', 'Bug Bash', 8),
  ('00000000-0000-0000-0000-000000000201', 'Architecture', 9),
  ('00000000-0000-0000-0000-000000000201', 'Coffee Chat', 10),
  ('00000000-0000-0000-0000-000000000201', 'Pull Request', 11),
  ('00000000-0000-0000-0000-000000000201', 'Standards Review', 12),
  ('00000000-0000-0000-0000-000000000201', 'Lunch and Learn', 13),
  ('00000000-0000-0000-0000-000000000201', 'Release Notes', 14),
  ('00000000-0000-0000-0000-000000000201', 'Security Review', 15),
  ('00000000-0000-0000-0000-000000000201', 'Accessibility', 16),
  ('00000000-0000-0000-0000-000000000201', 'Mentorship', 17),
  ('00000000-0000-0000-0000-000000000201', 'Knowledge Transfer', 18),
  ('00000000-0000-0000-0000-000000000201', 'Retest', 19),
  ('00000000-0000-0000-0000-000000000201', 'Backlog Grooming', 20),
  ('00000000-0000-0000-0000-000000000201', 'Pair Programming', 21),
  ('00000000-0000-0000-0000-000000000201', 'Environment Setup', 22),
  ('00000000-0000-0000-0000-000000000201', 'Ticket Refinement', 23),
  ('00000000-0000-0000-0000-000000000201', 'Design Review', 24),
  ('00000000-0000-0000-0000-000000000201', 'Production Check', 25),
  ('00000000-0000-0000-0000-000000000201', 'Team Win', 26)
ON CONFLICT (word_set_id, sort_order) DO UPDATE
SET word = EXCLUDED.word,
    is_active = true;

INSERT INTO game_templates (id, host_user_id, default_word_set_id, name, status, recurrence_rule, time_zone, min_players)
VALUES
  ('00000000-0000-0000-0000-000000000301', '00000000-0000-0000-0000-000000000101', '00000000-0000-0000-0000-000000000201', 'Friday Team Bingo', 'active', 'FREQ=WEEKLY;BYDAY=FR', 'America/Toronto', 6)
ON CONFLICT (id) DO UPDATE
SET name = EXCLUDED.name,
    status = EXCLUDED.status,
    default_word_set_id = EXCLUDED.default_word_set_id,
    updated_at = now();

INSERT INTO game_runs (id, template_id, host_user_id, word_set_id, code, name, status, scheduled_start_at, winning_pattern)
VALUES
  ('00000000-0000-0000-0000-000000000401', '00000000-0000-0000-0000-000000000301', '00000000-0000-0000-0000-000000000101', '00000000-0000-0000-0000-000000000201', 'LOCAL-DEV', 'Local Development Game', 'lobby_open', now() + interval '1 day', 'single_line')
ON CONFLICT (id) DO UPDATE
SET name = EXCLUDED.name,
    status = EXCLUDED.status,
    word_set_id = EXCLUDED.word_set_id,
    scheduled_start_at = EXCLUDED.scheduled_start_at,
    winning_pattern = EXCLUDED.winning_pattern,
    updated_at = now();

INSERT INTO allowed_players (game_run_id, email, display_name, source)
VALUES
  ('00000000-0000-0000-0000-000000000401', 'alex@example.local', 'Alex Local', 'seed'),
  ('00000000-0000-0000-0000-000000000401', 'sam@example.local', 'Sam Local', 'seed'),
  ('00000000-0000-0000-0000-000000000401', 'taylor@example.local', 'Taylor Local', 'seed')
ON CONFLICT (game_run_id, (lower(email))) DO UPDATE
SET display_name = EXCLUDED.display_name,
    source = EXCLUDED.source;

COMMIT;
