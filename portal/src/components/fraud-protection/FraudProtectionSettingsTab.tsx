import React, { useCallback, useContext, useMemo } from "react";
import { Context, FormattedMessage } from "../../intl";
import { FraudProtectionDecisionAction } from "../../types";
import FormTextField from "../../FormTextField";
import CustomTagPicker from "../../CustomTagPicker";
import { useMakeAlpha2Options } from "../../util/alpha2";
import { APIError } from "../../error/error";
import { ErrorParseRuleResult, ParsedAPIError } from "../../error/parse";
import ChoiceGroupWithDescriptions, {
  ChoiceGroupWithDescriptionOption,
} from "../common/ChoiceGroupWithDescriptions";
import styles from "./FraudProtectionSettingsTab.module.css";
import { IChoiceGroupOption, ITag } from "@fluentui/react";

export interface FraudProtectionSettingsTabProps {
  isModifiable: boolean;
  enforcementMode: FraudProtectionDecisionAction;
  ipAllowlist: string;
  phoneAllowlist: string;
  ipCountryAllowlist: string[];
  phoneCountryAllowlist: string[];
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
  onIPCountryAllowlistChange: (items?: ITag[]) => void;
  onPhoneCountryAllowlistChange: (items?: ITag[]) => void;
}

const FraudProtectionSettingsTab: React.VFC<FraudProtectionSettingsTabProps> =
  function FraudProtectionSettingsTab(props) {
    const {
      isModifiable,
      enforcementMode,
      ipAllowlist,
      phoneAllowlist,
      ipCountryAllowlist,
      phoneCountryAllowlist,
      onEnforcementModeChange,
      onIPAllowlistChange,
      onPhoneAllowlistChange,
      onIPCountryAllowlistChange,
      onPhoneCountryAllowlistChange,
    } = props;
    const { renderToString } = useContext(Context);
    const { alpha2Options } = useMakeAlpha2Options();

    const enforcementModeOptions = useMemo<
      ChoiceGroupWithDescriptionOption[]
    >(() => {
      return [
        {
          key: "record_only",
          text: renderToString(
            "FraudProtectionConfigurationScreen.enforcement.observe.label"
          ),
          description: (
            <FormattedMessage id="FraudProtectionConfigurationScreen.enforcement.observe.description" />
          ),
        },
        {
          key: "deny_if_any_warning",
          text: renderToString(
            "FraudProtectionConfigurationScreen.enforcement.protect.label"
          ),
          description: (
            <FormattedMessage id="FraudProtectionConfigurationScreen.enforcement.protect.description" />
          ),
        },
      ];
    }, [renderToString]);

    const splitRawItems = useMemo(() => {
      return (raw: string): string[] =>
        raw
          .split(/,|\n/)
          .map((item) => item.trim())
          .filter((item) => item !== "");
    }, []);

    const ipAllowlistFieldErrorRules = useMemo(
      () => [
        (apiError: APIError): ErrorParseRuleResult => {
          const parsedAPIErrors: ParsedAPIError[] = [];
          if (apiError.reason === "ValidationFailed") {
            for (const cause of apiError.info.causes) {
              const match =
                /\/fraud_protection\/decision\/always_allow\/ip_address\/cidrs\/(\d+)/.exec(
                  cause.location
                );
              if (
                match?.[1] &&
                cause.kind === "format" &&
                cause.details.format === "x_cidr"
              ) {
                const itemIndex = Number(match[1]);
                parsedAPIErrors.push({
                  messageID: "IPBlocklistForm.error.invalid-ip",
                  arguments: {
                    ipAddress: splitRawItems(ipAllowlist)[itemIndex],
                  },
                });
              }
            }
            return {
              parsedAPIErrors,
              fullyHandled:
                parsedAPIErrors.length === apiError.info.causes.length,
            };
          }
          return {
            parsedAPIErrors: [],
            fullyHandled: false,
          };
        },
      ],
      [ipAllowlist, splitRawItems]
    );

    const onResolveCountryCodeSuggestions = useCallback(
      (filter: string): ITag[] => {
        const matchedOptions = alpha2Options.filter(
          (opt) =>
            opt.key.startsWith(filter.toUpperCase()) ||
            opt.text.toLowerCase().includes(filter.toLowerCase())
        );
        if (matchedOptions.length > 0) {
          return matchedOptions.map((opt) => ({
            key: opt.key,
            name: opt.text,
          }));
        }
        if (filter.length === 2) {
          return [
            {
              key: filter.toUpperCase(),
              name: filter.toUpperCase(),
            },
          ];
        }
        return [];
      },
      [alpha2Options]
    );

    const selectedIPCountryTags: ITag[] = useMemo(
      () =>
        ipCountryAllowlist.map((alpha2) => ({
          key: alpha2,
          name: alpha2Options.find((opt) => opt.key === alpha2)?.text ?? alpha2,
        })),
      [ipCountryAllowlist, alpha2Options]
    );

    const selectedPhoneCountryTags: ITag[] = useMemo(
      () =>
        phoneCountryAllowlist.map((alpha2) => ({
          key: alpha2,
          name: alpha2Options.find((opt) => opt.key === alpha2)?.text ?? alpha2,
        })),
      [phoneCountryAllowlist, alpha2Options]
    );

    return (
      <section className={styles.section}>
        <ChoiceGroupWithDescriptions
          label={renderToString(
            "FraudProtectionConfigurationScreen.enforcement.mode.title"
          )}
          disabled={!isModifiable}
          selectedKey={enforcementMode}
          options={enforcementModeOptions}
          onChange={onEnforcementModeChange}
        />
        <FormTextField
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
          errorRules={ipAllowlistFieldErrorRules}
        />
        <CustomTagPicker
          label={renderToString(
            "FraudProtectionConfigurationScreen.allowlist.ip.country.label"
          )}
          disabled={!isModifiable}
          onResolveSuggestions={onResolveCountryCodeSuggestions}
          selectedItems={selectedIPCountryTags}
          onChange={onIPCountryAllowlistChange}
        />
        <FormTextField
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
        <CustomTagPicker
          label={renderToString(
            "FraudProtectionConfigurationScreen.allowlist.phone.country.label"
          )}
          disabled={!isModifiable}
          onResolveSuggestions={onResolveCountryCodeSuggestions}
          selectedItems={selectedPhoneCountryTags}
          onChange={onPhoneCountryAllowlistChange}
        />
      </section>
    );
  };

export default FraudProtectionSettingsTab;
