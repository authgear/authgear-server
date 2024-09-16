import deepEqual from "deep-equal";
import { produce, createDraft } from "immer";
import { getReducedClientConfig } from "../graphql/portal/EditOAuthClientForm";
import {
  OAuthClientConfig,
  OAuthClientSecret,
  PortalAPIAppConfig,
  PortalAPISecretConfig,
  PortalAPISecretConfigUpdateInstruction,
} from "../types";
import { clearEmptyObject } from "../util/misc";
import { useAppSecretConfigForm } from "./useAppSecretConfigForm";

interface FormState {
  publicOrigin: string;
  clients: OAuthClientConfig[];
  editedClient: OAuthClientConfig | null;
  removeClientByID?: string;
  clientSecretMap: Partial<Record<string, string>>;
}

function constructFormState(
  config: PortalAPIAppConfig,
  secrets: PortalAPISecretConfig
): FormState {
  const clientSecretMap: Partial<Record<string, string>> =
    secrets.oauthClientSecrets?.reduce<Record<string, string>>(
      (acc: Record<string, string>, currValue: OAuthClientSecret) => {
        if (currValue.keys?.length && currValue.keys.length >= 1) {
          acc[currValue.clientID] = currValue.keys[0].key;
        }
        return acc;
      },
      {}
    ) ?? {};
  return {
    publicOrigin: config.http?.public_origin ?? "",
    clients: config.oauth?.clients ?? [],
    editedClient: null,
    removeClientByID: undefined,
    clientSecretMap,
  };
}

function constructConfig(
  config: PortalAPIAppConfig,
  secrets: PortalAPISecretConfig,
  _initialState: FormState,
  currentState: FormState,
  _effectiveConfig: PortalAPIAppConfig
): [PortalAPIAppConfig, PortalAPISecretConfig] {
  const [newConfig, _] = produce(
    [config, currentState],
    ([config, currentState]) => {
      config.oauth ??= {};
      config.oauth.clients = currentState.clients;

      if (currentState.removeClientByID) {
        config.oauth.clients = config.oauth.clients.filter(
          (c) => c.client_id !== currentState.removeClientByID
        );
        clearEmptyObject(config);
        return;
      }

      const client = currentState.editedClient;
      if (client) {
        const index = config.oauth.clients.findIndex(
          (c) => c.client_id === client.client_id
        );
        if (
          index !== -1 &&
          !deepEqual(
            getReducedClientConfig(client),
            getReducedClientConfig(config.oauth.clients[index]),
            { strict: true }
          )
        ) {
          config.oauth.clients[index] = createDraft(client);
        }
      }
      clearEmptyObject(config);
    }
  );
  return [newConfig, secrets];
}

function constructSecretUpdateInstruction(
  _config: PortalAPIAppConfig,
  _secrets: PortalAPISecretConfig,
  currentState: FormState
): PortalAPISecretConfigUpdateInstruction | undefined {
  if (currentState.removeClientByID) {
    return {
      oauthClientSecrets: {
        action: "cleanup",
        cleanupData: {
          keepClientIDs: currentState.clients
            .filter((c) => c.client_id !== currentState.removeClientByID)
            .map((c) => c.client_id),
        },
      },
    };
  }

  return undefined;
}

export function useOAuthClientForm(
  appID: string,
  secretVisitToken: string | null
): ReturnType<typeof useAppSecretConfigForm<FormState>> {
  return useAppSecretConfigForm({
    appID,
    secretVisitToken,
    constructFormState,
    constructConfig,
    constructSecretUpdateInstruction,
  });
}
