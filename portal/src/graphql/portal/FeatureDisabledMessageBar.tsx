import React, { useMemo } from "react";
import { IMessageBarProps } from "@fluentui/react";
import { useParams } from "react-router-dom";
import { FormattedMessage, Values } from "@oursky/react-messageformat";
import BlueMessageBar from "../../BlueMessageBar";

export interface FeatureDisabledMessageBarProps extends IMessageBarProps {
  messageID: string;
  messageValues?: Values;
}

const FeatureDisabledMessageBar: React.VFC<FeatureDisabledMessageBarProps> =
  function FeatureDisabledMessageBar(props: FeatureDisabledMessageBarProps) {
    const { messageID, messageValues, ...rest } = props;
    const { appID } = useParams() as { appID: string };

    const values = useMemo(() => {
      return {
        planPagePath: `/project/${appID}/billing`,
        ...messageValues,
      };
    }, [appID, messageValues]);

    return (
      <BlueMessageBar {...rest}>
        <FormattedMessage id={messageID} values={values} />
      </BlueMessageBar>
    );
  };

export default FeatureDisabledMessageBar;
