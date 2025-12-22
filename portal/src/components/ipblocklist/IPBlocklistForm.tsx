import React, { useCallback, useContext, useMemo, useState } from "react";
import {
  Context as MessageContext,
  FormattedMessage,
} from "@oursky/react-messageformat";
import { useParams } from "react-router-dom";
import Toggle from "../../Toggle";
import TextField from "../../TextField";
import CustomTagPicker from "../../CustomTagPicker";
import { useMakeAlpha2Options } from "../../util/alpha2";
import { ITag, Label, MessageBar, MessageBarType } from "@fluentui/react";
import { useCheckIPMutation } from "../../graphql/portal/mutations/checkIPMutation";
import ButtonWithLoading from "../../ButtonWithLoading";
import ErrorRenderer from "../../ErrorRenderer";
import {
  ErrorParseRuleResult,
  parseAPIErrors,
  ParsedAPIError,
  parseRawError,
} from "../../error/parse";
import FormTextField from "../../FormTextField";
import { APIError } from "../../error/error";
import { Address4, Address6 } from "ip-address";

export interface IPBlocklistFormState {
  isEditAllowed: boolean;
  isEnabled: boolean;
  blockedIPCIDRs: string;
  blockedCountryAlpha2s: string[];
}

export interface IPBlocklistFormProps {
  state: IPBlocklistFormState;
  setState: (fn: (state: IPBlocklistFormState) => IPBlocklistFormState) => void;
}

interface IPCheckResult {
  ipAddress: string;
  result: boolean;
}

export function toCIDRs(blockedIPCIDRsStr: string): string[] {
  return blockedIPCIDRsStr
    .split(",")
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

  const { appID } = useParams() as { appID: string };
  const {
    checkIP,
    loading: checkingIP,
    error: checkIPError,
  } = useCheckIPMutation(appID);
  const [ipToCheck, setIPToCheck] = useState("");
  const [checkIPResult, setCheckIPResult] = useState<IPCheckResult | null>(
    null
  );

  const onIPToCheckChange = useCallback(
    (e: React.FormEvent<HTMLInputElement | HTMLTextAreaElement>) => {
      setIPToCheck(e.currentTarget.value);
    },
    []
  );

  const onCheckIP = useCallback(() => {
    checkIP(
      ipToCheck,
      state.blockedIPCIDRs
        .split(",")
        .map((cidr) => cidr.trim())
        .filter((cidr) => cidr !== ""),
      state.blockedCountryAlpha2s
    )
      .then((result) => {
        setCheckIPResult({
          ipAddress: ipToCheck,
          result: Boolean(result),
        });
      })
      .catch(() => {});
  }, [checkIP, ipToCheck, state.blockedIPCIDRs, state.blockedCountryAlpha2s]);

  const checkIPFieldError = useMemo(() => {
    if (checkIPError == null) {
      return undefined;
    }
    const apiErrors = parseRawError(checkIPError);
    const { topErrors } = parseAPIErrors(apiErrors, [], []);
    return <ErrorRenderer errors={topErrors} />;
  }, [checkIPError]);

  const cidrsFieldErrorRules = useMemo(
    () => [
      (apiError: APIError): ErrorParseRuleResult => {
        const parsedAPIErrors: ParsedAPIError[] = [];
        if (apiError.reason === "ValidationFailed") {
          for (const cause of apiError.info.causes) {
            if (
              cause.location.startsWith(
                "/network_protection/ip_filter/rules/0/source/cidrs/"
              ) &&
              cause.kind === "format" &&
              cause.details.format === "x_cidr"
            ) {
              const itemIndex = Number(
                cause.location.replace(
                  "/network_protection/ip_filter/rules/0/source/cidrs/",
                  ""
                )
              );
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
        <>
          <div className="mt-12">
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
          <div className="mt-6">
            <CustomTagPicker
              label={renderToString("IPBlocklistForm.block-country.label")}
              onResolveSuggestions={onResolveCountryCodeSuggestions}
              selectedItems={selectedCountryTags}
              onChange={onCountryItemChange}
            />
          </div>
          <div className="mt-6 flex flex-col gap-y-4">
            <div className="flex items-start gap-x-4">
              <TextField
                className="flex-1"
                label={renderToString("IPBlocklistForm.check-ip-address.label")}
                value={ipToCheck}
                onChange={onIPToCheckChange}
                errorMessage={checkIPFieldError}
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
        </>
      ) : null}
    </div>
  );
}
