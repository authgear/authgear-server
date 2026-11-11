import { describe, it, expect, jest } from "@jest/globals";
import { constructConfig, constructFormState } from "./useOAuthClientForm";
import {
  OAuthClientConfig,
  PortalAPIAppConfig,
  SAMLNameIDFormat,
  SAMLServiceProviderConfig,
} from "../types";

// Both modules pull in UI and API client code Jest cannot load, and neither is
// reached from constructFormState / constructConfig when editedClient is null.
jest.mock("../graphql/portal/EditOAuthClientForm", () => ({
  getReducedClientConfig: () => {
    throw new Error("unexpected call");
  },
}));
jest.mock("./useAppSecretConfigForm", () => ({
  useAppSecretConfigForm: () => {
    throw new Error("unexpected call");
  },
}));

const client = {
  client_id: "client",
  name: "Client",
  x_application_type: "confidential",
  redirect_uris: ["https://sp.example.com"],
} as OAuthClientConfig;

function makeConfig(
  sp: Partial<SAMLServiceProviderConfig>
): PortalAPIAppConfig {
  return {
    id: "app",
    oauth: { clients: [client] },
    saml: {
      service_providers: [
        {
          client_id: client.client_id,
          nameid_format: SAMLNameIDFormat.Unspecified,
          acs_urls: ["https://sp.example.com/acs"],
          ...sp,
        },
      ],
    },
  };
}

// Load the config into form state, apply an edit, and save it back.
function saveSP(
  sp: Partial<SAMLServiceProviderConfig>,
  edit: { audience?: string; audienceIncludesACSURL?: boolean }
): SAMLServiceProviderConfig {
  const config = makeConfig(sp);
  const state = constructFormState(config, {});
  const edited = {
    ...state,
    samlServiceProviders: state.samlServiceProviders.map((s) => ({
      ...s,
      ...edit,
    })),
  };
  const [newConfig] = constructConfig(config, {}, state, edited, config);
  return newConfig.saml!.service_providers![0];
}

describe("useOAuthClientForm audience_includes_acs_url", () => {
  it("does not save false when the audience is cleared", () => {
    const sp = saveSP(
      {
        audience: "https://sp.example.com",
        audience_includes_acs_url: false,
      },
      { audience: "" }
    );
    expect(sp.audience).toBe(undefined);
    expect(sp.audience_includes_acs_url).toBe(undefined);
  });

  it("does not save false when the audience is blank", () => {
    const sp = saveSP(
      { audience: "https://sp.example.com" },
      { audience: "   ", audienceIncludesACSURL: false }
    );
    expect(sp.audience).toBe(undefined);
    expect(sp.audience_includes_acs_url).toBe(undefined);
  });
});
