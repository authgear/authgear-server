import React, { useCallback, useContext, useMemo, useState } from "react";
import { useParams } from "react-router-dom";
import { Pivot, PivotItem } from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import { produce } from "immer";
import { parse } from "postcss";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import ManageLanguageWidget from "./ManageLanguageWidget";
import ThemeConfigurationWidget from "../../ThemeConfigurationWidget";
import ImageFilePicker from "../../ImageFilePicker";
import EditTemplatesWidget, {
  EditTemplatesWidgetSection,
} from "./EditTemplatesWidget";
import { useTemplateLocaleQuery } from "./query/templateLocaleQuery";
import { PortalAPIAppConfig } from "../../types";
import {
  ALL_EDITABLE_RESOURCES,
  ALL_TEMPLATES,
  renderPath,
  RESOURCE_FAVICON,
  RESOURCE_APP_LOGO,
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
  RESOURCE_AUTHGEAR_CSS,
  RESOURCE_AUTHGEAR_LIGHT_THEME_CSS,
  RESOURCE_AUTHGEAR_DARK_THEME_CSS,
} from "../../resources";
import {
  LanguageTag,
  LocaleInvalidReason,
  Resource,
  ResourceDefinition,
  ResourceSpecifier,
  specifierId,
  validateLocales,
} from "../../util/resource";
import {
  DEFAULT_LIGHT_THEME,
  DEFAULT_DARK_THEME,
  LightTheme,
  DarkTheme,
  getLightTheme,
  getDarkTheme,
  lightThemeToCSS,
  darkThemeToCSS,
} from "../../util/theme";

import styles from "./ResourceConfigurationScreen.module.scss";
import { useAppConfigForm } from "../../hook/useAppConfigForm";
import { clearEmptyObject } from "../../util/misc";
import { useResourceForm } from "../../hook/useResourceForm";
import FormContainer from "../../FormContainer";
import NavBreadcrumb, { BreadcrumbItem } from "../../NavBreadcrumb";

interface ConfigFormState {
  defaultLocale: string;
  darkThemeDisabled: boolean;
}

const NOOP = () => {};

