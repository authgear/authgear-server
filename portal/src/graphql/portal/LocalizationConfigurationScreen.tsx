import React, { useCallback, useContext, useMemo, useState } from "react";
import { useParams } from "react-router-dom";
import { Pivot, PivotItem } from "@fluentui/react";
import deepEqual from "deep-equal";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import { produce } from "immer";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import ManageLanguageWidget from "./ManageLanguageWidget";
import EditTemplatesWidget, {
  EditTemplatesWidgetSection,
} from "./EditTemplatesWidget";
import { PortalAPIAppConfig } from "../../types";
import {
  ALL_EDITABLE_RESOURCES,
  ALL_TEMPLATES,
  renderPath,
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
} from "../../resources";
import {
  LanguageTag,
  Resource,
  ResourceDefinition,
  ResourceSpecifier,
  specifierId,
} from "../../util/resource";
import { useAppConfigForm } from "../../hook/useAppConfigForm";
import { clearEmptyObject } from "../../util/misc";
import { useResourceForm } from "../../hook/useResourceForm";
import FormContainer from "../../FormContainer";
import NavBreadcrumb, { BreadcrumbItem } from "../../NavBreadcrumb";
import styles from "./LocalizationConfigurationScreen.module.scss";

interface ConfigFormState {
  supportedLanguages: string[];
  fallbackLanguage: string;
  darkThemeDisabled: boolean;
}

function constructConfigFormState(config: PortalAPIAppConfig): ConfigFormState {
  const fallbackLanguage = config.localization?.fallback_language ?? "en";
  return {
    fallbackLanguage,
    supportedLanguages: config.localization?.supported_languages ?? [
      fallbackLanguage,
    ],
    darkThemeDisabled: config.ui?.dark_theme_disabled ?? false,
  };
}

