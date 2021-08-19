import React, {
  useRef,
  useState,
  useCallback,
  useEffect,
  useContext,
} from "react";
import {
  Callout,
  DirectionalHint,
  IButtonProps,
  IIconProps,
  Text,
} from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import { copyToClipboard } from "../util/clipboard";

export interface CopyButtonProps {
  id: string;
  onClick: IButtonProps["onClick"];
  text: string;
  title: string;
  ariaLabel: string;
  onMouseLeave: IButtonProps["onMouseLeave"];
  iconProps: IIconProps;
}

export interface UseCopyFeedbackInput {
  textToCopy: string;
}

export interface UseCopyFeedbackOutput {
  copyButtonProps: CopyButtonProps;
  Feedback: React.FC;
}

const iconProps = {
  iconName: "Copy",
};

function useDelayedAction(delayed: () => void): (delay: number) => void {
  // tuple is used instead of number because we want to trigger the effect even when delay argument is the same
  const [delay, setDelay] = useState<[number] | null>(null);

  useEffect(() => {
    if (!delay) {
      return () => {};
    }
    const timer = setTimeout(() => {
      delayed();
    }, delay[0]);
    return () => {
      clearTimeout(timer);
    };
  }, [delay, delayed]);

  return (delay: number) => setDelay([delay]);
}

const CALLOUT_STYLES = {
  root: {
    padding: "8px",
  },
};

export function useCopyFeedback(
  input: UseCopyFeedbackInput
): UseCopyFeedbackOutput {
  const { current: id } = useRef("id-" + String(Math.random()).slice(2));
  const { textToCopy } = input;
  const [isCalloutVisible, setIsCalloutVisible] = useState(false);
  const dismissCallout = useCallback(() => setIsCalloutVisible(false), []);
  const scheduleCalloutDismiss = useDelayedAction(dismissCallout);
  const { renderToString } = useContext(Context);

  const onClick = useCallback(() => {
    copyToClipboard(textToCopy);
    setIsCalloutVisible(true);
    scheduleCalloutDismiss(2000);
  }, [textToCopy, scheduleCalloutDismiss]);

  const onMouseLeave = useCallback(() => {
    scheduleCalloutDismiss(500);
  }, [scheduleCalloutDismiss]);

  const title = renderToString("copy");
  const ariaLabel = renderToString("copy");
  const text = renderToString("copy");

  const Feedback = useCallback(() => {
    if (!isCalloutVisible) {
      return null;
    }
    return (
      <Callout
        target={"#" + id}
        directionalHint={DirectionalHint.topCenter}
        onDismiss={dismissCallout}
        styles={CALLOUT_STYLES}
      >
        <Text variant="small">
          <FormattedMessage id="copied-to-clipboard" />
        </Text>
      </Callout>
    );
  }, [isCalloutVisible, id, dismissCallout]);

  return {
    copyButtonProps: {
      id,
      text,
      title,
      ariaLabel,
      onClick,
      onMouseLeave,
      iconProps,
    },
    Feedback: Feedback,
  };
}
