-- +goose Up
-- Draft schema generated from models/db. Refine (constraints, indexes, FKs) as features land;
-- nothing is deployed yet, so editing this file is fine until the first release.
CREATE TABLE tracking_records (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  tracking_id uuid NOT NULL,
  email_id uuid NOT NULL,
  user_id uuid NOT NULL,
  sent_at timestamptz NOT NULL DEFAULT now(),
  first_open_at timestamptz,
  open_count integer NOT NULL DEFAULT 0,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE tracking_records;
