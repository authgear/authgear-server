-- Put downgrade SQL here
UPDATE config
SET config = jsonb_set(
  config,
  '{user_config,auth,login_id_keys}',
  jsonb_build_array('email', 'username', 'phone')
);
