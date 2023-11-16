import exec from "k6/execution";
import {
  makeNationalPhoneNumberForLogin,
  makePhoneAndEmail,
} from "./fixture.js";
import { authflowRun } from "./authflow.js";

export function setup() {
  // No need to return anything because the test function does not need anything.
  const vus = exec.test.options.scenarios.default.vus;
  for (let i = 1; i <= vus; ++i) {
    const nationalPhone = makeNationalPhoneNumberForLogin({ vu: i });
    const { email, phone } = makePhoneAndEmail(nationalPhone);
    authflowRun({
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
  const { phone, email } = makePhoneAndEmail(nationalPhone);
  authflowRun({
    phone,
    email,
    type: "login",
    name: "default",
  });
}
