import React, { useCallback, useMemo } from "react";
import { Callout } from "@radix-ui/themes";
import { InfoCircledIcon } from "@radix-ui/react-icons";
import { FormattedMessage } from "../../intl";
import CustomTagPicker, { Tag } from "../../CustomTagPicker";
import { useMakeAlpha2Options } from "../../util/alpha2";
import { ErrorParseRuleResult, ParsedAPIError } from "../../error/parse";
import { APIError } from "../../error/error";
import { Address4, Address6 } from "ip-address";
import { Toggle } from "../v2/Toggle/Toggle";
import { TextArea } from "../v2/TextArea/TextArea";
import { FormField } from "../v2/FormField/FormField";
import styles from "./IPBlocklistForm.module.css";

export interface IPBlocklistFormState {
  isEditAllowed: boolean;
  isEnabled: boolean;
  blockedIPCIDRs: string;
  blockedCountryAlpha2s: string[];
}

export interface IPCheckResult {
  ipAddress: string;
  result: boolean;
}

export interface IPBlocklistFormProps {
  state: IPBlocklistFormState;
  setState: (fn: (state: IPBlocklistFormState) => IPBlocklistFormState) => void;
}

export function toCIDRs(blockedIPCIDRsStr: string): string[] {
  return blockedIPCIDRsStr
    .split(/,|\n/)
    .map((s) => {
      const trimmed = s.trim();
      if (trimmed === "") {
        return "";
      }

      const hasSubnet = trimmed.includes("/");

      if (Address4.isValid(trimmed)) {
        if (!hasSubnet) {
          return `${trimmed}/32`;
        }
        return trimmed;
      }

      if (Address6.isValid(trimmed)) {
        if (!hasSubnet) {
          return `${trimmed}/128`;
        }
        return trimmed;
      }

      return trimmed;
    })
    .filter((s) => s !== "");
}

export function IPBlocklistForm({
  state,
  setState,
}: IPBlocklistFormProps): React.ReactElement {
  const { alpha2Options } = useMakeAlpha2Options();

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

  const onBlockedIPCIDRsChange = useCallback(
    (e: React.ChangeEvent<HTMLTextAreaElement>) => {
      const value = e.target.value;
      setState((prev) => ({
        ...prev,
        blockedIPCIDRs: value,
      }));
    },
    [setState]
  );

  const onCountryItemChange = useCallback(
    (items?: Tag[]) => {
      if (items == null) {
        return;
      }
      setState((prev) => ({
        ...prev,
        blockedCountryAlpha2s: items.map((it) => it.key as string),
      }));
    },
    [setState]
  );

  const onChangeEnabled = useCallback(
    (checked: boolean) => {
      setState((prev) => ({ ...prev, isEnabled: checked }));
    },
    [setState]
  );

  const selectedCountryTags: Tag[] = useMemo(() => {
    return state.blockedCountryAlpha2s.map((alpha2) => {
      const option = alpha2Options.find((opt) => opt.key === alpha2);
      return {
        key: alpha2,
        name: option?.text ?? alpha2,
      };
    });
  }, [state.blockedCountryAlpha2s, alpha2Options]);

  const cidrsFieldErrorRules = useMemo(
    () => [
      (apiError: APIError): ErrorParseRuleResult => {
        const parsedAPIErrors: ParsedAPIError[] = [];
        if (apiError.reason === "ValidationFailed") {
          for (const cause of apiError.info.causes) {
            const regex = /\/cidrs\/(\d+)/;
            const match = regex.exec(cause.location);
            if (
              match?.[1] &&
              cause.kind === "format" &&
              cause.details.format === "x_cidr"
            ) {
              const itemIndex = Number(match[1]);
              parsedAPIErrors.push({
                messageID: "IPBlocklistForm.error.invalid-ip",
                arguments: {
                  ipAddress: toCIDRs(state.blockedIPCIDRs)[itemIndex],
                },
              });
            }
          }
          return {
            parsedAPIErrors: parsedAPIErrors,
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
    [state.blockedIPCIDRs]
  );

  return (
    <>
      {!state.isEditAllowed ? (
        <Callout.Root color="blue" variant="surface" size="1">
          <Callout.Icon>
            <InfoCircledIcon />
          </Callout.Icon>
          <Callout.Text>
            <FormattedMessage id="IPBlocklistForm.error.edit-disabled" />
          </Callout.Text>
        </Callout.Root>
      ) : null}
      <div className={styles.enableToggle}>
        <Toggle
          checked={state.isEnabled}
          onCheckedChange={onChangeEnabled}
          disabled={!state.isEditAllowed}
          textWeight="medium"
          text={<FormattedMessage id="IPBlocklistForm.enable.label" />}
        />
      </div>
      {state.isEnabled && state.isEditAllowed ? (
        <>
          <TextArea
            size="2"
            labelSize="2"
            className={styles.textArea}
            label={<FormattedMessage id="IPBlocklistForm.ip-address.label" />}
            hint={
              <FormattedMessage id="IPBlocklistForm.ip-address.description" />
            }
            parentJSONPointer="/network_protection/ip_filter/rules/0/source"
            fieldName="cidrs"
            value={state.blockedIPCIDRs}
            onChange={onBlockedIPCIDRsChange}
            errorRules={cidrsFieldErrorRules}
          />
          <FormField
            size="2"
            labelSize="2"
            label={
              <FormattedMessage id="IPBlocklistForm.block-country.label" />
            }
            labelSpace="1"
          >
            <div className={styles.countryTagPicker}>
              <CustomTagPicker
                onResolveSuggestions={onResolveCountryCodeSuggestions}
                selectedItems={selectedCountryTags}
                onChange={onCountryItemChange}
              />
            </div>
          </FormField>
        </>
      ) : null}
    </>
  );
}
