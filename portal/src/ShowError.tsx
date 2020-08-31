import React, { useCallback } from "react";
import { MessageBar, MessageBarType, MessageBarButton } from "@fluentui/react";
import { FormattedMessage } from "@oursky/react-messageformat";

interface ShowErrorProps {
  error: unknown;
  onRetry?: (() => void) | null;
}

const ShowError: React.FC<ShowErrorProps> = function ShowError(
  props: ShowErrorProps
) {
  const { error, onRetry } = props;

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
    children.push(<br key="2" />);
    children.push(<React.Fragment key="3">{error.stack}</React.Fragment>);
  } else {
    children.push(<React.Fragment key="4">{String(error)}</React.Fragment>);
  }

  let actions;
  if (onRetry != null) {
    actions = (
      <MessageBarButton onClick={onClickRetry}>
        <FormattedMessage id="show-error.retry" />
      </MessageBarButton>
    );
  }

  return (
    <MessageBar messageBarType={MessageBarType.error} actions={actions}>
      {children}
    </MessageBar>
  );
};

export default ShowError;
