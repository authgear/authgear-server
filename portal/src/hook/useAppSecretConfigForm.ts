import { useCallback, useMemo, useState } from "react";
import deepEqual from "deep-equal";
import {
  PortalAPIAppConfig,
  PortalAPISecretConfig,
  PortalAPISecretConfigUpdateInstruction,
} from "../types";
import { useAppAndSecretConfigQuery } from "../graphql/portal/query/appAndSecretConfigQuery";
import { useUpdateAppAndSecretConfigMutation } from "../graphql/portal/mutations/updateAppAndSecretMutation";

export interface AppSecretConfigFormModel<State> {
  isLoading: boolean;
  isUpdating: boolean;
  isDirty: boolean;
  loadError: unknown;
  updateError: unknown;
  state: State;
  setState: (fn: (state: State) => State) => void;
  reload: () => void;
  reset: () => void;
  save: (ignoreConflict?: boolean) => Promise<void>;
}

export type StateConstructor<State> = (
  config: PortalAPIAppConfig,
  secrets: PortalAPISecretConfig
) => State;
export type ConfigConstructor<State> = (
  config: PortalAPIAppConfig,
  secrets: PortalAPISecretConfig,
  initialState: State,
  currentState: State,
  effectiveConfig: PortalAPIAppConfig
) => [PortalAPIAppConfig, PortalAPISecretConfig];
export type SecretUpdateInstructionConstructor<State> = (
  config: PortalAPIAppConfig,
  secrets: PortalAPISecretConfig,
  currentState: State
) => PortalAPISecretConfigUpdateInstruction | undefined;
export type InitialCurrentStateConstructor<State> = (state: State) => State;

interface UseAppSecretConfigFormOptions<State> {
  appID: string;
  secretVisitToken: string | null;
  constructFormState: StateConstructor<State>;
  constructConfig: ConfigConstructor<State>;
  constructSecretUpdateInstruction?: SecretUpdateInstructionConstructor<State>;
  constructInitialCurrentState?: InitialCurrentStateConstructor<State>;
}

export function useAppSecretConfigForm<State>(
  options: UseAppSecretConfigFormOptions<State>
): AppSecretConfigFormModel<State> {
  const {
    appID,
    secretVisitToken,
    constructFormState,
    constructConfig,
    constructSecretUpdateInstruction,
    constructInitialCurrentState,
  } = options;

  const {
    isLoading,
    loadError,
    rawAppConfig,
    rawAppConfigChecksum,
    effectiveAppConfig,
    secretConfig,
    secretConfigChecksum,
    refetch: reload,
  } = useAppAndSecretConfigQuery(appID, secretVisitToken);
  const {
    error: updateError,
    updateAppAndSecretConfig: updateConfig,
    resetError,
  } = useUpdateAppAndSecretConfigMutation(appID);

  const [isUpdating, setIsUpdating] = useState(false);

  const effectiveConfig = useMemo(
    () => effectiveAppConfig ?? { id: appID },
    [effectiveAppConfig, appID]
  );
  const secrets = useMemo(() => secretConfig ?? {}, [secretConfig]);

  const initialState = useMemo(
    () => constructFormState(effectiveConfig, secrets),
    [effectiveConfig, secrets, constructFormState]
  );
  const [currentState, setCurrentState] = useState<State | null>(
    constructInitialCurrentState != null
      ? constructInitialCurrentState(initialState)
      : null
  );

  const isDirty = useMemo(() => {
    if (!rawAppConfig || !currentState) {
      return false;
    }
    const originalConfig = constructConfig(
      rawAppConfig,
      secrets,
      initialState,
      initialState,
      effectiveConfig
    );
    const newConfig = constructConfig(
      rawAppConfig,
      secrets,
      initialState,
      currentState,
      effectiveConfig
    );
    const isConfigDirty = !deepEqual(originalConfig, newConfig, {
      strict: true,
    });
    if (isConfigDirty) {
      return true;
    }
    const secretUpdateInstruction = constructSecretUpdateInstruction
      ? constructSecretUpdateInstruction(
          newConfig[0],
          newConfig[1],
          currentState
        )
      : undefined;
    const isSecretDirty =
      secretUpdateInstruction != null &&
      Object.entries(secretUpdateInstruction).some(
        ([_, instruction]) => instruction != null
      );
    return isSecretDirty;
  }, [
    rawAppConfig,
    currentState,
    constructConfig,
    secrets,
    initialState,
    effectiveConfig,
    constructSecretUpdateInstruction,
  ]);

  const reset = useCallback(() => {
    if (isUpdating) {
      return;
    }
    resetError();
    setCurrentState(null);
  }, [isUpdating, resetError]);

  const save = useCallback(
    async (ignoreConflict: boolean = false) => {
      if (!rawAppConfig || !currentState) {
        return;
      } else if (!isDirty || isUpdating) {
        return;
      }

      const newConfig = constructConfig(
        rawAppConfig,
        secrets,
        initialState,
        currentState,
        effectiveConfig
      );

      // The app and secret config that pass to constructSecretUpdateInstruction
      // are the updated config that we are going to send to the server
      const secretUpdateInstruction = constructSecretUpdateInstruction
        ? constructSecretUpdateInstruction(
            newConfig[0],
            newConfig[1],
            currentState
          )
        : undefined;

      setIsUpdating(true);
      try {
        await updateConfig({
          appConfig: newConfig[0],
          appConfigChecksum: rawAppConfigChecksum,
          secretConfigUpdateInstructions: secretUpdateInstruction,
          secretConfigUpdateInstructionsChecksum: secretConfigChecksum,
          ignoreConflict,
        });
        await reload();
        setCurrentState(null);
      } finally {
        setIsUpdating(false);
      }
    },
    [
      rawAppConfig,
      rawAppConfigChecksum,
      currentState,
      isDirty,
      isUpdating,
      constructConfig,
      secrets,
      secretConfigChecksum,
      initialState,
      effectiveConfig,
      constructSecretUpdateInstruction,
      updateConfig,
      reload,
    ]
  );

  const state = currentState ?? initialState;
  const setState = useCallback(
    (fn: (state: State) => State) => {
      setCurrentState((s) => {
        const newState = fn(s ?? initialState);
        return newState;
      });
    },
    [initialState]
  );

  return {
    isLoading,
    isUpdating,
    isDirty,
    loadError,
    updateError,
    state,
    setState,
    // eslint-disable-next-line @typescript-eslint/no-misused-promises
    reload,
    reset,
    save,
  };
}
