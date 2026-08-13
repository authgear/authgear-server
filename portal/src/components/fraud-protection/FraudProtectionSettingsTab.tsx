import React, { useCallback, useContext, useMemo } from "react";
import { RadioGroup, Text } from "@radix-ui/themes";
import { Context, FormattedMessage } from "../../intl";
import { FraudProtectionDecisionAction } from "../../types";
import { TextArea } from "../v2/TextArea/TextArea";
import { FormField } from "../v2/FormField/FormField";
import { SettingsSectionCard } from "../v2/SettingsSectionCard/SettingsSectionCard";
import CustomTagPicker, { Tag } from "../../CustomTagPicker";
import { useMakeAlpha2Options } from "../../util/alpha2";
import { APIError } from "../../error/error";
import { ErrorParseRuleResult, ParsedAPIError } from "../../error/parse";
import styles from "./FraudProtectionSettingsTab.module.css";

export interface FraudProtectionSettingsTabProps {
  isModifiable: boolean;
  enforcementMode: FraudProtectionDecisionAction;
  ipAllowlist: string;
  phoneAllowlist: string;
  ipCountryAllowlist: string[];
  phoneCountryAllowlist: string[];
  onEnforcementModeChange: (value: FraudProtectionDecisionAction) => void;
  onIPAllowlistChange: (event: React.ChangeEvent<HTMLTextAreaElement>) => void;
  onPhoneAllowlistChange: (
    event: React.ChangeEvent<HTMLTextAreaElement>
  ) => void;
  onIPCountryAllowlistChange: (items?: Tag[]) => void;
  onPhoneCountryAllowlistChange: (items?: Tag[]) => void;
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
      {
        value: FraudProtectionDecisionAction;
        title: React.ReactNode;
        subtitle: React.ReactNode;
      }[]
    >(() => {
      return [
        {
          value: "record_only",
          title: (
            <FormattedMessage id="FraudProtectionConfigurationScreen.enforcement.observe.label" />
          ),
          subtitle: (
            <FormattedMessage id="FraudProtectionConfigurationScreen.enforcement.observe.description" />
          ),
        },
        {
          value: "deny_if_any_warning",
          title: (
            <FormattedMessage id="FraudProtectionConfigurationScreen.enforcement.protect.label" />
          ),
          subtitle: (
            <FormattedMessage id="FraudProtectionConfigurationScreen.enforcement.protect.description" />
          ),
        },
      ];
    }, []);

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
      (filter: string): Tag[] => {
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

    const selectedIPCountryTags: Tag[] = useMemo(
      () =>
        ipCountryAllowlist.map((alpha2) => ({
          key: alpha2,
          name: alpha2Options.find((opt) => opt.key === alpha2)?.text ?? alpha2,
        })),
      [ipCountryAllowlist, alpha2Options]
    );

    const selectedPhoneCountryTags: Tag[] = useMemo(
      () =>
        phoneCountryAllowlist.map((alpha2) => ({
          key: alpha2,
          name: alpha2Options.find((opt) => opt.key === alpha2)?.text ?? alpha2,
        })),
      [phoneCountryAllowlist, alpha2Options]
    );

    return (
      <SettingsSectionCard
        className={styles.card}
        title={
          <FormattedMessage id="FraudProtectionConfigurationScreen.tab.settings.title" />
        }
        contentClassName={styles.cardContent}
      >
        <FormField
          size="2"
          label={
            <FormattedMessage id="FraudProtectionConfigurationScreen.enforcement.mode.title" />
          }
        >
          <RadioGroup.Root
            className={styles.modeGroup}
            value={enforcementMode}
            onValueChange={(value) =>
              onEnforcementModeChange(value as FraudProtectionDecisionAction)
            }
          >
            {enforcementModeOptions.map((option) => (
              <label key={option.value} className={styles.modeOption}>
                <RadioGroup.Item
                  className={styles.modeRadio}
                  value={option.value}
                  disabled={!isModifiable}
                />
                <span className={styles.modeText}>
                  <Text
                    as="span"
                    size="2"
                    weight="medium"
                    className={styles.modeTitle}
                  >
                    {option.title}
                  </Text>
                  <Text as="span" size="2" className={styles.modeDescription}>
                    {option.subtitle}
                  </Text>
                </span>
              </label>
            ))}
          </RadioGroup.Root>
        </FormField>
        <TextArea
          size="2"
          parentJSONPointer="/fraud_protection/decision/always_allow/ip_address"
          fieldName="cidrs"
          label={renderToString(
            "FraudProtectionConfigurationScreen.allowlist.ip.label"
          )}
          hint={renderToString(
            "FraudProtectionConfigurationScreen.allowlist.ip.description"
          )}
          placeholder="127.0.0.1/32"
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
        <TextArea
          size="2"
          parentJSONPointer="/fraud_protection/decision/always_allow/phone_number"
          fieldName="regex"
          label={renderToString(
            "FraudProtectionConfigurationScreen.allowlist.phone.label"
          )}
          hint={renderToString(
            "FraudProtectionConfigurationScreen.allowlist.phone.description"
          )}
          placeholder="+1 555 123 4567"
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
      </SettingsSectionCard>
    );
  };

export default FraudProtectionSettingsTab;
