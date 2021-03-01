import React, { useMemo, useRef, useCallback } from "react";
import cn from "classnames";
import { Image, DefaultButton, PrimaryButton, ImageFit } from "@fluentui/react";
import { FormattedMessage } from "@oursky/react-messageformat";
import { useSystemConfig } from "./context/SystemConfigContext";
import {
  base64EncodedDataToDataURI,
  dataURIToBase64EncodedData,
} from "./util/uri";

import styles from "./ImageFilePicker.module.scss";

export type ImageFileExtension = ".jpeg" | ".png" | ".gif";

export interface ImageFilePickerProps {
  className?: string;
  base64EncodedData?: string;
  onChange?: (
    base64EncodedData: string | undefined,
    extension: ImageFileExtension | undefined
  ) => void;
}

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

const ImageFilePicker: React.FC<ImageFilePickerProps> = function ImageFilePicker(
  props: ImageFilePickerProps
) {
  const { className, base64EncodedData, onChange } = props;

  const hasImage = base64EncodedData != null;

  const { themes } = useSystemConfig();

  const src = useMemo(() => {
    if (base64EncodedData != null) {
      return base64EncodedDataToDataURI(base64EncodedData);
    }
    return undefined;
  }, [base64EncodedData]);

  const inputRef = useRef<HTMLInputElement | null>(null);

  const onClickRemoveImage = useCallback(
    (e: React.MouseEvent<HTMLElement>) => {
      e.preventDefault();
      e.stopPropagation();
      // Reset the input value so that onChange can fire again.
      // If we do not do this, the state of the input and the state of component
      // becomes out of sync.
      if (inputRef.current != null) {
        inputRef.current.value = "";
      }
      onChange?.(undefined, undefined);
    },
    [onChange]
  );

  const onClickSelectImage = useCallback((e: React.MouseEvent<HTMLElement>) => {
    e.preventDefault();
    e.stopPropagation();
    // Emulate a click on the input to open the user agent file select dialog.
    inputRef.current?.click();
  }, []);

  const onInputChange = useCallback(
    (e?: React.SyntheticEvent<HTMLInputElement>) => {
      const target = e?.target;
      if (target instanceof HTMLInputElement) {
        const file = target.files?.[0];
        if (file != null) {
          const extension = mediaTypeToExtension(file.type);
          const reader = new FileReader();
          reader.addEventListener("load", function () {
            const result = reader.result;
            if (typeof result === "string") {
              onChange?.(dataURIToBase64EncodedData(result), extension);
            }
          });
          reader.readAsDataURL(file);
        }
      }
    },
    [onChange]
  );

  const borderColor = themes.main.semanticColors.inputBorder;

  return (
    <div className={cn(className, styles.root)}>
      <input
        ref={inputRef}
        className={styles.input}
        type="file"
        accept="image/png, image/jpeg, image/gif"
        onChange={onInputChange}
      />
      <Image
        src={src}
        className={styles.image}
        styles={{
          root: {
            borderColor,
            borderWidth: "1px",
            borderStyle: "solid",
          },
        }}
        imageFit={ImageFit.centerContain}
        maximizeFrame={true}
      />
      {hasImage ? (
        <PrimaryButton
          className={styles.button}
          onClick={onClickRemoveImage}
          theme={themes.destructive}
        >
          <FormattedMessage id={"ImageFilePicker.remove"} />
        </PrimaryButton>
      ) : (
        <DefaultButton className={styles.button} onClick={onClickSelectImage}>
          <FormattedMessage id="ImageFilePicker.upload" />
        </DefaultButton>
      )}
    </div>
  );
};

export default ImageFilePicker;
