import React, { useCallback } from "react";
import { DefaultEffects } from "@fluentui/react";
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
  Option,
  Separator,
} from "./Components";
import { Alignment, AllAlignments } from "../../../model/themeAuthFlowV2";

import styles from "./DesignScreen.module.css";
import ScreenTitle from "../../../ScreenTitle";
import { FormattedMessage } from "@oursky/react-messageformat";
import ManageLanguageWidget from "../ManageLanguageWidget";

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
    return (
      <ConfigurationGroup labelKey="DesignScreen.configuration.link.label">
        <Configuration labelKey="DesignScreen.configuration.link.color.label">
          <ColorPicker
            color={designForm.state.customisableTheme.link.color}
            onChange={designForm.setLinkColor}
          />
        </Configuration>
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

interface ConfigurationPanelProps {
  designForm: BranchDesignForm;
}
const ConfigurationPanel: React.VFC<ConfigurationPanelProps> =
  function ConfigurationPanel(props) {
    const { designForm } = props;
    return (
      <div>
        <AlignmentConfiguration designForm={designForm} />
        <Separator />
        <BackgroundConfiguration designForm={designForm} />
        <Separator />
        <ButtonConfiguration designForm={designForm} />
        <Separator />
        <LinkConfiguration designForm={designForm} />
        <Separator />
        <InputConfiguration designForm={designForm} />
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
        <div className={cn("w-80", "p-6", "pt-4", "overflow-auto")}>
          <ConfigurationPanel designForm={form} />
        </div>
      </div>
    </FormContainer>
  );
};

export default DesignScreen;
