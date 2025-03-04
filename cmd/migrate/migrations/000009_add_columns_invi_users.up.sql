ALTER TABLE users
ADD COLUMN active bool DEFAULT false;

ALTER TABLE invitations
ADD COLUMN expires_at timestamp(0) NOT NULL DEFAULT NOW();
