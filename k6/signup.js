import { makeNationalPhoneNumberForSignup, makeLoginIDs } from "./fixture.js";
import { authflowRun } from "./authflow.js";

export default function () {
  const nationalPhone = makeNationalPhoneNumberForSignup();
  const { username, phone, email } = makeLoginIDs(nationalPhone);
  authflowRun({
    username,
    phone,
    email,
    type: "signup",
    name: "default",
  });
}
