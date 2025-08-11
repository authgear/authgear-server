import { useAppSecretConfigForm } from "./useAppSecretConfigForm";
import {
  PortalAPIAppConfig,
  PortalAPISecretConfig,
  PortalAPISecretConfigUpdateInstruction,
  OAuthClientSecret,
} from "../types";
import { useCallback } from "react";

interface FormState {
  generateSecretClientID: string | null;
  deleteSecretOptions: { clientID: string; keyID: string } | null;
  oauthClientSecrets: OAuthClientSecret[];
}

function constructFormState(
  _config: PortalAPIAppConfig,
  secrets: PortalAPISecretConfig
): FormState {
  return {
    generateSecretClientID: null,
    deleteSecretOptions: null,
    oauthClientSecrets: secrets.oauthClientSecrets ?? [],
  };
}

function constructConfig(
  config: PortalAPIAppConfig,
  secrets: PortalAPISecretConfig,
  _initialState: FormState,
  _currentState: FormState,
  _effectiveConfig: PortalAPIAppConfig
): [PortalAPIAppConfig, PortalAPISecretConfig] {
  return [config, secrets];
}

function constructSecretUpdateInstruction(
  _config: PortalAPIAppConfig,
  _secretConfig: PortalAPISecretConfig,
  currentState: FormState
): PortalAPISecretConfigUpdateInstruction | undefined {
  if (currentState.generateSecretClientID) {
    const instruction: PortalAPISecretConfigUpdateInstruction = {
      oauthClientSecrets: {
        action: "generate",
        generateData: { clientID: currentState.generateSecretClientID },
      },
    };
    return instruction;
  } else if (currentState.deleteSecretOptions) {
    const instruction: PortalAPISecretConfigUpdateInstruction = {
      oauthClientSecrets: {
        action: "delete",
        deleteData: currentState.deleteSecretOptions,
      },
    };
    return instruction;
  }
  return undefined;
}

export interface ClientSecretsHook {
  isLoading: boolean;
  isUpdating: boolean;
  generate: (clientID: string) => Promise<void>;
  delete: (clientID: string, keyID: string) => Promise<void>;
  loadError: unknown;
  reload: () => void;
  oauthClientSecrets: OAuthClientSecret[];
}

export function useGenerateClientSecret(
  appID: string,
  secretToken: string | null
): ClientSecretsHook {
  const { saveWithState, isLoading, isUpdating, loadError, reload, state } =
    useAppSecretConfigForm<FormState>({
      appID,
      secretVisitToken: secretToken,
      constructFormState,
      constructConfig,
      constructSecretUpdateInstruction,
    });

  const generate = useCallback(
    async (clientID: string) => {
      return saveWithState({
        generateSecretClientID: clientID,
        deleteSecretOptions: null,
        oauthClientSecrets: state.oauthClientSecrets,
      });
    },
    [saveWithState, state.oauthClientSecrets]
  );

  const deleteSecret = useCallback(
    async (clientID: string, keyID: string) => {
      return saveWithState({
        generateSecretClientID: null,
        deleteSecretOptions: { clientID, keyID },
        oauthClientSecrets: state.oauthClientSecrets,
      });
    },
    [saveWithState, state.oauthClientSecrets]
  );

  return {
    generate,
    delete: deleteSecret,
    isLoading,
    isUpdating,
    loadError,
    reload,
    oauthClientSecrets: state.oauthClientSecrets,
  };
}
