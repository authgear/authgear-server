import React, { useCallback, useContext, useMemo, useState } from "react";
import { useParams } from "react-router-dom";
import { Pivot, PivotItem } from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import { produce } from "immer";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import ScreenContent from "../../ScreenContent";
import ScreenTitle from "../../ScreenTitle";
import ScreenDescription from "../../ScreenDescription";
import WidgetTitle from "../../WidgetTitle";
import Widget from "../../Widget";
import ManageLanguageWidget from "./ManageLanguageWidget";
import EditTemplatesWidget, {
  EditTemplatesWidgetSection,
} from "./EditTemplatesWidget";
import { PortalAPIAppConfig } from "../../types";
import {
  ALL_LANGUAGES_TEMPLATES,
  RESOURCE_AUTHENTICATE_PRIMARY_OOB_EMAIL_HTML,
  RESOURCE_AUTHENTICATE_PRIMARY_OOB_EMAIL_TXT,
  RESOURCE_AUTHENTICATE_PRIMARY_OOB_SMS_TXT,
  RESOURCE_FORGOT_PASSWORD_EMAIL_HTML,
  RESOURCE_FORGOT_PASSWORD_EMAIL_TXT,
  RESOURCE_FORGOT_PASSWORD_SMS_TXT,
  RESOURCE_SETUP_PRIMARY_OOB_EMAIL_HTML,
  RESOURCE_SETUP_PRIMARY_OOB_EMAIL_TXT,
  RESOURCE_SETUP_PRIMARY_OOB_SMS_TXT,
  RESOURCE_TRANSLATION_JSON,
  TRANSLATION_JSON_KEY_EMAIL_AUTHENTICATE_PRIMARY_OOB_SUBJECT,
  TRANSLATION_JSON_KEY_EMAIL_FORGOT_PASSWORD_SUBJECT,
  TRANSLATION_JSON_KEY_EMAIL_SETUP_PRIMARY_OOB_SUBJECT,
} from "../../resources";
import {
  LanguageTag,
  Resource,
  ResourceDefinition,
  ResourceSpecifier,
  specifierId,
  expandSpecifier,
} from "../../util/resource";
import { useAppConfigForm } from "../../hook/useAppConfigForm";
import { clearEmptyObject } from "../../util/misc";
import { useResourceForm } from "../../hook/useResourceForm";
import FormContainer from "../../FormContainer";
import { useSystemConfig } from "../../context/SystemConfigContext";
import styles from "./LocalizationConfigurationScreen.module.css";
import cn from "classnames";

interface ConfigFormState {
  supportedLanguages: string[];
  fallbackLanguage: string;
}

function constructFormState(config: PortalAPIAppConfig): ConfigFormState {
  const fallbackLanguage = config.localization?.fallback_language ?? "en";
  return {
    fallbackLanguage,
    supportedLanguages: config.localization?.supported_languages ?? [
      fallbackLanguage,
    ],
  };
}

function constructConfig(
  config: PortalAPIAppConfig,
  _initialState: ConfigFormState,
  currentState: ConfigFormState
): PortalAPIAppConfig {
  return produce(config, (config) => {
    config.localization = config.localization ?? {};
    config.localization.fallback_language = currentState.fallbackLanguage;
    config.localization.supported_languages = currentState.supportedLanguages;
    clearEmptyObject(config);
  });
}

interface ResourcesFormState {
  resources: Partial<Record<string, Resource>>;
}

