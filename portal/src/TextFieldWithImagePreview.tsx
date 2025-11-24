import React, { useEffect, useMemo, useState } from "react";
import { Label } from "@fluentui/react";
import FormTextField, { FormTextFieldProps } from "./FormTextField";

import iconPreview from "./images/preview.svg";
import iconPreviewUnavailable from "./images/preview-unavailable.svg";

import cn from "classnames";

export interface TextFieldWithImagePreviewProps extends FormTextFieldProps {}

const TextFieldWithImagePreview: React.VFC<TextFieldWithImagePreviewProps> =
  function TextFieldWithImagePreview(props: TextFieldWithImagePreviewProps) {
    const { label, ...rest } = props;
    const [hasError, setHasError] = useState(false);

    useEffect(() => {
      setHasError(false);
    }, [rest.value]);

    const onError = () => {
      setHasError(true);
    };

    const renderImg = useMemo(() => {
      if (hasError || !rest.value) {
        return (
          <div
            className={cn(
              "h-10",
              "w-10",
              "flex",
              "flex-row",
              "justify-center",
              "items-center",
              "bg-neutral-lighter",
              "rounded"
            )}
          >
            {hasError ? (
              <img src={iconPreviewUnavailable} />
            ) : (
              <img src={iconPreview} />
            )}
          </div>
        );
      }
      return (
        <img
          src={rest.value}
          className={cn("max-h-12.5", "max-w-[95%]")}
          onError={onError}
        />
      );
    }, [hasError, rest.value]);

    return (
      <div className={cn("flex", "flex-col")}>
        <Label>{label}</Label>
        <div
          className={cn(
            "border",
            "border-solid",
            "border-neutral-light",
            "flex",
            "flex-row",
            "justify-center",
            "items-center",
            "h-29",
            "mb-1"
          )}
        >
          {renderImg}
        </div>
        <FormTextField {...rest} />
      </div>
    );
  };

export default TextFieldWithImagePreview;
