import {
  resourcePath,
  ResourceDefinition,
  LanguageTag,
  FALLBACK_EFFECTIVE_DATA,
} from "./util/resource";

export const DEFAULT_TEMPLATE_LOCALE: LanguageTag = "en";

export const IMAGE_EXTENSIONS: string[] = [".png", ".jpeg", ".gif"];

export const RESOURCE_TRANSLATION_JSON: ResourceDefinition = {
  resourcePath: resourcePath`templates/${"locale"}/translation.json`,
  type: "text",
  extensions: [],
  fallback: {
    kind: "Const",
    fallbackValue: "{}",
  },
};

export const RESOURCE_SETUP_PRIMARY_OOB_EMAIL_HTML: ResourceDefinition = {
  resourcePath: resourcePath`templates/${"locale"}/messages/setup_primary_oob_email.html`,
  type: "text",
  extensions: [],
  fallback: FALLBACK_EFFECTIVE_DATA,
};
export const RESOURCE_SETUP_PRIMARY_OOB_EMAIL_TXT: ResourceDefinition = {
  resourcePath: resourcePath`templates/${"locale"}/messages/setup_primary_oob_email.txt`,
  type: "text",
  extensions: [],
  fallback: FALLBACK_EFFECTIVE_DATA,
};
export const RESOURCE_SETUP_PRIMARY_OOB_SMS_TXT: ResourceDefinition = {
  resourcePath: resourcePath`templates/${"locale"}/messages/setup_primary_oob_sms.txt`,
  type: "text",
  extensions: [],
  fallback: FALLBACK_EFFECTIVE_DATA,
};

export const RESOURCE_AUTHENTICATE_PRIMARY_OOB_EMAIL_HTML: ResourceDefinition =
  {
    resourcePath: resourcePath`templates/${"locale"}/messages/authenticate_primary_oob_email.html`,
    type: "text",
    extensions: [],
    fallback: FALLBACK_EFFECTIVE_DATA,
  };
export const RESOURCE_AUTHENTICATE_PRIMARY_OOB_EMAIL_TXT: ResourceDefinition = {
  resourcePath: resourcePath`templates/${"locale"}/messages/authenticate_primary_oob_email.txt`,
  type: "text",
  extensions: [],
  fallback: FALLBACK_EFFECTIVE_DATA,
};
export const RESOURCE_AUTHENTICATE_PRIMARY_OOB_SMS_TXT: ResourceDefinition = {
  resourcePath: resourcePath`templates/${"locale"}/messages/authenticate_primary_oob_sms.txt`,
  type: "text",
  extensions: [],
  fallback: FALLBACK_EFFECTIVE_DATA,
};

export const RESOURCE_SETUP_PRIMARY_LOGIN_LINK_HTML: ResourceDefinition = {
  resourcePath: resourcePath`templates/${"locale"}/messages/setup_primary_login_link.html`,
  type: "text",
  extensions: [],
  fallback: FALLBACK_EFFECTIVE_DATA,
};
export const RESOURCE_SETUP_PRIMARY_LOGIN_LINK_TXT: ResourceDefinition = {
  resourcePath: resourcePath`templates/${"locale"}/messages/setup_primary_login_link.txt`,
  type: "text",
  extensions: [],
  fallback: FALLBACK_EFFECTIVE_DATA,
};

export const RESOURCE_AUTHENTICATE_PRIMARY_LOGIN_LINK_HTML: ResourceDefinition =
  {
    resourcePath: resourcePath`templates/${"locale"}/messages/authenticate_primary_login_link.html`,
    type: "text",
    extensions: [],
    fallback: FALLBACK_EFFECTIVE_DATA,
  };
export const RESOURCE_AUTHENTICATE_PRIMARY_LOGIN_LINK_TXT: ResourceDefinition =
  {
    resourcePath: resourcePath`templates/${"locale"}/messages/authenticate_primary_login_link.txt`,
    type: "text",
    extensions: [],
    fallback: FALLBACK_EFFECTIVE_DATA,
  };

