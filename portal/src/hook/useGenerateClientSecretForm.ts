import { useAppSecretConfigForm } from "./useAppSecretConfigForm";
import {
  PortalAPIAppConfig,
  PortalAPISecretConfig,
  PortalAPISecretConfigUpdateInstruction,
  OAuthClientSecret,
} from "../types";
import { useCallback } from "react";

export interface FormState {
  clientID: string | null;
  oauthClientSecrets: OAuthClientSecret[];
}

function constructFormState(
  _config: PortalAPIAppConfig,
  secrets: PortalAPISecretConfig
): FormState {
  return {
    clientID: null,
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
  if (currentState.clientID) {
    const instruction: PortalAPISecretConfigUpdateInstruction = {
      oauthClientSecrets: {
        action: "generate",
        generateData: { clientID: currentState.clientID },
      },
    };
    return instruction;
  }
  return undefined;
}

export interface GenerateClientSecretHook {
  isLoading: boolean;
  isUpdating: boolean;
  generate: (clientID: string) => Promise<void>;
  loadError: unknown;
  reload: () => void;
  oauthClientSecrets: OAuthClientSecret[];
}

export function useGenerateClientSecret(
  appID: string,
  secretToken: string | null
): GenerateClientSecretHook {
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
        clientID,
        oauthClientSecrets: state.oauthClientSecrets,
      });
    },
    [saveWithState, state.oauthClientSecrets]
  );

  return {
    generate,
    isLoading,
    isUpdating,
    loadError,
    reload,
    oauthClientSecrets: state.oauthClientSecrets,
  };
}