function constructConfigFormState(config: PortalAPIAppConfig): ConfigFormState {
  return {
    defaultLocale: config.localization?.fallback_language ?? "en",
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
    if (initialState.defaultLocale !== currentState.defaultLocale) {
      config.localization.fallback_language = currentState.defaultLocale;
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
  selectedLocale: string;
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
  locales: LanguageTag[];
  invalidLocaleReason: LocaleInvalidReason | null;
  invalidLocales: LanguageTag[];
}

const PIVOT_KEY_APPEARANCE = "appearance";
const PIVOT_KEY_CUSTOM_CSS = "custom-css";
const PIVOT_KEY_FORGOT_PASSWORD = "forgot_password";
const PIVOT_KEY_PASSWORDLESS = "passwordless";
const PIVOT_KEY_THEME = "theme";
const PIVOT_KEY_TRANSLATION_JSON = "translation.json";

const PIVOT_KEY_DEFAULT = PIVOT_KEY_APPEARANCE;

const ALL_PIVOT_KEYS = [
  PIVOT_KEY_APPEARANCE,
  PIVOT_KEY_CUSTOM_CSS,
  PIVOT_KEY_FORGOT_PASSWORD,
  PIVOT_KEY_PASSWORDLESS,
  PIVOT_KEY_TRANSLATION_JSON,
  PIVOT_KEY_THEME,
];

const ResourcesConfigurationContent: React.FC<ResourcesConfigurationContentProps> = function ResourcesConfigurationContent(
  props
) {
  const { state, setState } = props.form;
  const { locales, invalidLocaleReason, invalidLocales } = props;
  const { renderToString } = useContext(Context);

  const navBreadcrumbItems: BreadcrumbItem[] = useMemo(() => {
    return [
      {
        to: ".",
        label: <FormattedMessage id="ResourceConfigurationScreen.title" />,
      },
    ];
  }, []);

  const setDefaultLocale = useCallback(
    (defaultLocale: LanguageTag) => {
      setState((s) => ({ ...s, defaultLocale }));
    },
    [setState]
  );

  const setSelectedLocale = useCallback(
    (selectedLocale: LanguageTag) => {
      setState((s) => ({ ...s, selectedLocale }));
    },
    [setState]
  );

  const onChangeLocales = useCallback(
    (newLocales: LanguageTag[]) => {
      setState((prev) => {
        let { selectedLocale, resources } = prev;
        resources = { ...resources };

        // Reset selected locale to default if it's removed.
        if (!newLocales.includes(state.selectedLocale)) {
          selectedLocale = state.defaultLocale;
        }

        // Populate initial resources for added locales from default locale.
        const addedLocales = newLocales.filter((l) => !locales.includes(l));
        for (const locale of addedLocales) {
          for (const resource of ALL_TEMPLATES) {
            const defaultResource =
              state.resources[
                specifierId({ def: resource, locale: prev.defaultLocale })
              ];
            const newResource: Resource = {
              specifier: {
                def: resource,
                locale,
              },
              path: renderPath(resource.resourcePath, { locale }),
              value: defaultResource?.value ?? "",
            };
            resources[specifierId(newResource.specifier)] = newResource;
          }
        }

        // Remove resources of removed locales.
        const removedLocales = locales.filter((l) => !newLocales.includes(l));
        for (const [id, resource] of Object.entries(resources)) {
          const locale = resource?.specifier.locale;
          if (resource && locale && removedLocales.includes(locale)) {
            resources[id] = { ...resource, value: "" };
          }
        }

        return { ...prev, selectedLocale, resources };
      });
    },
    [locales, setState, state]
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

  const getValueIgnoreEmptyString = useCallback(
    (def: ResourceDefinition) => {
      const specifier: ResourceSpecifier = {
        def,
        locale: state.selectedLocale,
      };
      const resource = state.resources[specifierId(specifier)];
      if (resource == null || resource.value === "") {
        return undefined;
      }
      return resource.value;
    },
    [state.resources, state.selectedLocale]
  );

  const getValue = useCallback(
    (def: ResourceDefinition) => {
      const specifier: ResourceSpecifier = {
        def,
        locale: state.selectedLocale,
      };
      const resource = state.resources[specifierId(specifier)];
      return resource?.value ?? "";
    },
    [state.resources, state.selectedLocale]
  );

  const getOnChange = useCallback(
    (def: ResourceDefinition) => {
      const specifier: ResourceSpecifier = {
        def,
        locale: state.selectedLocale,
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
    [state.selectedLocale, setState]
  );

  const getOnChangeImage = useCallback(
    (def: ResourceDefinition) => {
      const specifier: ResourceSpecifier = {
        def,
        locale: state.selectedLocale,
      };
      return (base64EncodedData?: string, extension?: string) => {
        setState((prev) => {
          const updatedResources = { ...prev.resources };
          const resource: Resource = {
            specifier,
            path: renderPath(specifier.def.resourcePath, {
              locale: specifier.locale,
              extension,
            }),
            value: base64EncodedData ?? "",
          };
          updatedResources[specifierId(resource.specifier)] = resource;
          return { ...prev, resources: updatedResources };
        });
      };
    },
    [state.selectedLocale, setState]
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
          locale: state.selectedLocale,
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
    [setState, state.selectedLocale]
  );

  const setDarkTheme = useCallback(
    (newDarkTheme: DarkTheme) => {
      setState((prev) => {
        const specifier: ResourceSpecifier = {
          def: RESOURCE_AUTHGEAR_DARK_THEME_CSS,
          locale: state.selectedLocale,
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
    [setState, state.selectedLocale]
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
    "lightModePrimaryColor"
  );
  const onChangeLightModeTextColor = getOnChangeLightThemeColor(
    "lightModeTextColor"
  );
  const onChangeLightModeBackgroundColor = getOnChangeLightThemeColor(
    "lightModeBackgroundColor"
  );
  const onChangeDarkModePrimaryColor = getOnChangeDarkThemeColor(
    "darkModePrimaryColor"
  );
  const onChangeDarkModeTextColor = getOnChangeDarkThemeColor(
    "darkModeTextColor"
  );
  const onChangeDarkModeBackgroundColor = getOnChangeDarkThemeColor(
    "darkModeBackgroundColor"
  );

  const onChangeDarkModeEnabled = useCallback(
    (enabled) => {
      if (enabled) {
        // Become enabled, copy the light theme with text color and background color swapped.
        const base = lightTheme ?? DEFAULT_LIGHT_THEME;
        const newDarkTheme = {
          darkModePrimaryColor: base.lightModePrimaryColor,
          darkModeTextColor: base.lightModeBackgroundColor,
          darkModeBackgroundColor: base.lightModeTextColor,
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

  const sectionsCustomCSS: EditTemplatesWidgetSection[] = [
    {
      key: "custom-css",
      title: <FormattedMessage id="EditTemplatesWidget.custom-css.title" />,
      items: [
        {
          key: "custom-css",
          title: (
            <FormattedMessage id="EditTemplatesWidget.custom-css.subtitle" />
          ),
          language: "css",
          value: getValue(RESOURCE_AUTHGEAR_CSS),
          onChange: getOnChange(RESOURCE_AUTHGEAR_CSS),
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
        templateLocales={locales}
        onChangeTemplateLocales={onChangeLocales}
        templateLocale={state.selectedLocale}
        defaultTemplateLocale={state.defaultLocale}
        onSelectTemplateLocale={setSelectedLocale}
        onSelectDefaultTemplateLocale={setDefaultLocale}
        invalidLocaleReason={invalidLocaleReason}
        invalidTemplateLocales={invalidLocales}
      />
      <Pivot onLinkClick={onLinkClick} selectedKey={selectedKey}>
        <PivotItem
          headerText={renderToString(
            "ResourceConfigurationScreen.appearance.title"
          )}
          itemKey={PIVOT_KEY_APPEARANCE}
        >
          <div className={styles.pivotItemAppearance}>
            <ImageFilePicker
              title={renderToString("ResourceConfigurationScreen.favicon")}
              base64EncodedData={getValueIgnoreEmptyString(RESOURCE_FAVICON)}
              onChange={getOnChangeImage(RESOURCE_FAVICON)}
            />
            <ImageFilePicker
              title={renderToString("ResourceConfigurationScreen.app-logo")}
              base64EncodedData={getValueIgnoreEmptyString(RESOURCE_APP_LOGO)}
              onChange={getOnChangeImage(RESOURCE_APP_LOGO)}
            />
          </div>
        </PivotItem>
        <PivotItem
          headerText={renderToString("ResourceConfigurationScreen.theme.title")}
          itemKey={PIVOT_KEY_THEME}
        >
          <ThemeConfigurationWidget
            isDarkMode={false}
            darkModeEnabled={false}
            onChangeDarkModeEnabled={NOOP}
            primaryColor={
              lightTheme?.lightModePrimaryColor ??
              DEFAULT_LIGHT_THEME.lightModePrimaryColor
            }
            textColor={
              lightTheme?.lightModeTextColor ??
              DEFAULT_LIGHT_THEME.lightModeTextColor
            }
            backgroundColor={
              lightTheme?.lightModeBackgroundColor ??
              DEFAULT_LIGHT_THEME.lightModeBackgroundColor
            }
            onChangePrimaryColor={onChangeLightModePrimaryColor}
            onChangeTextColor={onChangeLightModeTextColor}
            onChangeBackgroundColor={onChangeLightModeBackgroundColor}
          />
          <ThemeConfigurationWidget
            isDarkMode={true}
            darkModeEnabled={!state.darkThemeDisabled}
            onChangeDarkModeEnabled={onChangeDarkModeEnabled}
            primaryColor={
              darkTheme?.darkModePrimaryColor ??
              DEFAULT_DARK_THEME.darkModePrimaryColor
            }
            textColor={
              darkTheme?.darkModeTextColor ??
              DEFAULT_DARK_THEME.darkModeTextColor
            }
            backgroundColor={
              darkTheme?.darkModeBackgroundColor ??
              DEFAULT_DARK_THEME.darkModeBackgroundColor
            }
            onChangePrimaryColor={onChangeDarkModePrimaryColor}
            onChangeTextColor={onChangeDarkModeTextColor}
            onChangeBackgroundColor={onChangeDarkModeBackgroundColor}
          />
        </PivotItem>
        <PivotItem
          headerText={renderToString(
            "ResourceConfigurationScreen.translationjson.title"
          )}
          itemKey={PIVOT_KEY_TRANSLATION_JSON}
        >
          <EditTemplatesWidget sections={sectionsTranslationJSON} />
        </PivotItem>
        <PivotItem
          headerText={renderToString(
            "ResourceConfigurationScreen.forgot-password.title"
          )}
          itemKey={PIVOT_KEY_FORGOT_PASSWORD}
        >
          <EditTemplatesWidget sections={sectionsForgotPassword} />
        </PivotItem>
        <PivotItem
          headerText={renderToString(
            "ResourceConfigurationScreen.passwordless-authenticator.title"
          )}
          itemKey={PIVOT_KEY_PASSWORDLESS}
        >
          <EditTemplatesWidget sections={sectionsPasswordless} />
        </PivotItem>
        <PivotItem
          headerText={renderToString(
            "ResourceConfigurationScreen.custom-css.title"
          )}
          itemKey={PIVOT_KEY_CUSTOM_CSS}
        >
          <EditTemplatesWidget sections={sectionsCustomCSS} />
        </PivotItem>
      </Pivot>
    </div>
  );
};

const ResourceConfigurationScreen: React.FC = function ResourceConfigurationScreen() {
  const { appID } = useParams();

  const {
    templateLocales: resourceLocales,
    loading: isLoadingLocales,
    error: loadLocalesError,
    refetch: reloadLocales,
  } = useTemplateLocaleQuery(appID);

  const specifiers = useMemo<ResourceSpecifier[]>(() => {
    const specifiers = [];
    for (const locale of resourceLocales) {
      for (const def of ALL_EDITABLE_RESOURCES) {
        specifiers.push({
          def,
          locale,
        });
      }
    }
    return specifiers;
  }, [resourceLocales]);

  const [selectedLocale, setSelectedLocale] = useState<LanguageTag | null>(
    null
  );

  const config = useAppConfigForm(
    appID,
    constructConfigFormState,
    constructConfig
  );
  const resources = useResourceForm(
    appID,
    specifiers,
    constructResourcesFormState,
    constructResources
  );
  const state = useMemo<FormState>(
    () => ({
      defaultLocale: config.state.defaultLocale,
      resources: resources.state.resources,
      selectedLocale: selectedLocale ?? config.state.defaultLocale,
      darkThemeDisabled: config.state.darkThemeDisabled,
    }),
    [
      config.state.defaultLocale,
      config.state.darkThemeDisabled,
      resources.state.resources,
      selectedLocale,
    ]
  );

  const form: FormModel = {
    isLoading: isLoadingLocales || config.isLoading || resources.isLoading,
    isUpdating: config.isUpdating || resources.isUpdating,
    isDirty: config.isDirty || resources.isDirty,
    loadError: loadLocalesError ?? config.loadError ?? resources.loadError,
    updateError: config.updateError ?? resources.updateError,
    state,
    setState: (fn) => {
      const newState = fn(state);
      config.setState(() => ({
        defaultLocale: newState.defaultLocale,
        darkThemeDisabled: newState.darkThemeDisabled,
      }));
      resources.setState(() => ({ resources: newState.resources }));
      setSelectedLocale(newState.selectedLocale);
    },
    reload: () => {
      reloadLocales().catch(() => {});
      config.reload();
      resources.reload();
    },
    reset: () => {
      config.reset();
      resources.reset();
      setSelectedLocale(config.state.defaultLocale);
    },
    save: () => {
      config.save();
      resources.save();
    },
  };

  const locales = useMemo<LanguageTag[]>(() => {
    const locales = new Set<LanguageTag>();
    for (const r of Object.values(resources.state.resources)) {
      if (r?.specifier.locale && r.value.length !== 0) {
        locales.add(r.specifier.locale);
      }
    }
    if (selectedLocale) {
      locales.add(selectedLocale);
    }
    locales.add(config.state.defaultLocale);
    // eslint-disable-next-line @typescript-eslint/require-array-sort-compare
    return Array.from(locales).sort();
  }, [selectedLocale, resources.state, config.state]);

  const { invalidReason: invalidLocaleReason, invalidLocales } = useMemo(
    () =>
      validateLocales(
        config.state.defaultLocale,
        locales,
        Object.values(resources.state.resources).filter(Boolean) as Resource[]
      ),
    [locales, config.state, resources.state]
  );

  if (form.isLoading) {
    return <ShowLoading />;
  }

  if (form.loadError) {
    return <ShowError error={form.loadError} onRetry={form.reload} />;
  }

  return (
    <FormContainer form={form} canSave={invalidLocaleReason == null}>
      <ResourcesConfigurationContent
        form={form}
        locales={locales}
        invalidLocaleReason={invalidLocaleReason}
        invalidLocales={invalidLocales}
      />
    </FormContainer>
  );
};

export default ResourceConfigurationScreen;