export const RESOURCE_FORGOT_PASSWORD_EMAIL_LINK_HTML: ResourceDefinition = {
  resourcePath: resourcePath`templates/${"locale"}/messages/forgot_password_email.html`,
  type: "text",
  extensions: [],
  fallback: FALLBACK_EFFECTIVE_DATA,
};
export const RESOURCE_FORGOT_PASSWORD_EMAIL_LINK_TXT: ResourceDefinition = {
  resourcePath: resourcePath`templates/${"locale"}/messages/forgot_password_email.txt`,
  type: "text",
  extensions: [],
  fallback: FALLBACK_EFFECTIVE_DATA,
};
export const RESOURCE_FORGOT_PASSWORD_SMS_LINK_TXT: ResourceDefinition = {
  resourcePath: resourcePath`templates/${"locale"}/messages/forgot_password_sms.txt`,
  type: "text",
  extensions: [],
  fallback: FALLBACK_EFFECTIVE_DATA,
};

export const RESOURCE_FORGOT_PASSWORD_EMAIL_CODE_HTML: ResourceDefinition = {
  resourcePath: resourcePath`templates/${"locale"}/messages/forgot_password_oob_email.html`,
  type: "text",
  extensions: [],
  fallback: FALLBACK_EFFECTIVE_DATA,
};
export const RESOURCE_FORGOT_PASSWORD_EMAIL_CODE_TXT: ResourceDefinition = {
  resourcePath: resourcePath`templates/${"locale"}/messages/forgot_password_oob_email.txt`,
  type: "text",
  extensions: [],
  fallback: FALLBACK_EFFECTIVE_DATA,
};
export const RESOURCE_FORGOT_PASSWORD_SMS_CODE_TXT: ResourceDefinition = {
  resourcePath: resourcePath`templates/${"locale"}/messages/forgot_password_oob_sms.txt`,
  type: "text",
  extensions: [],
  fallback: FALLBACK_EFFECTIVE_DATA,
};

export const RESOURCE_VERIFICATION_EMAIL_HTML: ResourceDefinition = {
  resourcePath: resourcePath`templates/${"locale"}/messages/verification_email.html`,
  type: "text",
  extensions: [],
  fallback: FALLBACK_EFFECTIVE_DATA,
};
export const RESOURCE_VERIFICATION_EMAIL_TXT: ResourceDefinition = {
  resourcePath: resourcePath`templates/${"locale"}/messages/verification_email.txt`,
  type: "text",
  extensions: [],
  fallback: FALLBACK_EFFECTIVE_DATA,
};
export const RESOURCE_VERIFICATION_SMS_TXT: ResourceDefinition = {
  resourcePath: resourcePath`templates/${"locale"}/messages/verification_sms.txt`,
  type: "text",
  extensions: [],
  fallback: FALLBACK_EFFECTIVE_DATA,
};

export const RESOURCE_APP_LOGO: ResourceDefinition = {
  resourcePath: resourcePath`static/${"locale"}/app_logo${"extension"}`,
  type: "binary",
  extensions: IMAGE_EXTENSIONS,
  optional: true,
};

export const RESOURCE_APP_LOGO_DARK: ResourceDefinition = {
  resourcePath: resourcePath`static/${"locale"}/app_logo_dark${"extension"}`,
  type: "binary",
  extensions: IMAGE_EXTENSIONS,
  optional: true,
};

export const RESOURCE_FAVICON: ResourceDefinition = {
  resourcePath: resourcePath`static/${"locale"}/favicon${"extension"}`,
  type: "binary",
  extensions: IMAGE_EXTENSIONS,
  optional: true,
};

export const RESOURCE_APP_BACKGROUND_IMAGE: ResourceDefinition = {
  resourcePath: resourcePath`static/app_background_image${"extension"}`,
  type: "binary",
  extensions: IMAGE_EXTENSIONS,
  optional: true,
};

export const RESOURCE_APP_BACKGROUND_IMAGE_DARK: ResourceDefinition = {
  resourcePath: resourcePath`static/app_background_image_dark${"extension"}`,
  type: "binary",
  extensions: IMAGE_EXTENSIONS,
  optional: true,
};

