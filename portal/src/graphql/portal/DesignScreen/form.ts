import { useMemo, useState } from "react";
import { parse as parseCSS } from "postcss";
import { produce } from "immer";
import { useResourceForm } from "../../../hook/useResourceForm";
import {
  Alignment,
  BorderRadiusStyle,
  CSSColor,
  CardAlignedSideDefaultSpacing,
  CssAstVisitor,
  CustomisableTheme,
  CustomisableThemeStyleGroup,
  DEFAULT_LIGHT_THEME,
  StyleCssVisitor,
  Theme,
  ThemeTargetSelector,
} from "../../../model/themeAuthFlowV2";
import {
  RESOURCE_APP_BACKGROUND_IMAGE,
  RESOURCE_APP_LOGO,
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

const LOCALE_BASED_RESOUCE_DEFINITIONS = [
  RESOURCE_TRANSLATION_JSON,
  RESOURCE_APP_LOGO,
  RESOURCE_FAVICON,
  RESOURCE_APP_BACKGROUND_IMAGE,
];

const THEME_RESOURCE_DEFINITIONS = [
  RESOURCE_AUTHGEAR_AUTHFLOW_V2_LIGHT_THEME_CSS,
];

const LightThemeResourceSpecifier = {
  def: RESOURCE_AUTHGEAR_AUTHFLOW_V2_LIGHT_THEME_CSS,
  locale: null,
  extension: null,
};

interface ConfigFormState {
  supportedLanguages: LanguageTag[];
  fallbackLanguage: LanguageTag;
  showAuthgearLogo: boolean;
}

interface ResourcesFormState {
  appName: string;
  appLogoBase64EncodedData: string | null;
  faviconBase64EncodedData: string | null;
  backgroundImageBase64EncodedData: string | null;
  customisableTheme: CustomisableTheme;

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
  theme: Theme;
} & ConfigFormState &
  ResourcesFormState &
  FeatureConfig;

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

  setAppName: (appName: string) => void;
  setAppLogo: (
    image: { base64EncodedData: string; extension: string } | null
  ) => void;
  setFavicon: (
    image: { base64EncodedData: string; extension: string } | null
  ) => void;
  setCardAlignment: (alignment: Alignment) => void;
  setBackgroundColor: (color: CSSColor) => void;
  setBackgroundImage: (
    image: { base64EncodedData: string; extension: string } | null
  ) => void;
  setPrimaryButtonBackgroundColor: (color: CSSColor) => void;
  setPrimaryButtonLabelColor: (color: CSSColor) => void;
  setPrimaryButtonBorderRadiusStyle: (
    borderRadiusStyle: BorderRadiusStyle
  ) => void;
  setLinkColor: (color: CSSColor) => void;
  setInputFieldBorderRadiusStyle: (
    borderRadiusStyle: BorderRadiusStyle
  ) => void;

  setPrivacyPolicyLink: (url: string) => void;
  setTermsOfServiceLink: (url: string) => void;
  setCustomerSupportLink: (url: string) => void;

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

export function useBrandDesignForm(appID: string): BranchDesignForm {
  const featureConfig = useAppFeatureConfigQuery(appID);
  const configForm = useAppConfigForm({
    appID,
    constructFormState: constructConfigFormState,
    constructConfig: constructConfigFromFormState,
  });
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

  const resourceForm = useResourceForm(appID, specifiers);

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
      def: ResourceDefinition
    ): string | null => {
      const specifiers = expandDef(def, selectedLanguage);
      const imageResouece = resolveResource(
        resourceForm.state.resources,
        specifiers
      );
      if (!imageResouece?.nullableValue) {
        return null;
      }
      return imageResouece.nullableValue;
    };

    const lightTheme = (() => {
      const lightThemeResource =
        resourceForm.state.resources[specifierId(LightThemeResourceSpecifier)];
      if (lightThemeResource?.nullableValue == null) {
        return DEFAULT_LIGHT_THEME;
      }
      const root = parseCSS(lightThemeResource.nullableValue);
      const styleCSSVisitor = new StyleCssVisitor(
        ThemeTargetSelector.Light,
        new CustomisableThemeStyleGroup()
      );
      return styleCSSVisitor.getStyle(root);
    })();

    return {
      appName: getValueFromTranslationJSON(TranslationKey.AppName),
      appLogoBase64EncodedData: getValueFromImageResource(RESOURCE_APP_LOGO),
      faviconBase64EncodedData: getValueFromImageResource(RESOURCE_FAVICON),
      backgroundImageBase64EncodedData: getValueFromImageResource(
        RESOURCE_APP_BACKGROUND_IMAGE
      ),
      customisableTheme: lightTheme,

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
  }, [resourceForm, selectedLanguage, configForm.state.fallbackLanguage]);

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
        resourceForm.setState((prev) => {
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
        updater: (prev: CustomisableTheme) => CustomisableTheme
      ) => {
        const newState = updater(resourcesState.customisableTheme);
        resourceForm.setState((s) => {
          return produce(s, (draft) => {
            const lightThemeResourceSpecifier = {
              def: RESOURCE_AUTHGEAR_AUTHFLOW_V2_LIGHT_THEME_CSS,
              locale: null,
              extension: null,
            };
            const lightThemeResource = draft.resources[
              specifierId(lightThemeResourceSpecifier)
            ] ?? {
              specifier: lightThemeResourceSpecifier,
              path: expandSpecifier(lightThemeResourceSpecifier),
            };
            lightThemeResource.nullableValue = (() => {
              const cssAstVisitor = new CssAstVisitor(
                ThemeTargetSelector.Light
              );
              const styleGroup = new CustomisableThemeStyleGroup(newState);
              styleGroup.acceptCssAstVisitor(cssAstVisitor);
              return cssAstVisitor.getCSS().toResult().css;
            })();

            draft.resources[specifierId(lightThemeResourceSpecifier)] =
              lightThemeResource;
          });
        });
      },
    };
  }, [
    resourcesState,
    resourceForm,
    selectedLanguage,
    configForm.state.fallbackLanguage,
  ]);

  const state: BranchDesignFormState = useMemo(
    () => ({
      theme: Theme.Light,
      selectedLanguage,
      ...configForm.state,
      ...resourcesState,
      whiteLabelingDisabled:
        featureConfig.effectiveFeatureConfig?.ui?.white_labeling?.disabled ??
        false,
    }),
    [
      selectedLanguage,
      configForm.state,
      resourcesState,
      featureConfig.effectiveFeatureConfig?.ui?.white_labeling?.disabled,
    ]
  );

  const errorRules: ErrorParseRule[] = useMemo(
    () => [
      makeImageSizeTooLargeErrorRule(
        Object.values(resourceForm.state.resources).filter(nonNullable)
      ),
    ],
    [resourceForm.state.resources]
  );

  const designForm = useMemo(
    (): BranchDesignForm => ({
      isLoading: configForm.isLoading || resourceForm.isLoading,
      isUpdating: configForm.isUpdating || resourceForm.isUpdating,
      isDirty: configForm.isDirty || resourceForm.isDirty,
      loadError: configForm.loadError ?? resourceForm.loadError,
      updateError: configForm.updateError ?? resourceForm.updateError,
      state,
      reload: () => {
        configForm.reload();
        resourceForm.reload();
      },
      reset: () => {
        configForm.reset();
        resourceForm.reset();
      },
      save: async (ignoreConflict: boolean = false) => {
        await configForm.save(ignoreConflict);
        await resourceForm.save(ignoreConflict);
      },
      errorRules,

      setSelectedLanguage,

      setAppName: (appName: string) => {
        resourceMutator.setTranslationValue(TranslationKey.AppName, appName);
      },
      setAppLogo: (image) => {
        resourceMutator.setImage(RESOURCE_APP_LOGO, image);
      },
      setFavicon: (image) => {
        resourceMutator.setImage(RESOURCE_FAVICON, image);
      },
      setCardAlignment: (alignment: Alignment) => {
        resourceMutator.updateCustomisableTheme((prev) => {
          return produce(prev, (draft) => {
            draft.card.alignment = alignment;
            if (alignment === "center") {
              draft.card.leftMargin = "0";
              draft.card.rightMargin = "0";
            } else {
              draft.card.leftMargin = CardAlignedSideDefaultSpacing;
              draft.card.rightMargin = CardAlignedSideDefaultSpacing;
            }
          });
        });
      },
      setBackgroundColor: (backgroundColor: CSSColor) => {
        resourceMutator.updateCustomisableTheme((prev) => {
          return produce(prev, (draft) => {
            draft.page.backgroundColor = backgroundColor;
          });
        });
      },
      setBackgroundImage: (image) => {
        resourceMutator.setImage(RESOURCE_APP_BACKGROUND_IMAGE, image);
      },
      setPrimaryButtonBackgroundColor: (backgroundColor: CSSColor) => {
        resourceMutator.updateCustomisableTheme((prev) => {
          return produce(prev, (draft) => {
            draft.primaryButton.backgroundColor = backgroundColor;
          });
        });
      },
      setPrimaryButtonLabelColor: (color: CSSColor) => {
        resourceMutator.updateCustomisableTheme((prev) => {
          return produce(prev, (draft) => {
            draft.primaryButton.labelColor = color;
          });
        });
      },
      setPrimaryButtonBorderRadiusStyle: (
        borderRadiusStyle: BorderRadiusStyle
      ) => {
        resourceMutator.updateCustomisableTheme((prev) => {
          return produce(prev, (draft) => {
            draft.primaryButton.borderRadius = borderRadiusStyle;
          });
        });
      },
      setLinkColor: (color: CSSColor) => {
        resourceMutator.updateCustomisableTheme((prev) => {
          return produce(prev, (draft) => {
            draft.link.color = color;
          });
        });
      },
      setInputFieldBorderRadiusStyle: (
        borderRadiusStyle: BorderRadiusStyle
      ) => {
        resourceMutator.updateCustomisableTheme((prev) => {
          return produce(prev, (draft) => {
            draft.inputField.borderRadius = borderRadiusStyle;
          });
        });
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
    }),
    [state, configForm, resourceForm, resourceMutator, errorRules]
  );

  return designForm;
}
