-- Put upgrade SQL here
UPDATE config
SET config = jsonb_set(
  config,
  '{user_config,auth,login_id_keys}',
  jsonb_build_object(
    'email', jsonb_build_object(
      'type', 'email'
    ),
    'username', jsonb_build_object(
      'type', 'raw'
    ),
    'phone', jsonb_build_object(
      'type', 'phone'
    )
  )
);
