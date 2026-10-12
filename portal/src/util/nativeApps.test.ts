import { describe, it, expect } from "@jest/globals";
import {
  nativeAppDisplayID,
  nativeAppFromDraft,
  nativeAppPlatformsForFramework,
  nativeAppToDraft,
  validateNativeApp,
} from "./nativeApps";
import { OAuthClientNativeAppConfig } from "../types";

const FINGERPRINT =
  "8B:BF:39:60:61:89:30:A4:45:F3:D7:09:1E:7B:1B:05:0F:8A:FD:AF:24:EB:F1:EB:2E:3D:13:88:09:FC:79:59";

const emptyDraft = {
  teamID: "",
  bundleID: "",
  packageName: "",
  fingerprints: "",
};

describe("nativeAppFromDraft", () => {
  it("builds an iOS entry with only iOS fields", () => {
    expect(
      nativeAppFromDraft("ios", {
        teamID: " abcde12345 ",
        bundleID: " com.example.app ",
        packageName: "ignored",
        fingerprints: "ignored",
      })
    ).toEqual({
      platform: "ios",
      team_id: "ABCDE12345",
      bundle_id: "com.example.app",
    });
  });

  it("builds an Android entry with only Android fields", () => {
    expect(
      nativeAppFromDraft("android", {
        teamID: "ignored",
        bundleID: "ignored",
        packageName: " com.example.app ",
        fingerprints: `\n ${FINGERPRINT.toLowerCase()} \r\n\n${FINGERPRINT}\n`,
      })
    ).toEqual({
      platform: "android",
      package_name: "com.example.app",
      sha256_cert_fingerprints: [FINGERPRINT, FINGERPRINT],
    });
  });

  it("round-trips through nativeAppToDraft", () => {
    const ios: OAuthClientNativeAppConfig = {
      platform: "ios",
      team_id: "ABCDE12345",
      bundle_id: "com.example.app",
    };
    const android: OAuthClientNativeAppConfig = {
      platform: "android",
      package_name: "com.example.app",
      sha256_cert_fingerprints: [FINGERPRINT, FINGERPRINT],
    };
    expect(nativeAppFromDraft("ios", nativeAppToDraft(ios))).toEqual(ios);
    expect(nativeAppFromDraft("android", nativeAppToDraft(android))).toEqual(
      android
    );
    expect(nativeAppToDraft(null)).toEqual(emptyDraft);
  });
});

describe("validateNativeApp", () => {
  it("requires every field of the platform", () => {
    expect(
      validateNativeApp(nativeAppFromDraft("ios", emptyDraft), [])
    ).toEqual([
      { field: "team_id", messageID: "NativeAppDialog.error.required" },
      { field: "bundle_id", messageID: "NativeAppDialog.error.required" },
    ]);
    expect(
      validateNativeApp(nativeAppFromDraft("android", emptyDraft), [])
    ).toEqual([
      { field: "package_name", messageID: "NativeAppDialog.error.required" },
      {
        field: "sha256_cert_fingerprints",
        messageID: "NativeAppDialog.error.required",
      },
    ]);
  });

  it("rejects a malformed team ID", () => {
    for (const teamID of ["ABCDE1234", "ABCDE-2345"]) {
      expect(
        validateNativeApp(
          { platform: "ios", team_id: teamID, bundle_id: "com.example.app" },
          []
        )
      ).toEqual([
        { field: "team_id", messageID: "NativeAppDialog.error.team-id-format" },
      ]);
    }
  });

  it("rejects a malformed fingerprint", () => {
    const short = FINGERPRINT.slice(0, -3);
    const noColon = FINGERPRINT.replace(":", "");
    expect(
      validateNativeApp(
        {
          platform: "android",
          package_name: "com.example.app",
          sha256_cert_fingerprints: [FINGERPRINT, short, noColon],
        },
        []
      )
    ).toEqual([
      {
        field: "sha256_cert_fingerprints",
        messageID: "NativeAppDialog.error.fingerprint-format",
        fingerprint: short,
      },
      {
        field: "sha256_cert_fingerprints",
        messageID: "NativeAppDialog.error.fingerprint-format",
        fingerprint: noColon,
      },
    ]);
  });

  it("rejects an app listed elsewhere on the same platform only", () => {
    const others: OAuthClientNativeAppConfig[] = [
      { platform: "ios", team_id: "ABCDE12345", bundle_id: "com.example.ios" },
      {
        platform: "android",
        package_name: "com.example.android",
        sha256_cert_fingerprints: [FINGERPRINT],
      },
    ];
    expect(
      validateNativeApp(
        {
          platform: "ios",
          team_id: "ZZZZZ99999",
          bundle_id: "com.example.ios",
        },
        others
      )
    ).toEqual([
      { field: "bundle_id", messageID: "NativeAppDialog.error.duplicated" },
    ]);
    expect(
      validateNativeApp(
        {
          platform: "android",
          package_name: "com.example.android",
          sha256_cert_fingerprints: [FINGERPRINT],
        },
        others
      )
    ).toEqual([
      { field: "package_name", messageID: "NativeAppDialog.error.duplicated" },
    ]);
    expect(
      validateNativeApp(
        {
          platform: "android",
          package_name: "com.example.ios",
          sha256_cert_fingerprints: [FINGERPRINT],
        },
        others
      )
    ).toEqual([]);
  });
});

describe("nativeAppDisplayID", () => {
  it("formats each platform", () => {
    expect(
      nativeAppDisplayID({
        platform: "ios",
        team_id: "ABCDE12345",
        bundle_id: "com.example.app",
      })
    ).toEqual("ABCDE12345.com.example.app");
    expect(
      nativeAppDisplayID({
        platform: "android",
        package_name: "com.example.app",
        sha256_cert_fingerprints: [FINGERPRINT],
      })
    ).toEqual("com.example.app");
  });
});

describe("nativeAppPlatformsForFramework", () => {
  it("limits single-platform frameworks to their platform", () => {
    expect(nativeAppPlatformsForFramework("ios")).toEqual(["ios"]);
    expect(nativeAppPlatformsForFramework("android")).toEqual(["android"]);
  });

  it("allows both for cross-platform frameworks and unset", () => {
    expect(nativeAppPlatformsForFramework("flutter")).toEqual([
      "ios",
      "android",
    ]);
    expect(nativeAppPlatformsForFramework(undefined)).toEqual([
      "ios",
      "android",
    ]);
  });
});
