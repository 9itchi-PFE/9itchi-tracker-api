-- +goose Up
CREATE TABLE tracking_records (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  tracking_id text NOT NULL,
  email_id uuid NOT NULL,
  user_id uuid NOT NULL,
  sent_at timestamptz NOT NULL DEFAULT now(),
  first_open_at timestamptz,
  last_open_at timestamptz,
  open_count integer NOT NULL DEFAULT 0 CHECK (open_count >= 0),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT tracking_records_tracking_id_key UNIQUE (tracking_id),
  CONSTRAINT tracking_records_email_id_key UNIQUE (email_id)
);
CREATE INDEX idx_tracking_records_user_id ON tracking_records (user_id);

CREATE TABLE tracking_opens (
  id bigserial PRIMARY KEY,
  tracking_id text NOT NULL REFERENCES tracking_records (tracking_id) ON DELETE CASCADE,
  opened_at timestamptz NOT NULL DEFAULT now(),
  user_agent varchar(255) NOT NULL DEFAULT '',
  ip_hash char(64) NOT NULL
);
CREATE INDEX idx_tracking_opens_tracking_id_opened_at ON tracking_opens (tracking_id, opened_at DESC);

-- +goose Down
DROP TABLE tracking_opens;
DROP TABLE tracking_records;
