import React, { PropsWithChildren } from "react";
import { FormattedMessage } from "@oursky/react-messageformat";
import WidgetSubtitle from "../../WidgetSubtitle";
import cn from "classnames";

interface ConfigurationProps {
  labelKey: string;
}
const Configuration: React.VFC<PropsWithChildren<ConfigurationProps>> =
  function Configuration(props) {
    const { labelKey } = props;
    return (
      <div>
        <WidgetSubtitle>
          <FormattedMessage id={labelKey} />
        </WidgetSubtitle>
        <div className={cn("mt-[0.3125rem]")}>{props.children}</div>
      </div>
    );
  };

export default Configuration;
