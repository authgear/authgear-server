import { useMemo } from "react";
import { parse as parseCSS } from "postcss";
import { produce } from "immer";
import {
  ResourceFormModel,
  useResourceForm,
} from "../../../hook/useResourceForm";
import {
  Alignment,
  CssAstVisitor,
  CustomisableTheme,
  CustomisableThemeStyleGroup,
  DEFAULT_LIGHT_THEME,
  StyleCssVisitor,
  ThemeTargetSelector,
} from "../../../model/themeAuthFlowV2";
import { RESOURCE_AUTHGEAR_AUTHFLOW_V2_LIGHT_THEME_CSS } from "../../../resources";
import {
  Resource,
  ResourceSpecifier,
  expandSpecifier,
} from "../../../util/resource";

const THEME_RESOURCE_DEFINITIONS = [
  RESOURCE_AUTHGEAR_AUTHFLOW_V2_LIGHT_THEME_CSS,
];

export interface BranchDesignFormState {
  orignalResources: Resource[];
  customisableTheme: CustomisableTheme;
}

export interface BranchDesignForm
  extends ResourceFormModel<BranchDesignFormState> {
  setCardAlignment: (alignment: Alignment) => void;
}

function constructResourcesFormStateFromResources(
  resources: Resource[]
): BranchDesignFormState {
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
  state: BranchDesignFormState
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

  const form = useResourceForm(
    appID,
    specifiers,
    constructResourcesFormStateFromResources,
    constructResourcesFromFormState
  );

  const designForm = useMemo(
    () => ({
      ...form,
      setCardAlignment: (alignment: Alignment) => {
        form.setState((prev) => {
          return produce(prev, (draft) => {
            draft.customisableTheme.cardAlignment = alignment;
          });
        });
      },
    }),

    [form]
  );

  return designForm;
}
