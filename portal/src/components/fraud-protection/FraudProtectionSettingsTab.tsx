import React, { useContext, useMemo } from "react";
import { ChoiceGroup, IChoiceGroupOption } from "@fluentui/react";
import { Context } from "../../intl";
import { FraudProtectionDecisionAction } from "../../types";
import FormTextField from "../../FormTextField";
import styles from "./FraudProtectionSettingsTab.module.css";

export interface FraudProtectionSettingsTabProps {
  isModifiable: boolean;
  enforcementMode: FraudProtectionDecisionAction;
  ipAllowlist: string;
  phoneAllowlist: string;
  onEnforcementModeChange: (
    event?: React.FormEvent<HTMLElement | HTMLInputElement>,
    option?: IChoiceGroupOption
  ) => void;
  onIPAllowlistChange: (
    event: React.FormEvent<HTMLInputElement | HTMLTextAreaElement>,
    newValue?: string
  ) => void;
  onPhoneAllowlistChange: (
    event: React.FormEvent<HTMLInputElement | HTMLTextAreaElement>,
    newValue?: string
  ) => void;
}

const FraudProtectionSettingsTab: React.VFC<FraudProtectionSettingsTabProps> =
  function FraudProtectionSettingsTab(props) {
    const {
      isModifiable,
      enforcementMode,
      ipAllowlist,
      phoneAllowlist,
      onEnforcementModeChange,
      onIPAllowlistChange,
      onPhoneAllowlistChange,
    } = props;
    const { renderToString } = useContext(Context);

    const enforcementModeOptions = useMemo<IChoiceGroupOption[]>(() => {
      return [
        {
          key: "record_only",
          text: renderToString(
            "FraudProtectionConfigurationScreen.enforcement.observe.label"
          ),
        },
        {
          key: "deny_if_any_warning",
          text: renderToString(
            "FraudProtectionConfigurationScreen.enforcement.protect.label"
          ),
        },
      ];
    }, [renderToString]);

    return (
      <section className={styles.section}>
        <ChoiceGroup
          label={renderToString(
            "FraudProtectionConfigurationScreen.enforcement.mode.title"
          )}
          disabled={!isModifiable}
          selectedKey={enforcementMode}
          options={enforcementModeOptions}
          onChange={onEnforcementModeChange}
        />
        <FormTextField
          className={styles.field}
          parentJSONPointer="/fraud_protection/decision/always_allow/ip_address"
          fieldName="cidrs"
          label={renderToString(
            "FraudProtectionConfigurationScreen.allowlist.ip.label"
          )}
          description={renderToString(
            "FraudProtectionConfigurationScreen.allowlist.ip.description"
          )}
          placeholder="127.0.0.1/32"
          multiline={true}
          resizable={false}
          disabled={!isModifiable}
          value={ipAllowlist}
          onChange={onIPAllowlistChange}
        />
        <FormTextField
          className={styles.field}
          parentJSONPointer="/fraud_protection/decision/always_allow/phone_number"
          fieldName="regex"
          label={renderToString(
            "FraudProtectionConfigurationScreen.allowlist.phone.label"
          )}
          description={renderToString(
            "FraudProtectionConfigurationScreen.allowlist.phone.description"
          )}
          placeholder="+1 555 123 4567"
          multiline={true}
          resizable={false}
          disabled={!isModifiable}
          value={phoneAllowlist}
          onChange={onPhoneAllowlistChange}
        />
      </section>
    );
  };

export default FraudProtectionSettingsTab;
