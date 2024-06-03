import { check, fail } from "k6";
import exec from "k6/execution";
import {
  makeNationalPhoneNumberForLogin,
  makeLoginIDs,
  getIndex,
} from "./fixture.js";
import { authflowRun } from "./authflow.js";
import { refreshAccessToken } from "./oauth.js";
import { ENDPOINT, CLIENT_ID } from "./env.js";

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
  const originalTokenResponse = data[index];
  const newTokenResponse = refreshAccessToken({
    endpoint: ENDPOINT,
    client_id: CLIENT_ID,
    refresh_token: originalTokenResponse.refresh_token,
  });
  const checkResult = check(
    { originalTokenResponse, newTokenResponse },
    {
      "access token is refreshed": (r) => {
        return (
          r.originalTokenResponse.access_token !==
          r.newTokenResponse.access_token
        );
      },
    },
  );
  if (!checkResult) {
    fail("failed to refresh access token");
  }
}
