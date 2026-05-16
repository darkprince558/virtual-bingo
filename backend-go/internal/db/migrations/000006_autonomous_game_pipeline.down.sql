DROP TABLE IF EXISTS theme_approvals;
DROP TABLE IF EXISTS themes;
DROP TABLE IF EXISTS theme_generation_jobs;
DROP TABLE IF EXISTS delivery_attempts;
DROP TABLE IF EXISTS delivery_batches;
DROP TABLE IF EXISTS caller_assets;
DROP TABLE IF EXISTS game_call_deck;

ALTER TABLE players
  DROP COLUMN IF EXISTS player_avatar_label,
  DROP COLUMN IF EXISTS player_avatar_color,
  DROP COLUMN IF EXISTS player_icon;
