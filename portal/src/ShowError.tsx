/* global process */
import React, { useCallback } from "react";
import { Button, Callout } from "@radix-ui/themes";
import { ExclamationTriangleIcon } from "@radix-ui/react-icons";
import { FormattedMessage } from "./intl";

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

  const children: React.ReactNode[] = [];
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

  return (
    <Callout.Root color="red" variant="surface" size="1">
      <Callout.Icon>
        <ExclamationTriangleIcon />
      </Callout.Icon>
      <Callout.Text>{children}</Callout.Text>
      {onRetry != null ? (
        <div>
          <Button
            type="button"
            size="1"
            variant="soft"
            color="red"
            onClick={onClickRetry}
          >
            <FormattedMessage id="show-error.retry" />
          </Button>
        </div>
      ) : null}
    </Callout.Root>
  );
};

export default ShowError;
