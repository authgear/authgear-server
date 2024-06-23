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
  Resource,
  ResourceSpecifier,
  expandSpecifier,
} from "../../../util/resource";
import { useAppConfigForm } from "../../../hook/useAppConfigForm";
import { PortalAPIAppConfig } from "../../../types";

const THEME_RESOURCE_DEFINITIONS = [
  RESOURCE_AUTHGEAR_AUTHFLOW_V2_LIGHT_THEME_CSS,
];

interface ConfigFormState {
  supportedLanguages: LanguageTag[];
  fallbackLanguage: LanguageTag;
}

interface ResourcesFormState {
  orignalResources: Resource[];
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

function constrcutConfigFromFormState(
  config: PortalAPIAppConfig
): PortalAPIAppConfig {
  return config;
}

function constructResourcesFormStateFromResources(
  resources: Resource[]
): ResourcesFormState {
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
    orignalResources: resources,
    customisableTheme: lightTheme,
  };
}

function constructResourcesFromFormState(
  state: ResourcesFormState
): Resource[] {
  const lightThemeResourceSpecifier = {
    def: RESOURCE_AUTHGEAR_AUTHFLOW_V2_LIGHT_THEME_CSS,
    locale: null,
    extension: null,
  };
  return [
    {
      specifier: lightThemeResourceSpecifier,
      path: expandSpecifier(lightThemeResourceSpecifier),
      nullableValue: (() => {
        const cssAstVisitor = new CssAstVisitor(ThemeTargetSelector.Light);
        const styleGroup = new CustomisableThemeStyleGroup(
          state.customisableTheme
        );
        styleGroup.acceptCssAstVisitor(cssAstVisitor);
        return cssAstVisitor.getCSS().toResult().css;
      })(),
    },
  ];
}

export function useBrandDesignForm(appID: string): BranchDesignForm {
  const configForm = useAppConfigForm({
    appID,
    constructFormState: constructConfigFormState,
    constructConfig: constrcutConfigFromFormState,
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
    return specifiers;
  }, []);

  const resourceForm = useResourceForm(
    appID,
    specifiers,
    constructResourcesFormStateFromResources,
    constructResourcesFromFormState
  );

  const state: BranchDesignFormState = useMemo(
    () => ({
      selectedLanguage,
      ...configForm.state,
      ...resourceForm.state,
    }),
    [selectedLanguage, configForm.state, resourceForm.state]
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
        resourceForm.setState((prev) => {
          return produce(prev, (draft) => {
            draft.customisableTheme.cardAlignment = alignment;
          });
        });
      },
      setBackgroundColor: (backgroundColor: string) => {
        resourceForm.setState((prev) => {
          return produce(prev, (draft) => {
            draft.customisableTheme.backgroundColor = backgroundColor;
          });
        });
      },
      setPrimaryButtonBackgroundColor: (backgroundColor: string) => {
        resourceForm.setState((prev) => {
          return produce(prev, (draft) => {
            draft.customisableTheme.primaryButton.backgroundColor =
              backgroundColor;
          });
        });
      },
      setPrimaryButtonLabelColor: (color: string) => {
        resourceForm.setState((prev) => {
          return produce(prev, (draft) => {
            draft.customisableTheme.primaryButton.labelColor = color;
          });
        });
      },
      setPrimaryButtonBorderRadiusStyle: (
        borderRadiusStyle: BorderRadiusStyle
      ) => {
        resourceForm.setState((prev) => {
          return produce(prev, (draft) => {
            draft.customisableTheme.primaryButton.borderRadius =
              borderRadiusStyle;
          });
        });
      },
      setLinkColor: (color: string) => {
        resourceForm.setState((prev) => {
          return produce(prev, (draft) => {
            draft.customisableTheme.link.color = color;
          });
        });
      },
      setInputFieldBorderRadiusStyle: (
        borderRadiusStyle: BorderRadiusStyle
      ) => {
        resourceForm.setState((prev) => {
          return produce(prev, (draft) => {
            draft.customisableTheme.inputField.borderRadius = borderRadiusStyle;
          });
        });
      },
    }),
    [state, configForm, resourceForm]
  );

  return designForm;
}
