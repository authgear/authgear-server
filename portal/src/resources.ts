import {
  ResourcePath,
  resourcePath,
  ResourceDefinition,
  LanguageTag,
} from "./util/resource";

export const DEFAULT_TEMPLATE_LOCALE: LanguageTag = "en";

export const IMAGE_EXTENSIONS: string[] = [".png", ".jpeg", ".gif"];

export const RESOURCE_TRANSLATION_JSON: ResourceDefinition = {
  resourcePath: resourcePath`templates/${"locale"}/translation.json`,
  type: "text",
  extensions: [],
  usesEffectiveDataAsFallbackValue: true,
};

export const RESOURCE_SETUP_PRIMARY_OOB_EMAIL_HTML: ResourceDefinition = {
  resourcePath: resourcePath`templates/${"locale"}/messages/setup_primary_oob_email.html`,
  type: "text",
  extensions: [],
  usesEffectiveDataAsFallbackValue: true,
};
export const RESOURCE_SETUP_PRIMARY_OOB_EMAIL_TXT: ResourceDefinition = {
  resourcePath: resourcePath`templates/${"locale"}/messages/setup_primary_oob_email.txt`,
  type: "text",
  extensions: [],
  usesEffectiveDataAsFallbackValue: true,
};
export const RESOURCE_SETUP_PRIMARY_OOB_SMS_TXT: ResourceDefinition = {
  resourcePath: resourcePath`templates/${"locale"}/messages/setup_primary_oob_sms.txt`,
  type: "text",
  extensions: [],
  usesEffectiveDataAsFallbackValue: true,
};

export const RESOURCE_AUTHENTICATE_PRIMARY_OOB_EMAIL_HTML: ResourceDefinition = {
  resourcePath: resourcePath`templates/${"locale"}/messages/authenticate_primary_oob_email.html`,
  type: "text",
  extensions: [],
  usesEffectiveDataAsFallbackValue: true,
};
export const RESOURCE_AUTHENTICATE_PRIMARY_OOB_EMAIL_TXT: ResourceDefinition = {
  resourcePath: resourcePath`templates/${"locale"}/messages/authenticate_primary_oob_email.txt`,
  type: "text",
  extensions: [],
  usesEffectiveDataAsFallbackValue: true,
};
export const RESOURCE_AUTHENTICATE_PRIMARY_OOB_SMS_TXT: ResourceDefinition = {
  resourcePath: resourcePath`templates/${"locale"}/messages/authenticate_primary_oob_sms.txt`,
  type: "text",
  extensions: [],
  usesEffectiveDataAsFallbackValue: true,
};

export const RESOURCE_FORGOT_PASSWORD_EMAIL_HTML: ResourceDefinition = {
  resourcePath: resourcePath`templates/${"locale"}/messages/forgot_password_email.html`,
  type: "text",
  extensions: [],
  usesEffectiveDataAsFallbackValue: true,
};
export const RESOURCE_FORGOT_PASSWORD_EMAIL_TXT: ResourceDefinition = {
  resourcePath: resourcePath`templates/${"locale"}/messages/forgot_password_email.txt`,
  type: "text",
  extensions: [],
  usesEffectiveDataAsFallbackValue: true,
};
export const RESOURCE_FORGOT_PASSWORD_SMS_TXT: ResourceDefinition = {
  resourcePath: resourcePath`templates/${"locale"}/messages/forgot_password_sms.txt`,
  type: "text",
  extensions: [],
  usesEffectiveDataAsFallbackValue: true,
};

export const RESOURCE_APP_LOGO: ResourceDefinition = {
  resourcePath: resourcePath`static/${"locale"}/app_logo${"extension"}`,
  type: "binary",
  extensions: IMAGE_EXTENSIONS,
  usesEffectiveDataAsFallbackValue: false,
  optional: true,
};

export const RESOURCE_FAVICON: ResourceDefinition = {
  resourcePath: resourcePath`static/${"locale"}/favicon${"extension"}`,
  type: "binary",
  extensions: IMAGE_EXTENSIONS,
  usesEffectiveDataAsFallbackValue: false,
  optional: true,
};

export const ALL_TEMPLATES = [
  RESOURCE_TRANSLATION_JSON,

  RESOURCE_SETUP_PRIMARY_OOB_EMAIL_HTML,
  RESOURCE_SETUP_PRIMARY_OOB_EMAIL_TXT,
  RESOURCE_SETUP_PRIMARY_OOB_SMS_TXT,

  RESOURCE_AUTHENTICATE_PRIMARY_OOB_EMAIL_HTML,
  RESOURCE_AUTHENTICATE_PRIMARY_OOB_EMAIL_TXT,
  RESOURCE_AUTHENTICATE_PRIMARY_OOB_SMS_TXT,

  RESOURCE_FORGOT_PASSWORD_EMAIL_HTML,
  RESOURCE_FORGOT_PASSWORD_EMAIL_TXT,
  RESOURCE_FORGOT_PASSWORD_SMS_TXT,
];

export const RESOURCE_AUTHGEAR_CSS: ResourceDefinition = {
  resourcePath: resourcePath`static/authgear.css`,
  type: "text",
  extensions: [],
  usesEffectiveDataAsFallbackValue: false,
  optional: true,
};

export const RESOURCE_AUTHGEAR_LIGHT_THEME_CSS: ResourceDefinition = {
  resourcePath: resourcePath`static/authgear-light-theme.css`,
  type: "text",
  extensions: [],
  usesEffectiveDataAsFallbackValue: false,
  optional: true,
};

export const RESOURCE_AUTHGEAR_DARK_THEME_CSS: ResourceDefinition = {
  resourcePath: resourcePath`static/authgear-dark-theme.css`,
  type: "text",
  extensions: [],
  usesEffectiveDataAsFallbackValue: false,
  optional: true,
};

export interface RenderPathArguments {
  locale?: LanguageTag;
  extension?: string;
}

export function renderPath(
  resourcePath: ResourcePath,
  args: RenderPathArguments
): string {
  const renderArgs: Record<string, string> = {};
  if (args.locale != null) {
    renderArgs["locale"] = args.locale;
  }
  if (args.extension != null) {
    renderArgs["extension"] = args.extension;
  }
  return resourcePath.render(renderArgs);
}
