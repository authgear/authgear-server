export interface PortalAPIApp {
  id: string;
  rawAppConfig?: Record<string, unknown>;
  effectiveAppConfig?: Record<string, unknown>;
  secretConfig?: Record<string, unknown>;
}

export type LoginIDKeyType = "raw" | "email" | "phone" | "username";
