/* global process */
import React, { useCallback } from "react";
import { MessageBar, MessageBarType } from "@fluentui/react";
import { FormattedMessage } from "@oursky/react-messageformat";
import MessageBarButton from "./MessageBarButton";

interface ShowErrorProps {
  error: unknown;
  onRetry?: (() => void) | null;
}

const ShowError: React.VFC<ShowErrorProps> = function ShowError(
  props: ShowErrorProps
) {
  const { error, onRetry } = props;

  const showErrorStack = process.env.NODE_ENV === "development";

  const onClickRetry = useCallback(
    (e: React.MouseEvent<HTMLElement>) => {
      e.stopPropagation();
      e.preventDefault();
      onRetry?.();
    },
    [onRetry]
  );

  const children = [];
  if (error instanceof Error) {
    children.push(
      <React.Fragment key="1">
        {error.name}: {error.message}
      </React.Fragment>
    );
    if (showErrorStack) {
      children.push(<br key="2" />);
      children.push(<React.Fragment key="3">{error.stack}</React.Fragment>);
    }
  } else {
    children.push(<React.Fragment key="4">{String(error)}</React.Fragment>);
  }

  let actions;
  if (onRetry != null) {
    actions = (
      <MessageBarButton
        onClick={onClickRetry}
        text={<FormattedMessage id="show-error.retry" />}
      />
    );
  }

  return (
    <MessageBar messageBarType={MessageBarType.error} actions={actions}>
      {children}
    </MessageBar>
  );
};

export default ShowError;
