import { describe, it, expect } from "@jest/globals";
import { resolveProjectSwitchPath, projectPath } from "./projectPath";

const OLD = "/project/old-id";
const NEW_BASE = "/project/new-id";

describe("resolveProjectSwitchPath", () => {
  const cases: { name: string; current: string; expected: string }[] = [
    // Landing pages stay on the same section.
    {
      name: "getting-started stays",
      current: `${OLD}/getting-started`,
      expected: `${NEW_BASE}/getting-started`,
    },
    {
      name: "a leaf config page stays",
      current: `${OLD}/advanced/smtp`,
      expected: `${NEW_BASE}/advanced/smtp`,
    },
    {
      name: "audit-log list stays (was missed by the old logic)",
      current: `${OLD}/audit-log`,
      expected: `${NEW_BASE}/audit-log`,
    },
    // Entity-detail pages truncate to their section list, because the id does
    // not exist in the target project.
    {
      name: "user details -> user list",
      current: `${OLD}/user-management/users/USERID/details`,
      expected: `${NEW_BASE}/user-management/users`,
    },
    {
      name: "user detail sub-form -> user list",
      current: `${OLD}/user-management/users/USERID/details/edit-email/IDID`,
      expected: `${NEW_BASE}/user-management/users`,
    },
    {
      name: "role details -> role list",
      current: `${OLD}/user-management/roles/ROLEID/details`,
      expected: `${NEW_BASE}/user-management/roles`,
    },
    {
      name: "group details -> group list",
      current: `${OLD}/user-management/groups/GROUPID/details`,
      expected: `${NEW_BASE}/user-management/groups`,
    },
    {
      name: "oauth client edit -> apps list",
      current: `${OLD}/configuration/apps/CLIENTID/edit`,
      expected: `${NEW_BASE}/configuration/apps`,
    },
    {
      name: "api resource scope -> api-resources list",
      current: `${OLD}/api-resources/RID/scopes/SID`,
      expected: `${NEW_BASE}/api-resources`,
    },
    {
      name: "sso edit -> external-oauth list",
      current: `${OLD}/configuration/authentication/external-oauth/edit/google/alias`,
      expected: `${NEW_BASE}/configuration/authentication/external-oauth`,
    },
    {
      name: "custom domain verify -> custom-domains list",
      current: `${OLD}/branding/custom-domains/DOMAINID/verify`,
      expected: `${NEW_BASE}/branding/custom-domains`,
    },
    {
      name: "custom attribute edit -> custom-attributes list",
      current: `${OLD}/configuration/user-profile/custom-attributes/0/edit`,
      expected: `${NEW_BASE}/configuration/user-profile/custom-attributes`,
    },
    {
      name: "audit-log entry -> audit-log list",
      current: `${OLD}/audit-log/LOGID/details`,
      expected: `${NEW_BASE}/audit-log`,
    },
    // Create/add forms truncate to their section list.
    {
      name: "add-user -> user list",
      current: `${OLD}/user-management/users/add-user`,
      expected: `${NEW_BASE}/user-management/users`,
    },
    {
      name: "portal-admins invite -> portal-admins list",
      current: `${OLD}/portal-admins/invite`,
      expected: `${NEW_BASE}/portal-admins`,
    },
    // Fallbacks.
    {
      name: "unknown path -> getting-started",
      current: `${OLD}/some/unknown/page`,
      expected: `${NEW_BASE}/getting-started`,
    },
    {
      name: "non-project path -> getting-started",
      current: `/projects`,
      expected: `${NEW_BASE}/getting-started`,
    },
  ];

  for (const c of cases) {
    it(c.name, () => {
      expect(resolveProjectSwitchPath(c.current, NEW_BASE)).toEqual(c.expected);
    });
  }

  it("does not match on a non-segment boundary", () => {
    // "api-resources-foo" must not be treated as the "api-resources" section.
    expect(
      resolveProjectSwitchPath(`${OLD}/api-resources-foo/bar`, NEW_BASE)
    ).toEqual(`${NEW_BASE}/getting-started`);
  });
});

describe("projectPath", () => {
  it("joins appID and section", () => {
    expect(projectPath("abc", "audit-log")).toEqual("/project/abc/audit-log");
  });
});
