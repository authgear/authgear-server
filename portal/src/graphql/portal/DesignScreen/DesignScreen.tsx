import React, { useCallback } from "react";
import { DefaultEffects } from "@fluentui/react";
import cn from "classnames";

import { useParams } from "react-router-dom";
import FormContainer from "../../../FormContainer";
import ShowError from "../../../ShowError";
import ShowLoading from "../../../ShowLoading";
import { BranchDesignForm, useBrandDesignForm } from "./form";
import {
  ButtonToggleGroup,
  Configuration,
  ConfigurationGroup,
  Option,
} from "./Components";
import { Alignment, AllAlignments } from "../../../model/themeAuthFlowV2";

import styles from "./DesignScreen.module.css";

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

interface ConfigurationPanelProps {
  designForm: BranchDesignForm;
}
const ConfigurationPanel: React.VFC<ConfigurationPanelProps> =
  function ConfigurationPanel(props) {
    const { designForm } = props;
    return (
      <div>
        <AlignmentConfiguration designForm={designForm} />
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
    <FormContainer className={cn("h-full")} form={form} canSave={true}>
      <div className={cn("h-full", "flex")}>
        <div className={cn("flex-1", "h-full", "p-6")}>
          <div
            className={cn("flex-1", "rounded-xl", "h-full")}
            style={{
              boxShadow: DefaultEffects.elevation4,
            }}
          >
            Preview
          </div>
        </div>
        <div className={cn("w-80", "p-6", "overflow-auto")}>
          <ConfigurationPanel designForm={form} />
        </div>
      </div>
    </FormContainer>
  );
};

export default DesignScreen;
