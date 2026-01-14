import React, {
  createContext,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "react";
import { ParsedAPIError } from "./error/parse";
import { MessageBar, MessageBarType, Text } from "@fluentui/react";
import { FormattedMessage } from "./intl";

interface ErrorMessageBarContext {
  readonly errors: readonly ParsedAPIError[];
  setErrors: (errors: readonly ParsedAPIError[]) => void;
}

const context = createContext<ErrorMessageBarContext | undefined>(undefined);

export interface ErrorMessageBarProps {
  children?: React.ReactNode;
}

export const ErrorMessageBar: React.VFC<ErrorMessageBarProps> = (
  props: ErrorMessageBarProps
) => {
  const ctx = useContext(context);
  if (ctx === undefined) {
    throw new Error("ErrorMessageBarContext not provided");
  }
  const { errors } = ctx;
  if (errors.length === 0) {
    return <>{props.children}</>;
  }

  return (
    <MessageBar messageBarType={MessageBarType.error}>
      {errors.map((err, i) => (
        <Text key={i}>
          {err.messageID ? (
            <FormattedMessage id={err.messageID ?? ""} values={err.arguments} />
          ) : (
            err.message ?? ""
          )}
        </Text>
      ))}
    </MessageBar>
  );
};

export const ErrorMessageBarContextProvider: React.VFC<
  React.PropsWithChildren<{
    readonly errors?: readonly ParsedAPIError[];
  }>
> = ({ errors: propsErrors, children }) => {
  const [errors, setErrors] = useState<readonly ParsedAPIError[]>([]);

  useEffect(() => {
    if (propsErrors !== undefined) {
      setErrors(propsErrors);
    }
  }, [propsErrors]);

  const value = useMemo<ErrorMessageBarContext>(() => {
    return {
      errors,
      setErrors,
    };
  }, [errors]);

  return <context.Provider value={value}>{children}</context.Provider>;
};

export function useErrorMessageBarContext(): ErrorMessageBarContext {
  const ctx = useContext(context);
  if (ctx === undefined) {
    throw new Error("ErrorMessageBarContext not provided");
  }
  return ctx;
}
