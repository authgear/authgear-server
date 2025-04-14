import React from "react";
import cn from "classnames";
import PhoneTextField, { PhoneTextFieldProps } from "../../PhoneTextField";
import styles from "./OnboardingSurveyPhoneInput.module.css";

export interface OnboardingSurveyPhoneInputProps
  extends Pick<
    PhoneTextFieldProps,
    "initialCountry" | "inputValue" | "onChange"
  > {}

export function OnboardingSurveyPhoneInput(
  props: OnboardingSurveyPhoneInputProps
): React.ReactElement {
  return (
    <PhoneTextField
      {...props}
      inputContainerClassName={cn(
        "rt-TextFieldRoot rt-r-size-2 rt-variant-surface",
        styles.onboardingSurveyPhoneInput__container
      )}
      inputClassNameOverride="rt-reset rt-TextFieldInput h-full"
    />
  );
}
