/* global describe, it, expect */
import { formatDeviceInfo } from "./deviceinfo";

describe("format device info", () => {
  it("android device info", () => {
    expect(
      formatDeviceInfo({
        android: {
          Build: {
            BOARD: "blueline",
            BRAND: "google",
            MODEL: "Pixel 3",
            DEVICE: "blueline",
            DISPLAY: "RQ1A.201205.003",
            PRODUCT: "blueline",
            HARDWARE: "blueline",
            MANUFACTURER: "Google",
          },
        },
      })
    ).toEqual("Google Pixel 3");
  });

  it("ios device info", () => {
    expect(
      formatDeviceInfo({
        ios: {
          uname: {
            machine: "iPhone13,1",
            release: "20.3.0",
            sysname: "Darwin",
            version:
              "Darwin Kernel Version 20.3.0: Tue Jan  5 18:34:42 PST 2021; root:xnu-7195.80.35~2/RELEASE_ARM64_T8101",
            nodename: "rfc1123",
          },
          UIDevice: {
            name: "rfc1123",
            model: "iPhone",
            systemName: "iOS",
            systemVersion: "14.4",
            userInterfaceIdiom: "phone",
          },
          NSProcessInfo: { isiOSAppOnMac: false, isMacCatalystApp: false },
        },
      })
    ).toEqual("iPhone 12 mini");
  });
});
