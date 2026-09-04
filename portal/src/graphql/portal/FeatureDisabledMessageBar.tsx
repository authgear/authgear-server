import React from "react";
import { Values } from "../../intl";
import {
  FeatureDisabledCallout,
  useFeatureDisabledMessageValues,
} from "../../components/v2/FeatureDisabledCallout/FeatureDisabledCallout";

export { useFeatureDisabledMessageValues };

export interface FeatureDisabledMessageBarProps {
  className?: string;
  messageID: string;
  messageValues?: Values;
}

// Kept for the existing import sites; renders the v2 FeatureDisabledCallout.
const FeatureDisabledMessageBar: React.VFC<FeatureDisabledMessageBarProps> =
  function FeatureDisabledMessageBar(props: FeatureDisabledMessageBarProps) {
    const { className, messageID, messageValues } = props;
    return (
      <FeatureDisabledCallout
        className={className}
        messageID={messageID}
        messageValues={messageValues}
      />
    );
  };

export default FeatureDisabledMessageBar;
