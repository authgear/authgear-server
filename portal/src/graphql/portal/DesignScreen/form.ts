import { useCallback, useMemo, useState } from "react";
import { parse as parseCSS } from "postcss";
import { produce } from "immer";
import {
  useResourceForm,
  ResourcesFormState as UseResourceFormState,
} from "../../../hook/useResourceForm";
import {
  Alignment,
  BorderRadiusStyle,
  CSSColor,
  CssAstVisitor,
  CustomisableThemeStyleGroup,
  EMPTY_THEME,
  PartialCustomisableTheme,
  StyleCssVisitor,
  TextDecorationType,
  Theme,
  getThemeTargetSelector,
  selectByTheme,
} from "../../../model/themeAuthFlowV2";
import {
  RESOURCE_APP_BACKGROUND_IMAGE,
  RESOURCE_APP_BACKGROUND_IMAGE_DARK,
  RESOURCE_APP_LOGO,
  RESOURCE_APP_LOGO_DARK,
  RESOURCE_AUTHGEAR_AUTHFLOW_V2_DARK_THEME_CSS,
  RESOURCE_AUTHGEAR_AUTHFLOW_V2_LIGHT_THEME_CSS,
  RESOURCE_FAVICON,
  RESOURCE_TRANSLATION_JSON,
} from "../../../resources";
import {
  LanguageTag,
  Resource,
  ResourceDefinition,
  ResourceSpecifier,
  expandDef,
  expandSpecifier,
  specifierId,
} from "../../../util/resource";
import { useAppConfigForm } from "../../../hook/useAppConfigForm";
import { PortalAPIAppConfig } from "../../../types";
import { ErrorParseRule } from "../../../error/parse";
import { useAppFeatureConfigQuery } from "../query/appFeatureConfigQuery";
import { makeImageSizeTooLargeErrorRule } from "../../../error/resources";
import { nonNullable } from "../../../util/types";
import { nullishCoalesce, or_ } from "../../../util/operators";
import { deriveColors } from "../../../util/theme";

const LOCALE_BASED_RESOUCE_DEFINITIONS = [
  RESOURCE_TRANSLATION_JSON,
  RESOURCE_APP_LOGO,
  RESOURCE_APP_LOGO_DARK,
  RESOURCE_FAVICON,
];

const THEME_RESOURCE_DEFINITIONS = [
  RESOURCE_AUTHGEAR_AUTHFLOW_V2_LIGHT_THEME_CSS,
  RESOURCE_AUTHGEAR_AUTHFLOW_V2_DARK_THEME_CSS,
];

const LightThemeResourceSpecifier = {
  def: RESOURCE_AUTHGEAR_AUTHFLOW_V2_LIGHT_THEME_CSS,
  locale: null,
  extension: null,
};

const DarkThemeResourceSpecifier = {
  def: RESOURCE_AUTHGEAR_AUTHFLOW_V2_DARK_THEME_CSS,
  locale: null,
  extension: null,
};

type ThemeOption = "lightOnly" | "darkOnly" | "auto";

interface ConfigFormState {
  supportedLanguages: LanguageTag[];
  fallbackLanguage: LanguageTag;
  showAuthgearLogo: boolean;
  defaultClientURI: string;
  themeOption: ThemeOption;
}

export interface AppLogoResource {
  base64EncodedData: string | null;
  fallbackBase64EncodedData: string | null;
}

interface ResourcesFormState {
  appName: string;

  // light
  customisableLightTheme: PartialCustomisableTheme;
  appLogo: AppLogoResource;
  backgroundImageBase64EncodedData: string | null;

  // dark
  customisableDarkTheme: PartialCustomisableTheme;
  appLogoDark: AppLogoResource;
  backgroundImageDarkBase64EncodedData: string | null;

  faviconBase64EncodedData: string | null;
  urls: {
    privacyPolicy: string;
    termsOfService: string;
    customerSupport: string;
  };
}

interface FeatureConfig {
  whiteLabelingDisabled: boolean;
}

export const enum TranslationKey {
  AppName = "app.name",
  PrivacyPolicy = "privacy-policy-link",
  TermsOfService = "terms-of-service-link",
  CustomerSupport = "customer-support-link",
}

