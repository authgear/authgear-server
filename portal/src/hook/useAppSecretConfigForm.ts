import { useCallback, useMemo, useState } from "react";
import deepEqual from "deep-equal";
import { PortalAPIAppConfig, PortalAPISecretConfig } from "../types";
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
  save: () => void;
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

export function useAppSecretConfigForm<State>(
  appID: string,
  constructState: StateConstructor<State>,
  constructConfig: ConfigConstructor<State>
): AppSecretConfigFormModel<State> {
  const {
    loading: isLoading,
    error: loadError,
    rawAppConfig,
    effectiveAppConfig,
    secretConfig,
    refetch: reload,
  } = useAppAndSecretConfigQuery(appID);
  const {
    loading: isUpdating,
    error: updateError,
    updateAppAndSecretConfig: updateConfig,
    resetError,
  } = useUpdateAppAndSecretConfigMutation(appID);

  const effectiveConfig = useMemo(
    () => effectiveAppConfig ?? { id: appID },
    [effectiveAppConfig, appID]
  );
  const secrets = useMemo(
    () => secretConfig ?? { secrets: [] },
    [secretConfig]
  );

  const initialState = useMemo(
    () => constructState(effectiveConfig, secrets),
    [effectiveConfig, secrets, constructState]
  );
  const [currentState, setCurrentState] = useState<State | null>(null);

  const isDirty = useMemo(() => {
    if (!rawAppConfig || !currentState) {
      return false;
    }
    return !deepEqual(
      constructConfig(
        rawAppConfig,
        secrets,
        initialState,
        initialState,
        effectiveConfig
      ),
      constructConfig(
        rawAppConfig,
        secrets,
        initialState,
        currentState,
        effectiveConfig
      ),
      { strict: true }
    );
  }, [
    constructConfig,
    rawAppConfig,
    secrets,
    effectiveConfig,
    initialState,
    currentState,
  ]);

  const reset = useCallback(() => {
    if (isUpdating) {
      return;
    }
    resetError();
    setCurrentState(null);
  }, [isUpdating, resetError]);

  const save = useCallback(() => {
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
    updateConfig(newConfig[0], newConfig[1])
      .then(() => setCurrentState(null))
      .catch(() => {});
  }, [
    isDirty,
    isUpdating,
    constructConfig,
    rawAppConfig,
    secrets,
    effectiveConfig,
    initialState,
    currentState,
    updateConfig,
  ]);

  const state = currentState ?? initialState;
  const setState = useCallback(
    (fn: (state: State) => State) => {
      setCurrentState(fn(state));
    },
    [state]
  );

  return {
    isLoading,
    isUpdating,
    isDirty,
    loadError,
    updateError,
    state,
    setState,
    reload,
    reset,
    save,
  };
}
