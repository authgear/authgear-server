import { describe, it, expect } from "@jest/globals";
import {
  ifNeededMessageIDs,
  optionLabel,
  destinationMessageID,
  optionalMessageIDs,
  pagelessNoteMessageIDs,
  phoneCodeSetupMessageID,
  stepDescriptionMessageID,
  verifyCodeMessageID,
} from "./labels";
import { KNOWN_STEP_TYPES } from "./tree";

describe("ifNeededMessageIDs", () => {
  it.each([
    "change_password",
    "terminate_other_sessions",
    "prompt_create_passkey",
  ])("gives %s a short condition and a full note", (type) => {
    expect(ifNeededMessageIDs(type)).toEqual({
      condition: `AuthFlowsScreen.if-needed.condition.${type}`,
      note: `AuthFlowsScreen.if-needed.${type}`,
    });
  });

  it("returns null for a step that always shows", () => {
    expect(ifNeededMessageIDs("identify")).toBeNull();
  });
});

describe("optionalMessageIDs", () => {
  it("names 2FA when every option is a second factor", () => {
    expect(
      optionalMessageIDs({
        type: "authenticate",
        optional: true,
        one_of: [
          { authentication: "secondary_totp" },
          { authentication: "secondary_oob_otp_email" },
          { authentication: "recovery_code" },
          { authentication: "device_token" },
        ],
      })
    ).toEqual({
      condition: "AuthFlowsScreen.optional.condition.2fa",
      note: "AuthFlowsScreen.optional.note.2fa",
    });
  });

  it("falls back to its options when they are not all second factors", () => {
    expect(
      optionalMessageIDs({
        type: "authenticate",
        optional: true,
        one_of: [
          { authentication: "primary_passkey" },
          { authentication: "secondary_totp" },
        ],
      })
    ).toEqual({
      condition: "AuthFlowsScreen.optional.condition.any",
      note: "AuthFlowsScreen.optional.note.any",
    });
  });
});

describe("stepDescriptionMessageID", () => {
  it("explains check_account_status", () => {
    expect(stepDescriptionMessageID("check_account_status")).toEqual(
      "AuthFlowsScreen.step-description.check_account_status"
    );
  });

  it.each([...KNOWN_STEP_TYPES])("describes %s", (type) => {
    expect(stepDescriptionMessageID(type)).toEqual(
      `AuthFlowsScreen.step-description.${type}`
    );
  });

  it("returns null for a step type this portal does not know", () => {
    expect(stepDescriptionMessageID("future_step")).toBeNull();
  });
});

describe("phoneCodeSetupMessageID", () => {
  it.each([
    ["sms", "AuthFlowsScreen.setup.oob_otp_sms.sms"],
    ["whatsapp", "AuthFlowsScreen.setup.oob_otp_sms.whatsapp"],
    ["whatsapp_sms", "AuthFlowsScreen.setup.oob_otp_sms.whatsapp_sms"],
  ] as const)("names the channel for %s", (mode, expected) => {
    expect(phoneCodeSetupMessageID(mode)).toEqual(expected);
  });

  it("says phone when the mode is unknown", () => {
    expect(phoneCodeSetupMessageID(undefined)).toEqual(
      "AuthFlowsScreen.setup.oob_otp_sms.any"
    );
  });
});

describe("optionLabel", () => {
  const renderToString = (id: string) => id;

  it.each([
    ["sms", "AuthFlowsScreen.authentication.oob_otp_sms.sms"],
    ["whatsapp", "AuthFlowsScreen.authentication.oob_otp_sms.whatsapp"],
    ["whatsapp_sms", "AuthFlowsScreen.authentication.oob_otp_sms.whatsapp_sms"],
  ] as const)("names the phone channel for %s", (mode, expected) => {
    expect(
      optionLabel(
        { authentication: "secondary_oob_otp_sms" },
        renderToString,
        mode
      )
    ).toEqual(expected);
  });

  it("says phone when the mode is unknown", () => {
    expect(
      optionLabel({ authentication: "primary_oob_otp_sms" }, renderToString)
    ).toEqual("AuthFlowsScreen.authentication.oob_otp_sms");
  });
});

describe("verifyCodeMessageID", () => {
  it("names email", () => {
    expect(verifyCodeMessageID("email", "sms")).toEqual(
      "AuthFlowsScreen.verify.email"
    );
  });

  it.each([
    ["sms", "AuthFlowsScreen.verify.phone.sms"],
    ["whatsapp", "AuthFlowsScreen.verify.phone.whatsapp"],
    ["whatsapp_sms", "AuthFlowsScreen.verify.phone.whatsapp_sms"],
  ] as const)("names the phone channel for %s", (mode, expected) => {
    expect(verifyCodeMessageID("phone", mode)).toEqual(expected);
  });

  it("stays generic otherwise", () => {
    expect(verifyCodeMessageID(null, "sms")).toEqual(
      "AuthFlowsScreen.sketch.enter-code"
    );
    expect(verifyCodeMessageID("phone", undefined)).toEqual(
      "AuthFlowsScreen.sketch.enter-code"
    );
  });
});

describe("pagelessNoteMessageIDs", () => {
  it("explains identify by ID token", () => {
    expect(
      pagelessNoteMessageIDs({
        type: "identify",
        one_of: [{ identification: "id_token" }],
      })
    ).toEqual({
      line: "AuthFlowsScreen.pageless.id_token",
      note: "AuthFlowsScreen.pageless.id_token.note",
    });
  });

  it("returns null for other steps", () => {
    expect(pagelessNoteMessageIDs({ type: "check_account_status" })).toBeNull();
  });
});

describe("destinationMessageID", () => {
  it.each([
    ["email", "link", "AuthFlowsScreen.destination.email.link"],
    ["email", "code", "AuthFlowsScreen.destination.email.code"],
    ["sms", "code", "AuthFlowsScreen.destination.sms.code"],
    ["sms", "link", "AuthFlowsScreen.destination.sms.link"],
    ["whatsapp", "code", "AuthFlowsScreen.destination.whatsapp.code"],
  ])("describes %s with %s", (channel, otpForm, expected) => {
    expect(destinationMessageID(channel, otpForm)).toEqual(expected);
  });

  it("returns null for an unknown channel", () => {
    expect(destinationMessageID("pigeon", "code")).toBeNull();
  });
});
