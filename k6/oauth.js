import { check, fail } from "k6";
import http from "k6/http";
import { URL } from "https://jslib.k6.io/url/1.0.0/index.js";

const response_type = "code";
const code_challenge_method = "S256";
// code_verifier must be at least 43 characters long.
// See https://github.com/authgear/authgear-server/pull/4126
const code_verifier = "the-quick-brown-fox-jumps-over-the-lazy-dog";
// code_challenge is S256 of code_verifier.
const code_challenge = "lSqEXx4ypW-y7Dj_NrquA6XliP_YgrqI9C1nkbNMIUs";
const scope = "openid offline_access https://authgear.com/scopes/full-access";
const x_suppress_idp_session_cookie = "true";
const x_sso_enabled = "false";

export function makeAuthenticationURL({ endpoint, client_id, redirect_uri }) {
  const url = new URL("/oauth2/authorize", endpoint);
  url.searchParams.set("response_type", response_type);
  url.searchParams.set("code_challenge_method", code_challenge_method);
  url.searchParams.set("code_challenge", code_challenge);
  url.searchParams.set("scope", scope);
  url.searchParams.set("client_id", client_id);
  url.searchParams.set("redirect_uri", redirect_uri);
  url.searchParams.set(
    "x_suppress_idp_session_cookie",
    x_suppress_idp_session_cookie,
  );
  url.searchParams.set("x_sso_enabled", x_sso_enabled);
  return url;
}

export function startAuthentication(url) {
  const response = http.get(url.toString(), { redirects: 0 });
  const checkResult = check(response, {
    "authentication request is of status 302": (r) => r.status === 302,
  });
  if (!checkResult) {
    fail("failed to start authentication request");
  }
  return response.headers["Location"];
}

export function extractAuthorizationCode(response) {
  const PREFIX = "0;url=";

  const doc = response.html();
  const content = doc.find("meta[http-equiv=refresh]").attr("content");
  const urlStr = content.slice(PREFIX.length);
  const url = new URL(urlStr);
  const code = url.searchParams.get("code");
  return code;
}

export function exchangeAuthorizationCode({
  endpoint,
  client_id,
  redirect_uri,
  code,
}) {
  const url = new URL("/oauth2/token", endpoint);
  const form = {
    grant_type: "authorization_code",
    client_id,
    code,
    redirect_uri,
    code_verifier,
  };
  const response = http.post(url.toString(), form);
  const checkResult = check(response, {
    "token request is of status 200": (r) => r.status === 200,
  });
  if (!checkResult) {
    fail("failed to exchange authorization code");
  }
  const json = response.json();
  return json;
}

export function refreshAccessToken({ endpoint, client_id, refresh_token }) {
  const url = new URL("/oauth2/token", endpoint);
  const form = {
    grant_type: "refresh_token",
    client_id,
    refresh_token,
  };
  const response = http.post(url.toString(), form);
  const checkResult = check(response, {
    "token request is of status 200": (r) => r.status === 200,
  });
  if (!checkResult) {
    fail("failed to refresh access token");
  }
  const json = response.json();
  return json;
}

export function getUserInfo({ endpoint, access_token }) {
  const url = new URL("/oauth2/userinfo", endpoint);
  const response = http.get(url.toString(), {
    headers: {
      Authorization: `Bearer ${access_token}`,
    },
  });
  const checkResult = check(response, {
    "user info request is of status 200": (r) => r.status === 200,
  });
  if (!checkResult) {
    fail("failed to get user info");
  }
  const json = response.json();
  return json;
}

export function getChallenge({ endpoint, purpose }) {
  const url = new URL("/oauth2/challenge", endpoint);
  const payload = JSON.stringify({
    purpose,
  });
  const params = {
    headers: {
      "Content-Type": "application/json",
    },
  };
  const response = http.post(url.toString(), payload, params);
  const checkResult = check(response, {
    "challenge request is of status 200": (r) => r.status === 200,
  });
  if (!checkResult) {
    fail("failed to get challenge");
  }

  return response.json().result.token;
}
