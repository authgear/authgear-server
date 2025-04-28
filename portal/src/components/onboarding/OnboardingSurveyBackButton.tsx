import React from "react";
import {
  TextButton,
  TextButtonIcon,
  TextButtonProps,
} from "../v2/TextButton/TextButton";
import { FormattedMessage } from "@oursky/react-messageformat";

export function OnboardingSurveyBackButton({
  onClick,
  disabled,
}: Pick<TextButtonProps, "onClick" | "disabled">): React.ReactElement {
  return (
    <TextButton
      size="4"
      variant="secondary"
      text={<FormattedMessage id="OnboardingSurveyScreen.actions.back" />}
      iconStart={TextButtonIcon.Back}
      onClick={onClick}
      disabled={disabled}
    />
  );
}
