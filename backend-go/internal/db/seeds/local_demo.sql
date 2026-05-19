INSERT INTO users (id, external_subject, display_name, email, role)
VALUES
  ('00000000-0000-0000-0000-000000000101', 'local-host', 'Local Demo Host', 'host@example.local', 'host')
ON CONFLICT (id) DO UPDATE
SET display_name = EXCLUDED.display_name,
    email = EXCLUDED.email,
    role = EXCLUDED.role,
    updated_at = now();

INSERT INTO word_sets (id, name, status, source, created_by_user_id, approved_by_user_id, approved_at)
VALUES
  ('00000000-0000-0000-0000-000000000201', 'Local Demo Workplace Bingo', 'approved', 'seed', '00000000-0000-0000-0000-000000000101', '00000000-0000-0000-0000-000000000101', now())
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
  ('00000000-0000-0000-0000-000000000201', 'Client Demo', 6),
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
  ('00000000-0000-0000-0000-000000000401', '00000000-0000-0000-0000-000000000301', '00000000-0000-0000-0000-000000000101', '00000000-0000-0000-0000-000000000201', 'LOCAL-DEMO', 'Local Demo Game', 'lobby_open', now() + interval '1 day', 'single_line')
ON CONFLICT (id) DO UPDATE
SET name = EXCLUDED.name,
    status = EXCLUDED.status,
    word_set_id = EXCLUDED.word_set_id,
    updated_at = now();

INSERT INTO allowed_players (game_run_id, email, display_name, source)
VALUES
  ('00000000-0000-0000-0000-000000000401', 'alex@example.local', 'Alex Demo', 'seed'),
  ('00000000-0000-0000-0000-000000000401', 'sam@example.local', 'Sam Demo', 'seed'),
  ('00000000-0000-0000-0000-000000000401', 'taylor@example.local', 'Taylor Demo', 'seed')
ON CONFLICT (game_run_id, (lower(email))) DO UPDATE
SET display_name = EXCLUDED.display_name,
    source = EXCLUDED.source;
