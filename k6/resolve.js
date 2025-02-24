import { check, fail } from "k6";
import http from "k6/http";
import exec from "k6/execution";
import {
  makeNationalPhoneNumberForLogin,
  makeLoginIDs,
  getIndex,
} from "./fixture.js";
import { authflowRun } from "./authflow.js";
import { ENDPOINT, RESOLVER_ENDPOINT } from "./env.js";

export function setup() {
  const data = [];
  const vus = exec.test.options.scenarios.default.vus;
  for (let i = 1; i <= vus; ++i) {
    const nationalPhone = makeNationalPhoneNumberForLogin({ vu: i });
    const { username, email, phone } = makeLoginIDs(nationalPhone);
    const tokenResponse = authflowRun({
      username,
      phone,
      email,
      type: "signup",
      name: "default",
    });
    data.push(tokenResponse);
  }
  return data;
}

export default function (data) {
  const index = getIndex();
  const tokenResponse = data[index];
  const origin = new URL(ENDPOINT);

  const url = new URL("/_resolver/resolve", RESOLVER_ENDPOINT);
  const response = http.get(url.toString(), {
    headers: {
      Host: origin.host,
      Authorization: `Bearer ${tokenResponse.access_token}`,
    },
  });
  const checkResult = check(response, {
    "resolve request is of status 200": (r) => r.status === 200,
    "X-Authgear-Session-Valid is true": (r) =>
      r.headers["X-Authgear-Session-Valid"] === "true",
    "X-Authgear-User-Id is non-empty": (r) =>
      typeof r.headers["X-Authgear-User-Id"] === "string" &&
      r.headers["X-Authgear-User-Id"].length > 0,
  });
  if (!checkResult) {
    fail("failed to call /_resolver/resolve");
  }
}
