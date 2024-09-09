import React, { useContext } from "react";
import { Context as MessageContext } from "@oursky/react-messageformat";
import CommandBarButton from "../../CommandBarButton";

interface ClearAllButtonProps {
  className?: string;
  onClick: () => void;
}

export const ClearAllButton: React.VFC<ClearAllButtonProps> =
  function ClearAllButton({ className, onClick }: ClearAllButtonProps) {
    const { renderToString } = useContext(MessageContext);

    return (
      <CommandBarButton
        className={className}
        key="clear"
        iconProps={{ iconName: "" }}
        onClick={onClick}
        text={renderToString("AuditLogScreen.clear-all-filters")}
        styles={{ root: { color: "rgba(89, 86, 83, 0.4)" } }}
      />
    );
  };
