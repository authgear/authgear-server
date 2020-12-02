import { useCallback, useMemo, useState } from "react";
import deepEqual from "deep-equal";
import {
  Resource,
  ResourceSpecifier,
  diffResourceUpdates,
  specifierId,
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

function cmp(a: Resource, b: Resource) {
  const ai = specifierId(a.specifier);
  const bi = specifierId(b.specifier);
  return ai === bi ? 0 : ai < bi ? -1 : 1;
}

function mergeResources(
  initialResources: Resource[],
  newResources: Resource[]
): Resource[] {
  const resources = new Map(
    initialResources.map((r) => [specifierId(r.specifier), r])
  );
  for (const r of newResources) {
    resources.set(specifierId(r.specifier), r);
  }
  return Array.from(resources.values()).sort(cmp);
}

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

  const isDirty = useMemo(() => {
    if (!currentState) {
      return false;
    }
    return !deepEqual(
      mergeResources(resources, constructResources(initialState)),
      mergeResources(resources, constructResources(currentState)),
      { strict: true }
    );
  }, [constructResources, resources, initialState, currentState]);

  const reset = useCallback(() => {
    if (isUpdating) {
      return;
    }
    resetError();
    setCurrentState(null);
  }, [isUpdating, resetError]);

  const save = useCallback(() => {
    if (!currentState) {
      return;
    } else if (isUpdating) {
      return;
    }

    const newResources = mergeResources(
      resources,
      constructResources(currentState)
    );
    const diff = diffResourceUpdates(resources, newResources);
    if (!diff.needUpdate) {
      setCurrentState(null);
      return;
    }

    updateResources(specifiers, [
      ...diff.newResources,
      ...diff.editedResources,
      ...diff.deletedResources,
    ])
      .then(() => setCurrentState(null))
      .catch(() => {});
  }, [
    isUpdating,
    constructResources,
    specifiers,
    resources,
    currentState,
    updateResources,
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
