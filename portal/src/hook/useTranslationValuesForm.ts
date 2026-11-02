import { useCallback, useMemo } from "react";
import { produce } from "immer";
import { RESOURCE_TRANSLATION_JSON } from "../resources";
import {
  ResourceSpecifier,
  expandSpecifier,
  specifierId,
} from "../util/resource";
import {
  escapeMessageFormatText,
  unescapeMessageFormatText,
} from "../util/messageFormat";
import { FormModel } from "../FormContainerBase";
import { AppConfigFormModel } from "./useAppConfigForm";
import {
  ResourceFormModel,
  ResourcesFormState,
  useResourceForm,
} from "./useResourceForm";

export interface TranslationValuesForm {
  form: ResourceFormModel<ResourcesFormState>;
  // The value of a key in the language's translation.json; "" when unset.
  getValue: (key: string) => string;
  // Sets a key; "" removes it.
  setValue: (key: string, value: string) => void;
}

// Plain-text values of a language's translation.json, the file Custom Text
// edits. language is null until the project's config has loaded.
export function useTranslationValuesForm(
  appID: string,
  language: string | null
): TranslationValuesForm {
  const specifier = useMemo<ResourceSpecifier | null>(
    () =>
      language != null
        ? { def: RESOURCE_TRANSLATION_JSON, locale: language, extension: null }
        : null,
    [language]
  );
  const specifiers = useMemo(
    () => (specifier != null ? [specifier] : []),
    [specifier]
  );
  const form = useResourceForm(
    specifier != null ? appID : undefined,
    specifiers
  );
  const { state, setState } = form;

  const read = useCallback(
    (s: ResourcesFormState): Record<string, unknown> => {
      if (specifier == null) {
        return {};
      }
      const value = s.resources[specifierId(specifier)]?.nullableValue;
      return value != null && value !== ""
        ? (JSON.parse(value) as Record<string, unknown>)
        : {};
    },
    [specifier]
  );

  const getValue = useCallback(
    (key: string): string => {
      const value = read(state)[key];
      return typeof value === "string" ? unescapeMessageFormatText(value) : "";
    },
    [read, state]
  );

  const setValue = useCallback(
    (key: string, value: string) => {
      if (specifier == null) {
        return;
      }
      setState((s) =>
        produce(s, (draft) => {
          const json = read(s);
          if (value === "") {
            delete json[key];
          } else {
            json[key] = escapeMessageFormatText(value);
          }
          draft.resources[specifierId(specifier)] = {
            specifier,
            path: expandSpecifier(specifier),
            nullableValue: JSON.stringify(json, null, 2),
          };
        })
      );
    },
    [specifier, read, setState]
  );

  return { form, getValue, setValue };
}

// A config form and a resource form saved as one. With config changes, both
// go in one request, so neither is saved without the other.
export function combineForms<State>(
  configForm: AppConfigFormModel<State>,
  resourceForm: ResourceFormModel<ResourcesFormState>
): AppConfigFormModel<State> & FormModel {
  return {
    ...configForm,
    isLoading: configForm.isLoading || resourceForm.isLoading,
    loadError: configForm.loadError ?? resourceForm.loadError,
    updateError: configForm.updateError ?? resourceForm.updateError,
    isUpdating: configForm.isUpdating || resourceForm.isUpdating,
    getIsDirty: () => configForm.getIsDirty() || resourceForm.getIsDirty(),
    reload: () => {
      configForm.reload();
      resourceForm.reload();
    },
    reset: () => {
      configForm.reset();
      resourceForm.reset();
    },
    save: async (ignoreConflict?: boolean) => {
      if (!configForm.getIsDirty()) {
        await resourceForm.save(ignoreConflict);
        return;
      }
      const diff = resourceForm.diff;
      const resourceUpdates = diff?.needUpdate
        ? [
            ...diff.newResources,
            ...diff.editedResources,
            ...diff.deletedResources,
          ]
        : undefined;
      await configForm.save(ignoreConflict, resourceUpdates);
      // The files were saved with the config, and refetched before save
      // resolved. A save that did not happen leaves the config dirty.
      if (resourceUpdates != null && !configForm.getIsDirty()) {
        resourceForm.reset();
      }
    },
  };
}
