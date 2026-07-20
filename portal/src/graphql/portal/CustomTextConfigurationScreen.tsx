import React, {
  useCallback,
  useContext,
  useMemo,
  useRef,
  useState,
} from "react";
import { useParams } from "react-router-dom";
import { Select, Text } from "@radix-ui/themes";
import { Context as MFContext, FormattedMessage } from "../../intl";
import cn from "classnames";
import ExternalLink from "../../ExternalLink";

import { useAppAndSecretConfigQuery } from "./query/appAndSecretConfigQuery";
import {
  LanguageTag,
  Resource,
  ResourceDefinition,
  ResourceSpecifier,
  expandSpecifier,
  specifierId,
} from "../../util/resource";
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
import CodeEditor from "../../CodeEditor";
import { useSystemConfig } from "../../context/SystemConfigContext";
import { SaveFunctionBar } from "../../components/v2/SaveFunctionBar/SaveFunctionBar";
import { useFormContainerBaseContext } from "../../FormContainerBase";

import styles from "./CustomTextConfigurationScreen.module.css";

interface FormState extends ResourcesFormState {
  supportedLanguages: string[];
  fallbackLanguage: string;
  selectedLanguage: string;
}

interface FormModel {
  isLoading: boolean;
  isUpdating: boolean;
  getIsDirty: () => boolean;
  loadError: unknown;
  updateError: unknown;
  state: FormState;
  setState: (fn: (state: FormState) => FormState) => void;
  reload: () => void;
  reset: () => void;
  save: () => Promise<void>;
}

interface LanguageSelectOption {
  value: LanguageTag;
  label: string;
  disabled: boolean;
}

interface CustomTextConfigurationContentProps {
  form: FormModel;
  languageOptions: LanguageSelectOption[];
  onChangeSelectedLanguage: (language: LanguageTag) => void;
  translationValue: string;
  onChangeTranslation: (value: string | undefined, e: unknown) => void;
  gitCommitHash: string;
  translationSheetLanguage: LanguageTag;
}

const CustomTextConfigurationContent: React.VFC<CustomTextConfigurationContentProps> =
  function CustomTextConfigurationContent(props) {
    const {
      form,
      languageOptions,
      onChangeSelectedLanguage,
      translationValue,
      onChangeTranslation,
      gitCommitHash,
      translationSheetLanguage,
    } = props;
    const { isDirty } = useFormContainerBaseContext();
    const contentWidthAnchorRef = useRef<HTMLDivElement>(null);

    return (
      <ScreenContent
        className={cn(isDirty ? styles.contentWithSaveBar : null)}
      >
        <div
          ref={contentWidthAnchorRef}
          className={cn(styles.widget, styles.pageHeader)}
        >
          <Text as="p" size="5" weight="bold" className={styles.pageTitle}>
            <FormattedMessage id="CustomTextConfigurationScreen.title" />
          </Text>
          <div className={styles.headerMeta}>
            <Text
              as="p"
              size="2"
              color="gray"
              className={styles.pageDescription}
            >
              <FormattedMessage id="CustomTextConfigurationScreen.description" />
            </Text>
            <Select.Root
              value={form.state.selectedLanguage}
              onValueChange={onChangeSelectedLanguage}
            >
              <Select.Trigger
                variant="surface"
                className={styles.languageSelectTrigger}
              />
              <Select.Content
                position="popper"
                className={styles.languageSelectContent}
              >
                {languageOptions.map((option) => (
                  <Select.Item
                    key={option.value}
                    value={option.value}
                    disabled={option.disabled}
                  >
                    {option.label}
                  </Select.Item>
                ))}
              </Select.Content>
            </Select.Root>
          </div>
        </div>

        <div
          className={cn(
            styles.widget,
            styles.editorCard,
            isDirty && styles.settingsCardSaveBarClearance
          )}
        >
          <div className={styles.editorCardHeader}>
            <Text as="p" size="3" weight="medium" className={styles.editorCardTitle}>
              <FormattedMessage id="CustomTextConfigurationScreen.editor.title" />
            </Text>
            <Text
              as="p"
              size="2"
              color="gray"
              className={styles.editorCardDescription}
            >
              <FormattedMessage
                id="EditTemplatesWidget.translationjson.subtitle"
                values={{
                  COMMIT: gitCommitHash,
                  language: translationSheetLanguage,
                  // eslint-disable-next-line react/no-unstable-nested-components
                  externalLink: (chunks: React.ReactNode) => (
                    <ExternalLink
                      href={`https://github.com/authgear/authgear-server/blob/${gitCommitHash}/resources/authgear/templates/${translationSheetLanguage}/translation.json`}
                    >
                      {chunks}
                    </ExternalLink>
                  ),
                  // eslint-disable-next-line react/no-unstable-nested-components
                  docLink: (chunks: React.ReactNode) => (
                    <ExternalLink href="https://docs.authgear.com/customization/ui-customization/built-in-ui/localization">
                      {chunks}
                    </ExternalLink>
                  ),
                }}
              />
            </Text>
          </div>
          <CodeEditor
            key={form.state.selectedLanguage}
            className={styles.codeEditor}
            language="json"
            value={translationValue}
            onChange={onChangeTranslation}
          />
        </div>
        <SaveFunctionBar anchorRef={contentWidthAnchorRef} />
      </ScreenContent>
    );
  };

