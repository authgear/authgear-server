import { useCallback, useMemo, useState } from "react";
import deepEqual from "deep-equal";
import { useAppAndSecretConfigQuery } from "../graphql/portal/query/appAndSecretConfigQuery";
import { useUpdateAppAndSecretConfigMutation } from "../graphql/portal/mutations/updateAppAndSecretMutation";
import { PortalAPIAppConfig } from "../types";
import { APIError } from "../error/error";

export interface AppConfigFormModel<State> {
  isLoading: boolean;
  isUpdating: boolean;
  isDirty: boolean;
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
  setCanSave: (canSave?: boolean) => void;
  effectiveConfig: PortalAPIAppConfig;
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
    loading: isLoading,
    error: loadError,
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
  const [currentState, setCurrentState] = useState<State | null>(
    constructInitialCurrentState != null
      ? constructInitialCurrentState(initialState)
      : null
  );

  const isDirty = useMemo(() => {
    if (!rawConfig || !currentState) {
      return false;
    }
    return !deepEqual(
      constructConfig(rawConfig, initialState, initialState, effectiveConfig),
      constructConfig(rawConfig, initialState, currentState, effectiveConfig),
      { strict: true }
    );
  }, [constructConfig, rawConfig, initialState, currentState, effectiveConfig]);

  const reset = useCallback(() => {
    if (isUpdating) {
      return;
    }
    setUpdateError(null);
    setCurrentState(null);
    setIsSubmitted(false);
  }, [isUpdating]);

  const save = useCallback(
    // eslint-disable-next-line complexity
    async (ignoreConflict: boolean = false) => {
      const allowSave = canSave !== undefined ? canSave : isDirty;
      if (!rawConfig || !initialState || secretConfig == null) {
        return;
      } else if (!allowSave || isUpdating) {
        return;
      }

      const err = validate?.(currentState ?? initialState);
      if (err) {
        setUpdateError(err);
        return;
      }

      const newConfig = constructConfig(
        rawConfig,
        initialState,
        currentState ?? initialState,
        effectiveConfig
      );

      setIsUpdating(true);
      setUpdateError(null);
      try {
        await updateConfig(
          newConfig,
          rawAppConfigChecksum,
          undefined,
          undefined,
          ignoreConflict
        );
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
      isDirty,
      isUpdating,
      constructConfig,
      rawConfig,
      rawAppConfigChecksum,
      effectiveConfig,
      initialState,
      currentState,
      updateConfig,
      secretConfig,
      validate,
      canSave,
    ]
  );

  const state = currentState ?? initialState;
  const setState = useCallback(
    (fn: (state: State) => State) => {
      setCurrentState((s) => fn(s ?? initialState));
    },
    [initialState]
  );

  return {
    isLoading,
    isUpdating,
    isDirty,
    isSubmitted,
    loadError,
    updateError,
    canSave,
    setCanSave,
    initialState,
    state,
    setState,
    // eslint-disable-next-line @typescript-eslint/no-misused-promises
    reload,
    reset,
    save,
    effectiveConfig,
  };
}
