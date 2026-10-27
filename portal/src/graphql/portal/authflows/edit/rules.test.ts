import { describe, it, expect } from "@jest/globals";
import {
  allowedAuthentications,
  allowedIdentifications,
  allowedStepTypes,
  authenticationKind,
  canHaveOptionForm,
  sendsCodes,
  newOption,
  newStep,
} from "./rules";

describe("allowedStepTypes", () => {
  it("lists the step types of each flow type", () => {
    expect(allowedStepTypes("signup")).toEqual([
      "identify",
      "create_authenticator",
      "verify",
      "fill_form",
      "view_recovery_code",
      "prompt_create_passkey",
    ]);
    expect(allowedStepTypes("promote")).toEqual(allowedStepTypes("signup"));
    expect(allowedStepTypes("login")).toEqual([
      "identify",
      "authenticate",
      "check_account_status",
      "terminate_other_sessions",
      "change_password",
      "prompt_create_passkey",
      "fill_form",
    ]);
    expect(allowedStepTypes("signup_login")).toEqual(["identify"]);
    expect(allowedStepTypes("reauth")).toEqual(["identify", "authenticate"]);
    expect(allowedStepTypes("account_recovery")).toEqual([
      "identify",
      "select_destination",
      "verify_account_recovery_code",
      "reset_password",
    ]);
  });
});

describe("allowedIdentifications", () => {
  it("follows the schema of each flow type", () => {
    expect(allowedIdentifications("signup")).toEqual([
      "email",
      "phone",
      "username",
      "oauth",
      "passkey",
      "ldap",
    ]);
    expect(allowedIdentifications("login")).toEqual([
      "email",
      "phone",
      "username",
      "oauth",
      "passkey",
      "ldap",
      "select_account",
    ]);
    expect(allowedIdentifications("reauth")).toEqual(["id_token"]);
    expect(allowedIdentifications("account_recovery")).toEqual([
      "email",
      "phone",
      "username",
    ]);
  });
});

describe("allowedAuthentications", () => {
  it("adds recovery code and device token to login authenticate only", () => {
    expect(allowedAuthentications("login", "authenticate")).toContain(
      "recovery_code"
    );
    expect(allowedAuthentications("reauth", "authenticate")).not.toContain(
      "recovery_code"
    );
    expect(allowedAuthentications("signup", "create_authenticator")).toEqual([
      "primary_password",
      "primary_oob_otp_email",
      "primary_oob_otp_sms",
      "secondary_password",
      "secondary_totp",
      "secondary_oob_otp_email",
      "secondary_oob_otp_sms",
    ]);
  });
});

describe("newStep", () => {
  it("starts a choice step with its first option", () => {
    expect(newStep("login", "identify", null)).toEqual({
      type: "identify",
      one_of: [{ identification: "email" }],
    });
    expect(newStep("login", "authenticate", null)).toEqual({
      type: "authenticate",
      one_of: [{ authentication: "primary_password" }],
    });
    expect(newStep("signup", "create_authenticator", null)).toEqual({
      type: "create_authenticator",
      one_of: [{ authentication: "primary_password" }],
    });
    expect(newStep("reauth", "identify", null)).toEqual({
      type: "identify",
      one_of: [{ identification: "id_token" }],
    });
  });

  it("hands signup_login options off to the default flows", () => {
    expect(newStep("signup_login", "identify", null)).toEqual({
      type: "identify",
      one_of: [
        {
          identification: "email",
          login_flow: "default",
          signup_flow: "default",
        },
      ],
    });
  });

  it("gives account recovery identify options a nested step", () => {
    expect(newStep("account_recovery", "identify", null)).toEqual({
      type: "identify",
      one_of: [
        { identification: "email", steps: [{ type: "select_destination" }] },
      ],
    });
  });

  it("sets target_step and a form field where required", () => {
    expect(newStep("signup", "verify", "identify")).toEqual({
      type: "verify",
      target_step: "identify",
    });
    expect(newStep("signup", "fill_form", null, "/x_team")).toEqual({
      type: "fill_form",
      form: { fields: [{ user_profile: { pointer: "/x_team" } }] },
    });
    expect(newStep("login", "check_account_status", null)).toEqual({
      type: "check_account_status",
    });
  });
});

describe("newOption", () => {
  it("takes the first value the step does not list", () => {
    expect(
      newOption("login", {
        type: "identify",
        one_of: [{ identification: "email" }, { identification: "phone" }],
      })
    ).toEqual({ identification: "username" });
    expect(
      newOption("login", {
        type: "authenticate",
        one_of: [{ authentication: "primary_password" }],
      })
    ).toEqual({ authentication: "primary_passkey" });
  });

  it("returns null when every value is used", () => {
    expect(
      newOption("reauth", {
        type: "identify",
        one_of: [{ identification: "id_token" }],
      })
    ).toBeNull();
  });

  it("returns null for a step that is not a choice", () => {
    expect(newOption("login", { type: "change_password" })).toBeNull();
  });
});

describe("newOption for signup_login", () => {
  it("does not give select_account a signup_flow", () => {
    const used = ["email", "phone", "username", "oauth", "passkey", "ldap"].map(
      (identification) => ({ identification })
    );
    expect(
      newOption("signup_login", { type: "identify", one_of: used })
    ).toEqual({ identification: "select_account", login_flow: "default" });
  });
});

describe("signup_login options that only log in", () => {
  it("does not give passkey a signup_flow", () => {
    const used = ["email", "phone", "username", "oauth"].map(
      (identification) => ({ identification })
    );
    expect(
      newOption("signup_login", { type: "identify", one_of: used })
    ).toEqual({ identification: "passkey", login_flow: "default" });
  });
});

describe("authenticationKind", () => {
  it("tells primary authenticators from 2FA ones", () => {
    expect(authenticationKind("primary_password")).toEqual("primary");
    expect(authenticationKind("primary_passkey")).toEqual("primary");
    expect(authenticationKind("secondary_password")).toEqual("secondary");
    expect(authenticationKind("secondary_totp")).toEqual("secondary");
  });

  it("counts recovery codes and remembered devices as 2FA", () => {
    expect(authenticationKind("recovery_code")).toEqual("secondary");
    expect(authenticationKind("device_token")).toEqual("secondary");
  });
});

describe("sendsCodes", () => {
  it("is true for authenticators that send codes to an email or phone", () => {
    expect(sendsCodes("primary_oob_otp_email")).toEqual(true);
    expect(sendsCodes("secondary_oob_otp_sms")).toEqual(true);
  });

  it("is false for the rest, which have no address to send to", () => {
    expect(sendsCodes("secondary_totp")).toEqual(false);
    expect(sendsCodes("primary_password")).toEqual(false);
    expect(sendsCodes(undefined)).toEqual(false);
  });
});

describe("canHaveOptionForm", () => {
  it("allows a form on a login ID option of a signup or promote flow", () => {
    expect(canHaveOptionForm("signup", "email")).toBe(true);
    expect(canHaveOptionForm("promote", "username")).toBe(true);
    expect(canHaveOptionForm("signup", "phone")).toBe(true);
  });

  it("does not allow one on other options or flow types", () => {
    expect(canHaveOptionForm("signup", "oauth")).toBe(false);
    expect(canHaveOptionForm("signup", undefined)).toBe(false);
    expect(canHaveOptionForm("login", "email")).toBe(false);
    expect(canHaveOptionForm("signup_login", "email")).toBe(false);
  });
});
