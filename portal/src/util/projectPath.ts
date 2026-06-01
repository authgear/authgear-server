// Single source of truth for the navigable "section" pages under a project,
// i.e. the landing pages that appear in ScreenNav. These are the only pages
// that make sense to land on directly when there is no project-specific
// context (e.g. when switching to another project).
//
// Keep this list in sync by ALWAYS referencing ProjectSectionPath when adding
// a navigable page. ScreenNav builds its links from these constants, and the
// project switcher resolves its redirect target from them, so a single edit
// here keeps both behaviours correct.
export const ProjectSectionPath = {
  gettingStarted: "getting-started",
  analytics: "analytics",
  users: "user-management/users",
  roles: "user-management/roles",
  groups: "user-management/groups",
  loginMethods: "configuration/authentication/login-methods",
  externalOAuth: "configuration/authentication/external-oauth",
  biometric: "configuration/authentication/biometric",
  mfa: "configuration/authentication/2fa",
  anonymousUsers: "configuration/authentication/anonymous-users",
  app2app: "configuration/authentication/app2app",
  clientApplications: "configuration/apps",
  apiResources: "api-resources",
  design: "branding/design",
  localization: "branding/localization",
  customDomains: "branding/custom-domains",
  customText: "branding/custom-text",
  languages: "configuration/languages",
  standardAttributes: "configuration/user-profile/standard-attributes",
  customAttributes: "configuration/user-profile/custom-attributes",
  botProtection: "attack-protection/bot-protection",
  fraudProtection: "attack-protection/fraud-protection",
  ipBlocklist: "attack-protection/ip-blocklist",
  integrations: "integrations",
  license: "license",
  billing: "billing",
  hooks: "advanced/hooks",
  adminAPI: "advanced/admin-api",
  accountDeletion: "advanced/account-deletion",
  accountAnonymization: "advanced/account-anonymization",
  session: "advanced/session",
  smtp: "advanced/smtp",
  smsGateway: "advanced/sms-gateway",
  endpointDirectAccess: "advanced/endpoint-direct-access",
  samlCertificate: "advanced/saml-certificate",
  editConfig: "edit-config",
  auditLog: "audit-log",
  portalAdmins: "portal-admins",
} as const;

export const PROJECT_SECTION_PATHS: readonly string[] =
  Object.values(ProjectSectionPath);

// The page to land on when no section can be resolved from the current path.
const DEFAULT_SECTION_PATH = ProjectSectionPath.gettingStarted;

const PROJECT_PATH_PATTERN = /^\/project\/[^/]+/;

// projectPath builds an absolute path to a section under a project.
export function projectPath(appID: string, section: string): string {
  return `/project/${appID}/${section}`;
}

// Returns true if `prefix` matches `path` on a path-segment boundary, so that
// "api-resources" matches "api-resources" and "api-resources/x" but never
// "api-resources-foo".
function isSegmentPrefix(prefix: string, path: string): boolean {
  return path === prefix || path.startsWith(prefix + "/");
}

// Given the path relative to /project/:id (e.g. "user-management/users/<id>/details"),
// returns the longest section path that is a segment-prefix of it, or null.
function findLongestSectionPrefix(relativePath: string): string | null {
  let best: string | null = null;
  for (const section of PROJECT_SECTION_PATHS) {
    if (isSegmentPrefix(section, relativePath)) {
      if (best == null || section.length > best.length) {
        best = section;
      }
    }
  }
  return best;
}

// resolveProjectSwitchPath computes where to navigate when switching to another
// project, given the current location's pathname and the new project's base
// path (e.g. "/project/<encoded-typed-id>").
//
// It keeps the user on the same section of the app, but truncates any
// project-specific suffix (entity IDs, create/edit forms) that would not exist
// in the target project. For example "/project/A/user-management/users/<id>/details"
// resolves to "/project/B/user-management/users" (the user list), because <id>
// belongs to project A. Paths that do not map to any known section fall back to
// the getting-started page.
export function resolveProjectSwitchPath(
  currentPathname: string,
  newProjectBasePath: string
): string {
  const match = PROJECT_PATH_PATTERN.exec(currentPathname);
  const relativePath =
    match != null
      ? currentPathname.slice(match[0].length).replace(/^\//, "")
      : "";
  const section =
    findLongestSectionPrefix(relativePath) ?? DEFAULT_SECTION_PATH;
  return `${newProjectBasePath}/${section}`;
}
