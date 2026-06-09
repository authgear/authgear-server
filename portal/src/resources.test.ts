import {
  ALL_LANGUAGES_TEMPLATES,
  RESOURCE_SETUP_SECONDARY_LOGIN_LINK_HTML,
  RESOURCE_SETUP_SECONDARY_LOGIN_LINK_TXT,
  RESOURCE_AUTHENTICATE_SECONDARY_LOGIN_LINK_HTML,
  RESOURCE_AUTHENTICATE_SECONDARY_LOGIN_LINK_TXT,
} from "./resources";

describe("secondary login-link email resources", () => {
  it("resolve to the correct template paths", () => {
    expect(
      RESOURCE_SETUP_SECONDARY_LOGIN_LINK_HTML.resourcePath.render({
        locale: "en",
      })
    ).toEqual("templates/en/messages/setup_secondary_login_link.html");
    expect(
      RESOURCE_SETUP_SECONDARY_LOGIN_LINK_TXT.resourcePath.render({
        locale: "en",
      })
    ).toEqual("templates/en/messages/setup_secondary_login_link.txt");
    expect(
      RESOURCE_AUTHENTICATE_SECONDARY_LOGIN_LINK_HTML.resourcePath.render({
        locale: "en",
      })
    ).toEqual("templates/en/messages/authenticate_secondary_login_link.html");
    expect(
      RESOURCE_AUTHENTICATE_SECONDARY_LOGIN_LINK_TXT.resourcePath.render({
        locale: "en",
      })
    ).toEqual("templates/en/messages/authenticate_secondary_login_link.txt");
  });

  it("are registered in ALL_LANGUAGES_TEMPLATES", () => {
    expect(ALL_LANGUAGES_TEMPLATES).toContain(
      RESOURCE_SETUP_SECONDARY_LOGIN_LINK_HTML
    );
    expect(ALL_LANGUAGES_TEMPLATES).toContain(
      RESOURCE_SETUP_SECONDARY_LOGIN_LINK_TXT
    );
    expect(ALL_LANGUAGES_TEMPLATES).toContain(
      RESOURCE_AUTHENTICATE_SECONDARY_LOGIN_LINK_HTML
    );
    expect(ALL_LANGUAGES_TEMPLATES).toContain(
      RESOURCE_AUTHENTICATE_SECONDARY_LOGIN_LINK_TXT
    );
  });
});
