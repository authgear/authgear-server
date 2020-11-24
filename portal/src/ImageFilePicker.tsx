import React, { useMemo, useRef, useCallback } from "react";
import cn from "classnames";
import {
  Image,
  DefaultButton,
  PrimaryButton,
  ImageFit,
  Label,
} from "@fluentui/react";
import { FormattedMessage } from "@oursky/react-messageformat";
import { useSystemConfig } from "./context/SystemConfigContext";

import styles from "./ImageFilePicker.module.scss";

export interface ImageFilePickerProps {
  title: string;
  className?: string;
  base64EncodedData?: string;
  onChange?: (base64EncodedData?: string) => void;
}

function base64EncodedDataToDataURI(base64EncodedData: string): string {
  return `data:;base64,${base64EncodedData}`;
}

function dataURIToBase64EncodedData(dataURI: string): string {
  const idx = dataURI.indexOf(",");
  if (idx < 0) {
    throw new Error("not a data URI: " + dataURI);
  }
  return dataURI.slice(idx + 1);
}

const ImageFilePicker: React.FC<ImageFilePickerProps> = function ImageFilePicker(
  props: ImageFilePickerProps
) {
  const { className, base64EncodedData, onChange, title } = props;

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
      onChange?.(undefined);
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
          const reader = new FileReader();
          reader.addEventListener("load", function () {
            const result = reader.result;
            if (typeof result === "string") {
              onChange?.(dataURIToBase64EncodedData(result));
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
      <Label className={styles.label}>{title}</Label>
      <Image
        src={src}
        className={styles.image}
        alt={title}
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
          <FormattedMessage id={"ImageFilePicker.remove-image"} />
        </PrimaryButton>
      ) : (
        <DefaultButton className={styles.button} onClick={onClickSelectImage}>
          <FormattedMessage id="ImageFilePicker.select-image" />
        </DefaultButton>
      )}
    </div>
  );
};

export default ImageFilePicker;
