import React, { useCallback, useContext } from "react";
import { DefaultEffects, Text } from "@fluentui/react";
import {
  Context as MFContext,
  FormattedMessage,
} from "@oursky/react-messageformat";
import cn from "classnames";

import { useParams } from "react-router-dom";
import FormContainer from "../../../FormContainer";
import ShowError from "../../../ShowError";
import ShowLoading from "../../../ShowLoading";
import { BranchDesignForm, useBrandDesignForm } from "./form";
import {
  BorderRadius,
  ButtonToggleGroup,
  ColorPicker,
  Configuration,
  ConfigurationDescription,
  ConfigurationGroup,
  FallbackDescription,
  ImagePicker,
  Option,
  Separator,
} from "./Components";
import { Alignment, AllAlignments } from "../../../model/themeAuthFlowV2";

import styles from "./DesignScreen.module.css";
import ScreenTitle from "../../../ScreenTitle";
import ManageLanguageWidget from "../ManageLanguageWidget";
import TextField from "../../../TextField";
import DefaultButton from "../../../DefaultButton";
import Toggle from "../../../Toggle";

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
        <FallbackDescription
          fallbackLanguage={designForm.state.fallbackLanguage}
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
    return (
      <ConfigurationGroup labelKey="DesignScreen.configuration.logo.label">
        <ConfigurationDescription labelKey="DesignScreen.configuration.logo.description" />
        <ImagePicker
          base64EncodedData={designForm.state.appLogoBase64EncodedData}
          onChange={designForm.setAppLogo}
        />
        <FallbackDescription
          fallbackLanguage={designForm.state.fallbackLanguage}
        />
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
        <FallbackDescription
          fallbackLanguage={designForm.state.fallbackLanguage}
        />
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
            value={designForm.state.customisableTheme.cardAlignment}
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
        <Configuration labelKey="DesignScreen.configuration.background.color.label">
          <ColorPicker
            color={designForm.state.customisableTheme.backgroundColor}
            onChange={designForm.setBackgroundColor}
          />
        </Configuration>
        <ImagePicker
          base64EncodedData={designForm.state.backgroundImageBase64EncodedData}
          onChange={designForm.setBackgroundImage}
        />
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
        <Configuration labelKey="DesignScreen.configuration.button.primaryButton.label">
          <ColorPicker
            color={
              designForm.state.customisableTheme.primaryButton.backgroundColor
            }
            onChange={designForm.setPrimaryButtonBackgroundColor}
          />
        </Configuration>
        <Configuration labelKey="DesignScreen.configuration.button.primaryButtonLabel.label">
          <ColorPicker
            color={designForm.state.customisableTheme.primaryButton.labelColor}
            onChange={designForm.setPrimaryButtonLabelColor}
          />
        </Configuration>
        <Configuration labelKey="DesignScreen.configuration.button.borderRadiusStyle.label">
          <BorderRadius
            value={
              designForm.state.customisableTheme.primaryButton.borderRadius
            }
            onChange={designForm.setPrimaryButtonBorderRadiusStyle}
          />
        </Configuration>
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
        <Configuration labelKey="DesignScreen.configuration.link.color.label">
          <ColorPicker
            color={designForm.state.customisableTheme.link.color}
            onChange={designForm.setLinkColor}
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
            value={designForm.state.customisableTheme.inputField.borderRadius}
            onChange={designForm.setInputFieldBorderRadiusStyle}
          />
        </Configuration>
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
      <div className={cn("w-80")}>
        <OrganisationConfiguration designForm={designForm} />
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
        <LinkConfiguration designForm={designForm} />
        <Separator />
        <InputConfiguration designForm={designForm} />
        <Separator />
        <AuthgearBrandingConfiguration appID={appID} designForm={designForm} />
      </div>
    );
  };

const DesignScreen: React.VFC = function DesignScreen() {
  const { appID } = useParams() as { appID: string };
  const form = useBrandDesignForm(appID);

  if (form.isLoading) {
    return <ShowLoading />;
  }

  if (form.loadError) {
    return <ShowError error={form.loadError} onRetry={form.reload} />;
  }

  return (
    <FormContainer
      className={cn("h-full", "flex", "flex-col")}
      form={form}
      canSave={true}
      errorRules={form.errorRules}
    >
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

        <ManageLanguageWidget
          existingLanguages={form.state.supportedLanguages}
          supportedLanguages={form.state.supportedLanguages}
          selectedLanguage={form.state.selectedLanguage}
          fallbackLanguage={form.state.fallbackLanguage}
          onChangeSelectedLanguage={form.setSelectedLanguage}
        />
      </div>
      <div className={cn("min-h-0", "flex-1", "flex")}>
        <div className={cn("flex-1", "h-full", "p-6", "pt-4")}>
          <div
            className={cn("rounded-xl", "h-full")}
            style={{
              boxShadow: DefaultEffects.elevation4,
            }}
          >
            Preview
          </div>
        </div>
        <div className={cn("p-6", "pt-4", "overflow-auto")}>
          <ConfigurationPanel appID={appID} designForm={form} />
        </div>
      </div>
    </FormContainer>
  );
};

export default DesignScreen;
