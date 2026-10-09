import { PortalAPIFeatureConfig } from "../types";

export const isGoogleTagManagerAvailable = (
  fc: PortalAPIFeatureConfig | null
): boolean => (fc?.google_tag_manager?.disabled ?? false) === false;
export const isAuditLogStreamingAvailable = (
  fc: PortalAPIFeatureConfig | null
): boolean =>
  (fc?.telemetry?.audit_logs?.streaming?.disabled ?? false) === false;
export const isIntegrationsAvailable = (
  fc: PortalAPIFeatureConfig | null
): boolean =>
  isGoogleTagManagerAvailable(fc) || isAuditLogStreamingAvailable(fc);
