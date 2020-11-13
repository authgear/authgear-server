// TODO(localizaton): allow localizing templates

export const TEMPLATE_SETUP_PRIMARY_OOB_EMAIL_HTML =
  "templates/en/messages/setup_primary_oob_email.html";
export const TEMPLATE_SETUP_PRIMARY_OOB_EMAIL_TEXT =
  "templates/en/messages/setup_primary_oob_email.txt";
export const TEMPLATE_SETUP_PRIMARY_OOB_SMS_TEXT =
  "templates/en/messages/setup_primary_oob_sms.txt";
export const SetupPrimaryOOBMessageTemplates = [
  TEMPLATE_SETUP_PRIMARY_OOB_EMAIL_HTML,
  TEMPLATE_SETUP_PRIMARY_OOB_EMAIL_TEXT,
  TEMPLATE_SETUP_PRIMARY_OOB_SMS_TEXT,
] as const;

export const TEMPLATE_AUTHENTICATE_PRIMARY_OOB_EMAIL_HTML =
  "templates/en/messages/authenticate_primary_oob_email.html";
export const TEMPLATE_AUTHENTICATE_PRIMARY_OOB_EMAIL_TEXT =
  "templates/en/messages/authenticate_primary_oob_email.txt";
export const TEMPLATE_AUTHENTICATE_PRIMARY_OOB_SMS_TEXT =
  "templates/en/messages/authenticate_primary_oob_sms.txt";
export const AuthenticatePrimaryOOBMessageTemplates = [
  TEMPLATE_AUTHENTICATE_PRIMARY_OOB_EMAIL_HTML,
  TEMPLATE_AUTHENTICATE_PRIMARY_OOB_EMAIL_TEXT,
  TEMPLATE_AUTHENTICATE_PRIMARY_OOB_SMS_TEXT,
] as const;

export const TEMPLATE_FORGOT_PASSWORD_EMAIL_HTML =
  "templates/en/messages/forgot_password_email.html";
export const TEMPLATE_FORGOT_PASSWORD_EMAIL_TEXT =
  "templates/en/messages/forgot_password_email.txt";
export const TEMPLATE_FORGOT_PASSWORD_SMS_TEXT =
  "templates/en/messages/forgot_password_sms.txt";
export const ForgotPasswordMessageTemplates = [
  TEMPLATE_FORGOT_PASSWORD_EMAIL_HTML,
  TEMPLATE_FORGOT_PASSWORD_EMAIL_TEXT,
  TEMPLATE_FORGOT_PASSWORD_SMS_TEXT,
] as const;

export const STATIC_AUTHGEAR_CSS = "static/authgear.css";