export const ALL_LANGUAGES_TEMPLATES = [
  RESOURCE_TRANSLATION_JSON,

  RESOURCE_SETUP_PRIMARY_OOB_EMAIL_HTML,
  RESOURCE_SETUP_PRIMARY_OOB_EMAIL_TXT,
  RESOURCE_SETUP_PRIMARY_OOB_SMS_TXT,

  RESOURCE_AUTHENTICATE_PRIMARY_OOB_EMAIL_HTML,
  RESOURCE_AUTHENTICATE_PRIMARY_OOB_EMAIL_TXT,
  RESOURCE_AUTHENTICATE_PRIMARY_OOB_SMS_TXT,

  RESOURCE_SETUP_PRIMARY_LOGIN_LINK_HTML,
  RESOURCE_SETUP_PRIMARY_LOGIN_LINK_TXT,

  RESOURCE_AUTHENTICATE_PRIMARY_LOGIN_LINK_HTML,
  RESOURCE_AUTHENTICATE_PRIMARY_LOGIN_LINK_TXT,

  RESOURCE_FORGOT_PASSWORD_EMAIL_LINK_HTML,
  RESOURCE_FORGOT_PASSWORD_EMAIL_LINK_TXT,
  RESOURCE_FORGOT_PASSWORD_SMS_LINK_TXT,

  RESOURCE_FORGOT_PASSWORD_EMAIL_CODE_HTML,
  RESOURCE_FORGOT_PASSWORD_EMAIL_CODE_TXT,
  RESOURCE_FORGOT_PASSWORD_SMS_CODE_TXT,

  RESOURCE_VERIFICATION_EMAIL_HTML,
  RESOURCE_VERIFICATION_EMAIL_TXT,
  RESOURCE_VERIFICATION_SMS_TXT,
];

export const RESOURCE_AUTHGEAR_LIGHT_THEME_CSS: ResourceDefinition = {
  resourcePath: resourcePath`static/authgear-light-theme.css`,
  type: "text",
  extensions: [],
  optional: true,
};

export const RESOURCE_AUTHGEAR_DARK_THEME_CSS: ResourceDefinition = {
  resourcePath: resourcePath`static/authgear-dark-theme.css`,
  type: "text",
  extensions: [],
  optional: true,
};

export const RESOURCE_AUTHGEAR_AUTHFLOW_V2_LIGHT_THEME_CSS: ResourceDefinition =
  {
    resourcePath: resourcePath`static/authgear-authflowv2-light-theme.css`,
    type: "text",
    extensions: [],
    optional: true,
  };

export const RESOURCE_AUTHGEAR_AUTHFLOW_V2_DARK_THEME_CSS: ResourceDefinition =
  {
    resourcePath: resourcePath`static/authgear-authflowv2-dark-theme.css`,
    type: "text",
    extensions: [],
    optional: true,
  };

export const RESOURCE_EMAIL_DOMAIN_BLOCKLIST: ResourceDefinition = {
  resourcePath: resourcePath`email_domain_blocklist.txt`,
  type: "text",
  extensions: [],
  optional: true,
};

export const RESOURCE_EMAIL_DOMAIN_ALLOWLIST: ResourceDefinition = {
  resourcePath: resourcePath`email_domain_allowlist.txt`,
  type: "text",
  extensions: [],
  optional: true,
};

export const RESOURCE_USERNAME_EXCLUDED_KEYWORDS_TXT: ResourceDefinition = {
  resourcePath: resourcePath`username_excluded_keywords.txt`,
  type: "text",
  extensions: [],
  optional: true,
};

export const RESOURCE_AUTHGEAR_YAML: ResourceDefinition = {
  resourcePath: resourcePath`authgear.yaml`,
  type: "text",
  extensions: [],
  optional: false,
};

export const TRANSLATION_JSON_KEY_EMAIL_FORGOT_PASSWORD_LINK_SUBJECT =
  "email.forgot-password.subject";
export const TRANSLATION_JSON_KEY_EMAIL_FORGOT_PASSWORD_CODE_SUBJECT =
  "email.forgot-password-oob.subject";
export const TRANSLATION_JSON_KEY_EMAIL_VERIFICATION_SUBJECT =
  "email.verification.subject";
export const TRANSLATION_JSON_KEY_EMAIL_SETUP_PRIMARY_OOB_SUBJECT =
  "email.setup-primary-oob.subject";
export const TRANSLATION_JSON_KEY_EMAIL_SETUP_PRIMARY_LOGIN_LINK_SUBJECT =
  "email.setup-primary-login-link.subject";
export const TRANSLATION_JSON_KEY_EMAIL_AUTHENTICATE_PRIMARY_LOGIN_LINK_SUBJECT =
  "email.authenticate-primary-login-link.subject";
export const TRANSLATION_JSON_KEY_EMAIL_AUTHENTICATE_PRIMARY_OOB_SUBJECT =
  "email.authenticate-primary-oob.subject";