const CustomTextConfigurationScreen: React.VFC =
  function CustomTextConfigurationScreen() {
    const { appID } = useParams() as { appID: string };
    const { renderToString } = useContext(MFContext);
    const { gitCommitHash, builtinLanguages } = useSystemConfig();
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
      const specifiers: ResourceSpecifier[] = [];

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
        isLoading: config.isLoading || resourceForm.isLoading,
        isUpdating: resourceForm.isUpdating,
        getIsDirty: resourceForm.getIsDirty,
        loadError: config.loadError ?? resourceForm.loadError,
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

    const languageOptions = useMemo<LanguageSelectOption[]>(() => {
      const options: LanguageSelectOption[] = [];
      const combinedLocales = new Set([
        ...initialSupportedLanguages,
        ...form.state.supportedLanguages,
      ]);

      for (const locale of combinedLocales) {
        const isNew = !initialSupportedLanguages.includes(locale);
        const isRemoved = !form.state.supportedLanguages.includes(locale);

        let localeDisplay = renderToString(`Locales.${locale}`);
        if (isRemoved) {
          localeDisplay = renderToString("ManageLanguageWidget.option-removed", {
            LANG: localeDisplay,
          });
        }

        options.push({
          value: locale,
          label: renderToString("ManageLanguageWidget.language-label", {
            LANG: localeDisplay,
            IS_FALLBACK: String(form.state.fallbackLanguage === locale),
          }),
          disabled: isRemoved || isNew,
        });
      }

      return options;
    }, [
      initialSupportedLanguages,
      form.state.supportedLanguages,
      form.state.fallbackLanguage,
      renderToString,
    ]);

    const getValueFromState = useCallback(
      (
        resources: Partial<Record<string, Resource>>,
        selectedLanguage: string,
        fallbackLanguage: string,
        def: ResourceDefinition,
        getValueFn: (
          resource: Resource | undefined
        ) => string | undefined | null
      ): string | undefined | null => {
        const specifier: ResourceSpecifier = {
          def,
          locale: selectedLanguage,
          extension: null,
        };
        const value = getValueFn(resources[specifierId(specifier)]);

        if (value == null) {
          const specifier: ResourceSpecifier = {
            def,
            locale: fallbackLanguage,
            extension: null,
          };
          return getValueFn(resources[specifierId(specifier)]);
        }

        return value;
      },
      []
    );

    const translationValue = useMemo(() => {
      const selectedValue = getValueFromState(
        form.state.resources,
        form.state.selectedLanguage,
        form.state.fallbackLanguage,
        RESOURCE_TRANSLATION_JSON,
        (res) => res?.nullableValue ?? res?.effectiveData
      );
      if (selectedValue != null) {
        return selectedValue;
      }

      return (
        getValueFromState(
          form.state.resources,
          DEFAULT_TEMPLATE_LOCALE,
          form.state.fallbackLanguage,
          RESOURCE_TRANSLATION_JSON,
          (res) => res?.effectiveData
        ) ?? ""
      );
    }, [form.state, getValueFromState]);

    const onChangeTranslation = useCallback(
      (value: string | undefined, _e: unknown) => {
        const specifier: ResourceSpecifier = {
          def: RESOURCE_TRANSLATION_JSON,
          locale: form.state.selectedLanguage,
          extension: null,
        };
        form.setState((prev) => {
          const updatedResources = { ...prev.resources };
          const resource: Resource = {
            specifier,
            path: expandSpecifier(specifier),
            nullableValue: value ?? "",
            effectiveData:
              prev.resources[specifierId(specifier)]?.effectiveData,
          };
          updatedResources[specifierId(resource.specifier)] = resource;
          return { ...prev, resources: updatedResources };
        });
      },
      [form]
    );

    const translationSheetLanguage = useMemo(() => {
      if (builtinLanguages.includes(state.selectedLanguage)) {
        return state.selectedLanguage;
      }
      return state.fallbackLanguage;
    }, [builtinLanguages, state.fallbackLanguage, state.selectedLanguage]);

    return (
      <FormContainer form={form} canSave={true} hideFooterComponent={true}>
        <CustomTextConfigurationContent
          form={form}
          languageOptions={languageOptions}
          onChangeSelectedLanguage={setSelectedLanguage}
          translationValue={translationValue}
          onChangeTranslation={onChangeTranslation}
          gitCommitHash={gitCommitHash}
          translationSheetLanguage={translationSheetLanguage}
        />
      </FormContainer>
    );
  };

export default CustomTextConfigurationScreen;
