-- This test fixes its app ID, because the Custom UI's origin has to be written
-- out in the config override and must sit on the same host as the public
-- origin. A fixed app ID means its data survives between runs against a live
-- environment, and a user who already has a passkey is not prompted to create
-- one — so the test would pass once and then fail.
--
-- Clearing the passkeys first makes it repeatable. CI starts from an empty
-- database and is unaffected either way.

DELETE FROM _auth_identity_passkey
WHERE id IN (
    SELECT id FROM _auth_identity
    WHERE app_id = 'e2e-pk-customui-v2' AND type = 'passkey'
);

DELETE FROM _auth_identity
WHERE app_id = 'e2e-pk-customui-v2' AND type = 'passkey';

DELETE FROM _auth_authenticator_passkey
WHERE id IN (
    SELECT id FROM _auth_authenticator
    WHERE app_id = 'e2e-pk-customui-v2' AND type = 'passkey'
);

DELETE FROM _auth_authenticator
WHERE app_id = 'e2e-pk-customui-v2' AND type = 'passkey';
