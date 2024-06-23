import { useMemo, useState } from "react";
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
  RESOURCE_AUTHGEAR_AUTHFLOW_V2_LIGHT_THEME_CSS,
  RESOURCE_TRANSLATION_JSON,
} from "../../../resources";
import {
  LanguageTag,
  Resource,
  ResourceSpecifier,
  expandDef,
  expandSpecifier,
  specifierId,
} from "../../../util/resource";
import { useAppConfigForm } from "../../../hook/useAppConfigForm";
import { PortalAPIAppConfig } from "../../../types";

const LOCALE_BASED_RESOUCE_DEFINITIONS = [RESOURCE_TRANSLATION_JSON];

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

  setSelectedLanguage: (lang: LanguageTag) => void;

  setAppName: (appName: string) => void;

  setCardAlignment: (alignment: Alignment) => void;
  setBackgroundColor: (color: string) => void;
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
      customisableTheme: lightTheme,
    };
  }, [resourceForm, selectedLanguage, configForm.state.fallbackLanguage]);

  const resourceMutator = useMemo(() => {
    return {
      setTransalationValue: (key: string, value: string) => {
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

      setSelectedLanguage,

      setAppName: (appName: string) => {
        resourceMutator.setTransalationValue("app.name", appName);
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
    [state, configForm, resourceForm, resourceMutator]
  );

  return designForm;
}
