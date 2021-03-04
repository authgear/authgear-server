import React, { useCallback, useMemo, useState, useContext } from "react";
import { useParams } from "react-router-dom";
import { TextField } from "@fluentui/react";
import deepEqual from "deep-equal";
import { FormattedMessage, Context } from "@oursky/react-messageformat";
import { produce } from "immer";
import { parse } from "postcss";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import ScreenContent from "../../ScreenContent";
import ScreenTitle from "../../ScreenTitle";
import ScreenDescription from "../../ScreenDescription";
import WidgetTitle from "../../WidgetTitle";
import Widget from "../../Widget";
import ManageLanguageWidget from "./ManageLanguageWidget";
import ThemeConfigurationWidget from "../../ThemeConfigurationWidget";
import {
  DEFAULT_LIGHT_THEME,
  DEFAULT_DARK_THEME,
} from "../../ThemePresetWidget";
import ImageFilePicker from "../../ImageFilePicker";
import { PortalAPIAppConfig } from "../../types";
import {
  renderPath,
  RESOURCE_TRANSLATION_JSON,
  RESOURCE_FAVICON,
  RESOURCE_APP_LOGO,
  RESOURCE_APP_LOGO_DARK,
  RESOURCE_AUTHGEAR_LIGHT_THEME_CSS,
  RESOURCE_AUTHGEAR_DARK_THEME_CSS,
} from "../../resources";
import {
  LanguageTag,
  Resource,
  ResourceDefinition,
  ResourceSpecifier,
  specifierId,
} from "../../util/resource";
import {
  LightTheme,
  DarkTheme,
  getLightTheme,
  getDarkTheme,
  lightThemeToCSS,
  darkThemeToCSS,
} from "../../util/theme";

import styles from "./UISettingsScreen.module.scss";
import { useAppConfigForm } from "../../hook/useAppConfigForm";
import { clearEmptyObject } from "../../util/misc";
import { useResourceForm } from "../../hook/useResourceForm";
import FormContainer from "../../FormContainer";

interface ConfigFormState {
  supportedLanguages: string[];
  fallbackLanguage: string;
  darkThemeDisabled: boolean;
}

const NOOP = () => {};

const RESOURCES_ON_THIS_SCREEN = [
  RESOURCE_TRANSLATION_JSON,
  RESOURCE_FAVICON,
  RESOURCE_APP_LOGO,
  RESOURCE_APP_LOGO_DARK,
  RESOURCE_AUTHGEAR_LIGHT_THEME_CSS,
  RESOURCE_AUTHGEAR_DARK_THEME_CSS,
];

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

