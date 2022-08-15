-- +migrate Up
CREATE TABLE _auth_identity_siwe
(
	id text PRIMARY KEY REFERENCES _auth_identity (id),
	app_id text NOT NULL,
	chain_id integer NOT NULL,
	address text NOT NULL,
	data jsonb NOT NULL
);
ALTER TABLE _auth_identity_siwe
	ADD CONSTRAINT _auth_identity_address UNIQUE (app_id, chain_id, address);

-- +migrate Down
DROP TABLE _auth_identity_siwe;
DELETE FROM _auth_identity WHERE "type" = 'siwe';