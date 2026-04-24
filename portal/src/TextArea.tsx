import React from "react";
import type { ITextFieldProps } from "@fluentui/react";
import TextField from "./TextField";

/** v1 multiline input; wraps {@link TextField} with `multiline` enabled. */
export type TextAreaProps = Omit<ITextFieldProps, "multiline">;

const TextArea: React.VFC<TextAreaProps> = function TextArea(props) {
  return <TextField {...props} multiline={true} />;
};

export default TextArea;
