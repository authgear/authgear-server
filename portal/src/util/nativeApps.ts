import {
  Framework,
  OAuthClientNativeAppConfig,
  OAuthClientNativeAppPlatform,
} from "../types";

// Mirrors OAuthClientNativeAppConfig in pkg/lib/config/oauth.go.
export const TEAM_ID_PATTERN = /^[A-Z0-9]{10}$/;
export const SHA256_FINGERPRINT_PATTERN = /^([0-9A-F]{2}:){31}[0-9A-F]{2}$/;

export interface NativeAppDraft {
  teamID: string;
  bundleID: string;
  packageName: string;
  fingerprints: string;
}

export type NativeAppFieldError =
  | { field: "team_id"; messageID: "NativeAppDialog.error.team-id-format" }
  | {
      field:
        | "team_id"
        | "bundle_id"
        | "package_name"
        | "sha256_cert_fingerprints";
      messageID: "NativeAppDialog.error.required";
    }
  | {
      field: "sha256_cert_fingerprints";
      messageID: "NativeAppDialog.error.fingerprint-format";
      fingerprint: string;
    }
  | {
      field: "bundle_id" | "package_name";
      messageID: "NativeAppDialog.error.duplicated";
    };

export function nativeAppToDraft(
  app: OAuthClientNativeAppConfig | null
): NativeAppDraft {
  if (app == null) {
    return { teamID: "", bundleID: "", packageName: "", fingerprints: "" };
  }
  switch (app.platform) {
    case "ios":
      return {
        teamID: app.team_id,
        bundleID: app.bundle_id,
        packageName: "",
        fingerprints: "",
      };
    case "android":
      return {
        teamID: "",
        bundleID: "",
        packageName: app.package_name,
        fingerprints: app.sha256_cert_fingerprints.join("\n"),
      };
  }
}

export function nativeAppFromDraft(
  platform: OAuthClientNativeAppPlatform,
  draft: NativeAppDraft
): OAuthClientNativeAppConfig {
  switch (platform) {
    case "ios":
      return {
        platform: "ios",
        team_id: draft.teamID.trim().toUpperCase(),
        bundle_id: draft.bundleID.trim(),
      };
    case "android":
      return {
        platform: "android",
        package_name: draft.packageName.trim(),
        sha256_cert_fingerprints: draft.fingerprints
          .split(/\r?\n/)
          .map((line) => line.trim().toUpperCase())
          .filter((line) => line !== ""),
      };
  }
}

export function validateNativeApp(
  app: OAuthClientNativeAppConfig,
  otherApps: OAuthClientNativeAppConfig[]
): NativeAppFieldError[] {
  const errors: NativeAppFieldError[] = [];
  switch (app.platform) {
    case "ios":
      if (app.team_id === "") {
        errors.push({
          field: "team_id",
          messageID: "NativeAppDialog.error.required",
        });
      } else if (!TEAM_ID_PATTERN.test(app.team_id)) {
        errors.push({
          field: "team_id",
          messageID: "NativeAppDialog.error.team-id-format",
        });
      }
      if (app.bundle_id === "") {
        errors.push({
          field: "bundle_id",
          messageID: "NativeAppDialog.error.required",
        });
      } else if (
        otherApps.some(
          (other) =>
            other.platform === "ios" && other.bundle_id === app.bundle_id
        )
      ) {
        errors.push({
          field: "bundle_id",
          messageID: "NativeAppDialog.error.duplicated",
        });
      }
      break;
    case "android":
      if (app.package_name === "") {
        errors.push({
          field: "package_name",
          messageID: "NativeAppDialog.error.required",
        });
      } else if (
        otherApps.some(
          (other) =>
            other.platform === "android" &&
            other.package_name === app.package_name
        )
      ) {
        errors.push({
          field: "package_name",
          messageID: "NativeAppDialog.error.duplicated",
        });
      }
      if (app.sha256_cert_fingerprints.length === 0) {
        errors.push({
          field: "sha256_cert_fingerprints",
          messageID: "NativeAppDialog.error.required",
        });
      }
      for (const fingerprint of app.sha256_cert_fingerprints) {
        if (!SHA256_FINGERPRINT_PATTERN.test(fingerprint)) {
          errors.push({
            field: "sha256_cert_fingerprints",
            messageID: "NativeAppDialog.error.fingerprint-format",
            fingerprint,
          });
        }
      }
      break;
  }
  return errors;
}

export function nativeAppDisplayID(app: OAuthClientNativeAppConfig): string {
  switch (app.platform) {
    case "ios":
      return `${app.team_id}.${app.bundle_id}`;
    case "android":
      return app.package_name;
  }
}

// The platforms an admin may add for a client: a single-platform framework
// allows only its own; cross-platform frameworks and unset allow both.
export function nativeAppPlatformsForFramework(
  framework: Framework | undefined
): OAuthClientNativeAppPlatform[] {
  switch (framework) {
    case "ios":
      return ["ios"];
    case "android":
      return ["android"];
    default:
      return ["ios", "android"];
  }
}
