import React, { useCallback } from "react";
import { Text } from "@fluentui/react";
import ErrorBoundSuspense, {
  type ErrorBoundSuspenseProps,
  type ErrorBoundaryFallbackProps,
} from "./ErrorBoundSuspense";
import ShowLoading from "./ShowLoading";
import PrimaryButton from "./PrimaryButton";
import styles from "./FlavoredErrorBoundSuspense.module.css";
import { FormattedMessage } from "@oursky/react-messageformat";

export interface FlavoredErrorBoundSuspenseProps {
  factory: ErrorBoundSuspenseProps["factory"];
  children: ErrorBoundSuspenseProps["children"];
}

function FallbackComponent(
  props: ErrorBoundaryFallbackProps
): React.ReactElement<any, any> {
  // The definition of resetError is
  //   resetError(): void;
  // instead of
  //   resetError: () => void;
  // This makes TypeScript think it is method that could access this.
  // But I think that is just a mistake in the type definition.
  // So it should be safe to ignore.
  // eslint-disable-next-line @typescript-eslint/unbound-method
  const { resetError } = props;
  const onClick = useCallback(
    (e: React.MouseEvent<unknown>) => {
      e.preventDefault();
      e.stopPropagation();
      resetError();
    },
    [resetError]
  );
  return (
    <div className={styles.fallbackContainer}>
      <Text as="p">
        <FormattedMessage id="FlavoredErrorBoundSuspense.message" />
      </Text>
      <PrimaryButton
        onClick={onClick}
        text={<FormattedMessage id="FlavoredErrorBoundSuspense.reload" />}
      />
    </div>
  );
}

function FlavoredErrorBoundSuspense(
  props: FlavoredErrorBoundSuspenseProps
): React.ReactElement<any, any> | null {
  return (
    <ErrorBoundSuspense
      {...props}
      suspenseFallback={<ShowLoading />}
      errorBoundaryFallback={FallbackComponent}
    />
  );
}

export default FlavoredErrorBoundSuspense;
