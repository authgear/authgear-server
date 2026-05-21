import { Viewer } from "../graphql/portal/globalTypes.generated";

export function isProjectQuotaReached(viewer: Viewer | null): boolean {
  if (viewer == null) {
    return false;
  }
  const { projectQuota, projectOwnerCount } = viewer;
  if (projectQuota == null) {
    return false;
  }
  return projectOwnerCount >= projectQuota;
}
