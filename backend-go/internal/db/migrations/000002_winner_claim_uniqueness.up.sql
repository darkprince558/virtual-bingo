ALTER TABLE winners
  ADD CONSTRAINT winners_claim_unique UNIQUE (claim_id);
