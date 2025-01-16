-- +migrate Up

UPDATE _auth_authenticator a SET type =
    CASE
        WHEN oob.channel = 'email' THEN 'oob_otp_email'
        WHEN oob.channel = 'sms' THEN 'oob_otp_sms'
    END
FROM 
    _auth_authenticator_oob oob
WHERE 
    oob.id = a.id AND a.type = 'oob_otp';

ALTER TABLE _auth_authenticator_oob DROP COLUMN channel;

-- +migrate Down

ALTER TABLE _auth_authenticator_oob ADD COLUMN channel text;

UPDATE _auth_authenticator_oob oob SET channel =
    CASE
        WHEN a.type = 'oob_otp_email' THEN 'email'
        WHEN a.type = 'oob_otp_sms' THEN 'sms'
    END
FROM 
    _auth_authenticator a
WHERE 
    oob.id = a.id;

ALTER TABLE _auth_authenticator_oob ALTER COLUMN channel SET NOT NULL;

UPDATE _auth_authenticator SET type = 'oob_otp'
WHERE 
    _auth_authenticator.type IN ('oob_otp_email', 'oob_otp_sms');
