import React, { useCallback } from "react";
import type { FallbackProps } from "react-error-boundary";
import { Text } from "@fluentui/react";
import ErrorBoundSuspense, {
  type ErrorBoundSuspenseProps,
} from "./ErrorBoundSuspense";
import ShowLoading from "./ShowLoading";
import PrimaryButton from "./PrimaryButton";
import styles from "./FlavoredErrorBoundSuspense.module.css";
import { FormattedMessage } from "@oursky/react-messageformat";

export interface FlavoredErrorBoundSuspenseProps {
  factory: ErrorBoundSuspenseProps["factory"];
  children: ErrorBoundSuspenseProps["children"];
}

function FallbackComponent(props: FallbackProps): React.ReactElement<any, any> {
  const { resetErrorBoundary } = props;
  const onClick = useCallback(
    (e: React.MouseEvent<unknown>) => {
      e.preventDefault();
      e.stopPropagation();
      resetErrorBoundary();
    },
    [resetErrorBoundary]
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
      errorBoundaryFallbackComponent={FallbackComponent}
    />
  );
}

export default FlavoredErrorBoundSuspense;
