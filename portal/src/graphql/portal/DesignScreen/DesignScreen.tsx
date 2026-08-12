import React, {
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import { ExternalLinkIcon, MoonIcon, SunIcon } from "@radix-ui/react-icons";
import {
  Button,
  Heading,
  Select,
  SegmentedControl,
  Tabs,
  Text,
} from "@radix-ui/themes";
import { Context as MFContext, FormattedMessage } from "../../../intl";
import cn from "classnames";

import { useParams } from "react-router-dom";
import FormContainer from "../../../FormContainer";
import ShowError from "../../../ShowError";
import ShowLoading from "../../../ShowLoading";
import {
  Alignment,
  DEFAULT_DARK_THEME,
  DEFAULT_LIGHT_THEME,
  Theme,
} from "../../../model/themeAuthFlowV2";

import ConfigurationGroup from "../../../components/design/ConfigurationGroup";
import FallbackDescription from "../../../components/design/FallbackDescription";
import ConfigurationDescription from "../../../components/design/ConfigurationDescription";
import AppLogoPicker from "../../../components/design/AppLogoPicker";
import { ImagePicker } from "../../../components/design/ImagePicker";
import Configuration from "../../../components/design/Configuration";
import { ColorPicker } from "../../../components/design/ColorPicker";
import BorderRadius from "../../../components/design/BorderRadius";
import TextDecoration from "../../../components/design/TextDecoration";
import Separator from "../../../components/design/Separator";
import { FormField } from "../../../components/v2/FormField/FormField";
import { TextField } from "../../../components/v2/TextField/TextField";
import { Toggle } from "../../../components/v2/Toggle/Toggle";
import { SecondaryButton } from "../../../components/v2/Button/SecondaryButton/SecondaryButton";
import { SaveFunctionBar } from "../../../components/v2/SaveFunctionBar/SaveFunctionBar";

import { BranchDesignForm, useBrandDesignForm } from "./form";
import styles from "./DesignScreen.module.css";
import { useAppAndSecretConfigQuery } from "../query/appAndSecretConfigQuery";
import { PortalAPIAppConfig } from "../../../types";
import {
  getSupportedPreviewPagesFromConfig,
  mapDesignFormStateToPreviewCustomisationMessage,
} from "./viewModel";
import { useFormContainerBaseContext } from "../../../FormContainerBase";
import AppLogoHeightSetter from "../../../components/design/AppLogoHeightSetter";
import { useTester } from "../../../hook/tester";
import { Tooltip } from "../../../components/v2/Tooltip/Tooltip";
import Link from "../../../Link";

interface OrganisationConfigurationProps {
  designForm: BranchDesignForm;
}
const OrganisationConfiguration: React.VFC<OrganisationConfigurationProps> =
  function OrganisationConfiguration(props) {
    const { designForm } = props;
    const { renderToString } = useContext(MFContext);
    const onChange = useCallback(
      (e: React.ChangeEvent<HTMLInputElement>) => {
        designForm.setAppName(e.target.value);
      },
      [designForm]
    );
    return (
      <ConfigurationGroup
        labelKey="DesignScreen.configuration.organisation.label"
        collapsible={true}
        defaultOpen={true}
      >
        <TextField
          size="2"
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

function useThemeOptionChange(designForm: BranchDesignForm) {
  return useCallback(
    (value: string) => {
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
}

interface AppearanceConfigurationProps {
  designForm: BranchDesignForm;
}

const AppearanceConfiguration: React.VFC<AppearanceConfigurationProps> =
  function AppearanceConfiguration(props) {
    const { designForm } = props;
    const { renderToString } = useContext(MFContext);
    const onValueChange = useThemeOptionChange(designForm);

    return (
      <FormField
        size="2"
        labelSpace="1"
        label={
          <FormattedMessage id="DesignScreen.configuration.appearance.label" />
        }
      >
        <Select.Root
          value={designForm.state.themeOption}
          onValueChange={onValueChange}
          size="2"
        >
          <Select.Trigger
            variant="surface"
            className={styles.appearanceSelect}
          />
          <Select.Content>
            <Select.Item value="auto">
              {renderToString(
                "DesignScreen.configuration.theme.autoAppearance"
              )}
            </Select.Item>
            <Select.Item value="lightOnly">
              {renderToString("DesignScreen.configuration.theme.lightOnly")}
            </Select.Item>
            <Select.Item value="darkOnly">
              {renderToString("DesignScreen.configuration.theme.darkOnly")}
            </Select.Item>
          </Select.Content>
        </Select.Root>
      </FormField>
    );
  };

interface AppLogoConfigurationProps {
  designForm: BranchDesignForm;
}
const AppLogoConfiguration: React.VFC<AppLogoConfigurationProps> =
  function AppLogoConfiguration(props) {
    const { designForm } = props;
    return (
      <ConfigurationGroup
        labelKey="DesignScreen.configuration.logo.label"
        collapsible={true}
      >
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
          </>
        ) : null}
        {designForm.state.themeOption !== "darkOnly" ? (
          <AppLogoHeightSetter
            sliderAriaLabel="light-logo-slider"
            value={
              designForm.state.customisableLightTheme.logo.height ??
              DEFAULT_LIGHT_THEME.logo.height
            }
            defaultValue={DEFAULT_LIGHT_THEME.logo.height}
            onChange={designForm.lightThemeSetters.setLogoHeight}
            labelKey={
              designForm.state.themeOption === "lightOnly"
                ? "DesignScreen.configuration.logo.height.label"
                : "DesignScreen.configuration.logo.height.label.light"
            }
          />
        ) : null}
        {designForm.state.themeOption !== "lightOnly" ? (
          <AppLogoHeightSetter
            sliderAriaLabel="dark-logo-slider"
            value={
              designForm.state.customisableDarkTheme.logo.height ??
              DEFAULT_DARK_THEME.logo.height
            }
            defaultValue={DEFAULT_DARK_THEME.logo.height}
            onChange={designForm.darkThemeSetters.setLogoHeight}
            labelKey={
              designForm.state.themeOption === "darkOnly"
                ? "DesignScreen.configuration.logo.height.label"
                : "DesignScreen.configuration.logo.height.label.dark"
            }
          />
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
      <ConfigurationGroup
        labelKey="DesignScreen.configuration.favicon.label"
        collapsible={true}
      >
        <ImagePicker
          sizeLimitInBytes={100 * 1000}
          base64EncodedData={designForm.state.faviconBase64EncodedData}
          descriptionKey="DesignScreen.configuration.favicon.uploadDescription"
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

const AlignmentIcon: React.VFC<{ alignment: Alignment }> =
  function AlignmentIcon(props) {
    const { alignment } = props;
    return (
      <span
        className={cn(
          styles.icAlignment,
          alignment === "start" && styles.icAlignmentLeft,
          alignment === "center" && styles.icAlignmentCenter,
          alignment === "end" && styles.icAlignmentRight
        )}
        aria-hidden={true}
      />
    );
  };

interface AlignmentConfigurationProps {
  designForm: BranchDesignForm;
}
const AlignmentConfiguration: React.VFC<AlignmentConfigurationProps> =
  function AlignmentConfiguration(props) {
    const { designForm } = props;
    const alignment =
      designForm.state.customisableLightTheme.card.alignment ??
      DEFAULT_LIGHT_THEME.card.alignment;

    return (
      <ConfigurationGroup
        labelKey="DesignScreen.configuration.card.label"
        collapsible={true}
        defaultOpen={true}
      >
        <Configuration labelKey="DesignScreen.configuration.card.alignment.label">
          <SegmentedControl.Root
            className={styles.cardAlignmentToggle}
            value={alignment}
            onValueChange={(value) =>
              designForm.setCardAlignment(value as Alignment)
            }
            size="1"
          >
            <SegmentedControl.Item value="start">
              <AlignmentIcon alignment="start" />
            </SegmentedControl.Item>
            <SegmentedControl.Item value="center">
              <AlignmentIcon alignment="center" />
            </SegmentedControl.Item>
            <SegmentedControl.Item value="end">
              <AlignmentIcon alignment="end" />
            </SegmentedControl.Item>
          </SegmentedControl.Root>
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
      <ConfigurationGroup
        labelKey="DesignScreen.configuration.background.label"
        collapsible={true}
      >
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
                sizeLimitInBytes={500 * 1000}
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
                sizeLimitInBytes={500 * 1000}
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
      <ConfigurationGroup
        labelKey="DesignScreen.configuration.button.label"
        collapsible={true}
      >
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
            parentJSONPointer="/primaryButton"
            fieldName="borderRadius"
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
      <ConfigurationGroup
        labelKey="DesignScreen.configuration.icon.label"
        collapsible={true}
      >
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
      (e: React.ChangeEvent<HTMLInputElement>) => {
        designForm.setPrivacyPolicyLink(e.target.value);
      },
      [designForm]
    );
    const onTermsOfServiceLinkChange = useCallback(
      (e: React.ChangeEvent<HTMLInputElement>) => {
        designForm.setTermsOfServiceLink(e.target.value);
      },
      [designForm]
    );
    const onCustomerSupportLinkChange = useCallback(
      (e: React.ChangeEvent<HTMLInputElement>) => {
        designForm.setCustomerSupportLink(e.target.value);
      },
      [designForm]
    );

    return (
      <ConfigurationGroup
        labelKey="DesignScreen.configuration.link.label"
        collapsible={true}
      >
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
          size="2"
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
          size="2"
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
          size="2"
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
      <ConfigurationGroup
        labelKey="DesignScreen.configuration.input.label"
        collapsible={true}
      >
        <Configuration labelKey="DesignScreen.configuration.input.border.label">
          <BorderRadius
            parentJSONPointer="/inputField"
            fieldName="borderRadius"
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
      (checked: boolean) => {
        if (checked) {
          designForm.setDefaultClientURI(uri);
        } else {
          designForm.setDefaultClientURI("");
        }
        setEnabled(checked);
      },
      [uri, designForm]
    );

    const onChangeURI = useCallback(
      (e: React.ChangeEvent<HTMLInputElement>) => {
        const value = e.target.value;
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
          onCheckedChange={onChangeEnableClientURI}
          text={renderToString(
            "DesignScreen.configuration.defaultClientURI.enable.description"
          )}
        />
        <TextField
          size="2"
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
    const onChangeDisplayAuthgearLogo = useCallback(
      (checked: boolean) => {
        designForm.setDisplayAuthgearLogo(checked);
      },
      [designForm]
    );
    return (
      <ConfigurationGroup labelKey="DesignScreen.configuration.authgearBranding.label">
        {designForm.state.whiteLabelingDisabled ? (
          <div className={styles.upgradeBanner}>
            <Text as="p" size="2" weight="medium">
              <FormattedMessage id="DesignScreen.configuration.authgearBranding.upgradeToHide" />
            </Text>
            <Button asChild={true} variant="outline" size="2" color="gray">
              <Link to={`/project/${appID}/billing`}>
                <FormattedMessage id="DesignScreen.configuration.authgearBranding.upgradeNow" />
              </Link>
            </Button>
          </div>
        ) : null}
        <Toggle
          checked={
            designForm.state.whiteLabelingDisabled ||
            designForm.state.showAuthgearLogo
          }
          onCheckedChange={onChangeDisplayAuthgearLogo}
          text={renderToString(
            "DesignScreen.configuration.authgearBranding.disableAuthgearLogo.label"
          )}
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
    const [activeTab, setActiveTab] = useState("branding");

    return (
      <Tabs.Root
        value={activeTab}
        onValueChange={setActiveTab}
        className={styles.configTabs}
      >
        <Tabs.List className={styles.configTabsList}>
          <Tabs.Trigger value="branding">
            <FormattedMessage id="DesignScreen.tabs.branding" />
          </Tabs.Trigger>
          <Tabs.Trigger value="style">
            <FormattedMessage id="DesignScreen.tabs.style" />
          </Tabs.Trigger>
          <Tabs.Trigger value="advance">
            <FormattedMessage id="DesignScreen.tabs.advance" />
          </Tabs.Trigger>
        </Tabs.List>
        <Tabs.Content
          value="branding"
          className={cn(
            styles.configTabContent,
            styles.configTabContentAccordion
          )}
        >
          <OrganisationConfiguration designForm={designForm} />
          <AppLogoConfiguration designForm={designForm} />
          <FaviconConfiguration designForm={designForm} />
        </Tabs.Content>
        <Tabs.Content
          value="style"
          className={cn(
            styles.configTabContent,
            styles.configTabContentAccordion
          )}
        >
          <AlignmentConfiguration designForm={designForm} />
          <ButtonConfiguration designForm={designForm} />
          <InputConfiguration designForm={designForm} />
          <BackgroundConfiguration designForm={designForm} />
          <IconConfiguration designForm={designForm} />
          <LinkConfiguration designForm={designForm} />
        </Tabs.Content>
        <Tabs.Content value="advance" className={styles.configTabContent}>
          <DefaultClientURIConfiguration designForm={designForm} />
          <Separator />
          <AuthgearBrandingConfiguration
            appID={appID}
            designForm={designForm}
          />
        </Tabs.Content>
      </Tabs.Root>
    );
  };

const PreviewLanguageSelect: React.VFC<{
  designForm: BranchDesignForm;
}> = function PreviewLanguageSelect(props) {
  const { designForm } = props;
  const { renderToString } = useContext(MFContext);

  const previewLanguageOptions = useMemo(() => {
    return designForm.state.supportedLanguages.map((lang) => ({
      value: lang,
      label: renderToString(`Locales.${lang}`),
    }));
  }, [designForm.state.supportedLanguages, renderToString]);

  return (
    <Select.Root
      value={designForm.state.selectedLanguage}
      onValueChange={designForm.setSelectedLanguage}
      size="2"
    >
      <Select.Trigger
        variant="surface"
        className={styles.headerLanguageSelect}
      />
      <Select.Content position="popper">
        {previewLanguageOptions.map((opt) => (
          <Select.Item key={opt.value} value={opt.value}>
            {opt.label}
          </Select.Item>
        ))}
      </Select.Content>
    </Select.Root>
  );
};

interface PreviewThemeToggleProps {
  activeTheme: Theme;
  setActiveTheme: (theme: Theme) => void;
  disabled: boolean;
}
const PreviewThemeToggle: React.VFC<PreviewThemeToggleProps> =
  function PreviewThemeToggle(props) {
    const { activeTheme, setActiveTheme, disabled } = props;
    return (
      <SegmentedControl.Root
        className={styles.previewThemeToggle}
        value={activeTheme}
        onValueChange={(v) => setActiveTheme(v as Theme)}
        size="1"
        disabled={disabled}
      >
        <SegmentedControl.Item value={Theme.Light}>
          <SunIcon className={styles.previewThemeIcon} aria-hidden={true} />
        </SegmentedControl.Item>
        <SegmentedControl.Item value={Theme.Dark}>
          <MoonIcon className={styles.previewThemeIcon} aria-hidden={true} />
        </SegmentedControl.Item>
      </SegmentedControl.Root>
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

  const supportedPreviewPages = useMemo(
    () => getSupportedPreviewPagesFromConfig(effectiveAppConfig),
    [effectiveAppConfig]
  );

  const [selectedPreviewPage, setSelectedPreviewPage] = useState(
    () => supportedPreviewPages[0].screen
  );

  const previewPageOptions = useMemo(() => {
    return supportedPreviewPages.map((page) => ({
      value: page.screen,
      label: renderToString(`DesignScreen.preview.pages.title.${page.key}`),
    }));
  }, [supportedPreviewPages, renderToString]);

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
    const message = mapDesignFormStateToPreviewCustomisationMessage(
      designForm.state
    );
    // We must use "*" as targetOrigin because the iframe is sandboxed with a unique origin (null).
    authUIIframeRef.current?.contentWindow?.postMessage(message, "*");
  }, [designForm.state]);

  const [prevSrc, setPrevSrc] = useState(src);
  if (prevSrc !== src) {
    setPrevSrc(src);
    setIsIframeLoading(true);
  }

  const onLoadIframe = useCallback(() => {
    const message = mapDesignFormStateToPreviewCustomisationMessage(
      designForm.state
    );
    setIsIframeLoading(false);
    // We must use "*" as targetOrigin because the iframe is sandboxed with a unique origin (null).
    authUIIframeRef.current?.contentWindow?.postMessage(message, "*");
  }, [designForm.state]);

  return (
    <div className={cn("flex", "flex-col", className)}>
      <div className={styles.previewToolbar}>
        <Select.Root
          value={selectedPreviewPage}
          onValueChange={(v) => setSelectedPreviewPage(v)}
          size="2"
        >
          <Select.Trigger
            variant="surface"
            className={styles.previewPageSelect}
          />
          <Select.Content position="popper">
            {previewPageOptions.map((opt) => (
              <Select.Item key={opt.value} value={opt.value}>
                {opt.label}
              </Select.Item>
            ))}
          </Select.Content>
        </Select.Root>
        <div className={styles.previewToolbarSpacer} />
        {designForm.state.themeOption === "auto" ? (
          <PreviewThemeToggle
            activeTheme={designForm.state.selectedTheme}
            setActiveTheme={designForm.setSelectedTheme}
            disabled={false}
          />
        ) : null}
      </div>
      {isIframeLoading ? <ShowLoading /> : null}
      {
        // It is strongly discouraged to use both `allow-scripts` and `allow-same-origin` in `sandbox` attribute.
        // https://developer.mozilla.org/en-US/docs/Web/HTML/Reference/Elements/iframe#allow-same-origin
      }
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
    const { canSave, getIsDirty } = useFormContainerBaseContext();
    const isDirty = useMemo(() => getIsDirty(), [getIsDirty]);
    const contentWidthAnchorRef = useRef<HTMLDivElement>(null);
    const { triggerTester, isLoading: isTryLoading } = useTester(
      appID,
      effectiveAppConfig.http?.public_origin ?? ""
    );
    const onTry = useCallback(() => {
      triggerTester().catch((err) => {
        console.error(err);
      });
    }, [triggerTester]);

    return (
      <>
        <div
          className={cn(
            styles.contentRoot,
            isDirty && styles.contentWithSaveBar
          )}
        >
          <div className={styles.saveBarScope}>
            <div ref={contentWidthAnchorRef} className={styles.saveBarAnchor}>
              <div className={styles.pageHeader}>
                <Heading as="h1" size="5" weight="medium">
                  <FormattedMessage id="DesignScreen.title" />
                </Heading>
                <div className={styles.titleActions}>
                  <PreviewLanguageSelect designForm={form} />
                  <Tooltip
                    disabled={!canSave}
                    content={
                      <FormattedMessage id="DesignScreen.action.try.disabledHint" />
                    }
                  >
                    {/* Wrap the button so the tooltip still triggers while
                        the button is disabled (a disabled element does not
                        fire the hover events Radix listens for). */}
                    <span className={styles.livePreviewTooltipTarget}>
                      <SecondaryButton
                        size="2"
                        text={
                          <span className={styles.livePreviewButton}>
                            <ExternalLinkIcon
                              className={styles.livePreviewIcon}
                            />
                            <FormattedMessage id="DesignScreen.action.livePreview" />
                          </span>
                        }
                        onClick={onTry}
                        disabled={canSave || isTryLoading}
                        loading={isTryLoading}
                      />
                    </span>
                  </Tooltip>
                </div>
              </div>
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
              <div className={styles.configColumn}>
                <div className={styles.appearancePanel}>
                  <AppearanceConfiguration designForm={form} />
                </div>
                <div className={styles.configPanel}>
                  <ConfigurationPanel appID={appID} designForm={form} />
                </div>
              </div>
            </div>
            <div className={cn("desktop:flex-1", "h-full", "p-6", "pt-4")}>
              <div
                className={cn(
                  "rounded-xl",
                  "h-full",
                  "tablet:h-178.5",
                  "overflow-hidden",
                  "border",
                  "border-solid",
                  "border-gray-a6"
                )}
              >
                <Preview
                  className={cn("h-full")}
                  effectiveAppConfig={effectiveAppConfig}
                  designForm={form}
                />
              </div>
            </div>
          </div>
        </div>
        <SaveFunctionBar anchorRef={contentWidthAnchorRef} />
      </>
    );
  };

const DesignScreen: React.VFC = function DesignScreen() {
  const { appID } = useParams() as { appID: string };
  const {
    effectiveAppConfig,
    isLoading: appConfigLoading,
    loadError: appConfigError,
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
      canSave={form.validationError == null}
      errorRules={form.errorRules}
      stickyFooterComponent={true}
      hideFooterComponent={true}
      localError={form.validationError}
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
