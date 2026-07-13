import { useCallback, useMemo, useRef, useState } from "react";
import deepEqual from "deep-equal";
import {
  PortalAPIAppConfig,
  PortalAPISecretConfig,
  PortalAPISecretConfigUpdateInstruction,
} from "../types";
import { useAppAndSecretConfigQuery } from "../graphql/portal/query/appAndSecretConfigQuery";
import { useUpdateAppAndSecretConfigMutation } from "../graphql/portal/mutations/updateAppAndSecretMutation";
import { useLiveState } from "./useSyncFormStates";

export interface AppSecretConfigFormModel<State> {
  isLoading: boolean;
  isUpdating: boolean;
  loadError: unknown;
  updateError: unknown;
  state: State;
  setState: (fn: (state: State) => State) => void;
  reload: () => void;
  reset: () => void;
  save: (ignoreConflict?: boolean) => Promise<void>;
  saveWithState: (state: State, ignoreConflict?: boolean) => Promise<void>;
  // Always-fresh dirty check, safe to call from anywhere (see
  // useSyncFormStates). For a render-time boolean (e.g. disabling a
  // button), derive one locally: useMemo(() => getIsDirty(), [getIsDirty]).
  getIsDirty: () => boolean;
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
  postSave?: (state: State) => Promise<void>;
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
    postSave,
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
  const [currentState, setCurrentState, getCurrentState] = useLiveState<
    State | null
  >(
    constructInitialCurrentState != null
      ? constructInitialCurrentState(initialState)
      : null
  );

  const computeIsDirty = useCallback(
    (current: State | null) => {
      if (!rawAppConfig || !current) {
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
        current,
        effectiveConfig
      );
      const isConfigDirty = !deepEqual(originalConfig, newConfig, {
        strict: true,
      });
      if (isConfigDirty) {
        return true;
      }
      const secretUpdateInstruction = constructSecretUpdateInstruction
        ? constructSecretUpdateInstruction(newConfig[0], newConfig[1], current)
        : undefined;
      const isSecretDirty =
        secretUpdateInstruction != null &&
        Object.entries(secretUpdateInstruction).some(
          ([_, instruction]) => instruction != null
        );
      return isSecretDirty;
    },
    [
      rawAppConfig,
      constructConfig,
      secrets,
      initialState,
      effectiveConfig,
      constructSecretUpdateInstruction,
    ]
  );

  const getIsDirty = useCallback(
    () => computeIsDirty(getCurrentState()),
    [computeIsDirty, getCurrentState]
  );

  const reset = useCallback(() => {
    if (isUpdating) {
      return;
    }
    resetError();
    setCurrentState(null);
  }, [isUpdating, resetError, setCurrentState]);

  const _save = useCallback(
    async (state: State, ignoreConflict: boolean) => {
      if (!rawAppConfig) {
        return;
      }

      const newConfig = constructConfig(
        rawAppConfig,
        secrets,
        initialState,
        state,
        effectiveConfig
      );

      // The app and secret config that pass to constructSecretUpdateInstruction
      // are the updated config that we are going to send to the server
      const secretUpdateInstruction = constructSecretUpdateInstruction
        ? constructSecretUpdateInstruction(newConfig[0], newConfig[1], state)
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
        await postSave?.(state);
        await reload();
        setCurrentState(null);
      } finally {
        setIsUpdating(false);
      }
    },
    [
      rawAppConfig,
      constructConfig,
      secrets,
      initialState,
      effectiveConfig,
      constructSecretUpdateInstruction,
      updateConfig,
      rawAppConfigChecksum,
      secretConfigChecksum,
      reload,
      postSave,
      setIsUpdating,
      setCurrentState,
    ]
  );

  const save = useCallback(
    async (ignoreConflict: boolean = false) => {
      if (!currentState) {
        return;
      }
      await _save(currentState, ignoreConflict);
    },
    [currentState, _save]
  );

  const saveWithState = useCallback(
    async (state: State, ignoreConflict: boolean = false) => {
      await _save(state, ignoreConflict);
    },
    [_save]
  );

  const state = currentState ?? initialState;

  const initialStateRef = useRef(initialState);
  // eslint-disable-next-line react-hooks/refs
  initialStateRef.current = initialState;
  const setState = useCallback((fn: (state: State) => State) => {
    setCurrentState((s) => {
      // setState can easily be captured by useCallback / useMemo causing stalled initialState
      // Use a ref to reference to the latest value to prevent this problem
      const newState = fn(s ?? initialStateRef.current);
      return newState;
    });
  }, [setCurrentState]);

  return {
    isLoading,
    isUpdating,
    loadError,
    updateError,
    state,
    setState,
    // eslint-disable-next-line @typescript-eslint/no-misused-promises, @typescript-eslint/strict-void-return
    reload,
    reset,
    save,
    saveWithState,
    getIsDirty,
  };
}
