import React, { useCallback, useRef } from "react";
import cn from "classnames";
import { dataURIToBase64EncodedData } from "../../util/uri";

export type ImageFileExtension = ".jpeg" | ".png" | ".gif";

function mediaTypeToExtension(mime: string): ImageFileExtension {
  switch (mime) {
    case "image/png":
      return ".png";
    case "image/jpeg":
      return ".jpeg";
    case "image/gif":
      return ".gif";
    default:
      throw new Error(`unsupported media type: ${mime}`);
  }
}

interface BaseImagePickerProps {
  className?: string;
  base64EncodedData: string | null;
  onChange: (
    image: {
      base64EncodedData: string;
      extension: ImageFileExtension;
    } | null
  ) => void;
  children?: (renderProps: {
    showFilePicker: () => void;
    clearImage: () => void;
  }) => React.ReactNode | null;
}
const BaseImagePicker: React.VFC<BaseImagePickerProps> =
  function BaseImagePicker(props) {
    const { className, base64EncodedData, onChange } = props;
    const inputRef = useRef<HTMLInputElement | null>(null);

    const onInputChange = useCallback(
      (e?: React.FormEvent<HTMLInputElement>) => {
        const target = e?.target;
        if (target instanceof HTMLInputElement) {
          const file = target.files?.[0];
          if (file != null) {
            const extension = mediaTypeToExtension(file.type);
            const reader = new FileReader();
            reader.addEventListener("load", function () {
              const result = reader.result;
              if (typeof result === "string") {
                onChange({
                  base64EncodedData: dataURIToBase64EncodedData(result),
                  extension,
                });
                if (inputRef.current) {
                  // The way that we use <input type=file> is we read the full inut as bytes.
                  // After reading, the value of the input is no longer relevant to us.
                  // Also in Chrome, if the same file is selected again, onChange will NOT be called.
                  // Therefore, every time we finish reading, we reset the value to empty.
                  inputRef.current.value = "";
                }
              }
            });
            reader.readAsDataURL(file);
          }
        }
      },
      [onChange]
    );

    const showFilePicker = useCallback(() => {
      if (base64EncodedData != null) {
        return;
      }
      inputRef.current?.click();
    }, [base64EncodedData]);

    const clearImage = useCallback(() => {
      onChange(null);
    }, [onChange]);

    return (
      <div className={className}>
        {props.children?.({
          showFilePicker,
          clearImage,
        })}
        <input
          ref={inputRef}
          className={cn("hidden")}
          type="file"
          accept="image/png, image/jpeg, image/gif"
          onChange={onInputChange}
        />
      </div>
    );
  };

export default BaseImagePicker;
