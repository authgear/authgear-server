import React, { useCallback, useMemo, useState, useContext } from "react";
import { useParams } from "react-router-dom";
import { FormattedMessage, Context } from "@oursky/react-messageformat";
import { produce } from "immer";
import { parse, Root } from "postcss";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import ScreenContent from "../../ScreenContent";
import ScreenTitle from "../../ScreenTitle";
import ScreenDescription from "../../ScreenDescription";
import WidgetTitle from "../../WidgetTitle";
import Widget from "../../Widget";
import TextField from "../../TextField";
import ManageLanguageWidget from "./ManageLanguageWidget";
import ThemeConfigurationWidget from "../../ThemeConfigurationWidget";
import {
  DEFAULT_LIGHT_THEME,
  DEFAULT_DARK_THEME,
} from "../../ThemePresetWidget";
import ImageFilePicker from "../../ImageFilePicker";
import FeatureDisabledMessageBar from "./FeatureDisabledMessageBar";
import { PortalAPIAppConfig } from "../../types";
import {
  ALL_LANGUAGES_TEMPLATES,
  RESOURCE_TRANSLATION_JSON,
  RESOURCE_FAVICON,
  RESOURCE_APP_LOGO,
  RESOURCE_APP_LOGO_DARK,
  RESOURCE_AUTHGEAR_LIGHT_THEME_CSS,
  RESOURCE_AUTHGEAR_DARK_THEME_CSS,
} from "../../resources";
import {
  expandSpecifier,
  expandDef,
  LanguageTag,
  Resource,
  ResourceDefinition,
  ResourceSpecifier,
  specifierId,
} from "../../util/resource";
import {
  LightTheme,
  DarkTheme,
  BannerConfiguration,
  getLightTheme,
  getDarkTheme,
  getLightBannerConfiguration,
  getDarkBannerConfiguration,
  addLightTheme,
  addDarkTheme,
  addLightBannerConfiguration,
  addDarkBannerConfiguration,
} from "../../util/theme";

import styles from "./UISettingsScreen.module.css";
import { useAppConfigForm } from "../../hook/useAppConfigForm";
import { clearEmptyObject } from "../../util/misc";
import {
  ResourcesFormState,
  useResourceForm,
} from "../../hook/useResourceForm";
import FormContainer from "../../FormContainer";
import { useAppFeatureConfigQuery } from "./query/appFeatureConfigQuery";
import WidgetDescription from "../../WidgetDescription";
import Toggle from "../../Toggle";
import { ErrorParseRule } from "../../error/parse";
import { makeImageSizeTooLargeErrorRule } from "../../error/resources";
import { nonNullable } from "../../util/types";

interface ConfigFormState {
  supportedLanguages: string[];
  fallbackLanguage: string;
  darkThemeDisabled: boolean;
  watermarkDisabled: boolean;

  default_client_uri: string;
}

interface FeatureConfigFormState {
  whiteLabelingDisabled: boolean;
}

const NOOP = () => {};

const ALL_LANGUAGES_TEMPLATES_AND_RESOURCES_ON_THIS_SCREEN = [
  ...ALL_LANGUAGES_TEMPLATES,

  RESOURCE_FAVICON,
  RESOURCE_APP_LOGO,
  RESOURCE_APP_LOGO_DARK,
  RESOURCE_AUTHGEAR_LIGHT_THEME_CSS,
  RESOURCE_AUTHGEAR_DARK_THEME_CSS,
];

function constructFormState(config: PortalAPIAppConfig): ConfigFormState {
  const fallbackLanguage = config.localization?.fallback_language ?? "en";
  return {
    fallbackLanguage,
    supportedLanguages: config.localization?.supported_languages ?? [
      fallbackLanguage,
    ],
    darkThemeDisabled: config.ui?.dark_theme_disabled ?? false,
    watermarkDisabled: config.ui?.watermark_disabled ?? false,
    default_client_uri: config.ui?.default_client_uri ?? "",
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

    config.ui = config.ui ?? {};
    config.ui.dark_theme_disabled = currentState.darkThemeDisabled;
    config.ui.watermark_disabled = currentState.watermarkDisabled;

    config.ui.default_client_uri = currentState.default_client_uri || undefined;

    clearEmptyObject(config);
  });
}

interface FormState
  extends ConfigFormState,
    ResourcesFormState,
    FeatureConfigFormState {
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
  initialSupportedLanguages: string[];
}

