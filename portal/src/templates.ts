import { UpdateAppTemplatesData } from "./graphql/portal/mutations/updateAppTemplatesMutation";
import { ResourcePath, resourcePath } from "./util/stringTemplate";

export type TemplateLocale = string;
export const DEFAULT_TEMPLATE_LOCALE: TemplateLocale = "en";
export type TemplateMap = Record<string, string>;

export const messageTemplatePath = resourcePath`templates/${"locale"}/messages/${"fileName"}`;

export const setupPrimaryOobEmailHtmlPath = resourcePath`templates/${"locale"}/messages/setup_primary_oob_email.html`;
export const setupPrimaryOobEmailTextPath = resourcePath`templates/${"locale"}/messages/setup_primary_oob_email.txt`;
export const setupPrimaryOobSmsTextPath = resourcePath`templates/${"locale"}/messages/setup_primary_oob_sms.txt`;

export const authenticatePrimaryOobEmailHtmlPath = resourcePath`templates/${"locale"}/messages/authenticate_primary_oob_email.html`;
export const authenticatePrimaryOobEmailTextPath = resourcePath`templates/${"locale"}/messages/authenticate_primary_oob_email.txt`;
export const authenticatePrimaryOobSmsTextPath = resourcePath`templates/${"locale"}/messages/authenticate_primary_oob_sms.txt`;

export const forgotPasswordEmailHtmlPath = resourcePath`templates/${"locale"}/messages/forgot_password_email.html`;
export const forgotPasswordEmailTextPath = resourcePath`templates/${"locale"}/messages/forgot_password_email.txt`;
export const forgotPasswordSmsTextPath = resourcePath`templates/${"locale"}/messages/forgot_password_sms.txt`;

export const ALL_TEMPLATE_PATHS = [
  setupPrimaryOobEmailHtmlPath,
  setupPrimaryOobEmailTextPath,
  setupPrimaryOobSmsTextPath,
  authenticatePrimaryOobEmailHtmlPath,
  authenticatePrimaryOobEmailTextPath,
  authenticatePrimaryOobSmsTextPath,
  forgotPasswordEmailHtmlPath,
  forgotPasswordEmailTextPath,
  forgotPasswordSmsTextPath,
];

export function getLocalizedTemplatePath(
  locale: TemplateLocale,
  pathTemplate: ResourcePath<"locale">
): string {
  return pathTemplate.render({ locale });
}

export function setUpdateTemplatesData(
  templateUpdates: UpdateAppTemplatesData,
  resourcePathTemplate: ResourcePath<"locale">,
  templateLocale: TemplateLocale,
  templateValue: string
): void {
  templateUpdates[
    getLocalizedTemplatePath(templateLocale, resourcePathTemplate)
  ] = templateValue !== "" ? templateValue : null;
}

export const STATIC_AUTHGEAR_CSS = "static/authgear.css";
