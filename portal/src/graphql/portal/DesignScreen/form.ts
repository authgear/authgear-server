import { useCallback, useMemo, useState } from "react";
import { parse as parseCSS } from "postcss";
import { produce } from "immer";
import { useResourceForm } from "../../../hook/useResourceForm";
import {
  Alignment,
  BorderRadiusStyle,
  CssAstVisitor,
  CustomisableTheme,
  CustomisableThemeStyleGroup,
  DEFAULT_LIGHT_THEME,
  StyleCssVisitor,
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
import { APIError } from "../../../error/error";
import { ErrorParseRule, ErrorParseRuleResult } from "../../../error/parse";

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
}

interface ResourcesFormState {
  appName: string;
  appLogoBase64EncodedData: string | null;
  faviconBase64EncodedData: string | null;
  backgroundImageBase64EncodedData: string | null;
  customisableTheme: CustomisableTheme;
}

export type BranchDesignFormState = {
  selectedLanguage: LanguageTag;
} & ConfigFormState &
  ResourcesFormState;

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
  setBackgroundColor: (color: string) => void;
  setBackgroundImage: (
    image: { base64EncodedData: string; extension: string } | null
  ) => void;
  setPrimaryButtonBackgroundColor: (color: string) => void;
  setPrimaryButtonLabelColor: (color: string) => void;
  setPrimaryButtonBorderRadiusStyle: (
    borderRadiusStyle: BorderRadiusStyle
  ) => void;
  setLinkColor: (color: string) => void;
  setInputFieldBorderRadiusStyle: (
    borderRadiusStyle: BorderRadiusStyle
  ) => void;
}

function constructConfigFormState(config: PortalAPIAppConfig): ConfigFormState {
  const fallbackLanguage = config.localization?.fallback_language ?? "en";
  return {
    fallbackLanguage,
    supportedLanguages: config.localization?.supported_languages ?? [
      fallbackLanguage,
    ],
  };
}

function constructConfigFromFormState(
  config: PortalAPIAppConfig
): PortalAPIAppConfig {
  return config;
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

const ImageMaxSizeInKB = 100;

export function useBrandDesignForm(appID: string): BranchDesignForm {
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
      appName: getValueFromTranslationJSON("app.name"),
      appLogoBase64EncodedData: getValueFromImageResource(RESOURCE_APP_LOGO),
      faviconBase64EncodedData: getValueFromImageResource(RESOURCE_FAVICON),
      backgroundImageBase64EncodedData: getValueFromImageResource(
        RESOURCE_APP_BACKGROUND_IMAGE
      ),
      customisableTheme: lightTheme,
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
      selectedLanguage,
      ...configForm.state,
      ...resourcesState,
    }),
    [selectedLanguage, configForm.state, resourcesState]
  );

  const imageSizeTooLargeErrorRule = useCallback(
    (apiError: APIError): ErrorParseRuleResult => {
      if (apiError.reason === "RequestEntityTooLarge") {
        // When the request is blocked by the load balancer due to RequestEntityTooLarge
        // We try to get the largest resource from the state
        // and construct the error message for display

        let path = "";
        let longestLength = 0;
        // get the largest resources from the state
        for (const r of Object.keys(resourceForm.state.resources)) {
          const l = resourceForm.state.resources[r]?.nullableValue?.length ?? 0;
          if (l > longestLength) {
            longestLength = l;
            path = resourceForm.state.resources[r]?.path ?? "";
          }
        }

        // parse resource type from resource path
        let resourceType = "other";
        if (path !== "") {
          const dir = path.split("/");
          const fileName = dir[dir.length - 1];
          if (fileName.lastIndexOf(".") !== -1) {
            resourceType = fileName.slice(0, fileName.lastIndexOf("."));
          } else {
            resourceType = fileName;
          }
        }

        return {
          parsedAPIErrors: [
            {
              messageID: "errors.resource-too-large",
              arguments: {
                maxSize: ImageMaxSizeInKB,
                resourceType,
              },
            },
          ],
          fullyHandled: true,
        };
      }
      return {
        parsedAPIErrors: [],
        fullyHandled: false,
      };
    },
    [resourceForm.state.resources]
  );

  const errorRules: ErrorParseRule[] = useMemo(
    () => [imageSizeTooLargeErrorRule],
    [imageSizeTooLargeErrorRule]
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
        resourceMutator.setTranslationValue("app.name", appName);
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
            draft.cardAlignment = alignment;
          });
        });
      },
      setBackgroundColor: (backgroundColor: string) => {
        resourceMutator.updateCustomisableTheme((prev) => {
          return produce(prev, (draft) => {
            draft.backgroundColor = backgroundColor;
          });
        });
      },
      setBackgroundImage: (image) => {
        resourceMutator.setImage(RESOURCE_APP_BACKGROUND_IMAGE, image);
      },
      setPrimaryButtonBackgroundColor: (backgroundColor: string) => {
        resourceMutator.updateCustomisableTheme((prev) => {
          return produce(prev, (draft) => {
            draft.primaryButton.backgroundColor = backgroundColor;
          });
        });
      },
      setPrimaryButtonLabelColor: (color: string) => {
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
      setLinkColor: (color: string) => {
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
    }),
    [state, configForm, resourceForm, resourceMutator, errorRules]
  );

  return designForm;
}
