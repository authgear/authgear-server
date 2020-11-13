import { UpdateAppTemplatesData } from "./graphql/portal/mutations/updateAppTemplatesMutation";
import {
  parseTemplateString,
  renderTemplateString,
} from "./util/stringTemplate";

export type TemplateLocale = string;
export const DEFAULT_TEMPLATE_LOCALE: TemplateLocale = "en";
export type TemplateMap = Record<string, string>;

export const TEMPLATE_MESSAGE_TEMPLATE =
  "templates/{{locale}}/messages/{{fileName}}";

export const TEMPLATE_SETUP_PRIMARY_OOB_EMAIL_HTML =
  "templates/{{locale}}/messages/setup_primary_oob_email.html";
export const TEMPLATE_SETUP_PRIMARY_OOB_EMAIL_TEXT =
  "templates/{{locale}}/messages/setup_primary_oob_email.txt";
export const TEMPLATE_SETUP_PRIMARY_OOB_SMS_TEXT =
  "templates/{{locale}}/messages/setup_primary_oob_sms.txt";
export const SetupPrimaryOOBMessageTemplatePaths = [
  TEMPLATE_SETUP_PRIMARY_OOB_EMAIL_HTML,
  TEMPLATE_SETUP_PRIMARY_OOB_EMAIL_TEXT,
  TEMPLATE_SETUP_PRIMARY_OOB_SMS_TEXT,
] as const;
export type SetupPrimaryOOBMessageTemplateKeys = typeof SetupPrimaryOOBMessageTemplatePaths[number];

export const TEMPLATE_AUTHENTICATE_PRIMARY_OOB_EMAIL_HTML =
  "templates/{{locale}}/messages/authenticate_primary_oob_email.html";
export const TEMPLATE_AUTHENTICATE_PRIMARY_OOB_EMAIL_TEXT =
  "templates/{{locale}}/messages/authenticate_primary_oob_email.txt";
export const TEMPLATE_AUTHENTICATE_PRIMARY_OOB_SMS_TEXT =
  "templates/{{locale}}/messages/authenticate_primary_oob_sms.txt";
export const AuthenticatePrimaryOOBMessageTemplatePaths = [
  TEMPLATE_AUTHENTICATE_PRIMARY_OOB_EMAIL_HTML,
  TEMPLATE_AUTHENTICATE_PRIMARY_OOB_EMAIL_TEXT,
  TEMPLATE_AUTHENTICATE_PRIMARY_OOB_SMS_TEXT,
] as const;
export type AuthenticatePrimaryOOBMessageTemplateKeys = typeof AuthenticatePrimaryOOBMessageTemplatePaths[number];

export const TEMPLATE_FORGOT_PASSWORD_EMAIL_HTML =
  "templates/{{locale}}/messages/forgot_password_email.html";
export const TEMPLATE_FORGOT_PASSWORD_EMAIL_TEXT =
  "templates/{{locale}}/messages/forgot_password_email.txt";
export const TEMPLATE_FORGOT_PASSWORD_SMS_TEXT =
  "templates/{{locale}}/messages/forgot_password_sms.txt";
export const ForgotPasswordMessageTemplatePaths = [
  TEMPLATE_FORGOT_PASSWORD_EMAIL_HTML,
  TEMPLATE_FORGOT_PASSWORD_EMAIL_TEXT,
  TEMPLATE_FORGOT_PASSWORD_SMS_TEXT,
] as const;
export type ForgotPasswordMessageTemplateKeys = typeof ForgotPasswordMessageTemplatePaths[number];

export type PathTemplate =
  | SetupPrimaryOOBMessageTemplateKeys
  | AuthenticatePrimaryOOBMessageTemplateKeys
  | ForgotPasswordMessageTemplateKeys;

export function getLocalizedTemplatePath(
  locale: TemplateLocale,
  pathTemplate: PathTemplate
): string {
  return renderTemplateString({ locale }, pathTemplate);
}

export function setUpdateTemplatesData(
  templateUpdates: UpdateAppTemplatesData,
  pathTemplate: PathTemplate,
  templateLocale: TemplateLocale,
  templateValue: string
): void {
  templateUpdates[getLocalizedTemplatePath(templateLocale, pathTemplate)] =
    templateValue !== "" ? templateValue : null;
}

export function getConfiguredLocales(
  configuredPaths: string[]
): TemplateLocale[] {
  const templateLocaleSet = new Set<TemplateLocale>();
  for (const path of configuredPaths) {
    const { locale } = parseTemplateString(path, TEMPLATE_MESSAGE_TEMPLATE);
    if (locale != null) {
      templateLocaleSet.add(locale);
    }
  }
  return Array.from(templateLocaleSet);
}

export const STATIC_AUTHGEAR_CSS = "static/authgear.css";