const ResourcesConfigurationContent: React.FC<ResourcesConfigurationContentProps> = function ResourcesConfigurationContent(
  props
) {
  const { state, setState } = props.form;
  const { supportedLanguages } = props;

  const { renderToString } = useContext(Context);

  const setSelectedLanguage = useCallback(
    (selectedLanguage: LanguageTag) => {
      setState((s) => ({ ...s, selectedLanguage }));
    },
    [setState]
  );

  const getValueIgnoreEmptyString = useCallback(
    (def: ResourceDefinition) => {
      const specifier: ResourceSpecifier = {
        def,
        locale: state.selectedLanguage,
      };
      const resource = state.resources[specifierId(specifier)];
      if (resource == null || resource.value === "") {
        return undefined;
      }
      return resource.value;
    },
    [state.resources, state.selectedLanguage]
  );

  const getOnChangeImage = useCallback(
    (def: ResourceDefinition) => {
      const specifier: ResourceSpecifier = {
        def,
        locale: state.selectedLanguage,
      };
      return (base64EncodedData?: string, extension?: string) => {
        setState((prev) => {
          const updatedResources = { ...prev.resources };

          // We always remove the old one first.
          const oldResource = prev.resources[specifierId(specifier)];
          if (oldResource != null) {
            updatedResources[specifierId(specifier)] = {
              ...oldResource,
              value: "",
            };
          }

          // Add the new one.
          if (base64EncodedData != null && extension != null) {
            const resource: Resource = {
              specifier,
              path: renderPath(specifier.def.resourcePath, {
                locale: specifier.locale,
                extension,
              }),
              value: base64EncodedData,
            };
            updatedResources[specifierId(specifier)] = resource;
          }

          return { ...prev, resources: updatedResources };
        });
      };
    },
    [state.selectedLanguage, setState]
  );

  const valueForTranslationJSON = useCallback(
    (key: string) => {
      const specifier: ResourceSpecifier = {
        def: RESOURCE_TRANSLATION_JSON,
        locale: state.selectedLanguage,
      };
      const resource = state.resources[specifierId(specifier)];
      if (resource == null) {
        return "";
      }
      const jsonValue = JSON.parse(resource.value);
      return jsonValue[key] ?? "";
    },
    [state.selectedLanguage, state.resources]
  );

  const onChangeForTranslationJSON = useCallback(
    (key: string) => {
      const specifier: ResourceSpecifier = {
        def: RESOURCE_TRANSLATION_JSON,
        locale: state.selectedLanguage,
      };
      return (
        _e: React.FormEvent<HTMLInputElement | HTMLTextAreaElement>,
        value?: string
      ) => {
        if (value == null) {
          return;
        }
        setState((prev) => {
          const updatedResources = { ...prev.resources };
          const oldResource = prev.resources[specifierId(specifier)];
          if (oldResource == null) {
            return prev;
          }
          const jsonValue = JSON.parse(oldResource.value);
          jsonValue[key] = value;
          updatedResources[specifierId(specifier)] = {
            ...oldResource,
            value: JSON.stringify(jsonValue, null, 2),
          };
          return { ...prev, resources: updatedResources };
        });
      };
    },
    [state.selectedLanguage, setState]
  );

  const lightTheme = useMemo(() => {
    let lightTheme = null;
    for (const r of Object.values(state.resources)) {
      if (r != null && r.specifier.def === RESOURCE_AUTHGEAR_LIGHT_THEME_CSS) {
        const root = parse(r.value);
        lightTheme = getLightTheme(root.nodes);
      }
    }

    return lightTheme;
  }, [state.resources]);

  const darkTheme = useMemo(() => {
    let darkTheme = null;
    for (const r of Object.values(state.resources)) {
      if (r != null && r.specifier.def === RESOURCE_AUTHGEAR_DARK_THEME_CSS) {
        const root = parse(r.value);
        darkTheme = getDarkTheme(root.nodes);
      }
    }
    return darkTheme;
  }, [state.resources]);

  const setLightTheme = useCallback(
    (newLightTheme: LightTheme) => {
      setState((prev) => {
        const specifier: ResourceSpecifier = {
          def: RESOURCE_AUTHGEAR_LIGHT_THEME_CSS,
          locale: state.selectedLanguage,
        };
        const updatedResources = { ...prev.resources };
        const css = lightThemeToCSS(newLightTheme);
        const newResource: Resource = {
          specifier,
          path: renderPath(specifier.def.resourcePath, {
            locale: specifier.locale,
          }),
          value: css,
        };
        updatedResources[specifierId(newResource.specifier)] = newResource;
        return {
          ...prev,
          resources: updatedResources,
        };
      });
    },
    [setState, state.selectedLanguage]
  );

  const setDarkTheme = useCallback(
    (newDarkTheme: DarkTheme) => {
      setState((prev) => {
        const specifier: ResourceSpecifier = {
          def: RESOURCE_AUTHGEAR_DARK_THEME_CSS,
          locale: state.selectedLanguage,
        };
        const updatedResources = { ...prev.resources };
        const css = darkThemeToCSS(newDarkTheme);
        const newResource: Resource = {
          specifier,
          path: renderPath(specifier.def.resourcePath, {
            locale: specifier.locale,
          }),
          value: css,
        };
        updatedResources[specifierId(newResource.specifier)] = newResource;
        return {
          ...prev,
          resources: updatedResources,
        };
      });
    },
    [setState, state.selectedLanguage]
  );

  const getOnChangeLightThemeColor = useCallback(
    (key: keyof LightTheme) => {
      return (color: string) => {
        const newLightTheme: LightTheme = {
          ...(lightTheme ?? DEFAULT_LIGHT_THEME),
          [key]: color,
        };
        setLightTheme(newLightTheme);
      };
    },
    [lightTheme, setLightTheme]
  );

  const getOnChangeDarkThemeColor = useCallback(
    (key: keyof DarkTheme) => {
      return (color: string) => {
        const newDarkTheme: DarkTheme = {
          ...(darkTheme ?? DEFAULT_DARK_THEME),
          [key]: color,
        };
        setDarkTheme(newDarkTheme);
      };
    },
    [darkTheme, setDarkTheme]
  );

  const onChangeLightModePrimaryColor = getOnChangeLightThemeColor(
    "primaryColor"
  );
  const onChangeLightModeTextColor = getOnChangeLightThemeColor("textColor");
  const onChangeLightModeBackgroundColor = getOnChangeLightThemeColor(
    "backgroundColor"
  );
  const onChangeDarkModePrimaryColor = getOnChangeDarkThemeColor(
    "primaryColor"
  );
  const onChangeDarkModeTextColor = getOnChangeDarkThemeColor("textColor");
  const onChangeDarkModeBackgroundColor = getOnChangeDarkThemeColor(
    "backgroundColor"
  );

  const onChangeDarkModeEnabled = useCallback(
    (enabled) => {
      if (enabled) {
        // Become enabled, copy the light theme with text color and background color swapped.
        const base = lightTheme ?? DEFAULT_LIGHT_THEME;
        const newDarkTheme: DarkTheme = {
          isDarkTheme: true,
          primaryColor: base.primaryColor,
          textColor: base.backgroundColor,
          backgroundColor: base.textColor,
        };
        setDarkTheme(newDarkTheme);
      }

      setState((prev) => {
        return {
          ...prev,
          darkThemeDisabled: !enabled,
        };
      });
    },
    [setState, lightTheme, setDarkTheme]
  );

  return (
    <ScreenContent className={styles.root}>
      <div className={styles.titleContainer}>
        <ScreenTitle>
          <FormattedMessage id="UISettingsScreen.title" />
        </ScreenTitle>
        <ManageLanguageWidget
          selectOnly={true}
          supportedLanguages={supportedLanguages}
          selectedLanguage={state.selectedLanguage}
          fallbackLanguage={state.fallbackLanguage}
          onChangeSelectedLanguage={setSelectedLanguage}
        />
      </div>
      <ScreenDescription className={styles.widget}>
        <FormattedMessage id="UISettingsScreen.description" />
      </ScreenDescription>
      <Widget className={styles.widget}>
        <WidgetTitle>
          <FormattedMessage id="UISettingsScreen.app-name-title" />
        </WidgetTitle>
        <TextField
          className={styles.textField}
          label={renderToString("UISettingsScreen.app-name-label")}
          value={valueForTranslationJSON("app.name")}
          onChange={onChangeForTranslationJSON("app.name")}
        />
      </Widget>
      <Widget className={styles.widget}>
        <WidgetTitle>
          <FormattedMessage id="UISettingsScreen.link-settings-title" />
        </WidgetTitle>
        <TextField
          className={styles.textField}
          label={renderToString("UISettingsScreen.privacy-policy-link-label")}
          description={renderToString(
            "UISettingsScreen.privacy-policy-link-description"
          )}
          value={valueForTranslationJSON("privacy-policy-link")}
          onChange={onChangeForTranslationJSON("privacy-policy-link")}
        />
        <TextField
          className={styles.textField}
          label={renderToString("UISettingsScreen.terms-of-service-link-label")}
          description={renderToString(
            "UISettingsScreen.terms-of-service-link-description"
          )}
          value={valueForTranslationJSON("terms-of-service-link")}
          onChange={onChangeForTranslationJSON("terms-of-service-link")}
        />
      </Widget>
      <Widget className={styles.widget}>
        <WidgetTitle>
          <FormattedMessage id="UISettingsScreen.favicon-title" />
        </WidgetTitle>
        <ImageFilePicker
          className={styles.faviconImagePicker}
          base64EncodedData={getValueIgnoreEmptyString(RESOURCE_FAVICON)}
          onChange={getOnChangeImage(RESOURCE_FAVICON)}
        />
      </Widget>
      <ThemeConfigurationWidget
        className={styles.widget}
        darkTheme={darkTheme}
        lightTheme={lightTheme}
        isDarkMode={false}
        darkModeEnabled={false}
        appLogoValue={getValueIgnoreEmptyString(RESOURCE_APP_LOGO)}
        onChangeAppLogo={getOnChangeImage(RESOURCE_APP_LOGO)}
        onChangeDarkModeEnabled={NOOP}
        onChangeLightTheme={setLightTheme}
        onChangeDarkTheme={setDarkTheme}
        onChangePrimaryColor={onChangeLightModePrimaryColor}
        onChangeTextColor={onChangeLightModeTextColor}
        onChangeBackgroundColor={onChangeLightModeBackgroundColor}
      />
      <ThemeConfigurationWidget
        className={styles.widget}
        darkTheme={darkTheme}
        lightTheme={lightTheme}
        isDarkMode={true}
        darkModeEnabled={!state.darkThemeDisabled}
        appLogoValue={getValueIgnoreEmptyString(RESOURCE_APP_LOGO_DARK)}
        onChangeAppLogo={getOnChangeImage(RESOURCE_APP_LOGO_DARK)}
        onChangeLightTheme={setLightTheme}
        onChangeDarkTheme={setDarkTheme}
        onChangeDarkModeEnabled={onChangeDarkModeEnabled}
        onChangePrimaryColor={onChangeDarkModePrimaryColor}
        onChangeTextColor={onChangeDarkModeTextColor}
        onChangeBackgroundColor={onChangeDarkModeBackgroundColor}
      />
    </ScreenContent>
  );
};

const UISettingsScreen: React.FC = function UISettingsScreen() {
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
      for (const def of RESOURCES_ON_THIS_SCREEN) {
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

export default UISettingsScreen;
