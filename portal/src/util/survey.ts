import { AppListItem, Viewer } from "../graphql/portal/globalTypes.generated";
import { SystemConfig } from "../system-config";

export function shouldShowSurvey(
  systemConfig: SystemConfig,
  apps: AppListItem[] | null,
  viewer: Viewer
): boolean {
  if (
    (apps === null || apps.length === 0) &&
    !viewer.isOnboardingSurveyCompleted &&
    !systemConfig.isAuthgearOnce
  ) {
    return true;
  }
  return false;
}
