import { check, fail } from "k6";
import exec from "k6/execution";
import {
  makeNationalPhoneNumberForLogin,
  makeLoginIDs,
  getIndex,
} from "./fixture.js";
import { authflowRun } from "./authflow.js";
import { enableBiometric, authenticateBiometric } from "./biometric.js";
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
    const jwk = enableBiometric({
      endpoint: ENDPOINT,
      client_id: CLIENT_ID,
      access_token: tokenResponse.access_token,
    });
    data.push(jwk);
  }
  return data;
}

export default function (data) {
  const index = getIndex();
  const jwk = data[index];
  const tokenResponse = authenticateBiometric({
    endpoint: ENDPOINT,
    client_id: CLIENT_ID,
    jwk,
  });
  const checkResult = check(tokenResponse, {
    "refresh_token is present": (r) =>
      typeof r.refresh_token === "string" && r.refresh_token !== "",
    "access_token is present": (r) =>
      typeof r.access_token === "string" && r.access_token !== "",
  });
  if (!checkResult) {
    fail("failed to authenticate with biometric");
  }
}
