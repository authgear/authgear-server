import React from "react";
import { Callout } from "@radix-ui/themes";
import { InfoCircledIcon } from "@radix-ui/react-icons";
import { FormattedMessage, Values } from "../../../intl";
import { useFeatureDisabledMessageValues } from "../../../graphql/portal/FeatureDisabledMessageBar";

export interface FeatureDisabledCalloutProps {
  className?: string;
  messageID: string;
  messageValues?: Values;
}

// The v2 counterpart of FeatureDisabledMessageBar: the same
// FeatureConfig.*.disabled message (with plan-page / contact-us links)
// rendered as a Radix Callout, matching the info callouts used across the
// migrated Advanced-settings screens.
export function FeatureDisabledCallout({
  className,
  messageID,
  messageValues,
}: FeatureDisabledCalloutProps): React.ReactElement {
  const values = useFeatureDisabledMessageValues(messageValues);

  return (
    <Callout.Root className={className} color="blue" variant="surface" size="1">
      <Callout.Icon>
        <InfoCircledIcon />
      </Callout.Icon>
      <Callout.Text>
        <FormattedMessage id={messageID} values={values} />
      </Callout.Text>
    </Callout.Root>
  );
}
