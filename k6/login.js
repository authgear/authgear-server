import exec from "k6/execution";
import { makeNationalPhoneNumberForLogin, makeLoginIDs } from "./fixture.js";
import { authflowRun } from "./authflow.js";

export function setup() {
  // No need to return anything because the test function does not need anything.
  const vus = exec.test.options.scenarios.default.vus;
  for (let i = 1; i <= vus; ++i) {
    const nationalPhone = makeNationalPhoneNumberForLogin({ vu: i });
    const { username, email, phone } = makeLoginIDs(nationalPhone);
    authflowRun({
      username,
      phone,
      email,
      type: "signup",
      name: "default",
    });
  }
}

export default function () {
  const nationalPhone = makeNationalPhoneNumberForLogin({
    vu: exec.vu.idInTest,
  });
  const { username, phone, email } = makeLoginIDs(nationalPhone);
  authflowRun({
    username,
    phone,
    email,
    type: "login",
    name: "default",
  });
}
