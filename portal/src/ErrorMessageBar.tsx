import React, {
  createContext,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import { ParsedAPIError } from "./error/parse";
import { Text } from "@radix-ui/themes";
import { useCalloutToast } from "./components/v2/Callout/Callout";
import { FormattedMessage } from "./intl";
import { Link } from "react-router-dom";
import ExternalLink from "./ExternalLink";

interface ErrorMessageBarContext {
  readonly errors: readonly ParsedAPIError[];
  setErrors: (errors: readonly ParsedAPIError[]) => void;
}

const context = createContext<ErrorMessageBarContext | undefined>(undefined);

export interface ErrorMessageBarProps {
  children?: React.ReactNode;
}

function renderError(err: ParsedAPIError, key: number): React.ReactElement {
  return (
    <Text as="p" size="2" key={key}>
      {err.messageID ? (
        <FormattedMessage
          id={err.messageID ?? ""}
          values={{
            ...err.arguments,

            reactRouterLink: (chunks: React.ReactNode) => (
              <Link to={err.arguments?.to ?? err.arguments?.href}>
                {chunks}
              </Link>
            ),

            externalLink: (chunks: React.ReactNode) => (
              <ExternalLink
                href={err.arguments?.to ?? err.arguments?.href}
                target="_blank"
                rel="noreferrer"
              >
                {chunks}
              </ExternalLink>
            ),

            docLink: (chunks: React.ReactNode) => (
              <ExternalLink href={err.arguments?.to ?? err.arguments?.href}>
                {chunks}
              </ExternalLink>
            ),

            b: (chunks: React.ReactNode) => <b>{chunks}</b>,

            strong: (chunks: React.ReactNode) => <strong>{chunks}</strong>,

            code: (chunks: React.ReactNode) => <code>{chunks}</code>,
          }}
        />
      ) : (
        err.message ?? ""
      )}
    </Text>
  );
}

// Despite the name, errors are surfaced as an error toast in the bottom
// right corner; the children are always rendered in place.
export const ErrorMessageBar: React.VFC<ErrorMessageBarProps> = (
  props: ErrorMessageBarProps
) => {
  const ctx = useContext(context);
  if (ctx === undefined) {
    throw new Error("ErrorMessageBarContext not provided");
  }
  const { errors } = ctx;
  const { showToast } = useCalloutToast();

  const shownErrorsRef = useRef<readonly ParsedAPIError[]>([]);
  useEffect(() => {
    if (errors.length > 0 && errors !== shownErrorsRef.current) {
      showToast({
        type: "error",
        text: errors.map((err, i) => renderError(err, i)),
      });
    }
    shownErrorsRef.current = errors;
  }, [errors, showToast]);

  return <>{props.children}</>;
};

export const ErrorMessageBarContextProvider: React.VFC<
  React.PropsWithChildren<{
    readonly errors?: readonly ParsedAPIError[];
  }>
> = ({ errors: propsErrors, children }) => {
  const [errors, setErrors] = useState<readonly ParsedAPIError[]>([]);

  useEffect(() => {
    if (propsErrors !== undefined) {
      // eslint-disable-next-line react-hooks/set-state-in-effect
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
