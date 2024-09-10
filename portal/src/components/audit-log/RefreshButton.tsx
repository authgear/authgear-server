import React, { useContext, useMemo } from "react";
import { Context as MessageContext } from "@oursky/react-messageformat";
import CommandBarButton from "../../CommandBarButton";
import {
  DirectionalHint,
  ITooltipHostStyles,
  ITooltipProps,
  TooltipHost,
} from "@fluentui/react";
import { useId } from "@fluentui/react-hooks";
import { DateTime } from "luxon";

interface RefreshButtonProps {
  onClick: () => void;
  lastUpdatedAt: Date;
}

export const RefreshButton: React.VFC<RefreshButtonProps> =
  function RefreshButton({ onClick, lastUpdatedAt }: RefreshButtonProps) {
    const tooltipStyle: Partial<ITooltipHostStyles> = {
      root: {
        display: "flex",
        padding: "auto",
        height: "100%",
        // mobile
        "@media(max-width: 640px)": {
          height: "2.75rem", // 44px
        },
      },
    };
    const tooltipId = useId("refreshTooltip");
    const tooltipCalloutProps = {
      gapSpace: 0,
    };

    const { renderToString, locale } = useContext(MessageContext);

    const tooltipProps: ITooltipProps = useMemo(() => {
      return {
        // eslint-disable-next-line react/no-unstable-nested-components
        onRenderContent: () => {
          const tooltipcontent = renderToString(
            "AuditLogScreen.last-update-at",
            {
              datetime: DateTime.fromJSDate(lastUpdatedAt).toRelative({
                locale,
              }),
            }
          );
          return <>{tooltipcontent}</>;
        },
      };
    }, [lastUpdatedAt, locale, renderToString]);

    return (
      <TooltipHost
        styles={tooltipStyle}
        id={tooltipId}
        calloutProps={tooltipCalloutProps}
        directionalHint={DirectionalHint.bottomCenter}
        tooltipProps={tooltipProps}
      >
        <CommandBarButton
          key="refresh"
          text={renderToString("AuditLogScreen.refresh")}
          iconProps={{ iconName: "Sync" }}
          onClick={onClick}
        />
      </TooltipHost>
    );
  };
