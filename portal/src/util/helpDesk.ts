const HELP_DESK_URL = "https://www.authgear.com/help-desk/";

export type HelpDeskPurpose =
  | "project-limit"
  | "sms-otp"
  | "resume-suspended"
  | "other";

export interface HelpDeskURLOptions {
  purpose: HelpDeskPurpose;
  projectID?: string | null;
  email?: string | null;
  fullName?: string | null;
}

// Unknown values are omitted so the help desk form leaves those fields blank.
export function makeHelpDeskURL(options: HelpDeskURLOptions): string {
  const { purpose, projectID, email, fullName } = options;
  const url = new URL(HELP_DESK_URL);
  if (projectID) {
    url.searchParams.set("project", projectID);
  }
  if (email) {
    url.searchParams.set("email", email);
  }
  if (fullName) {
    url.searchParams.set("name", fullName);
  }
  url.searchParams.set("purpose", purpose);
  return url.toString();
}