function constructConfig(
  config: PortalAPIAppConfig,
  initialState: ConfigFormState,
  currentState: ConfigFormState
): PortalAPIAppConfig {
  return produce(config, (config) => {
    config.localization = config.localization ?? {};

    if (initialState.fallbackLanguage !== currentState.fallbackLanguage) {
      config.localization.fallback_language = currentState.fallbackLanguage;
    }
    if (
      !deepEqual(
        initialState.supportedLanguages,
        currentState.supportedLanguages,
        { strict: true }
      )
    ) {
      config.localization.supported_languages = currentState.supportedLanguages;
    }

    config.ui = config.ui ?? {};
    if (initialState.darkThemeDisabled !== currentState.darkThemeDisabled) {
      config.ui.dark_theme_disabled = currentState.darkThemeDisabled;
    }
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
    if ((resourceMap[id]?.value ?? "") === "") {
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
  save: () => void;
}

interface ResourcesConfigurationContentProps {
  form: FormModel;
  supportedLanguages: LanguageTag[];
}

const PIVOT_KEY_FORGOT_PASSWORD = "forgot_password";
const PIVOT_KEY_PASSWORDLESS = "passwordless";
const PIVOT_KEY_TRANSLATION_JSON = "translation.json";

const PIVOT_KEY_DEFAULT = PIVOT_KEY_TRANSLATION_JSON;

const ALL_PIVOT_KEYS = [
  PIVOT_KEY_FORGOT_PASSWORD,
  PIVOT_KEY_PASSWORDLESS,
  PIVOT_KEY_TRANSLATION_JSON,
];

const ResourcesConfigurationContent: React.FC<ResourcesConfigurationContentProps> = function ResourcesConfigurationContent(
  props
) {
  const { state, setState } = props.form;
  const { supportedLanguages } = props;
  const { renderToString } = useContext(Context);

  const navBreadcrumbItems: BreadcrumbItem[] = useMemo(() => {
    return [
      {
        to: ".",
        label: <FormattedMessage id="LocalizationConfigurationScreen.title" />,
      },
    ];
  }, []);

  const setFallbackLanguage = useCallback(
    (fallbackLanguage: LanguageTag) => {
      setState((s) => ({ ...s, fallbackLanguage }));
    },
    [setState]
  );

  const setSelectedLanguage = useCallback(
    (selectedLanguage: LanguageTag) => {
      setState((s) => ({ ...s, selectedLanguage }));
    },
    [setState]
  );

  const setSupportedLanguages = useCallback(
    (supportedLanguages: LanguageTag[]) => {
      setState((prev) => {
        // Reset selected language to fallback language if it was removed.
        let { selectedLanguage, resources } = prev;
        resources = { ...resources };
        if (!supportedLanguages.includes(selectedLanguage)) {
          selectedLanguage = prev.fallbackLanguage;
        }

        // Populate initial values for added languages from fallback language.
        const addedLanguages = supportedLanguages.filter(
          (l) => !prev.supportedLanguages.includes(l)
        );
        for (const language of addedLanguages) {
          for (const def of ALL_TEMPLATES) {
            const defaultResource =
              prev.resources[
                specifierId({ def, locale: prev.fallbackLanguage })
              ];
            const newResource: Resource = {
              specifier: {
                def,
                locale: language,
              },
              path: renderPath(def.resourcePath, { locale: language }),
              value: defaultResource?.value ?? "",
            };
            resources[specifierId(newResource.specifier)] = newResource;
          }
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
            resources[id] = { ...resource, value: "" };
          }
        }

        return { ...prev, selectedLanguage, supportedLanguages, resources };
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

  const getValue = useCallback(
    (def: ResourceDefinition) => {
      const specifier: ResourceSpecifier = {
        def,
        locale: state.selectedLanguage,
      };
      return state.resources[specifierId(specifier)]?.value ?? "";
    },
    [state.resources, state.selectedLanguage]
  );

  const getOnChange = useCallback(
    (def: ResourceDefinition) => {
      const specifier: ResourceSpecifier = {
        def,
        locale: state.selectedLanguage,
      };
      return (_e: unknown, value?: string) => {
        setState((prev) => {
          const updatedResources = { ...prev.resources };
          const resource: Resource = {
            specifier,
            path: renderPath(specifier.def.resourcePath, {
              locale: specifier.locale,
            }),
            value: value ?? "",
          };
          updatedResources[specifierId(resource.specifier)] = resource;
          return { ...prev, resources: updatedResources };
        });
      };
    },
    [state.selectedLanguage, setState]
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
            <FormattedMessage id="EditTemplatesWidget.translationjson.subtitle" />
          ),
          language: "json",
          value: getValue(RESOURCE_TRANSLATION_JSON),
          onChange: getOnChange(RESOURCE_TRANSLATION_JSON),
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
          key: "html-email",
          title: <FormattedMessage id="EditTemplatesWidget.html-email" />,
          language: "html",
          value: getValue(RESOURCE_FORGOT_PASSWORD_EMAIL_HTML),
          onChange: getOnChange(RESOURCE_FORGOT_PASSWORD_EMAIL_HTML),
        },
        {
          key: "plaintext-email",
          title: <FormattedMessage id="EditTemplatesWidget.plaintext-email" />,
          language: "plaintext",
          value: getValue(RESOURCE_FORGOT_PASSWORD_EMAIL_TXT),
          onChange: getOnChange(RESOURCE_FORGOT_PASSWORD_EMAIL_TXT),
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
        },
      ],
    },
  ];

  const sectionsPasswordless: EditTemplatesWidgetSection[] = [
    {
      key: "setup",
      title: (
        <FormattedMessage id="EditTemplatesWidget.passwordless.setup.title" />
      ),
      items: [
        {
          key: "html-email",
          title: <FormattedMessage id="EditTemplatesWidget.html-email" />,
          language: "html",
          value: getValue(RESOURCE_SETUP_PRIMARY_OOB_EMAIL_HTML),
          onChange: getOnChange(RESOURCE_SETUP_PRIMARY_OOB_EMAIL_HTML),
        },
        {
          key: "plaintext-email",
          title: <FormattedMessage id="EditTemplatesWidget.plaintext-email" />,
          language: "plaintext",
          value: getValue(RESOURCE_SETUP_PRIMARY_OOB_EMAIL_TXT),
          onChange: getOnChange(RESOURCE_SETUP_PRIMARY_OOB_EMAIL_TXT),
        },
        {
          key: "sms",
          title: <FormattedMessage id="EditTemplatesWidget.sms-body" />,
          language: "plaintext",
          value: getValue(RESOURCE_SETUP_PRIMARY_OOB_SMS_TXT),
          onChange: getOnChange(RESOURCE_SETUP_PRIMARY_OOB_SMS_TXT),
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
          key: "html-email",
          title: <FormattedMessage id="EditTemplatesWidget.html-email" />,
          language: "html",
          value: getValue(RESOURCE_AUTHENTICATE_PRIMARY_OOB_EMAIL_HTML),
          onChange: getOnChange(RESOURCE_AUTHENTICATE_PRIMARY_OOB_EMAIL_HTML),
        },
        {
          key: "plaintext-email",
          title: <FormattedMessage id="EditTemplatesWidget.plaintext-email" />,
          language: "plaintext",
          value: getValue(RESOURCE_AUTHENTICATE_PRIMARY_OOB_EMAIL_TXT),
          onChange: getOnChange(RESOURCE_AUTHENTICATE_PRIMARY_OOB_EMAIL_TXT),
        },
        {
          key: "sms",
          title: <FormattedMessage id="EditTemplatesWidget.sms-body" />,
          language: "plaintext",
          value: getValue(RESOURCE_AUTHENTICATE_PRIMARY_OOB_SMS_TXT),
          onChange: getOnChange(RESOURCE_AUTHENTICATE_PRIMARY_OOB_SMS_TXT),
        },
      ],
    },
  ];

  return (
    <div className={styles.root}>
      <NavBreadcrumb items={navBreadcrumbItems} />
      <ManageLanguageWidget
        supportedLanguages={supportedLanguages}
        onChangeSupportedLanguages={setSupportedLanguages}
        selectedLanguage={state.selectedLanguage}
        onChangeSelectedLanguage={setSelectedLanguage}
        fallbackLanguage={state.fallbackLanguage}
        onChangeFallbackLanguage={setFallbackLanguage}
      />
      <Pivot onLinkClick={onLinkClick} selectedKey={selectedKey}>
        <PivotItem
          headerText={renderToString(
            "LocalizationConfigurationScreen.translationjson.title"
          )}
          itemKey={PIVOT_KEY_TRANSLATION_JSON}
        >
          <EditTemplatesWidget sections={sectionsTranslationJSON} />
        </PivotItem>
        <PivotItem
          headerText={renderToString(
            "LocalizationConfigurationScreen.forgot-password.title"
          )}
          itemKey={PIVOT_KEY_FORGOT_PASSWORD}
        >
          <EditTemplatesWidget sections={sectionsForgotPassword} />
        </PivotItem>
        <PivotItem
          headerText={renderToString(
            "LocalizationConfigurationScreen.passwordless-authenticator.title"
          )}
          itemKey={PIVOT_KEY_PASSWORDLESS}
        >
          <EditTemplatesWidget sections={sectionsPasswordless} />
        </PivotItem>
      </Pivot>
    </div>
  );
};

const LocalizationConfigurationScreen: React.FC = function LocalizationConfigurationScreen() {
  const { appID } = useParams();
  const [selectedLanguage, setSelectedLanguage] = useState<LanguageTag | null>(
    null
  );

  const config = useAppConfigForm(
    appID,
    constructConfigFormState,
    constructConfig
  );

  const initialSupportedLanguages = useMemo(() => {
    return (
      config.effectiveConfig.localization?.supported_languages ?? [
        config.effectiveConfig.localization?.fallback_language ?? "en",
      ]
    );
  }, [config.effectiveConfig.localization]);

  const specifiers = useMemo<ResourceSpecifier[]>(() => {
    const specifiers = [];
    for (const locale of initialSupportedLanguages) {
      for (const def of ALL_EDITABLE_RESOURCES) {
        specifiers.push({
          def,
          locale,
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
      darkThemeDisabled: config.state.darkThemeDisabled,
    }),
    [
      config.state.supportedLanguages,
      config.state.fallbackLanguage,
      config.state.darkThemeDisabled,
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
        darkThemeDisabled: newState.darkThemeDisabled,
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
    save: () => {
      config.save();
      resources.save();
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
      />
    </FormContainer>
  );
};

export default LocalizationConfigurationScreen;
