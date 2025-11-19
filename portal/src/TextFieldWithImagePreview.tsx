import React from "react";
import { Label } from "@fluentui/react";
import FormTextField, { FormTextFieldProps } from "./FormTextField";

export interface TextFieldWithImagePreviewProps extends FormTextFieldProps {}

const TextFieldWithImagePreview: React.VFC<TextFieldWithImagePreviewProps> =
  function TextFieldWithImagePreview(props: TextFieldWithImagePreviewProps) {
    const { label, ...rest } = props;

    return (
      <div style={{ display: "flex", flexDirection: "column" }}>
        <Label>{label}</Label>
        <FormTextField {...rest} />
      </div>
    );
  };

export default TextFieldWithImagePreview;
