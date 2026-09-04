import React, { useContext, useMemo } from "react";
import { Button } from "@radix-ui/themes";
import { UpdateIcon } from "@radix-ui/react-icons";
import { DateTime } from "luxon";
import { Context as MessageContext } from "../../intl";
import { Tooltip } from "../v2/Tooltip/Tooltip";
import styles from "./RefreshButton.module.css";

interface RefreshButtonProps {
  onClick: () => void;
  lastUpdatedAt: Date;
}

export const RefreshButton: React.VFC<RefreshButtonProps> =
  function RefreshButton({ onClick, lastUpdatedAt }: RefreshButtonProps) {
    const { renderToString, locale } = useContext(MessageContext);

    const tooltipContent = useMemo(() => {
      return renderToString("AuditLogScreen.last-update-at", {
        datetime: DateTime.fromJSDate(lastUpdatedAt).toRelative({
          locale,
        }),
      });
    }, [lastUpdatedAt, locale, renderToString]);

    return (
      <Tooltip content={tooltipContent}>
        <Button
          type="button"
          className={styles.refreshButton}
          size="2"
          variant="ghost"
          color="gray"
          highContrast={true}
          onClick={onClick}
        >
          <UpdateIcon />
          {renderToString("AuditLogScreen.refresh")}
        </Button>
      </Tooltip>
    );
  };
