import React, { useMemo, useState } from "react";
import { useParams } from "react-router-dom";
import { FormattedMessage } from "@oursky/react-messageformat";
import cn from "classnames";

import { useAppAndSecretConfigQuery } from "./query/appAndSecretConfigQuery";
import { LanguageTag, ResourceSpecifier } from "../../util/resource";
import {
  DEFAULT_TEMPLATE_LOCALE,
  RESOURCE_TRANSLATION_JSON,
} from "../../resources";
import {
  ResourcesFormState,
  useResourceForm,
} from "../../hook/useResourceForm";
import FormContainer from "../../FormContainer";
import ScreenContent from "../../ScreenContent";
import ScreenTitle from "../../ScreenTitle";
import ManageLanguageWidget from "./ManageLanguageWidget";

interface FormState extends ResourcesFormState {
  supportedLanguages: string[];
  fallbackLanguage: string;
  selectedLanguage: string;
}

interface FormModel {
  isLoading: boolean;
  isUpdating: boolean;
  isDirty: boolean;
  loadError: unknown;
  updateError: unknown;
  state: FormState;
  setState: (fn: (state: FormState) => FormState) => void;
  reload: () => void;
  reset: () => void;
  save: () => Promise<void>;
}

const CustomTextConfigurationScreen: React.VFC =
  function CustomTextConfigurationScreen() {
    const { appID } = useParams() as { appID: string };
    const config = useAppAndSecretConfigQuery(appID);

    const initialSupportedLanguages = useMemo(() => {
      return (
        config.effectiveAppConfig?.localization?.supported_languages ?? [
          config.effectiveAppConfig?.localization?.fallback_language ??
            DEFAULT_TEMPLATE_LOCALE,
        ]
      );
    }, [config.effectiveAppConfig?.localization]);

    const specifiers = useMemo<ResourceSpecifier[]>(() => {
      const specifiers = [];

      const supportedLanguages = [...initialSupportedLanguages];
      if (!supportedLanguages.includes(DEFAULT_TEMPLATE_LOCALE)) {
        supportedLanguages.push(DEFAULT_TEMPLATE_LOCALE);
      }

      for (const locale of supportedLanguages) {
        specifiers.push({
          def: RESOURCE_TRANSLATION_JSON,
          locale,
          extension: null,
        });
      }
      return specifiers;
    }, [initialSupportedLanguages]);

    const resourceForm = useResourceForm(appID, specifiers);

    const [selectedLanguage, setSelectedLanguage] =
      useState<LanguageTag | null>(null);
    const state = useMemo<FormState>(() => {
      const fallbackLanguage =
        config.effectiveAppConfig?.localization?.fallback_language ??
        DEFAULT_TEMPLATE_LOCALE;
      return {
        supportedLanguages: config.effectiveAppConfig?.localization
          ?.supported_languages ?? [fallbackLanguage],
        fallbackLanguage,
        resources: resourceForm.state.resources,
        selectedLanguage: selectedLanguage ?? fallbackLanguage,
      };
    }, [
      config.effectiveAppConfig?.localization,
      resourceForm.state.resources,
      selectedLanguage,
    ]);

    const form: FormModel = useMemo(
      () => ({
        isLoading: config.loading || resourceForm.isLoading,
        isUpdating: resourceForm.isUpdating,
        isDirty: resourceForm.isDirty,
        loadError: config.error ?? resourceForm.loadError,
        updateError: resourceForm.updateError,
        state,
        setState: (fn) => {
          const newState = fn(state);
          resourceForm.setState(() => ({ resources: newState.resources }));
          setSelectedLanguage(newState.selectedLanguage);
        },
        reload: () => {
          // Previously is also a floating promise, so just log the error out
          // to make linter happy
          config.refetch().catch((err) => {
            console.error("Reload config error", err);
            throw err;
          });
          resourceForm.reload();
        },
        reset: () => {
          resourceForm.reset();
          setSelectedLanguage(state.fallbackLanguage);
        },
        save: async (ignoreConflict: boolean = false) => {
          await resourceForm.save(ignoreConflict);
        },
      }),
      [config, resourceForm, state]
    );

    return (
      <FormContainer form={form} canSave={true}>
        <ScreenContent>
          <ScreenTitle className={cn("col-span-8", "tablet:col-span-full")}>
            <FormattedMessage id="CustomTextConfigurationScreen.title" />
          </ScreenTitle>
          <div
            className={cn(
              "pt-1",
              "col-span-8",
              "tablet:col-span-full",
              "flex",
              "justify-end"
            )}
          >
            <ManageLanguageWidget
              showLabel={false}
              existingLanguages={initialSupportedLanguages}
              supportedLanguages={initialSupportedLanguages}
              selectedLanguage={state.selectedLanguage}
              fallbackLanguage={state.fallbackLanguage}
              onChangeSelectedLanguage={setSelectedLanguage}
            />
          </div>
        </ScreenContent>
      </FormContainer>
    );
  };

export default CustomTextConfigurationScreen;