export type BranchDesignFormState = {
  selectedLanguage: LanguageTag;
  selectedTheme: Theme;
} & ConfigFormState &
  ResourcesFormState &
  FeatureConfig;

interface ThemeSetters {
  setAppLogo: (
    image: { base64EncodedData: string; extension: string } | null
  ) => void;
  setLogoHeight: (height: string | undefined) => void;
  setBackgroundColor: (color: CSSColor | undefined) => void;
  setBackgroundImage: (
    image: { base64EncodedData: string; extension: string } | null
  ) => void;
  setPrimaryButtonBackgroundColor: (color: CSSColor | undefined) => void;
  setPrimaryButtonLabelColor: (color: CSSColor | undefined) => void;
  setIconColor: (color: CSSColor | undefined) => void;
  setLinkColor: (color: CSSColor | undefined) => void;
}

export interface BranchDesignForm {
  isLoading: boolean;
  isUpdating: boolean;
  isDirty: boolean;
  loadError: unknown;
  updateError: unknown;
  state: BranchDesignFormState;
  reload: () => void;
  reset: () => void;
  save: () => Promise<void>;

  errorRules: ErrorParseRule[];

  setSelectedLanguage: (lang: LanguageTag) => void;
  setSelectedTheme: (theme: Theme) => void;
  setThemeOption: (themeOption: ThemeOption) => void;
  setAppName: (appName: string) => void;
  lightThemeSetters: ThemeSetters;
  darkThemeSetters: ThemeSetters;
  setFavicon: (
    image: { base64EncodedData: string; extension: string } | null
  ) => void;
  setCardAlignment: (alignment: Alignment) => void;
  setPrimaryButtonBorderRadiusStyle: (
    borderRadiusStyle: BorderRadiusStyle | undefined
  ) => void;
  setInputFieldBorderRadiusStyle: (
    borderRadiusStyle: BorderRadiusStyle | undefined
  ) => void;
  setLinkTextDecorationStyle: (
    textDecoration: TextDecorationType | undefined
  ) => void;
  setPrivacyPolicyLink: (url: string) => void;
  setTermsOfServiceLink: (url: string) => void;
  setCustomerSupportLink: (url: string) => void;
  setDefaultClientURI: (url: string) => void;
  setDisplayAuthgearLogo: (disabled: boolean) => void;
}

function constructConfigFormState(config: PortalAPIAppConfig): ConfigFormState {
  const fallbackLanguage = config.localization?.fallback_language ?? "en";
  return {
    fallbackLanguage,
    supportedLanguages: config.localization?.supported_languages ?? [
      fallbackLanguage,
    ],
    showAuthgearLogo: !(config.ui?.watermark_disabled ?? false),
    defaultClientURI: config.ui?.default_client_uri ?? "",
    themeOption:
      !config.ui?.dark_theme_disabled && config.ui?.light_theme_disabled
        ? "darkOnly"
        : config.ui?.dark_theme_disabled && !config.ui.light_theme_disabled
        ? "lightOnly"
        : "auto",
  };
}

function constructConfigFromFormState(
  config: PortalAPIAppConfig,
  _initialState: ConfigFormState,
  currentState: ConfigFormState
): PortalAPIAppConfig {
  return produce(config, (draft) => {
    if (draft.ui == null) {
      draft.ui = {};
    }
    draft.ui.watermark_disabled = !currentState.showAuthgearLogo;
    draft.ui.default_client_uri = currentState.defaultClientURI || undefined;
    if (currentState.themeOption === "lightOnly") {
      draft.ui.dark_theme_disabled = true;
      draft.ui.light_theme_disabled = undefined;
    } else if (currentState.themeOption === "darkOnly") {
      draft.ui.dark_theme_disabled = undefined;
      draft.ui.light_theme_disabled = true;
    } else {
      draft.ui.dark_theme_disabled = undefined;
      draft.ui.light_theme_disabled = undefined;
    }
  });
}

