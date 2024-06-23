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
import { RESOURCE_AUTHGEAR_AUTHFLOW_V2_LIGHT_THEME_CSS } from "../../../resources";
import {
  LanguageTag,
  ResourceSpecifier,
  expandSpecifier,
  specifierId,
} from "../../../util/resource";
import { useAppConfigForm } from "../../../hook/useAppConfigForm";
import { PortalAPIAppConfig } from "../../../types";
import { nonNullable } from "../../../util/types";

const LOCALE_BASED_RESOUCE_DEFINITIONS = [RESOURCE_TRANSLATION_JSON];

const THEME_RESOURCE_DEFINITIONS = [
  RESOURCE_AUTHGEAR_AUTHFLOW_V2_LIGHT_THEME_CSS,
];

interface ConfigFormState {
  supportedLanguages: LanguageTag[];
  fallbackLanguage: LanguageTag;
}

interface ResourcesFormState {
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
    const resources = Object.values(resourceForm.state.resources).filter(
      nonNullable
    );
    const lightTheme = (() => {
      const lightThemeResource = resources.find((r) => {
        return (
          r.nullableValue != null &&
          r.specifier.def === RESOURCE_AUTHGEAR_AUTHFLOW_V2_LIGHT_THEME_CSS
        );
      });
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
      customisableTheme: lightTheme,
    };
  }, [resourceForm]);

  const resourceMutator = useMemo(() => {
    return {
      updateCustomisableTheme: (
        updater: (prev: CustomisableTheme) => CustomisableTheme
      ) => {
        const newState = updater(resourcesState.customisableTheme);
        resourceForm.setState((s) => {
          return produce(s, (draft) => {
            const resources = Object.values(draft.resources).filter(
              nonNullable
            );
            const lightThemeResourceSpecifier = {
              def: RESOURCE_AUTHGEAR_AUTHFLOW_V2_LIGHT_THEME_CSS,
              locale: null,
              extension: null,
            };
            const lightThemeResource = resources.find((r) => {
              return (
                r.nullableValue != null &&
                r.specifier.def ===
                  RESOURCE_AUTHGEAR_AUTHFLOW_V2_LIGHT_THEME_CSS
              );
            }) ?? {
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
  }, [resourcesState, resourceForm]);

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
