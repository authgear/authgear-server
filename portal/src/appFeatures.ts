import { PortalAPIFeatureConfig } from "./types";
import { SystemConfig } from "./system-config";

export interface AppFeatureContext {
  featureConfig: PortalAPIFeatureConfig | null;
  systemConfig: SystemConfig;
}

interface AppFeatureDefinition {
  isAvailable: (ctx: AppFeatureContext) => boolean;
}

// APP_FEATURES defines when each feature is available. Which routes a feature
// gates is declared by `feature` on AppRoute in AppRoot.tsx.
export const APP_FEATURES = {
  fraudProtection: {
    isAvailable: ({ featureConfig }) =>
      featureConfig?.fraud_protection?.is_modifiable ?? false,
  },
} satisfies Record<string, AppFeatureDefinition>;

export type AppFeature = keyof typeof APP_FEATURES;
