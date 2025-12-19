import React, { useCallback, useContext, useMemo } from "react";
import { Context as MessageContext } from "@oursky/react-messageformat";
import Toggle from "../../Toggle";
import TextField from "../../TextField";
import CustomTagPicker from "../../CustomTagPicker";
import { useMakeAlpha2Options } from "../../util/alpha2";
import { ITag } from "@fluentui/react";

export interface IPBlocklistFormState {
  isEnabled: boolean;
  blockedIPCIDRs: string;
  blockedCountryAlpha2s: string[];
}

export interface IPBlocklistFormProps {
  state: IPBlocklistFormState;
  setState: (fn: (state: IPBlocklistFormState) => IPBlocklistFormState) => void;
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
      return [
        {
          key: filter.toUpperCase(),
          name: filter.toUpperCase(),
        },
      ];
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

  return (
    <div className="p-6 max-w-180">
      <Toggle
        label={renderToString("IPBlocklistForm.enable.label")}
        inlineLabel={false}
        checked={state.isEnabled}
        onChange={useCallback(
          (_, checked) => {
            setState((prev) => ({ ...prev, isEnabled: !!checked }));
          },
          [setState]
        )}
      />
      {state.isEnabled ? (
        <>
          <div className="mt-12">
            <TextField
              className="h-37"
              label={renderToString("IPBlocklistForm.ip-address.label")}
              multiline={true}
              resizable={false}
              description={renderToString(
                "IPBlocklistForm.ip-address.description"
              )}
              value={state.blockedIPCIDRs}
              onChange={onBlockedIPCIDRsChange}
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
          <div className="mt-6">
            <TextField
              label={renderToString("IPBlocklistForm.check-ip-address.label")}
            />
          </div>
        </>
      ) : null}
    </div>
  );
}
