import { PortalAPIFeatureConfig } from "./types";
import { SystemConfig } from "./system-config";

export interface AppFeatureContext {
  featureConfig: PortalAPIFeatureConfig | null;
  systemConfig: SystemConfig;
}

interface AppFeatureDefinition {
  isAvailable: (ctx: AppFeatureContext) => boolean;
}

const isGoogleTagManagerAvailable = ({ featureConfig }: AppFeatureContext) =>
  (featureConfig?.google_tag_manager?.disabled ?? false) === false;
const isAuditLogStreamingAvailable = ({ featureConfig }: AppFeatureContext) =>
  (featureConfig?.telemetry?.audit_logs?.streaming?.disabled ?? false) ===
  false;

// APP_FEATURES defines when each feature is available. Which routes a feature
// gates is declared by `feature` on AppRoute in AppRoot.tsx.
export const APP_FEATURES = {
  fraudProtection: {
    isAvailable: ({ featureConfig }) =>
      featureConfig?.fraud_protection?.is_modifiable ?? false,
  },
  auditLog: {
    isAvailable: ({ systemConfig }) => systemConfig.auditLogEnabled,
  },
  analytics: {
    isAvailable: ({ systemConfig }) => systemConfig.analyticEnabled,
  },
  app2app: {
    isAvailable: ({ featureConfig }) =>
      featureConfig?.oauth?.client?.app2app_enabled ?? false,
  },
  customUI: {
    isAvailable: ({ featureConfig }) =>
      featureConfig?.oauth?.client?.custom_ui_enabled ?? false,
  },
  googleTagManager: {
    isAvailable: isGoogleTagManagerAvailable,
  },
  auditLogStreaming: {
    isAvailable: isAuditLogStreamingAvailable,
  },
  // The Integrations page lists the integrations above that are available.
  integrations: {
    isAvailable: (ctx) =>
      isGoogleTagManagerAvailable(ctx) || isAuditLogStreamingAvailable(ctx),
  },
} satisfies Record<string, AppFeatureDefinition>;

export type AppFeature = keyof typeof APP_FEATURES;
