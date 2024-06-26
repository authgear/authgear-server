import React, { PropsWithChildren } from "react";
import { FormattedMessage } from "@oursky/react-messageformat";
import cn from "classnames";
import WidgetTitle from "../../WidgetTitle";

interface ConfigurationGroupProps {
  labelKey: string;
}
const ConfigurationGroup: React.VFC<
  PropsWithChildren<ConfigurationGroupProps>
> = function ConfigurationGroup(props) {
  const { labelKey } = props;
  return (
    <div className={cn("space-y-4")}>
      <WidgetTitle>
        <FormattedMessage id={labelKey} />
      </WidgetTitle>
      {props.children}
    </div>
  );
};

export default ConfigurationGroup;
