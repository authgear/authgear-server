import { useCallback, useMemo, useState } from "react";
import {
  Resource,
  ResourceSpecifier,
  ResourcesDiffResult,
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
  save: (ignoreConflict?: boolean) => Promise<void>;
  diff: ResourcesDiffResult | null;
}

export type StateConstructor<State> = (resources: Resource[]) => State;
export type ResourcesConstructor<State> = (state: State) => Resource[];

export interface ResourcesFormState {
  resources: Partial<Record<string, Resource>>;
}

function constructResourcesFormStateFromResources(
  resources: Resource[]
): ResourcesFormState {
  const resourceMap: Partial<Record<string, Resource>> = {};
  for (const r of resources) {
    const id = specifierId(r.specifier);
    // Multiple resources may use same specifier ID (images),
    // use the first resource with non-empty values.
    if ((resourceMap[id]?.nullableValue ?? "") === "") {
      resourceMap[specifierId(r.specifier)] = r;
    }
  }

  return { resources: resourceMap };
}

function constructResourcesFromResourcesFormState(
  state: ResourcesFormState
): Resource[] {
  return Object.values(state.resources).filter(Boolean) as Resource[];
}
export function useResourceForm(
  appID: string,
  specifiers: ResourceSpecifier[]
): ResourceFormModel<ResourcesFormState>;
export function useResourceForm<State>(
  appID: string,
  specifiers: ResourceSpecifier[],
  constructState: StateConstructor<State>,
  constructResources: ResourcesConstructor<State>
): ResourceFormModel<State>;
export function useResourceForm<State>(
  appID: string,
  specifiers: ResourceSpecifier[],
  constructState: StateConstructor<any> = constructResourcesFormStateFromResources,
  constructResources: ResourcesConstructor<any> = constructResourcesFromResourcesFormState
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

  const initialState = useMemo(
    () => constructState(resources),
    [resources, constructState]
  );
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

  const save = useCallback(
    async (ignoreConflict: boolean = false) => {
      if (!diff) {
        return;
      } else if (!diff.needUpdate) {
        // In the case that a builtin language is added,
        // and no changes have been made to the existing resources,
        // we need to reset current state to null to make
        // state consistent.
        setCurrentState(null);
        return;
      } else if (isUpdating) {
        return;
      }

      try {
        await updateResources(
          [
            ...diff.newResources,
            ...diff.editedResources,
            ...diff.deletedResources,
          ],
          ignoreConflict
        );
        setCurrentState(null);
      } finally {
      }
    },
    [diff, isUpdating, updateResources]
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
    loadError,
    updateError,
    state,
    setState,
    // eslint-disable-next-line @typescript-eslint/no-misused-promises
    reload,
    reset,
    save,
    diff,
  };
}
