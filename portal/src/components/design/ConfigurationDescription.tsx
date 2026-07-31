import React from "react";
import { Text } from "@radix-ui/themes";
import { FormattedMessage } from "../../intl";

interface ConfigurationDescriptionProps {
  labelKey: string;
}
const ConfigurationDescription: React.VFC<ConfigurationDescriptionProps> =
  function ConfigurationDescription(props) {
    const { labelKey } = props;
    return (
      <Text as="p" size="2" color="gray">
        <FormattedMessage id={labelKey} />
      </Text>
    );
  };

export default ConfigurationDescription;
