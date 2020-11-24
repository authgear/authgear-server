import {
  ResourcePath,
  resourcePath,
  ResourceDefinition,
  LanguageTag,
} from "./util/resource";

export const DEFAULT_TEMPLATE_LOCALE: LanguageTag = "en";

export const RESOURCE_TRANSLATION_JSON: ResourceDefinition = {
  resourcePath: resourcePath`templates/${"locale"}/translation.json`,
  type: "text",
};

export const RESOURCE_SETUP_PRIMARY_OOB_EMAIL_HTML: ResourceDefinition = {
  resourcePath: resourcePath`templates/${"locale"}/messages/setup_primary_oob_email.html`,
  type: "text",
};
export const RESOURCE_SETUP_PRIMARY_OOB_EMAIL_TXT: ResourceDefinition = {
  resourcePath: resourcePath`templates/${"locale"}/messages/setup_primary_oob_email.txt`,
  type: "text",
};
export const RESOURCE_SETUP_PRIMARY_OOB_SMS_TXT: ResourceDefinition = {
  resourcePath: resourcePath`templates/${"locale"}/messages/setup_primary_oob_sms.txt`,
  type: "text",
};

export const RESOURCE_AUTHENTICATE_PRIMARY_OOB_EMAIL_HTML: ResourceDefinition = {
  resourcePath: resourcePath`templates/${"locale"}/messages/authenticate_primary_oob_email.html`,
  type: "text",
};
export const RESOURCE_AUTHENTICATE_PRIMARY_OOB_EMAIL_TXT: ResourceDefinition = {
  resourcePath: resourcePath`templates/${"locale"}/messages/authenticate_primary_oob_email.txt`,
  type: "text",
};
export const RESOURCE_AUTHENTICATE_PRIMARY_OOB_SMS_TXT: ResourceDefinition = {
  resourcePath: resourcePath`templates/${"locale"}/messages/authenticate_primary_oob_sms.txt`,
  type: "text",
};

export const RESOURCE_FORGOT_PASSWORD_EMAIL_HTML: ResourceDefinition = {
  resourcePath: resourcePath`templates/${"locale"}/messages/forgot_password_email.html`,
  type: "text",
};
export const RESOURCE_FORGOT_PASSWORD_EMAIL_TXT: ResourceDefinition = {
  resourcePath: resourcePath`templates/${"locale"}/messages/forgot_password_email.txt`,
  type: "text",
};
export const RESOURCE_FORGOT_PASSWORD_SMS_TXT: ResourceDefinition = {
  resourcePath: resourcePath`templates/${"locale"}/messages/forgot_password_sms.txt`,
  type: "text",
};

export const ALL_RESOURCES = [
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

export interface RenderPathArguments {
  locale: LanguageTag;
  extension?: string;
}

export function renderPath(
  resourcePath: ResourcePath,
  args: RenderPathArguments
): string {
  const renderArgs: Record<string, string> = {};
  renderArgs["locale"] = args.locale;
  if (args.extension != null) {
    renderArgs["extension"] = args.extension;
  }
  return resourcePath.render(renderArgs);
}

export const STATIC_AUTHGEAR_CSS = "static/authgear.css";
