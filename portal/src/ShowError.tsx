/* global process */
import React, { useCallback, useMemo } from "react";
import { Button, Callout } from "@radix-ui/themes";
import { ExclamationTriangleIcon } from "@radix-ui/react-icons";
import { FormattedMessage } from "./intl";
import { parseAPIErrors, parseRawError } from "./error/parse";
import ErrorRenderer from "./ErrorRenderer";

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

  const topErrors = useMemo(() => {
    const parsed = parseAPIErrors(parseRawError(error), [], []).topErrors;
    // Never render an empty box.
    return parsed.length > 0
      ? parsed
      : parseAPIErrors(parseRawError(new Error()), [], []).topErrors;
  }, [error]);

  const children: React.ReactNode[] = [
    <ErrorRenderer key="message" errors={topErrors} />,
  ];
  // The raw error is for developers only; users get the parsed message.
  if (showErrorStack && error instanceof Error) {
    children.push(<br key="br" />);
    children.push(
      <React.Fragment key="raw">
        {error.name}: {error.message}
        <br />
        {error.stack}
      </React.Fragment>
    );
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
