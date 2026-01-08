import React, { useCallback, useContext, useMemo } from "react";
import {
  Context as MessageContext,
  FormattedMessage,
} from "@oursky/react-messageformat";
import Toggle from "../../Toggle";
import CustomTagPicker from "../../CustomTagPicker";
import { useMakeAlpha2Options } from "../../util/alpha2";
import { ITag, Label, MessageBar, MessageBarType } from "@fluentui/react";
import ButtonWithLoading from "../../ButtonWithLoading";
import { ErrorParseRuleResult, ParsedAPIError } from "../../error/parse";
import FormTextField from "../../FormTextField";
import { APIError } from "../../error/error";
import { Address4, Address6 } from "ip-address";

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
  ipToCheck: string;
  onIPToCheckChange: (
    e: React.FormEvent<HTMLInputElement | HTMLTextAreaElement>
  ) => void;
  onCheckIP: () => void;
  checkingIP: boolean;
  checkIPResult: IPCheckResult | null;
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
  ipToCheck,
  onIPToCheckChange,
  onCheckIP,
  checkingIP,
  checkIPResult,
}: IPBlocklistFormProps): React.ReactElement {
  const { renderToString } = useContext(MessageContext);

  const { alpha2Options } = useMakeAlpha2Options();

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

  const onBlockedIPCIDRsChange = useCallback(
    (e: React.FormEvent<HTMLInputElement | HTMLTextAreaElement>) => {
      const value = e.currentTarget.value;
      setState((prev) => {
        return {
          ...prev,
          blockedIPCIDRs: value,
        };
      });
    },
    [setState]
  );

  const onCountryItemChange = useCallback(
    (items?: ITag[]) => {
      if (items == null) {
        return;
      }
      setState((prev) => {
        return {
          ...prev,
          blockedCountryAlpha2s: items.map((it) => it.key as string),
        };
      });
    },
    [setState]
  );

  const selectedCountryTags: ITag[] = useMemo(() => {
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
    <div className="p-6 max-w-180">
      {!state.isEditAllowed ? (
        <div className="mb-6">
          <MessageBar messageBarType={MessageBarType.info}>
            <FormattedMessage id="IPBlocklistForm.error.edit-disabled" />
          </MessageBar>
        </div>
      ) : null}
      <Toggle
        label={renderToString("IPBlocklistForm.enable.label")}
        inlineLabel={false}
        disabled={!state.isEditAllowed}
        checked={state.isEnabled}
        onChange={useCallback(
          (_, checked) => {
            setState((prev) => ({ ...prev, isEnabled: !!checked }));
          },
          [setState]
        )}
      />
      {state.isEnabled && state.isEditAllowed ? (
        <div className="mt-12 flex flex-col gap-y-6">
          <div>
            <FormTextField
              parentJSONPointer="/network_protection/ip_filter/rules/0/source"
              fieldName="cidrs"
              className="h-37"
              label={renderToString("IPBlocklistForm.ip-address.label")}
              multiline={true}
              resizable={false}
              description={renderToString(
                "IPBlocklistForm.ip-address.description"
              )}
              value={state.blockedIPCIDRs}
              onChange={onBlockedIPCIDRsChange}
              errorRules={cidrsFieldErrorRules}
            />
          </div>
          <div>
            <CustomTagPicker
              label={renderToString("IPBlocklistForm.block-country.label")}
              onResolveSuggestions={onResolveCountryCodeSuggestions}
              selectedItems={selectedCountryTags}
              onChange={onCountryItemChange}
            />
          </div>
          <div className="h-px w-full bg-separator" />
          <div className="flex flex-col gap-y-4 p-4 bg-[#FAF9F8]">
            <div className="flex items-start gap-x-4">
              <FormTextField
                parentJSONPointer=""
                fieldName="ipAddress"
                className="flex-1"
                label={renderToString("IPBlocklistForm.check-ip-address.label")}
                value={ipToCheck}
                onChange={onIPToCheckChange}
              />
              <div>
                {/* Add a empty label to align the button */}
                <Label>&nbsp;</Label>
                <ButtonWithLoading
                  labelId="IPBlocklistForm.check-ip-address.button"
                  onClick={onCheckIP}
                  loading={checkingIP}
                />
              </div>
            </div>
            {checkIPResult != null ? (
              checkIPResult.result ? (
                <MessageBar messageBarType={MessageBarType.error}>
                  <FormattedMessage
                    id="IPBlocklistForm.check-ip-address.result.is-blocked"
                    values={{
                      ipAddress: checkIPResult.ipAddress,
                    }}
                  />
                </MessageBar>
              ) : (
                <MessageBar messageBarType={MessageBarType.info}>
                  <FormattedMessage
                    id="IPBlocklistForm.check-ip-address.result.is-not-blocked"
                    values={{
                      ipAddress: checkIPResult.ipAddress,
                    }}
                  />
                </MessageBar>
              )
            ) : null}
          </div>
        </div>
      ) : null}
    </div>
  );
}
