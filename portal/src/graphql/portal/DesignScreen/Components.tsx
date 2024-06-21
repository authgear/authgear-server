import React, { PropsWithChildren } from "react";
import cn from "classnames";
import { FormattedMessage } from "@oursky/react-messageformat";
import WidgetTitle from "../../../WidgetTitle";
import WidgetSubtitle from "../../../WidgetSubtitle";

export const Separator: React.VFC = function Separator() {
  return <div className={cn("h-px", "my-12", "bg-separator")}></div>;
};

interface ConfigurationGroupProps {
  labelKey: string;
}
export const ConfigurationGroup: React.VFC<
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

interface ConfigurationProps {
  labelKey: string;
}
export const Configuration: React.VFC<ConfigurationProps> =
  function Configuration(props) {
    const { labelKey } = props;
    return (
      <div>
        <WidgetSubtitle>
          <FormattedMessage id={labelKey} />
        </WidgetSubtitle>
      </div>
    );
  };

