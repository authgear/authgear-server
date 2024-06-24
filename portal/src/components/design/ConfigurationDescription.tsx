import React from "react";
import { FormattedMessage } from "@oursky/react-messageformat";
import WidgetDescription from "../../WidgetDescription";

interface ConfigurationDescriptionProps {
  labelKey: string;
}
const ConfigurationDescription: React.VFC<ConfigurationDescriptionProps> =
  function ConfigurationDescription(props) {
    const { labelKey } = props;
    return (
      <WidgetDescription>
        <FormattedMessage id={labelKey} />
      </WidgetDescription>
    );
  };

export default ConfigurationDescription;
