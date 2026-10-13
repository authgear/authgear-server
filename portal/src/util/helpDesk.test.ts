import { describe, it, expect } from "@jest/globals";
import { makeHelpDeskURL } from "./helpDesk";

describe("makeHelpDeskURL", () => {
  it("includes every known value", () => {
    expect(
      makeHelpDeskURL({
        purpose: "sms-otp",
        projectID: "myapp",
        email: "user+1@example.com",
        fullName: "Chan Tai Man",
      })
    ).toEqual(
      "https://www.authgear.com/help-desk/?project=myapp&email=user%2B1%40example.com&name=Chan+Tai+Man&purpose=sms-otp"
    );
  });

  it("omits unknown values", () => {
    expect(
      makeHelpDeskURL({
        purpose: "project-limit",
        projectID: null,
        email: "user@example.com",
        fullName: "",
      })
    ).toEqual(
      "https://www.authgear.com/help-desk/?email=user%40example.com&purpose=project-limit"
    );
  });
});
