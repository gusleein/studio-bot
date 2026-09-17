ALTER TABLE rents DROP CONSTRAINT rents_type_check;
ALTER TABLE rents ADD CONSTRAINT rents_type_check CHECK (type IN ('dj', 'production', 'admin', 'owner'));