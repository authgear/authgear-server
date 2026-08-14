import React, { useMemo } from "react";
import { IMessageBarProps } from "@fluentui/react";
import { useParams } from "react-router-dom";
import { FormattedMessage, Values } from "../../intl";
import BlueMessageBar from "../../BlueMessageBar";
import ReactRouterLink from "../../ReactRouterLink";
import ExternalLink from "../../ExternalLink";

export interface FeatureDisabledMessageBarProps extends IMessageBarProps {
  messageID: string;
  messageValues?: Values;
}

// useFeatureDisabledMessageValues provides the standard rich-text values
// (plan-page link, contact-us link, bold) for FeatureConfig.*.disabled
// messages. Shared by the Fluent message bar below and the v2
// FeatureDisabledCallout.
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

const FeatureDisabledMessageBar: React.VFC<FeatureDisabledMessageBarProps> =
  function FeatureDisabledMessageBar(props: FeatureDisabledMessageBarProps) {
    const { messageID, messageValues, ...rest } = props;
    const values = useFeatureDisabledMessageValues(messageValues);

    return (
      <BlueMessageBar {...rest}>
        <FormattedMessage id={messageID} values={values} />
      </BlueMessageBar>
    );
  };

export default FeatureDisabledMessageBar;
