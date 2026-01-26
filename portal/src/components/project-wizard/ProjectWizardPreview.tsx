import React, { useCallback, useEffect, useMemo, useRef } from "react";
import cn from "classnames";
import { useFormContainerBaseContext } from "../../FormContainerBase";
import {
  FormState,
  ProjectWizardFormModel,
} from "../../screens/v2/ProjectWizard/form";
import { useSystemConfig } from "../../context/SystemConfigContext";
import { PreviewCustomisationMessage } from "../../model/preview";
import {
  CssAstVisitor,
  CustomisableThemeStyleGroup,
  DEFAULT_LIGHT_THEME,
  PartialCustomisableTheme,
  Theme,
  getThemeTargetSelector,
} from "../../model/themeAuthFlowV2";
import { TranslationKey } from "../../model/translations";
import { deriveColors } from "../../util/theme";

interface ProjectWizardPreviewProps {
  className?: string;
}

export function ProjectWizardPreview({
  className,
}: ProjectWizardPreviewProps): React.ReactElement {
  const { authgearEndpoint } = useSystemConfig();
  const authUIIframeRef = useRef<HTMLIFrameElement>(null);

  const { form } = useFormContainerBaseContext<ProjectWizardFormModel>();

  const src = useMemo(() => {
    const url = new URL(authgearEndpoint);
    url.pathname = "/noproject/preview/widget";
    return url.toString();
  }, [authgearEndpoint]);

  const targetOrigin = useMemo(() => {
    return new URL(src).origin;
  }, [src]);

  useEffect(() => {
    const message = mapProjectWizardFormStateToPreviewCustomisationMessage(
      form.state
    );
    authUIIframeRef.current?.contentWindow?.postMessage(message, targetOrigin);
  }, [form.state, targetOrigin]);

  const onLoadIframe = useCallback(() => {
    const message = mapProjectWizardFormStateToPreviewCustomisationMessage(
      form.state
    );
    authUIIframeRef.current?.contentWindow?.postMessage(message, targetOrigin);
  }, [form.state, targetOrigin]);

  return (
    <div className={cn(className, "flex-col flex")}>
      <iframe
        ref={authUIIframeRef}
        className={cn("w-full", "min-h-0", "flex-1", "border-none")}
        src={src}
        sandbox="allow-scripts"
        onLoad={onLoadIframe}
      ></iframe>
    </div>
  );
}

function mapProjectWizardFormStateToPreviewCustomisationMessage(
  state: FormState
): PreviewCustomisationMessage {
  const theme = Theme.Light;
  const cssAstVisitor = new CssAstVisitor(getThemeTargetSelector(theme));
  const defaultStyleGroup = new CustomisableThemeStyleGroup(
    DEFAULT_LIGHT_THEME
  );
  defaultStyleGroup.acceptCssAstVisitor(cssAstVisitor);
  const newTheme = mapFormStateToPartialCustomisableTheme(state);
  const styleGroup = new CustomisableThemeStyleGroup(newTheme);
  styleGroup.acceptCssAstVisitor(cssAstVisitor);
  const declarations = cssAstVisitor.getDeclarations();
  const cssVars: Record<string, string> = {};
  for (const declaration of declarations) {
    cssVars[declaration.prop] = declaration.value;
  }

  const images: Record<string, string | null> = {};

  images["brand-logo-light"] = state.logo
    ? `data:;base64,${state.logo.base64EncodedData}`
    : null;

  const translations = {
    [TranslationKey.AppName]: state.projectName,
  };
  return {
    type: "PreviewCustomisationMessage",
    theme,
    cssVars,
    images,
    translations,
    data: {
      previewWidgetLoginMethods: JSON.stringify(state.loginMethods),
    },
  };
}

function mapFormStateToPartialCustomisableTheme(
  state: FormState
): PartialCustomisableTheme {
  const newTheme: PartialCustomisableTheme = {
    page: {},
    card: {},
    primaryButton: {},
    secondaryButton: {},
    inputField: {},
    phoneInputField: {},
    icon: {},
    link: {},
    logo: {},
  };
  if (state.buttonAndLinkColor) {
    const color = state.buttonAndLinkColor;
    const derivedColors = deriveColors(color);
    newTheme.primaryButton.backgroundColor = color;
    newTheme.primaryButton.backgroundColorActive = derivedColors?.variant;
    newTheme.primaryButton.backgroundColorHover = derivedColors?.variant;

    newTheme.icon.color = color;

    newTheme.link.color = color;
    newTheme.link.colorActive = derivedColors?.variant;
    newTheme.link.colorHover = derivedColors?.variant;
  }
  if (state.buttonLabelColor) {
    newTheme.primaryButton.labelColor = state.buttonLabelColor;
  }
  return newTheme;
}
