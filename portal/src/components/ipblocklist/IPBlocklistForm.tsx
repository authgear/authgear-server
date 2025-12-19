import React, { useCallback, useContext } from "react";
import { Context as MessageContext } from "@oursky/react-messageformat";
import Toggle from "../../Toggle";
import TextField from "../../TextField";
import CustomTagPicker from "../../CustomTagPicker";
import { useMakeAlpha2Options } from "../../util/alpha2";
import { ITag } from "@fluentui/react";

export interface IPBlocklistFormProps {}

export function IPBlocklistForm({}: IPBlocklistFormProps): React.ReactElement {
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

  return (
    <div className="p-6 max-w-180">
      <Toggle
        label={renderToString("IPBlocklistForm.enable.label")}
        inlineLabel={false}
      />
      <div className="mt-12">
        <TextField
          className="h-37"
          label={renderToString("IPBlocklistForm.ip-address.label")}
          multiline={true}
          resizable={false}
          description={renderToString("IPBlocklistForm.ip-address.description")}
        />
      </div>
      <div className="mt-6">
        <CustomTagPicker
          label={renderToString("IPBlocklistForm.block-country.label")}
          onResolveSuggestions={onResolveCountryCodeSuggestions}
        />
      </div>
      <div className="mt-6">
        <TextField label={renderToString("IPBlocklistForm.check-ip-address.label")} />
      </div>
    </div>
  );
}
