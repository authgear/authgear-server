{{ $userID := (uuidv4) }}
{{ $identityID := (uuidv4) }}

INSERT INTO _auth_user (
  "id",
  "app_id",
  "created_at",
  "updated_at",
  "is_disabled",
  "standard_attributes"
) VALUES (
  '{{ $userID }}',
  '{{ .AppID }}',
  NOW(),
  NOW(),
  FALSE,
  '{"testuserid": "1"}'
);

INSERT INTO _auth_identity (
  "id",
  "app_id",
  "type",
  "user_id",
  "created_at",
  "updated_at"
) VALUES (
  '{{ $identityID }}',
  '{{ .AppID }}',
  'anonymous',
  '{{ $userID }}',
  NOW(),
  NOW()
);

INSERT INTO _auth_identity_anonymous (
  "id",
  "app_id",
  "key_id",
  "key"
) VALUES (
  '{{ $identityID }}',
  '{{ .AppID }}',
  '4ABDD8A7-2102-4AA0-A229-299456A45B33',
  $${
    "kid": "4ABDD8A7-2102-4AA0-A229-299456A45B33",
    "alg": "RS256",
    "kty": "RSA",
    "n": "gqP1JZx3RJmCuEfGznR9Yhqrh78Ty9vRAT-FauLzpIOMQ1u0S2L_rfQqAwiI2S73uWGXjnDoJ_lnp72b6Mi_ZagbnAbJQ7lWWX8LxgYwWAm8AxX32Q-gQyxcEAhhlUxDsWhknpBakdDS06hoTVSrUbt60I7EhMMaQuhz1Js4KTGSoBn0QXASBcLDxd0jUAc0frCW0SDvy5bJCUKUTHhmXDjYDc_hRm9PYGrccC8lDpXxoLldCshWUZeurmhUNaNwmkMhlf95_lB2WX7fUNwYb36J2vegoVDvtymUHOpYKNKvyGm7QYHaSj8u8dgOP7z5IrXctrztYOSuLZkL7iSC3Q",
    "e": "AQAB"
  }$$
);
