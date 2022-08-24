import React from "react";
import {
  IButtonProps,
  // eslint-disable-next-line no-restricted-imports
  MessageBarButton as FluentUIMessageBarButton,
} from "@fluentui/react";

export interface MessageBarButtonProps
  extends Omit<IButtonProps, "children" | "text"> {
  text?: React.ReactNode;
}

const MessageBarButton: React.VFC<MessageBarButtonProps> =
  function MessageBarButton(props: MessageBarButtonProps) {
    // @ts-expect-error
    return <FluentUIMessageBarButton {...props} />;
  };

export default MessageBarButton;
