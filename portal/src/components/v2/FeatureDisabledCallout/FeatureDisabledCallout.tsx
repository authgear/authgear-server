import React, { useMemo } from "react";
import { Callout } from "@radix-ui/themes";
import { InfoCircledIcon } from "@radix-ui/react-icons";
import { useParams } from "react-router-dom";
import { FormattedMessage, Values } from "../../../intl";
import ReactRouterLink from "../../../ReactRouterLink";
import ExternalLink from "../../../ExternalLink";

export interface FeatureDisabledCalloutProps {
  className?: string;
  messageID: string;
  messageValues?: Values;
}

// useFeatureDisabledMessageValues provides the standard rich-text values
// (plan-page link, contact-us link, bold) for FeatureConfig.*.disabled
// messages.
export function useFeatureDisabledMessageValues(
  messageValues?: Values
): Values {
  const { appID } = useParams() as { appID: string };

  return useMemo(() => {
    const planPagePath = `/project/${appID}/billing`;
    const contactUsHref =
      "https://www.authgear.com/schedule-demo?utm_source=portal&utm_medium=link&utm_campaign=additional_order";
    return {
      planPagePath,
      contactUsHref,

      b: (chunks: React.ReactNode) => <b>{chunks}</b>,

      ReactRouterLink: (chunks: React.ReactNode) => (
        <ReactRouterLink to={planPagePath} target="_blank">
          {chunks}
        </ReactRouterLink>
      ),

      ExternalLink: (chunks: React.ReactNode) => (
        <ExternalLink href={contactUsHref}>{chunks}</ExternalLink>
      ),
      ...messageValues,
    };
  }, [appID, messageValues]);
}

// The standard FeatureConfig.*.disabled message (with plan-page /
// contact-us links) rendered as a Radix Callout, matching the info
// callouts used across the migrated screens.
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
