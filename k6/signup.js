import {
  makeNationalPhoneNumberForSignup,
  makePhoneAndEmail,
} from "./fixture.js";
import { authflowRun } from "./authflow.js";

export default function () {
  const nationalPhone = makeNationalPhoneNumberForSignup();
  const { phone, email } = makePhoneAndEmail(nationalPhone);
  authflowRun({
    phone,
    email,
    type: "signup",
    name: "default",
  });
}
