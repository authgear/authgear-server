import { check, fail } from "k6";
import http from "k6/http";
import { URL } from "https://jslib.k6.io/url/1.0.0/index.js";
import authgear from "k6/x/authgear";
import { getChallenge } from "./oauth.js";

const device_info = {
  ios: {
    uname: {
      machine: "iPhone13,1",
      release: "22.0.0",
      sysname: "Darwin",
      version:
        "Darwin Kernel Version 22.0.0: Tue Aug 16 20:52:01 PDT 2022; root:xnu-8792.2.11.0.1~1/RELEASE_ARM64_T8101",
      nodename: "iPhone",
    },
    NSBundle: {
      CFBundleName: "reactNativeExample",
      CFBundleVersion: "1",
      CFBundleExecutable: "reactNativeExample",
      CFBundleIdentifier: "com.authgear.exampleapp.reactnative",
      CFBundleDisplayName: "reactNativeExample",
      CFBundleShortVersionString: "1.0",
    },
    UIDevice: {
      name: "iPhone",
      model: "iPhone",
      systemName: "iOS",
      systemVersion: "16.0.2",
      userInterfaceIdiom: "phone",
    },
    NSProcessInfo: { isiOSAppOnMac: false, isMacCatalystApp: false },
  },
};

export function enableBiometric({ endpoint, client_id, access_token }) {
  const kid = authgear.uuid();
  const privateKeyPEM = authgear.generateRSAPrivateKeyInPKCS8PEM(2048);

  const jwk = authgear.jwkKeyFromPKCS8PEM(privateKeyPEM);
  jwk["kid"] = kid;

  const challenge = getChallenge({ endpoint, purpose: "biometric_request" });

  const now = Math.floor(new Date().getTime() / 1000);

  const payload = {
    iat: now,
    exp: now + 300,
    challenge,
    action: "setup",
    device_info,
  };

  const headers = {
    jwk,
    typ: "vnd.authgear.biometric-request",
  };

  const jwt = authgear.signJWT("RS256", jwk, payload, headers);

  const url = new URL("/oauth2/token", endpoint);
  const form = {
    grant_type: "urn:authgear:params:oauth:grant-type:biometric-request",
    jwt,
    client_id,
  };
  const params = {
    headers: {
      Authorization: `Bearer ${access_token}`,
    },
  };
  const response = http.post(url.toString(), form, params);
  const checkResult = check(response, {
    "biometric request is of status 200": (r) => r.status === 200,
  });
  if (!checkResult) {
    fail("failed to perform biometric request");
  }

  return jwk;
}

export function authenticateBiometric({ endpoint, client_id, jwk }) {
  const challenge = getChallenge({ endpoint, purpose: "biometric_request" });

  const now = Math.floor(new Date().getTime() / 1000);

  const payload = {
    iat: now,
    exp: now + 300,
    challenge,
    action: "authenticate",
    device_info,
  };

  const headers = {
    kid: jwk["kid"],
    typ: "vnd.authgear.biometric-request",
  };

  const jwt = authgear.signJWT("RS256", jwk, payload, headers);

  const url = new URL("/oauth2/token", endpoint);
  const form = {
    grant_type: "urn:authgear:params:oauth:grant-type:biometric-request",
    jwt,
    client_id,
  };
  const response = http.post(url.toString(), form);
  const checkResult = check(response, {
    "biometric request is of status 200": (r) => r.status === 200,
  });
  if (!checkResult) {
    fail("failed to perform biometric request");
  }

  const tokenResponse = response.json();
  return tokenResponse;
}
