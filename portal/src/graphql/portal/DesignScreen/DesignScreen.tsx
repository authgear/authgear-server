import React, {
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import {
  ChoiceGroup,
  DefaultEffects,
  Dropdown,
  IChoiceGroupOption,
  IDropdownOption,
  IDropdownStyleProps,
  IDropdownStyles,
  IStyleFunctionOrObject,
  Text,
} from "@fluentui/react";
import {
  Context as MFContext,
  FormattedMessage,
} from "@oursky/react-messageformat";
import cn from "classnames";

import { useParams } from "react-router-dom";
import FormContainer from "../../../FormContainer";
import ShowError from "../../../ShowError";
import ShowLoading from "../../../ShowLoading";
import {
  Alignment,
  AllAlignments,
  DEFAULT_DARK_THEME,
  DEFAULT_LIGHT_THEME,
  Theme,
} from "../../../model/themeAuthFlowV2";

import ScreenTitle from "../../../ScreenTitle";
import ManageLanguageWidget from "../ManageLanguageWidget";
import FormTextField from "../../../FormTextField";
import TextField from "../../../TextField";
import DefaultButton from "../../../DefaultButton";
import Toggle from "../../../Toggle";
import ConfigurationGroup from "../../../components/design/ConfigurationGroup";
import FallbackDescription from "../../../components/design/FallbackDescription";
import ConfigurationDescription from "../../../components/design/ConfigurationDescription";
import AppLogoPicker from "../../../components/design/AppLogoPicker";
import { ImagePicker } from "../../../components/design/ImagePicker";
import ButtonToggleGroup, {
  Option,
} from "../../../components/common/ButtonToggleGroup";
import Configuration from "../../../components/design/Configuration";
import { ColorPicker } from "../../../components/design/ColorPicker";
import BorderRadius from "../../../components/design/BorderRadius";
import TextDecoration from "../../../components/design/TextDecoration";
import Separator from "../../../components/design/Separator";

import { BranchDesignForm, useBrandDesignForm } from "./form";
import styles from "./DesignScreen.module.css";
import { useAppAndSecretConfigQuery } from "../query/appAndSecretConfigQuery";
import { PortalAPIAppConfig } from "../../../types";
import {
  PreviewPageType,
  getSupportedPreviewPagesFromConfig,
  mapDesignFormStateToPreviewCustomisationMessage,
} from "./viewModel";
import PrimaryButton from "../../../PrimaryButton";
import { useFormContainerBaseContext } from "../../../FormContainerBase";

interface OrganisationConfigurationProps {
  designForm: BranchDesignForm;
}
const OrganisationConfiguration: React.VFC<OrganisationConfigurationProps> =
  function OrganisationConfiguration(props) {
    const { designForm } = props;
    const { renderToString } = useContext(MFContext);
    const onChange = useCallback(
      (_: React.FormEvent<any>, value?: string) => {
        if (value == null) {
          return;
        }
        designForm.setAppName(value);
      },
      [designForm]
    );
    return (
      <ConfigurationGroup labelKey="DesignScreen.configuration.organisation.label">
        <TextField
          label={renderToString(
            "DesignScreen.configuration.organisation.name.label"
          )}
          value={designForm.state.appName}
          onChange={onChange}
        />
        {designForm.state.selectedLanguage !==
        designForm.state.fallbackLanguage ? (
          <FallbackDescription
            fallbackLanguage={designForm.state.fallbackLanguage}
          />
        ) : null}
      </ConfigurationGroup>
    );
  };

interface ThemeConfigurationProps {
  designForm: BranchDesignForm;
}

const ThemeConfiguration: React.VFC<ThemeConfigurationProps> =
  function ThemeConfiguration(props) {
    const { designForm } = props;
    const { renderToString } = useContext(MFContext);
    const onChange = useCallback(
      (_event, options?: IChoiceGroupOption) => {
        const value = options?.key;
        if (value !== "lightOnly" && value !== "darkOnly" && value !== "auto") {
          return;
        }
        designForm.setThemeOption(value);
        if (value === "lightOnly") {
          designForm.setSelectedTheme(Theme.Light);
        } else if (value === "darkOnly") {
          designForm.setSelectedTheme(Theme.Dark);
        } else {
          designForm.setSelectedTheme(
            window.matchMedia("(prefers-color-scheme: dark)").matches
              ? Theme.Dark
              : Theme.Light
          );
        }
      },
      [designForm]
    );
    const options: IChoiceGroupOption[] = useMemo(
      () => [
        {
          key: "lightOnly",
          text: renderToString("DesignScreen.configuration.theme.lightOnly"),
        },
        {
          key: "darkOnly",
          text: renderToString("DesignScreen.configuration.theme.darkOnly"),
        },
        {
          key: "auto",
          text: renderToString("DesignScreen.configuration.theme.auto"),
        },
      ],
      [renderToString]
    );
    return (
      <ConfigurationGroup labelKey="DesignScreen.configuration.theme.label">
        <ChoiceGroup
          selectedKey={designForm.state.themeOption}
          options={options}
          onChange={onChange}
        />
      </ConfigurationGroup>
    );
  };

interface AppLogoConfigurationProps {
  designForm: BranchDesignForm;
}
const AppLogoConfiguration: React.VFC<AppLogoConfigurationProps> =
  function AppLogoConfiguration(props) {
    const { designForm } = props;
    const { renderToString } = useContext(MFContext);
    const onLogoHeightChangeLight = useCallback(
      (_: React.FormEvent, value?: string) => {
        designForm.lightThemeSetters.setLogoHeight(value);
      },
      [designForm]
    );
    const onLogoHeightChangeDark = useCallback(
      (_: React.FormEvent, value?: string) => {
        designForm.darkThemeSetters.setLogoHeight(value);
      },
      [designForm]
    );
    return (
      <ConfigurationGroup labelKey="DesignScreen.configuration.logo.label">
        <ConfigurationDescription labelKey="DesignScreen.configuration.logo.description" />
        {designForm.state.themeOption !== "darkOnly" ? (
          <>
            <Configuration labelKey="DesignScreen.configuration.logo.light">
              <AppLogoPicker
                logo={designForm.state.appLogo}
                onChange={designForm.lightThemeSetters.setAppLogo}
              />
            </Configuration>
            {designForm.state.selectedLanguage !==
            designForm.state.fallbackLanguage ? (
              <FallbackDescription
                fallbackLanguage={designForm.state.fallbackLanguage}
              />
            ) : null}
            <TextField
              label={renderToString(
                "DesignScreen.configuration.logo.height.label.light"
              )}
              placeholder={DEFAULT_LIGHT_THEME.logo.height}
              value={designForm.state.customisableLightTheme.logo.height}
              onChange={onLogoHeightChangeLight}
            />
          </>
        ) : null}
        {designForm.state.themeOption !== "lightOnly" ? (
          <>
            <Configuration labelKey="DesignScreen.configuration.logo.dark">
              <AppLogoPicker
                logo={designForm.state.appLogoDark}
                onChange={designForm.darkThemeSetters.setAppLogo}
              />
            </Configuration>
            {designForm.state.selectedLanguage !==
            designForm.state.fallbackLanguage ? (
              <FallbackDescription
                fallbackLanguage={designForm.state.fallbackLanguage}
              />
            ) : null}
            <TextField
              label={renderToString(
                "DesignScreen.configuration.logo.height.label.dark"
              )}
              placeholder={DEFAULT_DARK_THEME.logo.height}
              value={designForm.state.customisableDarkTheme.logo.height}
              onChange={onLogoHeightChangeDark}
            />
          </>
        ) : null}
      </ConfigurationGroup>
    );
  };

interface FaviconConfigurationProps {
  designForm: BranchDesignForm;
}
const FaviconConfiguration: React.VFC<FaviconConfigurationProps> =
  function FaviconConfiguration(props) {
    const { designForm } = props;
    return (
      <ConfigurationGroup labelKey="DesignScreen.configuration.favicon.label">
        <ConfigurationDescription labelKey="DesignScreen.configuration.favicon.description" />
        <ImagePicker
          base64EncodedData={designForm.state.faviconBase64EncodedData}
          onChange={designForm.setFavicon}
        />
        {designForm.state.selectedLanguage !==
        designForm.state.fallbackLanguage ? (
          <FallbackDescription
            fallbackLanguage={designForm.state.fallbackLanguage}
          />
        ) : null}
      </ConfigurationGroup>
    );
  };

const AlignmentOptions = AllAlignments.map((value) => ({ value }));
interface AlignmentConfigurationProps {
  designForm: BranchDesignForm;
}
const AlignmentConfiguration: React.VFC<AlignmentConfigurationProps> =
  function AlignmentConfiguration(props) {
    const { designForm } = props;
    const onSelectOption = useCallback(
      (option: Option<Alignment>) => {
        designForm.setCardAlignment(option.value);
      },
      [designForm]
    );
    const renderOption = useCallback(
      (option: Option<Alignment>, selected: boolean) => {
        return (
          <span
            className={cn(
              styles.icAlignment,
              (() => {
                switch (option.value) {
                  case "start":
                    return styles.icAlignmentLeft;
                  case "center":
                    return styles.icAlignmentCenter;
                  case "end":
                    return styles.icAlignmentRight;
                  default:
                    return undefined;
                }
              })(),
              selected && styles.selected
            )}
          ></span>
        );
      },
      []
    );
    return (
      <ConfigurationGroup labelKey="DesignScreen.configuration.card.label">
        <Configuration labelKey="DesignScreen.configuration.card.alignment.label">
          <ButtonToggleGroup
            className={cn("mt-2")}
            value={
              designForm.state.customisableLightTheme.card.alignment ??
              DEFAULT_LIGHT_THEME.card.alignment
            }
            options={AlignmentOptions}
            onSelectOption={onSelectOption}
            renderOption={renderOption}
          ></ButtonToggleGroup>
        </Configuration>
      </ConfigurationGroup>
    );
  };

interface BackgroundConfigurationProps {
  designForm: BranchDesignForm;
}
const BackgroundConfiguration: React.VFC<BackgroundConfigurationProps> =
  function BackgroundConfiguration(props) {
    const { designForm } = props;
    return (
      <ConfigurationGroup labelKey="DesignScreen.configuration.background.label">
        <ConfigurationDescription labelKey="DesignScreen.configuration.background.description" />
        {designForm.state.themeOption !== "darkOnly" ? (
          <>
            <Configuration labelKey="DesignScreen.configuration.background.color.label.light">
              <ColorPicker
                color={
                  designForm.state.customisableLightTheme.page.backgroundColor
                }
                placeholderColor={DEFAULT_LIGHT_THEME.page.backgroundColor}
                onChange={designForm.lightThemeSetters.setBackgroundColor}
              />
            </Configuration>
          </>
        ) : null}
        {designForm.state.themeOption !== "lightOnly" ? (
          <>
            <Configuration labelKey="DesignScreen.configuration.background.color.label.dark">
              <ColorPicker
                color={
                  designForm.state.customisableDarkTheme.page.backgroundColor
                }
                placeholderColor={DEFAULT_DARK_THEME.page.backgroundColor}
                onChange={designForm.darkThemeSetters.setBackgroundColor}
              />
            </Configuration>
          </>
        ) : null}
        {designForm.state.themeOption !== "darkOnly" ? (
          <>
            <Configuration labelKey="DesignScreen.configuration.background.image.label.light">
              <ImagePicker
                base64EncodedData={
                  designForm.state.backgroundImageBase64EncodedData
                }
                onChange={designForm.lightThemeSetters.setBackgroundImage}
              />
            </Configuration>
          </>
        ) : null}
        {designForm.state.themeOption !== "lightOnly" ? (
          <>
            <Configuration labelKey="DesignScreen.configuration.background.image.label.dark">
              <ImagePicker
                base64EncodedData={
                  designForm.state.backgroundImageDarkBase64EncodedData
                }
                onChange={designForm.darkThemeSetters.setBackgroundImage}
              />
            </Configuration>
          </>
        ) : null}
      </ConfigurationGroup>
    );
  };

interface ButtonConfigurationProps {
  designForm: BranchDesignForm;
}
const ButtonConfiguration: React.VFC<ButtonConfigurationProps> =
  function ButtonConfiguration(props) {
    const { designForm } = props;
    return (
      <ConfigurationGroup labelKey="DesignScreen.configuration.button.label">
        {designForm.state.themeOption !== "darkOnly" ? (
          <>
            <Configuration labelKey="DesignScreen.configuration.button.primaryButton.label.light">
              <ColorPicker
                color={
                  designForm.state.customisableLightTheme.primaryButton
                    .backgroundColor
                }
                placeholderColor={
                  DEFAULT_LIGHT_THEME.primaryButton.backgroundColor
                }
                onChange={
                  designForm.lightThemeSetters.setPrimaryButtonBackgroundColor
                }
              />
            </Configuration>
          </>
        ) : null}
        {designForm.state.themeOption !== "lightOnly" ? (
          <>
            <Configuration labelKey="DesignScreen.configuration.button.primaryButton.label.dark">
              <ColorPicker
                color={
                  designForm.state.customisableDarkTheme.primaryButton
                    .backgroundColor
                }
                placeholderColor={
                  DEFAULT_DARK_THEME.primaryButton.backgroundColor
                }
                onChange={
                  designForm.darkThemeSetters.setPrimaryButtonBackgroundColor
                }
              />
            </Configuration>
          </>
        ) : null}

        {designForm.state.themeOption !== "darkOnly" ? (
          <>
            <Configuration labelKey="DesignScreen.configuration.button.primaryButtonLabel.label.light">
              <ColorPicker
                color={
                  designForm.state.customisableLightTheme.primaryButton
                    .labelColor
                }
                placeholderColor={DEFAULT_LIGHT_THEME.primaryButton.labelColor}
                onChange={
                  designForm.lightThemeSetters.setPrimaryButtonLabelColor
                }
              />
            </Configuration>
          </>
        ) : null}
        {designForm.state.themeOption !== "lightOnly" ? (
          <>
            <Configuration labelKey="DesignScreen.configuration.button.primaryButtonLabel.label.dark">
              <ColorPicker
                color={
                  designForm.state.customisableDarkTheme.primaryButton
                    .labelColor
                }
                placeholderColor={DEFAULT_DARK_THEME.primaryButton.labelColor}
                onChange={
                  designForm.darkThemeSetters.setPrimaryButtonLabelColor
                }
              />
            </Configuration>
          </>
        ) : null}
        <Configuration labelKey="DesignScreen.configuration.button.borderRadiusStyle.label">
          <BorderRadius
            value={
              designForm.state.customisableLightTheme.primaryButton
                .borderRadius ?? DEFAULT_LIGHT_THEME.primaryButton.borderRadius
            }
            onChange={designForm.setPrimaryButtonBorderRadiusStyle}
          />
        </Configuration>
      </ConfigurationGroup>
    );
  };

interface IconConfigurationProps {
  designForm: BranchDesignForm;
}
const IconConfiguration: React.VFC<IconConfigurationProps> =
  function IconConfiguration(props) {
    const { designForm } = props;
    return (
      <ConfigurationGroup labelKey="DesignScreen.configuration.icon.label">
        <ConfigurationDescription labelKey="DesignScreen.configuration.icon.description" />
        {designForm.state.themeOption !== "darkOnly" ? (
          <>
            <Configuration labelKey="DesignScreen.configuration.icon.color.label.light">
              <ColorPicker
                color={designForm.state.customisableLightTheme.icon.color}
                placeholderColor={DEFAULT_LIGHT_THEME.icon.color}
                onChange={designForm.lightThemeSetters.setIconColor}
              />
            </Configuration>
          </>
        ) : null}
        {designForm.state.themeOption !== "lightOnly" ? (
          <>
            <Configuration labelKey="DesignScreen.configuration.icon.color.label.dark">
              <ColorPicker
                color={designForm.state.customisableDarkTheme.icon.color}
                placeholderColor={DEFAULT_DARK_THEME.icon.color}
                onChange={designForm.darkThemeSetters.setIconColor}
              />
            </Configuration>
          </>
        ) : null}
      </ConfigurationGroup>
    );
  };

interface LinkConfigurationProps {
  designForm: BranchDesignForm;
}
const LinkConfiguration: React.VFC<LinkConfigurationProps> =
  function LinkConfiguration(props) {
    const { designForm } = props;
    const { renderToString } = useContext(MFContext);

    const onPrivacyPolicyLinkChange = useCallback(
      (_: React.FormEvent, value?: string) => {
        designForm.setPrivacyPolicyLink(value ?? "");
      },
      [designForm]
    );
    const onTermsOfServiceLinkChange = useCallback(
      (_: React.FormEvent, value?: string) => {
        designForm.setTermsOfServiceLink(value ?? "");
      },
      [designForm]
    );
    const onCustomerSupportLinkChange = useCallback(
      (_: React.FormEvent, value?: string) => {
        designForm.setCustomerSupportLink(value ?? "");
      },
      [designForm]
    );

    return (
      <ConfigurationGroup labelKey="DesignScreen.configuration.link.label">
        {designForm.state.themeOption !== "darkOnly" ? (
          <>
            <Configuration labelKey="DesignScreen.configuration.link.color.label.light">
              <ColorPicker
                color={designForm.state.customisableLightTheme.link.color}
                placeholderColor={DEFAULT_LIGHT_THEME.link.color}
                onChange={designForm.lightThemeSetters.setLinkColor}
              />
            </Configuration>
          </>
        ) : null}
        {designForm.state.themeOption !== "lightOnly" ? (
          <>
            <Configuration labelKey="DesignScreen.configuration.link.color.label.dark">
              <ColorPicker
                color={designForm.state.customisableDarkTheme.link.color}
                placeholderColor={DEFAULT_DARK_THEME.link.color}
                onChange={designForm.darkThemeSetters.setLinkColor}
              />
            </Configuration>
          </>
        ) : null}
        <Configuration labelKey="DesignScreen.configuration.link.textDecoration.label">
          <TextDecoration
            value={
              designForm.state.customisableLightTheme.link.textDecoration ??
              DEFAULT_LIGHT_THEME.link.textDecoration
            }
            onChange={designForm.setLinkTextDecorationStyle}
          />
        </Configuration>
        <Separator className={cn(styles.linkConfigurationSeparator)} />
        <TextField
          label={renderToString(
            "DesignScreen.configuration.link.urls.privacyPolicy.label"
          )}
          placeholder={renderToString(
            "DesignScreen.configuration.link.urls.privacyPolicy.placeholder"
          )}
          value={designForm.state.urls.privacyPolicy}
          onChange={onPrivacyPolicyLinkChange}
        />
        <TextField
          label={renderToString(
            "DesignScreen.configuration.link.urls.termsOfService.label"
          )}
          placeholder={renderToString(
            "DesignScreen.configuration.link.urls.termsOfService.placeholder"
          )}
          value={designForm.state.urls.termsOfService}
          onChange={onTermsOfServiceLinkChange}
        />
        <TextField
          label={renderToString(
            "DesignScreen.configuration.link.urls.customerSupport.label"
          )}
          placeholder={renderToString(
            "DesignScreen.configuration.link.urls.customerSupport.placeholder"
          )}
          value={designForm.state.urls.customerSupport}
          onChange={onCustomerSupportLinkChange}
        />
        <FallbackDescription
          fallbackLanguage={designForm.state.fallbackLanguage}
        />
      </ConfigurationGroup>
    );
  };

interface InputConfigurationProps {
  designForm: BranchDesignForm;
}
const InputConfiguration: React.VFC<InputConfigurationProps> =
  function InputConfiguration(props) {
    const { designForm } = props;
    return (
      <ConfigurationGroup labelKey="DesignScreen.configuration.input.label">
        <Configuration labelKey="DesignScreen.configuration.input.border.label">
          <BorderRadius
            value={
              designForm.state.customisableLightTheme.inputField.borderRadius ??
              DEFAULT_LIGHT_THEME.inputField.borderRadius
            }
            onChange={designForm.setInputFieldBorderRadiusStyle}
          />
        </Configuration>
      </ConfigurationGroup>
    );
  };

interface DefaultClientURIConfigurationProps {
  designForm: BranchDesignForm;
}
const DefaultClientURIConfiguration: React.VFC<DefaultClientURIConfigurationProps> =
  function DefaultClientURIConfiguration(props) {
    const { designForm } = props;
    const { renderToString } = useContext(MFContext);
    const [uri, setURI] = useState(() => designForm.state.defaultClientURI);
    const [enabled, setEnabled] = useState(
      () => designForm.state.defaultClientURI !== ""
    );
    const onChangeEnableClientURI = useCallback(
      (_: React.MouseEvent, checked?: boolean) => {
        const enabled_ = checked ?? false;
        if (enabled_) {
          designForm.setDefaultClientURI(uri);
        } else {
          designForm.setDefaultClientURI("");
        }
        setEnabled(enabled_);
      },
      [uri, designForm]
    );

    const onChangeURI = useCallback(
      (_: React.FormEvent<any>, value?: string) => {
        if (value == null) {
          return;
        }
        setURI(value);
        designForm.setDefaultClientURI(value);
      },
      [designForm]
    );

    return (
      <ConfigurationGroup labelKey="DesignScreen.configuration.defaultClientURI.label">
        <ConfigurationDescription labelKey="DesignScreen.configuration.defaultClientURI.description" />
        <Toggle
          checked={enabled}
          onChange={onChangeEnableClientURI}
          label={renderToString(
            "DesignScreen.configuration.defaultClientURI.enable.description"
          )}
          inlineLabel={true}
        />
        <FormTextField
          fieldName="default_client_uri"
          parentJSONPointer="/ui"
          disabled={!enabled}
          placeholder="https://example.com"
          value={uri}
          onChange={onChangeURI}
        />
      </ConfigurationGroup>
    );
  };

interface AuthgearBrandingConfigurationProps {
  appID: string;
  designForm: BranchDesignForm;
}
const AuthgearBrandingConfiguration: React.VFC<AuthgearBrandingConfigurationProps> =
  function AuthgearBrandingConfiguration(props) {
    const { appID, designForm } = props;
    const { renderToString } = useContext(MFContext);
    const onChangeDisableWatermark = useCallback(
      (_: React.MouseEvent, checked?: boolean) => {
        designForm.setDisplayAuthgearLogo(checked ?? true);
      },
      [designForm]
    );
    return (
      <ConfigurationGroup labelKey="DesignScreen.configuration.authgearBranding.label">
        {designForm.state.whiteLabelingDisabled ? (
          <div
            className={cn(
              "flex",
              "items-center",
              "p-4",
              "border",
              "border-solid",
              "border-neutral-quaternaryAlt"
            )}
          >
            <Text
              className={cn("leading-5", "font-semibold", "text-neutral-dark")}
            >
              <FormattedMessage id="DesignScreen.configuration.authgearBranding.upgradeToHide" />
            </Text>
            <DefaultButton
              className={cn(styles.upgradeNowButton, "ml-3", "flex-none")}
              href={`/project/${appID}/billing`}
              text={
                <FormattedMessage id="DesignScreen.configuration.authgearBranding.upgradeNow" />
              }
            />
          </div>
        ) : null}
        <Toggle
          checked={
            designForm.state.whiteLabelingDisabled ||
            designForm.state.showAuthgearLogo
          }
          onChange={onChangeDisableWatermark}
          label={renderToString(
            "DesignScreen.configuration.authgearBranding.disableAuthgearLogo.label"
          )}
          inlineLabel={true}
          disabled={designForm.state.whiteLabelingDisabled}
        />
      </ConfigurationGroup>
    );
  };

interface ConfigurationPanelProps {
  appID: string;
  designForm: BranchDesignForm;
}
const ConfigurationPanel: React.VFC<ConfigurationPanelProps> =
  function ConfigurationPanel(props) {
    const { appID, designForm } = props;
    return (
      <div>
        <OrganisationConfiguration designForm={designForm} />
        <Separator />
        <ThemeConfiguration designForm={designForm} />
        <Separator />
        <AppLogoConfiguration designForm={designForm} />
        <Separator />
        <FaviconConfiguration designForm={designForm} />
        <Separator />
        <AlignmentConfiguration designForm={designForm} />
        <Separator />
        <BackgroundConfiguration designForm={designForm} />
        <Separator />
        <ButtonConfiguration designForm={designForm} />
        <Separator />
        <IconConfiguration designForm={designForm} />
        <Separator />
        <LinkConfiguration designForm={designForm} />
        <Separator />
        <InputConfiguration designForm={designForm} />
        <Separator />
        <DefaultClientURIConfiguration designForm={designForm} />
        <Separator />
        <AuthgearBrandingConfiguration appID={appID} designForm={designForm} />
      </div>
    );
  };

const PreviewPageDropdownStyles: IStyleFunctionOrObject<
  IDropdownStyleProps,
  IDropdownStyles
> = {
  dropdown: {
    width: "180px",
    selectors: {
      "::after": {
        display: "none",
      },
    },
  },
  title: {
    border: "none",
    textAlign: "left",
  },
};

interface PreviewThemeToggleProps {
  activeTheme: Theme;
  setActiveTheme: (theme: Theme) => void;
  disabled: boolean;
}
const PreviewThemeToggleOptions = [Theme.Light, Theme.Dark].map((value) => ({
  value,
}));
const PreviewThemeToggle: React.VFC<PreviewThemeToggleProps> =
  function PreviewThemeToggle(props) {
    const { activeTheme, setActiveTheme, disabled } = props;
    const renderOption = useCallback(
      (option: Option<Theme>, selected: boolean) => {
        return (
          <span
            className={cn(
              styles.icTheme,
              (() => {
                switch (option.value) {
                  case Theme.Light:
                    return styles.icLightMode;
                  case Theme.Dark:
                    return styles.icDarkMode;
                  default:
                    return undefined;
                }
              })(),
              selected && styles.selected
            )}
          ></span>
        );
      },
      []
    );
    const onSelectOption = useCallback(
      (option: Option<Theme>) => {
        setActiveTheme(option.value);
      },
      [setActiveTheme]
    );
    return (
      <ButtonToggleGroup
        value={activeTheme}
        options={PreviewThemeToggleOptions}
        onSelectOption={onSelectOption}
        renderOption={renderOption}
        disabled={disabled}
        withBorder={false}
      ></ButtonToggleGroup>
    );
  };

interface PreviewProps {
  className?: string;
  effectiveAppConfig: PortalAPIAppConfig;
  designForm: BranchDesignForm;
}
const Preview: React.VFC<PreviewProps> = function Preview(props) {
  const { className, designForm, effectiveAppConfig } = props;
  const { renderToString } = useContext(MFContext);

  const authUIIframeRef = useRef<HTMLIFrameElement | null>(null);

  const [isIframeLoading, setIsIframeLoading] = useState(true);

  useEffect(() => {
    const message = mapDesignFormStateToPreviewCustomisationMessage(
      designForm.state
    );
    authUIIframeRef.current?.contentWindow?.postMessage(message, "*");
  }, [designForm.state]);

  const supportedPreviewPages = useMemo(
    () => getSupportedPreviewPagesFromConfig(effectiveAppConfig),
    [effectiveAppConfig]
  );

  const [selectedPreviewPage, setSelectedPreviewPage] = useState(
    () => supportedPreviewPages[0].screen
  );

  const previewPageOptions = useMemo((): IDropdownOption[] => {
    return supportedPreviewPages.map(
      (page): IDropdownOption => ({
        key: page.screen,
        text: renderToString(`DesignScreen.preview.pages.title.${page.key}`),
      })
    );
  }, [supportedPreviewPages, renderToString]);
  const onChangePreviewPageOption = useCallback(
    (_e: unknown, option?: IDropdownOption) => {
      if (option == null) {
        return;
      }
      setSelectedPreviewPage(option.key as PreviewPageType);
    },
    []
  );

  const src = useMemo(() => {
    const url = new URL(effectiveAppConfig.http?.public_origin ?? "");
    url.pathname = selectedPreviewPage;
    url.searchParams.append("ui_locales", designForm.state.selectedLanguage);
    return url.toString();
  }, [
    effectiveAppConfig.http?.public_origin,
    designForm.state.selectedLanguage,
    selectedPreviewPage,
  ]);

  useEffect(() => {
    setIsIframeLoading(true);
  }, [src]);

  const onLoadIframe = useCallback(() => {
    const message = mapDesignFormStateToPreviewCustomisationMessage(
      designForm.state
    );
    setIsIframeLoading(false);
    authUIIframeRef.current?.contentWindow?.postMessage(message, "*");
  }, [designForm.state]);

  return (
    <div className={cn("flex", "flex-col", className)}>
      <div
        className={cn(
          "flex",
          "justify-between",
          "content-center",
          "px-6",
          "py-1",
          "border-x-0",
          "border-t-0",
          "border-b",
          "border-solid",
          "border-neutral-light"
        )}
      >
        <Dropdown
          className={styles.previewDropdown}
          styles={PreviewPageDropdownStyles}
          selectedKey={selectedPreviewPage}
          options={previewPageOptions}
          onChange={onChangePreviewPageOption}
        />
        {designForm.state.themeOption === "auto" ? (
          <PreviewThemeToggle
            activeTheme={designForm.state.selectedTheme}
            setActiveTheme={designForm.setSelectedTheme}
            disabled={false}
          />
        ) : null}
      </div>
      {isIframeLoading ? <ShowLoading /> : null}
      <iframe
        ref={authUIIframeRef}
        className={cn("w-full", "min-h-0", "flex-1", "border-none")}
        src={src}
        sandbox="allow-scripts"
        onLoad={onLoadIframe}
      ></iframe>
    </div>
  );
};

interface DesignScreenContentProps {
  appID: string;
  effectiveAppConfig: PortalAPIAppConfig;
  form: BranchDesignForm;
}
const DesignScreenContent: React.VFC<DesignScreenContentProps> =
  function DesignScreenContent(props) {
    const { appID, effectiveAppConfig, form } = props;
    const { canSave, onSave } = useFormContainerBaseContext();
    const { renderToString } = useContext(MFContext);
    return (
      <>
        <div
          className={cn(
            "pt-6",
            "px-6",
            "flex",
            "items-center",
            "justify-between"
          )}
        >
          <ScreenTitle>
            <FormattedMessage id="DesignScreen.title" />
          </ScreenTitle>
          <div className={styles.titleActions}>
            <ManageLanguageWidget
              existingLanguages={form.state.supportedLanguages}
              supportedLanguages={form.state.supportedLanguages}
              selectedLanguage={form.state.selectedLanguage}
              fallbackLanguage={form.state.fallbackLanguage}
              onChangeSelectedLanguage={form.setSelectedLanguage}
            />
            <PrimaryButton
              text={renderToString("save")}
              disabled={!canSave}
              onClick={onSave}
            />
          </div>
        </div>
        <div
          className={cn(
            "min-h-0",
            "flex-1",
            "flex",
            "flex-row-reverse",
            "tablet:flex-col",
            "tablet:overflow-auto"
          )}
        >
          <div className={cn("p-6", "pt-4", "desktop:overflow-auto")}>
            <div className={cn("desktop:w-80")}>
              <ConfigurationPanel appID={appID} designForm={form} />
            </div>
          </div>
          <div className={cn("desktop:flex-1", "h-full", "p-6", "pt-4")}>
            <div
              className={cn(
                "rounded-xl",
                "h-full",
                "tablet:h-178.5",
                "overflow-hidden"
              )}
              style={{
                boxShadow: DefaultEffects.elevation4,
              }}
            >
              <Preview
                className={cn("h-full")}
                effectiveAppConfig={effectiveAppConfig}
                designForm={form}
              />
            </div>
          </div>
        </div>
      </>
    );
  };

const DesignScreen: React.VFC = function DesignScreen() {
  const { appID } = useParams() as { appID: string };
  const {
    effectiveAppConfig,
    loading: appConfigLoading,
    error: appConfigError,
    refetch: reloadConfig,
  } = useAppAndSecretConfigQuery(appID);
  const form = useBrandDesignForm(appID);

  const reloadData = useCallback(() => {
    form.reload();
    reloadConfig().catch((error) => {
      console.error(error);
    });
  }, [form, reloadConfig]);

  useEffect(() => {
    const onChange = (ev: MediaQueryListEvent) => {
      if (form.state.themeOption === "auto") {
        form.setSelectedTheme(ev.matches ? Theme.Dark : Theme.Light);
      }
    };
    const watcher = window.matchMedia("(prefers-color-scheme: dark)");
    watcher.addEventListener("change", onChange);
    return () => {
      watcher.removeEventListener("change", onChange);
    };
  }, [form, form.state.themeOption]);

  const isLoading =
    form.isLoading || appConfigLoading || effectiveAppConfig == null;
  if (isLoading) {
    return <ShowLoading />;
  }

  const loadError = form.loadError ?? appConfigError;
  if (loadError != null) {
    return <ShowError error={loadError} onRetry={reloadData} />;
  }

  return (
    <FormContainer
      className={cn("h-full", "flex", "flex-col")}
      form={form}
      canSave={true}
      errorRules={form.errorRules}
      stickyFooterComponent={true}
      hideFooterComponent={true}
    >
      <DesignScreenContent
        appID={appID}
        effectiveAppConfig={effectiveAppConfig}
        form={form}
      />
    </FormContainer>
  );
};

export default DesignScreen;
