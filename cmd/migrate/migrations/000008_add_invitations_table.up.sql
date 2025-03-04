CREATE TABLE IF NOT EXISTS invitations (
    id bigserial PRIMARY KEY,
    user_id bigint NOT NULL,
    token bytea NOT NULL,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),

    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_invitations_token ON invitations (token);


