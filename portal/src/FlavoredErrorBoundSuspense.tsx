import React from "react";
import { Text } from "@fluentui/react";
import ErrorBoundSuspense, {
  type ErrorBoundSuspenseProps,
  type ErrorBoundaryFallbackProps,
} from "./ErrorBoundSuspense";
import ShowLoading from "./ShowLoading";
import styles from "./FlavoredErrorBoundSuspense.module.css";

export interface FlavoredErrorBoundSuspenseProps {
  factory: ErrorBoundSuspenseProps["factory"];
  children: ErrorBoundSuspenseProps["children"];
}

export function FallbackComponent(
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
  const { resetError: _resetError } = props;

  // It is intentionally not using FormattedMessage so that it can always show something.
  return (
    <div className={styles.fallbackContainer}>
      <Text as="p">{"Something went wrong. Please refresh this page."}</Text>
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
