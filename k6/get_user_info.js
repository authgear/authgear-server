import { check, fail } from "k6";
import exec from "k6/execution";
import {
  makeNationalPhoneNumberForLogin,
  makeLoginIDs,
  getIndex,
} from "./fixture.js";
import { authflowRun } from "./authflow.js";
import { getUserInfo } from "./oauth.js";
import { ENDPOINT } from "./env.js";

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
  const userInfo = getUserInfo({
    endpoint: ENDPOINT,
    access_token: tokenResponse.access_token,
  });
  const checkResult = check(userInfo, {
    "sub is present in user info": (r) =>
      typeof r.sub === "string" && r.sub !== "",
  });
  if (!checkResult) {
    fail("failed to get user info");
  }
}
