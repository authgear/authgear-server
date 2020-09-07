import React from "react";

export const useTextField = (
  initialValue: string
): { value: string; onChange: (_event: any, value?: string) => void } => {
  const [textFieldValue, setTextFieldValue] = React.useState(initialValue);
  const onChange = React.useCallback(
    (_event, value?: string) => {
      setTextFieldValue(value ?? "");
    },
    [setTextFieldValue]
  );
  return {
    value: textFieldValue,
    onChange,
  };
};

export const useCheckbox = (
  initialValue: boolean
): { value: boolean; onChange: (_event: any, value?: boolean) => void } => {
  const [checked, setChecked] = React.useState(initialValue);
  const onChange = React.useCallback(
    (_event, value?: boolean) => {
      setChecked(!!value);
    },
    [setChecked]
  );
  return {
    value: checked,
    onChange,
  };
};