const ResourcesConfigurationContent: React.VFC<ResourcesConfigurationContentProps> =
  function ResourcesConfigurationContent(props) {
    const { initialSupportedLanguages } = props;
    const { state, setState } = props.form;
    const { supportedLanguages } = state;

    const { renderToString } = useContext(Context);

    const setSelectedLanguage = useCallback(
      (selectedLanguage: LanguageTag) => {
        setState((s) => ({ ...s, selectedLanguage }));
      },
      [setState]
    );

    const getValueIgnoreEmptyString = useCallback(
      (def: ResourceDefinition) => {
        for (const extension of def.extensions) {
          const specifier: ResourceSpecifier = {
            def,
            extension,
            locale: state.selectedLanguage,
          };
          const value = state.resources[specifierId(specifier)]?.nullableValue;
          if (value != null && value !== "") {
            return value;
          }
        }
        return undefined;
      },
      [state.resources, state.selectedLanguage]
    );

    const getOnChangeImage = useCallback(
      (def: ResourceDefinition) => {
        return (
          image: { base64EncodedData: string; extension: string } | null
        ) => {
          setState((prev) => {
            const updatedResources = { ...prev.resources };

            // We always remove the old one first.
            for (const extension of def.extensions) {
              const specifier: ResourceSpecifier = {
                def,
                extension,
                locale: state.selectedLanguage,
              };
              const oldResource = prev.resources[specifierId(specifier)];
              if (oldResource != null) {
                updatedResources[specifierId(specifier)] = {
                  ...oldResource,
                  nullableValue: "",
                };
              }
            }

            // Add the new one.
            if (image != null) {
              const specifier = {
                def,
                extension: image.extension,
                locale: state.selectedLanguage,
              };
              const resource: Resource = {
                specifier,
                path: expandSpecifier(specifier),
                nullableValue: image.base64EncodedData,
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
          extension: null,
        };

        const value = state.resources[specifierId(specifier)]?.nullableValue;

        if (value == null || value === "") {
          const specifier: ResourceSpecifier = {
            def: RESOURCE_TRANSLATION_JSON,
            locale: state.fallbackLanguage,
            extension: null,
          };
          const value = state.resources[specifierId(specifier)]?.nullableValue;
          if (value == null || value === "") {
            return "";
          }
          const jsonValue = JSON.parse(value);
          return jsonValue[key] ?? "";
        }

        const jsonValue = JSON.parse(value);
        return jsonValue[key] ?? "";
      },
      [state.fallbackLanguage, state.selectedLanguage, state.resources]
    );

    const onChangeForTranslationJSON = useCallback(
      (key: string) => {
        const selectedSpecifier: ResourceSpecifier = {
          def: RESOURCE_TRANSLATION_JSON,
          locale: state.selectedLanguage,
          extension: null,
        };
        const fallbackSpecifier: ResourceSpecifier = {
          def: RESOURCE_TRANSLATION_JSON,
          locale: state.fallbackLanguage,
          extension: null,
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

            let oldValue =
              prev.resources[specifierId(selectedSpecifier)]?.nullableValue;
            if (oldValue == null || oldValue === "") {
              oldValue =
                prev.resources[specifierId(fallbackSpecifier)]?.nullableValue;
              if (oldValue == null || oldValue === "") {
                return prev;
              }
            }

            const jsonValue = JSON.parse(oldValue);

            // If the value is an empty string,
            // interpret as using default value.
            // This interpretation is only present on this screen.
            // LocalizationConfigurationScreen still allows saving empty strings.
            if (value === "") {
              delete jsonValue[key];
            } else {
              jsonValue[key] = value;
            }

            updatedResources[specifierId(selectedSpecifier)] = {
              specifier: selectedSpecifier,
              path: expandSpecifier(selectedSpecifier),
              nullableValue: JSON.stringify(jsonValue, null, 2),
            };
            return { ...prev, resources: updatedResources };
          });
        };
      },
      [state.selectedLanguage, state.fallbackLanguage, setState]
    );

    const onChangeDefaultClientURI = useCallback(
      (
        _e: React.FormEvent<HTMLInputElement | HTMLTextAreaElement>,
        value?: string
      ) => {
        if (value == null) {
          return;
        }
        setState((prev) => {
          return {
            ...prev,
            default_client_uri: value,
          };
        });
      },
      [setState]
    );

    const lightTheme = useMemo(() => {
      let lightTheme = null;
      for (const r of Object.values(state.resources)) {
        if (
          r?.nullableValue != null &&
          r.specifier.def === RESOURCE_AUTHGEAR_LIGHT_THEME_CSS
        ) {
          const root = parse(r.nullableValue);
          lightTheme = getLightTheme(root.nodes);
        }
      }

      return lightTheme;
    }, [state.resources]);

    const darkTheme = useMemo(() => {
      let darkTheme = null;
      for (const r of Object.values(state.resources)) {
        if (
          r?.nullableValue != null &&
          r.specifier.def === RESOURCE_AUTHGEAR_DARK_THEME_CSS
        ) {
          const root = parse(r.nullableValue);
          darkTheme = getDarkTheme(root.nodes);
        }
      }
      return darkTheme;
    }, [state.resources]);

    const lightBannerConfiguration = useMemo(() => {
      let bannerConfiguration = null;
      for (const r of Object.values(state.resources)) {
        if (
          r?.nullableValue != null &&
          r.specifier.def === RESOURCE_AUTHGEAR_LIGHT_THEME_CSS
        ) {
          const root = parse(r.nullableValue);
          bannerConfiguration = getLightBannerConfiguration(root.nodes);
        }
      }
      return bannerConfiguration;
    }, [state.resources]);

    const darkBannerConfiguration = useMemo(() => {
      let bannerConfiguration = null;
      for (const r of Object.values(state.resources)) {
        if (
          r?.nullableValue != null &&
          r.specifier.def === RESOURCE_AUTHGEAR_DARK_THEME_CSS
        ) {
          const root = parse(r.nullableValue);
          bannerConfiguration = getDarkBannerConfiguration(root.nodes);
        }
      }
      return bannerConfiguration;
    }, [state.resources]);

    const setLightThemeAndBannerConfiguration = useCallback(
      (
        newLightTheme: LightTheme | null,
        bannerConfiguration: BannerConfiguration | null
      ) => {
        setState((prev) => {
          const specifier: ResourceSpecifier = {
            def: RESOURCE_AUTHGEAR_LIGHT_THEME_CSS,
            locale: state.selectedLanguage,
            extension: null,
          };
          const updatedResources = { ...prev.resources };
          const root = new Root();
          if (newLightTheme != null) {
            addLightTheme(root, newLightTheme);
          }
          if (bannerConfiguration != null) {
            addLightBannerConfiguration(root, bannerConfiguration);
          }
          const css = root.toResult().css;
          const newResource: Resource = {
            specifier,
            path: expandSpecifier(specifier),
            nullableValue: css,
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

    const setDarkThemeAndBannerConfiguration = useCallback(
      (
        newDarkTheme: DarkTheme | null,
        bannerConfiguration: BannerConfiguration | null
      ) => {
        setState((prev) => {
          const specifier: ResourceSpecifier = {
            def: RESOURCE_AUTHGEAR_DARK_THEME_CSS,
            locale: state.selectedLanguage,
            extension: null,
          };
          const updatedResources = { ...prev.resources };
          const root = new Root();
          if (newDarkTheme != null) {
            addDarkTheme(root, newDarkTheme);
          }
          if (bannerConfiguration != null) {
            addDarkBannerConfiguration(root, bannerConfiguration);
          }
          const css = root.toResult().css;
          const newResource: Resource = {
            specifier,
            path: expandSpecifier(specifier),
            nullableValue: css,
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

    const setLightTheme = useCallback(
      (newLightTheme: LightTheme) => {
        setLightThemeAndBannerConfiguration(
          newLightTheme,
          lightBannerConfiguration
        );
      },
      [lightBannerConfiguration, setLightThemeAndBannerConfiguration]
    );

    const setLightBannerConfiguration = useCallback(
      (bannerConfiguration: BannerConfiguration) => {
        setLightThemeAndBannerConfiguration(lightTheme, bannerConfiguration);
      },
      [lightTheme, setLightThemeAndBannerConfiguration]
    );

    const setDarkTheme = useCallback(
      (newDarkTheme: DarkTheme) => {
        setDarkThemeAndBannerConfiguration(
          newDarkTheme,
          darkBannerConfiguration
        );
      },
      [darkBannerConfiguration, setDarkThemeAndBannerConfiguration]
    );

    const setDarkBannerConfiguration = useCallback(
      (bannerConfiguration: BannerConfiguration) => {
        setDarkThemeAndBannerConfiguration(darkTheme, bannerConfiguration);
      },
      [darkTheme, setDarkThemeAndBannerConfiguration]
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

    const onChangeLightModePrimaryColor =
      getOnChangeLightThemeColor("primaryColor");
    const onChangeLightModeTextColor = getOnChangeLightThemeColor("textColor");
    const onChangeLightModeBackgroundColor =
      getOnChangeLightThemeColor("backgroundColor");
    const onChangeDarkModePrimaryColor =
      getOnChangeDarkThemeColor("primaryColor");
    const onChangeDarkModeTextColor = getOnChangeDarkThemeColor("textColor");
    const onChangeDarkModeBackgroundColor =
      getOnChangeDarkThemeColor("backgroundColor");

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

    const onChangeWatermarkEnabled = useCallback(
      (_event, checked?: boolean) => {
        setState((prev) => {
          return {
            ...prev,
            watermarkDisabled: !checked,
          };
        });
      },
      [setState]
    );

    const watermarkEnabled = useMemo(() => {
      return state.whiteLabelingDisabled || !state.watermarkDisabled;
    }, [state.whiteLabelingDisabled, state.watermarkDisabled]);

    return (
      <ScreenContent>
        <div className={styles.titleContainer}>
          <ScreenTitle>
            <FormattedMessage id="UISettingsScreen.title" />
          </ScreenTitle>
          <ManageLanguageWidget
            existingLanguages={initialSupportedLanguages}
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
            label={renderToString("UISettingsScreen.privacy-policy-link-label")}
            description={renderToString(
              "UISettingsScreen.privacy-policy-link-description"
            )}
            value={valueForTranslationJSON("privacy-policy-link")}
            onChange={onChangeForTranslationJSON("privacy-policy-link")}
          />
          <TextField
            label={renderToString(
              "UISettingsScreen.terms-of-service-link-label"
            )}
            description={renderToString(
              "UISettingsScreen.terms-of-service-link-description"
            )}
            value={valueForTranslationJSON("terms-of-service-link")}
            onChange={onChangeForTranslationJSON("terms-of-service-link")}
          />
          <TextField
            label={renderToString(
              "UISettingsScreen.customer-support-link-label"
            )}
            description={renderToString(
              "UISettingsScreen.customer-support-link-description"
            )}
            value={valueForTranslationJSON("customer-support-link")}
            onChange={onChangeForTranslationJSON("customer-support-link")}
          />
          <TextField
            label={renderToString("UISettingsScreen.default-client-uri-label")}
            description={renderToString(
              "UISettingsScreen.default-client-uri-description"
            )}
            value={state.default_client_uri}
            onChange={onChangeDefaultClientURI}
          />
        </Widget>
        <Widget className={styles.widget}>
          <WidgetTitle>
            <FormattedMessage id="UISettingsScreen.favicon-title" />
          </WidgetTitle>
          <WidgetDescription>
            <FormattedMessage id="UISettingsScreen.favicon-description" />
          </WidgetDescription>
          <ImageFilePicker
            sizeLimitInBytes={100 * 1000}
            base64EncodedData={getValueIgnoreEmptyString(RESOURCE_FAVICON)}
            onChange={getOnChangeImage(RESOURCE_FAVICON)}
          />
        </Widget>
        <Widget className={styles.widget}>
          <WidgetTitle>
            <FormattedMessage id="UISettingsScreen.branding.title" />
          </WidgetTitle>
          {state.whiteLabelingDisabled ? (
            <FeatureDisabledMessageBar messageID="FeatureConfig.white-labeling.disabled" />
          ) : null}
          <Toggle
            checked={watermarkEnabled}
            onChange={onChangeWatermarkEnabled}
            label={renderToString(
              "UISettingsScreen.branding.disable-authgear-logo.label"
            )}
            inlineLabel={true}
            disabled={state.whiteLabelingDisabled}
          />
        </Widget>
        <ThemeConfigurationWidget
          className={styles.widget}
          darkTheme={darkTheme}
          lightTheme={lightTheme}
          isDarkMode={false}
          darkModeEnabled={false}
          watermarkEnabled={watermarkEnabled}
          appLogoValue={getValueIgnoreEmptyString(RESOURCE_APP_LOGO)}
          onChangeAppLogo={getOnChangeImage(RESOURCE_APP_LOGO)}
          onChangeDarkModeEnabled={NOOP}
          onChangeLightTheme={setLightTheme}
          onChangeDarkTheme={setDarkTheme}
          onChangePrimaryColor={onChangeLightModePrimaryColor}
          onChangeTextColor={onChangeLightModeTextColor}
          onChangeBackgroundColor={onChangeLightModeBackgroundColor}
          bannerConfiguration={lightBannerConfiguration}
          onChangeBannerConfiguration={setLightBannerConfiguration}
        />
        <ThemeConfigurationWidget
          className={styles.widget}
          darkTheme={darkTheme}
          lightTheme={lightTheme}
          isDarkMode={true}
          darkModeEnabled={!state.darkThemeDisabled}
          watermarkEnabled={watermarkEnabled}
          appLogoValue={getValueIgnoreEmptyString(RESOURCE_APP_LOGO_DARK)}
          onChangeAppLogo={getOnChangeImage(RESOURCE_APP_LOGO_DARK)}
          onChangeLightTheme={setLightTheme}
          onChangeDarkTheme={setDarkTheme}
          onChangeDarkModeEnabled={onChangeDarkModeEnabled}
          onChangePrimaryColor={onChangeDarkModePrimaryColor}
          onChangeTextColor={onChangeDarkModeTextColor}
          onChangeBackgroundColor={onChangeDarkModeBackgroundColor}
          bannerConfiguration={darkBannerConfiguration}
          onChangeBannerConfiguration={setDarkBannerConfiguration}
        />
      </ScreenContent>
    );
  };

const UISettingsScreen: React.VFC = function UISettingsScreen() {
  const { appID } = useParams() as { appID: string };
  const [selectedLanguage, setSelectedLanguage] = useState<LanguageTag | null>(
    null
  );

  const config = useAppConfigForm({
    appID,
    constructFormState,
    constructConfig,
  });

  const featureConfig = useAppFeatureConfigQuery(appID);

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
      for (const def of ALL_LANGUAGES_TEMPLATES_AND_RESOURCES_ON_THIS_SCREEN) {
        specifiers.push(...expandDef(def, locale));
      }
    }
    return specifiers;
  }, [initialSupportedLanguages]);

  const resources = useResourceForm(appID, specifiers);

  const state = useMemo<FormState>(
    () => ({
      supportedLanguages: config.state.supportedLanguages,
      fallbackLanguage: config.state.fallbackLanguage,
      resources: resources.state.resources,
      selectedLanguage: selectedLanguage ?? config.state.fallbackLanguage,
      darkThemeDisabled: config.state.darkThemeDisabled,
      watermarkDisabled: config.state.watermarkDisabled,
      default_client_uri: config.state.default_client_uri,
      whiteLabelingDisabled:
        featureConfig.effectiveFeatureConfig?.ui?.white_labeling?.disabled ??
        false,
    }),
    [
      config.state.supportedLanguages,
      config.state.fallbackLanguage,
      config.state.darkThemeDisabled,
      config.state.watermarkDisabled,
      config.state.default_client_uri,
      resources.state.resources,
      selectedLanguage,
      featureConfig.effectiveFeatureConfig?.ui?.white_labeling?.disabled,
    ]
  );

  const form: FormModel = useMemo(
    () => ({
      isLoading:
        config.isLoading || resources.isLoading || featureConfig.isLoading,
      isUpdating: config.isUpdating || resources.isUpdating,
      isDirty: config.isDirty || resources.isDirty,
      loadError:
        config.loadError ?? resources.loadError ?? featureConfig.loadError,
      updateError: config.updateError ?? resources.updateError,
      state,
      setState: (fn) => {
        const newState = fn(state);
        config.setState((configState) => ({
          ...configState,
          darkThemeDisabled: newState.darkThemeDisabled,
          watermarkDisabled: newState.watermarkDisabled,
          default_client_uri: newState.default_client_uri,
        }));
        resources.setState(() => ({ resources: newState.resources }));
        setSelectedLanguage(newState.selectedLanguage);
      },
      reload: () => {
        config.reload();
        resources.reload();
        featureConfig.refetch().finally(() => {});
      },
      reset: () => {
        config.reset();
        resources.reset();
        setSelectedLanguage(config.state.fallbackLanguage);
      },
      save: async (ignoreConflict: boolean = false) => {
        await config.save(ignoreConflict);
        await resources.save(ignoreConflict);
      },
    }),
    [config, featureConfig, resources, state]
  );

  const errorRules: ErrorParseRule[] = useMemo(
    () => [
      makeImageSizeTooLargeErrorRule(
        Object.values(form.state.resources).filter(nonNullable)
      ),
    ],
    [form.state.resources]
  );

  if (form.isLoading) {
    return <ShowLoading />;
  }

  if (form.loadError) {
    return <ShowError error={form.loadError} onRetry={form.reload} />;
  }

  return (
    <FormContainer form={form} canSave={true} errorRules={errorRules}>
      <ResourcesConfigurationContent
        form={form}
        initialSupportedLanguages={initialSupportedLanguages}
      />
    </FormContainer>
  );
};

export default UISettingsScreen;
