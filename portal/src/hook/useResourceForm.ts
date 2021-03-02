import { useCallback, useMemo, useState } from "react";
import {
  Resource,
  ResourceSpecifier,
  ResourcesDiffResult,
  diffResourceUpdates,
} from "../util/resource";
import { useAppTemplatesQuery } from "../graphql/portal/query/appTemplatesQuery";
import { useUpdateAppTemplatesMutation } from "../graphql/portal/mutations/updateAppTemplatesMutation";

export interface ResourceFormModel<State> {
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

export type StateConstructor<State> = (resources: Resource[]) => State;
export type ResourcesConstructor<State> = (state: State) => Resource[];

export function useResourceForm<State>(
  appID: string,
  specifiers: ResourceSpecifier[],
  constructState: StateConstructor<State>,
  constructResources: ResourcesConstructor<State>
): ResourceFormModel<State> {
  const {
    resources,
    loading: isLoading,
    error: loadError,
    refetch: reload,
  } = useAppTemplatesQuery(appID, specifiers);
  const {
    loading: isUpdating,
    error: updateError,
    updateAppTemplates: updateResources,
    resetError,
  } = useUpdateAppTemplatesMutation(appID);

  const initialState = useMemo(() => constructState(resources), [
    resources,
    constructState,
  ]);
  const [currentState, setCurrentState] = useState<State | null>(null);

  const newResources: Resource[] | null = useMemo(() => {
    if (!currentState) {
      return null;
    }
    return constructResources(currentState);
  }, [currentState, constructResources]);

  const diff: ResourcesDiffResult | null = useMemo(() => {
    if (newResources == null) {
      return null;
    }
    return diffResourceUpdates(resources, newResources);
  }, [resources, newResources]);

  const isDirty = useMemo(() => {
    if (diff == null) {
      return false;
    }
    return diff.needUpdate;
  }, [diff]);

  const reset = useCallback(() => {
    if (isUpdating) {
      return;
    }
    resetError();
    setCurrentState(null);
  }, [isUpdating, resetError]);

  const save = useCallback(() => {
    if (!diff) {
      return;
    } else if (!diff.needUpdate) {
      return;
    } else if (isUpdating) {
      return;
    }

    updateResources([
      ...diff.newResources,
      ...diff.editedResources,
      ...diff.deletedResources,
    ])
      .then(() => setCurrentState(null))
      .catch(() => {});
  }, [diff, isUpdating, updateResources]);

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