function resolveResource(
  resources: Partial<Record<string, Resource>>,
  specifiers: [ResourceSpecifier] | ResourceSpecifier[]
): Resource | null {
  for (const specifier of specifiers) {
    const resource = resources[specifierId(specifier)];
    if (resource?.nullableValue) {
      return resource;
    }
  }
  return resources[specifierId(specifiers[specifiers.length - 1])] ?? null;
}

function getThemeFromResourceFormState(
  state: UseResourceFormState,
  theme: Theme
): PartialCustomisableTheme {
  const themeResource =
    state.resources[
      specifierId(
        selectByTheme(
          {
            [Theme.Light]: LightThemeResourceSpecifier,
            [Theme.Dark]: DarkThemeResourceSpecifier,
          },
          theme
        )
      )
    ];
  if (themeResource?.nullableValue == null) {
    return EMPTY_THEME;
  }
  const root = parseCSS(themeResource.nullableValue);
  const styleCSSVisitor = new StyleCssVisitor(
    getThemeTargetSelector(theme),
    new CustomisableThemeStyleGroup()
  );
  return styleCSSVisitor.getStyle(root);
}

export function useBrandDesignForm(appID: string): BranchDesignForm {
  const featureConfig = useAppFeatureConfigQuery(appID);
  const configForm = useAppConfigForm({
    appID,
    constructFormState: constructConfigFormState,
    constructConfig: constructConfigFromFormState,
  });
  const [selectedTheme, setSelectedTheme] = useState(
    configForm.state.themeOption === "darkOnly" ||
      (configForm.state.themeOption === "auto" &&
        window.matchMedia("(prefers-color-scheme: dark)").matches)
      ? Theme.Dark
      : Theme.Light
  );
  const [selectedLanguage, setSelectedLanguage] = useState(
    configForm.state.fallbackLanguage
  );

  const specifiers = useMemo<ResourceSpecifier[]>(() => {
    const specifiers: ResourceSpecifier[] = [];
    for (const def of THEME_RESOURCE_DEFINITIONS) {
      specifiers.push({
        def,
        locale: null,
        extension: null,
      });
    }
    for (const locale of configForm.state.supportedLanguages) {
      for (const def of LOCALE_BASED_RESOUCE_DEFINITIONS) {
        specifiers.push(...expandDef(def, locale));
      }
    }
    return specifiers;
  }, [configForm.state.supportedLanguages]);

  const backgroundImageSpecifiers = useMemo(() => {
    const specifiers: ResourceSpecifier[] = [];
    specifiers.push(...expandDef(RESOURCE_APP_BACKGROUND_IMAGE, ""));
    specifiers.push(...expandDef(RESOURCE_APP_BACKGROUND_IMAGE_DARK, ""));
    return specifiers;
  }, []);

  const resourceForm = useResourceForm(appID, specifiers);
  const backgroundImageResourceForm = useResourceForm(
    appID,
    backgroundImageSpecifiers
  );

  const getResourceFormByResourceDefinition = useCallback(
    (def: ResourceDefinition) => {
      if (
        def === RESOURCE_APP_BACKGROUND_IMAGE ||
        def === RESOURCE_APP_BACKGROUND_IMAGE_DARK
      ) {
        return backgroundImageResourceForm;
      }
      return resourceForm;
    },
    [resourceForm, backgroundImageResourceForm]
  );

  const resourcesState: ResourcesFormState = useMemo(() => {
    const getValueFromTranslationJSON = (key: string): string => {
      const specifier: ResourceSpecifier = {
        def: RESOURCE_TRANSLATION_JSON,
        locale: selectedLanguage,
        extension: null,
      };
      const fallbackSpecifier: ResourceSpecifier = {
        def: RESOURCE_TRANSLATION_JSON,
        locale: configForm.state.fallbackLanguage,
        extension: null,
      };
      const translationResource = resolveResource(
        resourceForm.state.resources,
        [specifier, fallbackSpecifier]
      );
      if (!translationResource?.nullableValue) {
        return "";
      }
      const jsonValue = JSON.parse(translationResource.nullableValue);
      return jsonValue[key] ?? "";
    };

    const getValueFromImageResource = (
      def: ResourceDefinition,
      language: string
    ): string | null => {
      const form = getResourceFormByResourceDefinition(def);
      const specifiers = expandDef(def, language);
      const imageResouece = resolveResource(form.state.resources, specifiers);
      if (!imageResouece?.nullableValue) {
        return null;
      }
      return imageResouece.nullableValue;
    };

    const lightTheme = getThemeFromResourceFormState(
      resourceForm.state,
      Theme.Light
    );
    const darkTheme = getThemeFromResourceFormState(
      resourceForm.state,
      Theme.Dark
    );

    return {
      appName: getValueFromTranslationJSON(TranslationKey.AppName),
      appLogo: {
        base64EncodedData: getValueFromImageResource(
          RESOURCE_APP_LOGO,
          selectedLanguage
        ),
        fallbackBase64EncodedData: getValueFromImageResource(
          RESOURCE_APP_LOGO,
          configForm.state.fallbackLanguage
        ),
      },
      appLogoDark: {
        base64EncodedData: getValueFromImageResource(
          RESOURCE_APP_LOGO_DARK,
          selectedLanguage
        ),
        fallbackBase64EncodedData: getValueFromImageResource(
          RESOURCE_APP_LOGO_DARK,
          configForm.state.fallbackLanguage
        ),
      },
      faviconBase64EncodedData: getValueFromImageResource(
        RESOURCE_FAVICON,
        selectedLanguage
      ),
      backgroundImageBase64EncodedData: getValueFromImageResource(
        RESOURCE_APP_BACKGROUND_IMAGE,
        selectedLanguage
      ),
      backgroundImageDarkBase64EncodedData: getValueFromImageResource(
        RESOURCE_APP_BACKGROUND_IMAGE_DARK,
        selectedLanguage
      ),
      customisableLightTheme: lightTheme,
      customisableDarkTheme: darkTheme,

      urls: {
        privacyPolicy: getValueFromTranslationJSON(
          TranslationKey.PrivacyPolicy
        ),
        termsOfService: getValueFromTranslationJSON(
          TranslationKey.TermsOfService
        ),
        customerSupport: getValueFromTranslationJSON(
          TranslationKey.CustomerSupport
        ),
      },
    };
  }, [
    resourceForm.state,
    selectedLanguage,
    configForm.state.fallbackLanguage,
    getResourceFormByResourceDefinition,
  ]);

  const resourceMutator = useMemo(() => {
    return {
      setTranslationValue: (key: string, value: string) => {
        resourceForm.setState((s) => {
          return produce(s, (draft) => {
            const specifier: ResourceSpecifier = {
              def: RESOURCE_TRANSLATION_JSON,
              locale: selectedLanguage,
              extension: null,
            };
            const fallbackSpecifier: ResourceSpecifier = {
              def: RESOURCE_TRANSLATION_JSON,
              locale: configForm.state.fallbackLanguage,
              extension: null,
            };
            const translationResource = resolveResource(
              resourceForm.state.resources,
              [specifier, fallbackSpecifier]
            );
            if (!translationResource?.nullableValue) {
              return;
            }
            const jsonValue = JSON.parse(translationResource.nullableValue);
            if (value === "") {
              delete jsonValue[key];
            } else {
              jsonValue[key] = value;
            }
            draft.resources[specifierId(specifier)] = {
              specifier: specifier,
              path: expandSpecifier(specifier),
              nullableValue: JSON.stringify(jsonValue, null, 2),
            };
          });
        });
      },
      setImage: (
        def: ResourceDefinition,
        image: {
          base64EncodedData: string;
          extension: string;
        } | null
      ) => {
        const form = getResourceFormByResourceDefinition(def);
        form.setState((prev) => {
          return produce(prev, (draft) => {
            const specifiers = expandDef(def, selectedLanguage);
            for (const specifier of specifiers) {
              const resource = draft.resources[specifierId(specifier)];
              if (resource != null) {
                resource.nullableValue = "";
              }
            }
            if (image == null) {
              return;
            }
            const specifier = {
              def,
              extension: image.extension,
              locale: selectedLanguage,
            };
            const resource: Resource = {
              specifier,
              path: expandSpecifier(specifier),
              nullableValue: image.base64EncodedData,
            };
            draft.resources[specifierId(specifier)] = resource;
          });
        });
      },
      updateCustomisableTheme: (
        updater: (prev: PartialCustomisableTheme) => PartialCustomisableTheme,
        targetTheme: Theme
      ) => {
        resourceForm.setState((s) => {
          return produce(s, (draft) => {
            const newState = updater(
              getThemeFromResourceFormState(s, targetTheme)
            );
            const resourceSpecifier = selectByTheme(
              {
                [Theme.Light]: LightThemeResourceSpecifier,
                [Theme.Dark]: DarkThemeResourceSpecifier,
              },
              targetTheme
            );
            const themeResource = draft.resources[
              specifierId(resourceSpecifier)
            ] ?? {
              specifier: resourceSpecifier,
              path: expandSpecifier(resourceSpecifier),
            };

            themeResource.nullableValue = (() => {
              const cssAstVisitor = new CssAstVisitor(
                getThemeTargetSelector(targetTheme)
              );
              const styleGroup = new CustomisableThemeStyleGroup(newState);
              styleGroup.acceptCssAstVisitor(cssAstVisitor);
              if (cssAstVisitor.getDeclarations().length <= 0) {
                return "";
              }
              return cssAstVisitor.getCSS().toResult().css;
            })();

            draft.resources[specifierId(resourceSpecifier)] = themeResource;
          });
        });
      },
    };
  }, [
    resourceForm,
    selectedLanguage,
    configForm.state.fallbackLanguage,
    getResourceFormByResourceDefinition,
  ]);

  const state: BranchDesignFormState = useMemo(
    () => ({
      selectedTheme,
      selectedLanguage,
      ...configForm.state,
      ...resourcesState,
      whiteLabelingDisabled:
        featureConfig.effectiveFeatureConfig?.ui?.white_labeling?.disabled ??
        false,
    }),
    [
      selectedTheme,
      selectedLanguage,
      configForm.state,
      resourcesState,
      featureConfig.effectiveFeatureConfig?.ui?.white_labeling?.disabled,
    ]
  );

  const errorRules: ErrorParseRule[] = useMemo(
    () => [
      makeImageSizeTooLargeErrorRule(
        Object.values(resourceForm.state.resources)
          .concat(Object.values(backgroundImageResourceForm.state.resources))
          .filter(nonNullable)
      ),
    ],
    [resourceForm.state.resources, backgroundImageResourceForm.state.resources]
  );

  const _setAppLogo = useCallback(
    (image, targetTheme: Theme) => {
      resourceMutator.setImage(
        selectByTheme(
          {
            light: RESOURCE_APP_LOGO,
            dark: RESOURCE_APP_LOGO_DARK,
          },
          targetTheme
        ),
        image
      );
    },
    [resourceMutator]
  );

  const _setLogoHeight = useCallback(
    (height: string | undefined, targetTheme: Theme) => {
      resourceMutator.updateCustomisableTheme((prev) => {
        return produce(prev, (draft) => {
          draft.logo.height = height;
        });
      }, targetTheme);
    },
    [resourceMutator]
  );

  const _setBackgroundColor = useCallback(
    (backgroundColor: CSSColor | undefined, targetTheme: Theme) => {
      resourceMutator.updateCustomisableTheme((prev) => {
        return produce(prev, (draft) => {
          draft.page.backgroundColor = backgroundColor;
        });
      }, targetTheme);
    },
    [resourceMutator]
  );

  const _setBackgroundImage = useCallback(
    (image, targetTheme: Theme) => {
      resourceMutator.setImage(
        selectByTheme(
          {
            light: RESOURCE_APP_BACKGROUND_IMAGE,
            dark: RESOURCE_APP_BACKGROUND_IMAGE_DARK,
          },
          targetTheme
        ),
        image
      );
    },
    [resourceMutator]
  );

  const _setPrimaryButtonBackgroundColor = useCallback(
    (backgroundColor: CSSColor | undefined, targetTheme: Theme) => {
      resourceMutator.updateCustomisableTheme((prev) => {
        return produce(prev, (draft) => {
          if (backgroundColor == null) {
            draft.primaryButton.backgroundColor = undefined;
            draft.primaryButton.backgroundColorActive = undefined;
            draft.primaryButton.backgroundColorHover = undefined;
            return;
          }
          draft.primaryButton.backgroundColor = backgroundColor;

          const derivedColors = deriveColors(backgroundColor);
          if (derivedColors == null) {
            draft.primaryButton.backgroundColor = undefined;
            draft.primaryButton.backgroundColorActive = undefined;
            draft.primaryButton.backgroundColorHover = undefined;
            return;
          }
          draft.primaryButton.backgroundColorActive = derivedColors.variant;
          draft.primaryButton.backgroundColorHover = derivedColors.variant;
        });
      }, targetTheme);
    },
    [resourceMutator]
  );

  const _setPrimaryButtonLabelColor = useCallback(
    (color: CSSColor | undefined, targetTheme: Theme) => {
      resourceMutator.updateCustomisableTheme((prev) => {
        return produce(prev, (draft) => {
          draft.primaryButton.labelColor = color;
        });
      }, targetTheme);
    },
    [resourceMutator]
  );

  const _setIconColor = useCallback(
    (color: CSSColor | undefined, targetTheme: Theme) => {
      resourceMutator.updateCustomisableTheme((prev) => {
        return produce(prev, (draft) => {
          if (color == null) {
            draft.icon.color = undefined;
            return;
          }
          draft.icon.color = color;
        });
      }, targetTheme);
    },
    [resourceMutator]
  );

  const _setLinkColor = useCallback(
    (color: CSSColor | undefined, targetTheme: Theme) => {
      resourceMutator.updateCustomisableTheme((prev) => {
        return produce(prev, (draft) => {
          if (color == null) {
            draft.link.color = undefined;
            draft.link.colorActive = undefined;
            draft.link.colorHover = undefined;
            return;
          }
          draft.link.color = color;

          const derivedColors = deriveColors(color);
          if (derivedColors == null) {
            draft.link.color = undefined;
            draft.link.colorActive = undefined;
            draft.link.colorHover = undefined;
            return;
          }
          draft.link.colorActive = derivedColors.variant;
          draft.link.colorHover = derivedColors.variant;
        });
      }, targetTheme);
    },
    [resourceMutator]
  );

  const designForm = useMemo(
    (): BranchDesignForm => ({
      isLoading: or_(
        configForm.isLoading,
        resourceForm.isLoading,
        backgroundImageResourceForm.isLoading
      ),
      isUpdating: or_(
        configForm.isUpdating,
        resourceForm.isUpdating,
        backgroundImageResourceForm.isUpdating
      ),
      isDirty: or_(
        configForm.isDirty,
        resourceForm.isDirty,
        backgroundImageResourceForm.isDirty
      ),
      loadError: nullishCoalesce(
        configForm.loadError,
        resourceForm.loadError,
        backgroundImageResourceForm.loadError
      ),
      updateError: nullishCoalesce(
        configForm.updateError,
        resourceForm.updateError,
        backgroundImageResourceForm.updateError
      ),
      state,
      reload: () => {
        configForm.reload();
        resourceForm.reload();
        backgroundImageResourceForm.reload();
      },
      reset: () => {
        configForm.reset();
        resourceForm.reset();
        backgroundImageResourceForm.reset();
      },
      save: async (ignoreConflict: boolean = false) => {
        await configForm.save(ignoreConflict);
        await resourceForm.save(ignoreConflict);
        await backgroundImageResourceForm.save(ignoreConflict);
      },
      errorRules,

      setSelectedLanguage,

      setSelectedTheme,

      setThemeOption: (themeOption: ThemeOption) => {
        configForm.setState((prev) => {
          return produce(prev, (draft) => {
            draft.themeOption = themeOption;
          });
        });
      },

      setAppName: (appName: string) => {
        resourceMutator.setTranslationValue(TranslationKey.AppName, appName);
      },
      setFavicon: (image) => {
        resourceMutator.setImage(RESOURCE_FAVICON, image);
      },
      setCardAlignment: (alignment: Alignment) => {
        resourceMutator.updateCustomisableTheme((prev) => {
          return produce(prev, (draft) => {
            draft.card.alignment = alignment;
          });
        }, Theme.Light);
      },
      setPrimaryButtonBorderRadiusStyle: (
        borderRadiusStyle: BorderRadiusStyle | undefined
      ) => {
        resourceMutator.updateCustomisableTheme((prev) => {
          return produce(prev, (draft) => {
            draft.primaryButton.borderRadius = borderRadiusStyle;
            // NOTE: DEV-1541 Apply border radius to secondary button as well
            draft.secondaryButton.borderRadius = borderRadiusStyle;
          });
        }, Theme.Light);
      },
      setInputFieldBorderRadiusStyle: (
        borderRadiusStyle: BorderRadiusStyle | undefined
      ) => {
        resourceMutator.updateCustomisableTheme((prev) => {
          return produce(prev, (draft) => {
            draft.inputField.borderRadius = borderRadiusStyle;
            draft.phoneInputField.borderRadius = borderRadiusStyle;
          });
        }, Theme.Light);
      },
      setLinkTextDecorationStyle: (
        textDecoration: TextDecorationType | undefined
      ) => {
        resourceMutator.updateCustomisableTheme((prev) => {
          return produce(prev, (draft) => {
            draft.link.textDecoration = textDecoration;
          });
        }, Theme.Light);
      },
      setPrivacyPolicyLink: (link: string) => {
        resourceMutator.setTranslationValue(TranslationKey.PrivacyPolicy, link);
      },
      setTermsOfServiceLink: (link: string) => {
        resourceMutator.setTranslationValue(
          TranslationKey.TermsOfService,
          link
        );
      },
      setCustomerSupportLink: (link: string) => {
        resourceMutator.setTranslationValue(
          TranslationKey.CustomerSupport,
          link
        );
      },
      setDisplayAuthgearLogo: (visible: boolean) => {
        configForm.setState((prev) => {
          return produce(prev, (draft) => {
            draft.showAuthgearLogo = visible;
          });
        });
      },
      setDefaultClientURI: (uri: string) => {
        configForm.setState((prev) => {
          return produce(prev, (draft) => {
            draft.defaultClientURI = uri;
          });
        });
      },
      lightThemeSetters: {
        setAppLogo: (image) => _setAppLogo(image, Theme.Light),
        setLogoHeight: (height) => _setLogoHeight(height, Theme.Light),
        setBackgroundColor: (color) => _setBackgroundColor(color, Theme.Light),
        setBackgroundImage: (image) => _setBackgroundImage(image, Theme.Light),
        setPrimaryButtonBackgroundColor: (color) =>
          _setPrimaryButtonBackgroundColor(color, Theme.Light),
        setPrimaryButtonLabelColor: (color) =>
          _setPrimaryButtonLabelColor(color, Theme.Light),
        setIconColor: (color) => _setIconColor(color, Theme.Light),
        setLinkColor: (color) => _setLinkColor(color, Theme.Light),
      },
      darkThemeSetters: {
        setAppLogo: (image) => _setAppLogo(image, Theme.Dark),
        setLogoHeight: (height) => _setLogoHeight(height, Theme.Dark),
        setBackgroundColor: (color) => _setBackgroundColor(color, Theme.Dark),
        setBackgroundImage: (image) => _setBackgroundImage(image, Theme.Dark),
        setPrimaryButtonBackgroundColor: (color) =>
          _setPrimaryButtonBackgroundColor(color, Theme.Dark),
        setPrimaryButtonLabelColor: (color) =>
          _setPrimaryButtonLabelColor(color, Theme.Dark),
        setIconColor: (color) => _setIconColor(color, Theme.Dark),
        setLinkColor: (color) => _setLinkColor(color, Theme.Dark),
      },
    }),
    [
      configForm,
      resourceForm,
      backgroundImageResourceForm,
      state,
      errorRules,
      resourceMutator,
      _setAppLogo,
      _setLogoHeight,
      _setBackgroundColor,
      _setBackgroundImage,
      _setPrimaryButtonBackgroundColor,
      _setPrimaryButtonLabelColor,
      _setIconColor,
      _setLinkColor,
    ]
  );

  return designForm;
}
