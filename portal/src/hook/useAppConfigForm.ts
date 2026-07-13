import { useCallback, useMemo, useState } from "react";
import deepEqual from "deep-equal";
import { useAppAndSecretConfigQuery } from "../graphql/portal/query/appAndSecretConfigQuery";
import { useUpdateAppAndSecretConfigMutation } from "../graphql/portal/mutations/updateAppAndSecretMutation";
import { PortalAPIAppConfig } from "../types";
import { APIError } from "../error/error";
import { useLiveState } from "./useSyncFormStates";

export interface AppConfigFormModel<State> {
  isLoading: boolean;
  isUpdating: boolean;
  isSubmitted: boolean;
  canSave?: boolean;
  loadError: unknown;
  updateError: unknown;
  initialState: State;
  state: State;
  setState: (fn: (state: State) => State) => void;
  reload: () => void;
  reset: () => void;
  save: (ignoreConflict?: boolean) => Promise<void>;
  saveWith: (
    fn: (state: State) => State,
    ignoreConflict?: boolean
  ) => Promise<void>;
  setCanSave: (canSave?: boolean) => void;
  effectiveConfig: PortalAPIAppConfig;
  // Always-fresh dirty check, safe to call from anywhere (see
  // useSyncFormStates). For a render-time boolean (e.g. disabling a
  // button), derive one locally: useMemo(() => getIsDirty(), [getIsDirty]).
  getIsDirty: () => boolean;
}

export type StateConstructor<State> = (config: PortalAPIAppConfig) => State;
export type ConfigConstructor<State> = (
  config: PortalAPIAppConfig,
  initialState: State,
  currentState: State,
  effectiveConfig: PortalAPIAppConfig
) => PortalAPIAppConfig;
export type InitialCurrentStateConstructor<State> = (state: State) => State;

interface UseAppConfigFormOptions<State> {
  appID: string;
  constructFormState: StateConstructor<State>;
  constructConfig: ConfigConstructor<State>;
  constructInitialCurrentState?: InitialCurrentStateConstructor<State>;
  validate?: (state: State) => APIError | null;
  initialCanSave?: boolean;
}

export function useAppConfigForm<State>(
  options: UseAppConfigFormOptions<State>
): AppConfigFormModel<State> {
  const {
    appID,
    constructFormState,
    constructConfig,
    constructInitialCurrentState,
    validate,
    initialCanSave,
  } = options;

  const {
    isLoading,
    loadError,
    effectiveAppConfig,
    rawAppConfig: rawConfig,
    rawAppConfigChecksum,
    secretConfig,
    refetch: reload,
  } = useAppAndSecretConfigQuery(appID);
  const { updateAppAndSecretConfig: updateConfig } =
    useUpdateAppAndSecretConfigMutation(appID);
  const [isUpdating, setIsUpdating] = useState(false);
  const [updateError, setUpdateError] = useState<unknown>(null);
  const [isSubmitted, setIsSubmitted] = useState(false);
  const [canSave, setCanSave] = useState<boolean | undefined>(initialCanSave);

  const effectiveConfig = useMemo(
    () => effectiveAppConfig ?? { id: appID },
    [effectiveAppConfig, appID]
  );

  const initialState = useMemo(
    () => constructFormState(effectiveConfig),
    [effectiveConfig, constructFormState]
  );
  const [currentState, setCurrentState, getCurrentState] = useLiveState<
    State | null
  >(
    constructInitialCurrentState != null
      ? constructInitialCurrentState(initialState)
      : null
  );

  const getIsDirty = useCallback(() => {
    const current = getCurrentState();
    if (!rawConfig || !current) {
      return false;
    }
    return !deepEqual(
      constructConfig(rawConfig, initialState, initialState, effectiveConfig),
      constructConfig(rawConfig, initialState, current, effectiveConfig),
      { strict: true }
    );
  }, [constructConfig, rawConfig, initialState, effectiveConfig, getCurrentState]);

  const reset = useCallback(() => {
    if (isUpdating) {
      return;
    }
    setUpdateError(null);
    setCurrentState(null);
    setIsSubmitted(false);
  }, [isUpdating, setCurrentState]);

  const performSave = useCallback(
    async (stateToSave: State, ignoreConflict: boolean) => {
      const err = validate?.(stateToSave);
      if (err) {
        setUpdateError(err);
        return;
      }

      const newConfig = constructConfig(
        rawConfig!,
        initialState,
        stateToSave,
        effectiveConfig
      );

      setIsUpdating(true);
      setUpdateError(null);
      try {
        await updateConfig({
          appConfig: newConfig,
          appConfigChecksum: rawAppConfigChecksum,
          ignoreConflict,
        });
        await reload();
        setCurrentState(null);
        setIsSubmitted(true);
      } catch (e: unknown) {
        setUpdateError(e);
        throw e;
      } finally {
        setIsUpdating(false);
      }
    },
    [
      validate,
      constructConfig,
      rawConfig,
      initialState,
      effectiveConfig,
      updateConfig,
      rawAppConfigChecksum,
      reload,
      setCurrentState,
    ]
  );

  const save = useCallback(
    async (ignoreConflict: boolean = false) => {
      const allowSave = canSave !== undefined ? canSave : getIsDirty();
      if (!rawConfig || !initialState || secretConfig == null) {
        return;
      } else if (!allowSave || isUpdating) {
        return;
      }
      await performSave(currentState ?? initialState, ignoreConflict);
    },
    [
      canSave,
      getIsDirty,
      rawConfig,
      initialState,
      secretConfig,
      isUpdating,
      currentState,
      performSave,
    ]
  );

  const saveWith = useCallback(
    // eslint-disable-next-line react-hooks/preserve-manual-memoization
    async (fn: (state: State) => State, ignoreConflict: boolean = false) => {
      if (!rawConfig || !initialState || secretConfig == null || isUpdating) {
        return;
      }
      const newState = fn(currentState ?? initialState);
      setCurrentState(newState);
      await performSave(newState, ignoreConflict);
    },
    [
      rawConfig,
      initialState,
      secretConfig,
      isUpdating,
      currentState,
      performSave,
      setCurrentState,
    ]
  );

  const state = currentState ?? initialState;
  const setState = useCallback(
    (fn: (state: State) => State) => {
      setCurrentState((s) => fn(s ?? initialState));
    },
    [initialState, setCurrentState]
  );

  return {
    isLoading,
    isUpdating,
    isSubmitted,
    loadError,
    updateError,
    canSave,
    setCanSave,
    initialState,
    state,
    setState,
    // eslint-disable-next-line @typescript-eslint/no-misused-promises, @typescript-eslint/strict-void-return
    reload,
    reset,
    save,
    saveWith,
    effectiveConfig,
    getIsDirty,
  };
}