function constructResourcesFormState(
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

function constructResources(state: ResourcesFormState): Resource[] {
  return Object.values(state.resources).filter(Boolean) as Resource[];
}

interface FormState extends ConfigFormState, ResourcesFormState {
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

interface ResourcesConfigurationContentProps {
  form: FormModel;
  supportedLanguages: LanguageTag[];
  passwordlessViaEmailEnabled: boolean;
  passwordlessViaSMSEnabled: boolean;
}

const PIVOT_KEY_FORGOT_PASSWORD = "forgot_password";
const PIVOT_KEY_PASSWORDLESS_VIA_EMAIL = "passwordless_via_email";
const PIVOT_KEY_PASSWORDLESS_VIA_SMS = "passwordless_via_sms";
const PIVOT_KEY_TRANSLATION_JSON = "translation.json";

const PIVOT_KEY_DEFAULT = PIVOT_KEY_FORGOT_PASSWORD;

const ALL_PIVOT_KEYS = [
  PIVOT_KEY_FORGOT_PASSWORD,
  PIVOT_KEY_PASSWORDLESS_VIA_EMAIL,
  PIVOT_KEY_PASSWORDLESS_VIA_SMS,
  PIVOT_KEY_TRANSLATION_JSON,
];

const ResourcesConfigurationContent: React.FC<ResourcesConfigurationContentProps> =
  function ResourcesConfigurationContent(props) {
    const { state, setState } = props.form;
    const {
      supportedLanguages,
      passwordlessViaEmailEnabled,
      passwordlessViaSMSEnabled,
    } = props;
    const { renderToString } = useContext(Context);
    const { gitCommitHash } = useSystemConfig();

    const setSelectedLanguage = useCallback(
      (selectedLanguage: LanguageTag) => {
        setState((s) => ({ ...s, selectedLanguage }));
      },
      [setState]
    );

    const onChangeLanguages = useCallback(
      (supportedLanguages: LanguageTag[], fallbackLanguage: LanguageTag) => {
        setState((prev) => {
          // Reset selected language to fallback language if it was removed.
          let { selectedLanguage, resources } = prev;
          resources = { ...resources };
          if (!supportedLanguages.includes(selectedLanguage)) {
            selectedLanguage = fallbackLanguage;
          }

          // Remove resources of removed languges
          const removedLanguages = prev.supportedLanguages.filter(
            (l) => !supportedLanguages.includes(l)
          );
          for (const [id, resource] of Object.entries(resources)) {
            const language = resource?.specifier.locale;
            if (
              resource != null &&
              language != null &&
              removedLanguages.includes(language)
            ) {
              resources[id] = { ...resource, nullableValue: "" };
            }
          }

          return {
            ...prev,
            selectedLanguage,
            supportedLanguages,
            fallbackLanguage,
            resources,
          };
        });
      },
      [setState]
    );

    const [selectedKey, setSelectedKey] = useState<string>(PIVOT_KEY_DEFAULT);
    const onLinkClick = useCallback((item?: PivotItem) => {
      const itemKey = item?.props.itemKey;
      if (itemKey != null) {
        const idx = ALL_PIVOT_KEYS.indexOf(itemKey);
        if (idx >= 0) {
          setSelectedKey(itemKey);
        }
      }
    }, []);

    const getValueFromState = useCallback(
      (
        resources: Partial<Record<string, Resource>>,
        selectedLanguage: string,
        fallbackLanguage: string,
        def: ResourceDefinition,
        getValueFn: (
          resource: Resource | undefined
        ) => string | undefined | null
      ) => {
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
          return getValueFn(resources[specifierId(specifier)]) ?? "";
        }

        return value;
      },
      []
    );

    const getValue = useCallback(
      (def: ResourceDefinition) => {
        return getValueFromState(
          state.resources,
          state.selectedLanguage,
          state.fallbackLanguage,
          def,
          (res) => res?.nullableValue
        );
      },
      [
        state.resources,
        state.selectedLanguage,
        state.fallbackLanguage,
        getValueFromState,
      ]
    );

    const getOnChange = useCallback(
      (def: ResourceDefinition) => {
        const specifier: ResourceSpecifier = {
          def,
          locale: state.selectedLanguage,
          extension: null,
        };
        return (value: string | undefined, _e: unknown) => {
          setState((prev) => {
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
        };
      },
      [state.selectedLanguage, setState]
    );

    const getTranslationValue = useCallback(
      (key: string) => {
        // get from the translation json first
        const translationJSONStr = getValueFromState(
          state.resources,
          state.selectedLanguage,
          state.fallbackLanguage,
          RESOURCE_TRANSLATION_JSON,
          (res) => res?.nullableValue
        );
        try {
          const translationJSON = JSON.parse(translationJSONStr);
          if (translationJSON[key] != null) {
            return translationJSON[key];
          }
        } catch (_e: unknown) {
          // if failed to decode the translation.json, use the effective data
        }
        // fallback to the effective data
        const effTranslationJSONStr = getValueFromState(
          state.resources,
          state.selectedLanguage,
          state.fallbackLanguage,
          RESOURCE_TRANSLATION_JSON,
          (res) => res?.effectiveData
        );
        const jsonValue = JSON.parse(effTranslationJSONStr);
        return jsonValue[key] ?? "";
      },
      [
        state.resources,
        state.selectedLanguage,
        state.fallbackLanguage,
        getValueFromState,
      ]
    );

    const getTranslationOnChange = useCallback(
      (key: string) => {
        const specifier: ResourceSpecifier = {
          def: RESOURCE_TRANSLATION_JSON,
          locale: state.selectedLanguage,
          extension: null,
        };
        return (value: string | undefined, _e: unknown) => {
          setState((prev) => {
            // get the translation JSON, decode and alter
            const translationJSONStr = getValueFromState(
              prev.resources,
              prev.selectedLanguage,
              prev.fallbackLanguage,
              RESOURCE_TRANSLATION_JSON,
              (res) => res?.nullableValue
            );

            let resultTranslationJSON;
            try {
              const translationJSON = JSON.parse(translationJSONStr);
              if (value) {
                translationJSON[key] = value;
              } else {
                delete translationJSON[key];
              }
              resultTranslationJSON = JSON.stringify(translationJSON, null, 2);
            } catch (error: unknown) {
              // if failed to decode the translation.json, don't update it
              console.error(error);
              return prev;
            }

            // get the translation JSON effective data, decode and alter
            const effTranslationJSONStr = getValueFromState(
              prev.resources,
              prev.selectedLanguage,
              prev.fallbackLanguage,
              RESOURCE_TRANSLATION_JSON,
              (res) => res?.effectiveData
            );
            const effTranslationJSON = JSON.parse(effTranslationJSONStr);
            if (value) {
              effTranslationJSON[key] = value;
            } else {
              delete effTranslationJSON[key];
            }
            const resultEffTranslationJSON = JSON.stringify(
              effTranslationJSON,
              null,
              2
            );

            // When the value is updated, both value and effective data of
            // translation.json need to be updated
            // Otherwise there will be inconsistent in the ui
            const updatedResources = { ...prev.resources };
            const resource: Resource = {
              specifier,
              path: expandSpecifier(specifier),
              nullableValue: resultTranslationJSON,
              effectiveData: resultEffTranslationJSON,
            };
            updatedResources[specifierId(resource.specifier)] = resource;
            return { ...prev, resources: updatedResources };
          });
        };
      },
      [setState, getValueFromState, state.selectedLanguage]
    );

    const sectionsTranslationJSON: EditTemplatesWidgetSection[] = [
      {
        key: "translation.json",
        title: (
          <FormattedMessage id="EditTemplatesWidget.translationjson.title" />
        ),
        items: [
          {
            key: "translation.json",
            title: (
              <FormattedMessage
                id="EditTemplatesWidget.translationjson.subtitle"
                values={{
                  COMMIT: gitCommitHash,
                }}
              />
            ),
            language: "json",
            value: getValue(RESOURCE_TRANSLATION_JSON),
            onChange: getOnChange(RESOURCE_TRANSLATION_JSON),
            editor: "code",
          },
        ],
      },
    ];

    const sectionsForgotPassword: EditTemplatesWidgetSection[] = [
      {
        key: "email",
        title: <FormattedMessage id="EditTemplatesWidget.email" />,
        items: [
          {
            key: "email-subject",
            title: <FormattedMessage id="EditTemplatesWidget.email-subject" />,
            language: "plaintext",
            value: getTranslationValue(
              TRANSLATION_JSON_KEY_EMAIL_FORGOT_PASSWORD_SUBJECT
            ),
            onChange: getTranslationOnChange(
              TRANSLATION_JSON_KEY_EMAIL_FORGOT_PASSWORD_SUBJECT
            ),
            editor: "textfield",
          },
          {
            key: "html-email",
            title: <FormattedMessage id="EditTemplatesWidget.html-email" />,
            language: "html",
            value: getValue(RESOURCE_FORGOT_PASSWORD_EMAIL_HTML),
            onChange: getOnChange(RESOURCE_FORGOT_PASSWORD_EMAIL_HTML),
            editor: "code",
          },
          {
            key: "plaintext-email",
            title: (
              <FormattedMessage id="EditTemplatesWidget.plaintext-email" />
            ),
            language: "plaintext",
            value: getValue(RESOURCE_FORGOT_PASSWORD_EMAIL_TXT),
            onChange: getOnChange(RESOURCE_FORGOT_PASSWORD_EMAIL_TXT),
            editor: "code",
          },
        ],
      },
      {
        key: "sms",
        title: <FormattedMessage id="EditTemplatesWidget.sms" />,
        items: [
          {
            key: "sms",
            title: <FormattedMessage id="EditTemplatesWidget.sms-body" />,
            language: "plaintext",
            value: getValue(RESOURCE_FORGOT_PASSWORD_SMS_TXT),
            onChange: getOnChange(RESOURCE_FORGOT_PASSWORD_SMS_TXT),
            editor: "code",
          },
        ],
      },
    ];

    const sectionsPasswordlessViaEmail: EditTemplatesWidgetSection[] = [
      {
        key: "setup",
        title: (
          <FormattedMessage id="EditTemplatesWidget.passwordless.setup.title" />
        ),
        items: [
          {
            key: "email-subject",
            title: <FormattedMessage id="EditTemplatesWidget.email-subject" />,
            language: "plaintext",
            value: getTranslationValue(
              TRANSLATION_JSON_KEY_EMAIL_SETUP_PRIMARY_OOB_SUBJECT
            ),
            onChange: getTranslationOnChange(
              TRANSLATION_JSON_KEY_EMAIL_SETUP_PRIMARY_OOB_SUBJECT
            ),
            editor: "textfield",
          },
          {
            key: "html-email",
            title: <FormattedMessage id="EditTemplatesWidget.html-email" />,
            language: "html",
            value: getValue(RESOURCE_SETUP_PRIMARY_OOB_EMAIL_HTML),
            onChange: getOnChange(RESOURCE_SETUP_PRIMARY_OOB_EMAIL_HTML),
            editor: "code",
          },
          {
            key: "plaintext-email",
            title: (
              <FormattedMessage id="EditTemplatesWidget.plaintext-email" />
            ),
            language: "plaintext",
            value: getValue(RESOURCE_SETUP_PRIMARY_OOB_EMAIL_TXT),
            onChange: getOnChange(RESOURCE_SETUP_PRIMARY_OOB_EMAIL_TXT),
            editor: "code",
          },
        ],
      },
      {
        key: "login",
        title: (
          <FormattedMessage id="EditTemplatesWidget.passwordless.login.title" />
        ),
        items: [
          {
            key: "email-subject",
            title: <FormattedMessage id="EditTemplatesWidget.email-subject" />,
            language: "plaintext",
            value: getTranslationValue(
              TRANSLATION_JSON_KEY_EMAIL_AUTHENTICATE_PRIMARY_OOB_SUBJECT
            ),
            onChange: getTranslationOnChange(
              TRANSLATION_JSON_KEY_EMAIL_AUTHENTICATE_PRIMARY_OOB_SUBJECT
            ),
            editor: "textfield",
          },
          {
            key: "html-email",
            title: <FormattedMessage id="EditTemplatesWidget.html-email" />,
            language: "html",
            value: getValue(RESOURCE_AUTHENTICATE_PRIMARY_OOB_EMAIL_HTML),
            onChange: getOnChange(RESOURCE_AUTHENTICATE_PRIMARY_OOB_EMAIL_HTML),
            editor: "code",
          },
          {
            key: "plaintext-email",
            title: (
              <FormattedMessage id="EditTemplatesWidget.plaintext-email" />
            ),
            language: "plaintext",
            value: getValue(RESOURCE_AUTHENTICATE_PRIMARY_OOB_EMAIL_TXT),
            onChange: getOnChange(RESOURCE_AUTHENTICATE_PRIMARY_OOB_EMAIL_TXT),
            editor: "code",
          },
        ],
      },
    ];

    const sectionsPasswordlessViaSMS: EditTemplatesWidgetSection[] = [
      {
        key: "setup",
        title: (
          <FormattedMessage id="EditTemplatesWidget.passwordless.setup.title" />
        ),
        items: [
          {
            key: "sms",
            title: <FormattedMessage id="EditTemplatesWidget.sms-body" />,
            language: "plaintext",
            value: getValue(RESOURCE_SETUP_PRIMARY_OOB_SMS_TXT),
            onChange: getOnChange(RESOURCE_SETUP_PRIMARY_OOB_SMS_TXT),
            editor: "code",
          },
        ],
      },
      {
        key: "login",
        title: (
          <FormattedMessage id="EditTemplatesWidget.passwordless.login.title" />
        ),
        items: [
          {
            key: "sms",
            title: <FormattedMessage id="EditTemplatesWidget.sms-body" />,
            language: "plaintext",
            value: getValue(RESOURCE_AUTHENTICATE_PRIMARY_OOB_SMS_TXT),
            onChange: getOnChange(RESOURCE_AUTHENTICATE_PRIMARY_OOB_SMS_TXT),
            editor: "code",
          },
        ],
      },
    ];

    return (
      <ScreenContent>
        <div
          className={cn(
            styles.titleContainer,
            "mobile:flex-col mobile:items-stretch mobile:col-span-full mobile:gap-y-5"
          )}
        >
          <ScreenTitle className="mobile:w-full">
            <FormattedMessage id="LocalizationConfigurationScreen.title" />
          </ScreenTitle>
          <ManageLanguageWidget
            className="mobile:w-full"
            supportedLanguages={supportedLanguages}
            selectedLanguage={state.selectedLanguage}
            onChangeSelectedLanguage={setSelectedLanguage}
            fallbackLanguage={state.fallbackLanguage}
            onChangeLanguages={onChangeLanguages}
          />
        </div>
        <ScreenDescription className={styles.widget}>
          <FormattedMessage id="LocalizationConfigurationScreen.description" />
        </ScreenDescription>
        <Widget className={styles.widget}>
          <WidgetTitle>
            <FormattedMessage id="LocalizationConfigurationScreen.template-content-title" />
          </WidgetTitle>
          <Pivot
            overflowBehavior="menu"
            onLinkClick={onLinkClick}
            selectedKey={selectedKey}
          >
            <PivotItem
              headerText={renderToString(
                "LocalizationConfigurationScreen.forgot-password.title"
              )}
              itemKey={PIVOT_KEY_FORGOT_PASSWORD}
            >
              <EditTemplatesWidget sections={sectionsForgotPassword} />
            </PivotItem>
            {passwordlessViaEmailEnabled && (
              <PivotItem
                headerText={renderToString(
                  "LocalizationConfigurationScreen.passwordless-via-email.title"
                )}
                itemKey={PIVOT_KEY_PASSWORDLESS_VIA_EMAIL}
              >
                <EditTemplatesWidget sections={sectionsPasswordlessViaEmail} />
              </PivotItem>
            )}
            {passwordlessViaSMSEnabled && (
              <PivotItem
                headerText={renderToString(
                  "LocalizationConfigurationScreen.passwordless-via-sms.title"
                )}
                itemKey={PIVOT_KEY_PASSWORDLESS_VIA_SMS}
              >
                <EditTemplatesWidget sections={sectionsPasswordlessViaSMS} />
              </PivotItem>
            )}
            <PivotItem
              headerText={renderToString(
                "LocalizationConfigurationScreen.translationjson.title"
              )}
              itemKey={PIVOT_KEY_TRANSLATION_JSON}
            >
              <EditTemplatesWidget sections={sectionsTranslationJSON} />
            </PivotItem>
          </Pivot>
        </Widget>
      </ScreenContent>
    );
  };

const LocalizationConfigurationScreen: React.FC =
  function LocalizationConfigurationScreen() {
    const { appID } = useParams() as { appID: string };
    const [selectedLanguage, setSelectedLanguage] =
      useState<LanguageTag | null>(null);

    const config = useAppConfigForm({
      appID,
      constructFormState,
      constructConfig,
    });

    const initialSupportedLanguages = useMemo(() => {
      return (
        config.effectiveConfig.localization?.supported_languages ?? [
          config.effectiveConfig.localization?.fallback_language ?? "en",
        ]
      );
    }, [config.effectiveConfig.localization]);

    const passwordlessViaEmailEnabled = useMemo(() => {
      return (
        config.effectiveConfig.authentication?.primary_authenticators?.indexOf(
          "oob_otp_email"
        ) !== -1
      );
    }, [config.effectiveConfig]);

    const passwordlessViaSMSEnabled = useMemo(() => {
      return (
        config.effectiveConfig.authentication?.primary_authenticators?.indexOf(
          "oob_otp_sms"
        ) !== -1
      );
    }, [config.effectiveConfig]);

    const specifiers = useMemo<ResourceSpecifier[]>(() => {
      const specifiers = [];
      for (const locale of initialSupportedLanguages) {
        for (const def of ALL_LANGUAGES_TEMPLATES) {
          specifiers.push({
            def,
            locale,
            extension: null,
          });
        }
      }
      return specifiers;
    }, [initialSupportedLanguages]);

    const resources = useResourceForm(
      appID,
      specifiers,
      constructResourcesFormState,
      constructResources
    );

    const state = useMemo<FormState>(
      () => ({
        supportedLanguages: config.state.supportedLanguages,
        fallbackLanguage: config.state.fallbackLanguage,
        resources: resources.state.resources,
        selectedLanguage: selectedLanguage ?? config.state.fallbackLanguage,
      }),
      [
        config.state.supportedLanguages,
        config.state.fallbackLanguage,
        resources.state.resources,
        selectedLanguage,
      ]
    );

    const form: FormModel = {
      isLoading: config.isLoading || resources.isLoading,
      isUpdating: config.isUpdating || resources.isUpdating,
      isDirty: config.isDirty || resources.isDirty,
      loadError: config.loadError ?? resources.loadError,
      updateError: config.updateError ?? resources.updateError,
      state,
      setState: (fn) => {
        const newState = fn(state);
        config.setState(() => ({
          supportedLanguages: newState.supportedLanguages,
          fallbackLanguage: newState.fallbackLanguage,
        }));
        resources.setState(() => ({ resources: newState.resources }));
        setSelectedLanguage(newState.selectedLanguage);
      },
      reload: () => {
        config.reload();
        resources.reload();
      },
      reset: () => {
        config.reset();
        resources.reset();
        setSelectedLanguage(config.state.fallbackLanguage);
      },
      save: async () => {
        await config.save();
        await resources.save();
      },
    };

    if (form.isLoading) {
      return <ShowLoading />;
    }

    if (form.loadError) {
      return <ShowError error={form.loadError} onRetry={form.reload} />;
    }

    return (
      <FormContainer form={form} canSave={true}>
        <ResourcesConfigurationContent
          form={form}
          supportedLanguages={config.state.supportedLanguages}
          passwordlessViaEmailEnabled={passwordlessViaEmailEnabled}
          passwordlessViaSMSEnabled={passwordlessViaSMSEnabled}
        />
      </FormContainer>
    );
  };

export default LocalizationConfigurationScreen;
