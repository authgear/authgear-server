import React, { useMemo } from "react";
import { useParams } from "react-router-dom";
import { useResourceForm } from "../../hook/useResourceForm";
import { RESOURCE_AUTHGEAR_AUTHFLOW_V2_LIGHT_THEME_CSS } from "../../resources";
import {
  Resource,
  ResourceSpecifier,
  expandSpecifier,
} from "../../util/resource";
import { parse as parseCSS } from "postcss";
import {
  CssAstVisitor,
  CustomisableTheme,
  CustomisableThemeStyleGroup,
  DEFAULT_LIGHT_THEME,
  StyleCssVisitor,
  ThemeTargetSelector,
} from "../../model/themeAuthFlowV2";
import FormContainer from "../../FormContainer";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";

const THEME_RESOURCE_DEFINITIONS = [
  RESOURCE_AUTHGEAR_AUTHFLOW_V2_LIGHT_THEME_CSS,
];

interface ResourcesFormState {
  orignalResources: Resource[];
  customisableTheme: CustomisableTheme;
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

const DesignScreen: React.VFC = function DesignScreen() {
  const { appID } = useParams() as { appID: string };
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

  if (form.isLoading) {
    return <ShowLoading />;
  }

  if (form.loadError) {
    return <ShowError error={form.loadError} onRetry={form.reload} />;
  }

  return <FormContainer form={form} canSave={true}></FormContainer>;
};

export default DesignScreen;
