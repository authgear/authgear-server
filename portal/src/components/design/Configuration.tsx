import React, { PropsWithChildren } from "react";
import { Text } from "@radix-ui/themes";
import { FormattedMessage } from "../../intl";
import cn from "classnames";

interface ConfigurationProps {
  labelKey: string;
}
const Configuration: React.VFC<PropsWithChildren<ConfigurationProps>> =
  function Configuration(props) {
    const { labelKey } = props;
    return (
      <div className={cn("flex", "flex-col", "gap-2")}>
        <Text as="p" size="2" weight="medium">
          <FormattedMessage id={labelKey} />
        </Text>
        {props.children}
      </div>
    );
  };

export default Configuration;
