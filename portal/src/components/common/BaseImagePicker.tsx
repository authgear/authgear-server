import React, { useCallback, useRef, useState, useMemo } from "react";
import cn from "classnames";
import { Button, Dialog } from "@radix-ui/themes";
import { FormattedMessage } from "../../intl";
import { dataURIToBase64EncodedData } from "../../util/uri";
import { SecondaryButton } from "../v2/Button/SecondaryButton/SecondaryButton";
import styles from "./BaseImagePicker.module.css";

export type ImageFileExtension = ".jpeg" | ".png" | ".gif";

function mediaTypeToExtension(mime: string): ImageFileExtension | null {
  switch (mime) {
    case "image/png":
      return ".png";
    case "image/jpeg":
      return ".jpeg";
    case "image/gif":
      return ".gif";
    default:
      return null;
  }
}

interface BaseImagePickerProps {
  sizeLimitInBytes: number;
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

type BaseImagePickerError = "size" | "load" | "media_type";

const BaseImagePicker: React.VFC<BaseImagePickerProps> =
  function BaseImagePicker(props) {
    const { sizeLimitInBytes, className, base64EncodedData, onChange } = props;
    const [error, setError] = useState<BaseImagePickerError | null>(null);
    const [isErrorDialogHidden, setIsErrorDialogHidden] =
      useState<boolean>(true);
    const inputRef = useRef<HTMLInputElement | null>(null);

    // Note 1
    // The way that we use <input type=file> is we read the full input as bytes.
    // After reading, the value of the input is no longer relevant to us.
    // Also in Chrome, if the same file is selected again, onChange will NOT be called.
    // Therefore, every time we finish reading, we reset the value to empty.

    const onInputChange = useCallback(
      (e?: React.FormEvent<HTMLInputElement>) => {
        const target = e?.target;
        if (target instanceof HTMLInputElement) {
          const file = target.files?.[0];
          if (file != null) {
            if (file.size > sizeLimitInBytes) {
              setError("size");
              setIsErrorDialogHidden(false);
              // See Note 1
              if (inputRef.current) {
                inputRef.current.value = "";
              }
              return;
            }

            const extension = mediaTypeToExtension(file.type);
            if (extension == null) {
              setError("media_type");
              setIsErrorDialogHidden(false);
              // See Note 1
              if (inputRef.current) {
                inputRef.current.value = "";
              }
              return;
            }
            const reader = new FileReader();
            reader.addEventListener("error", () => {
              setError("load");
              setIsErrorDialogHidden(false);
              // See Note 1
              if (inputRef.current) {
                inputRef.current.value = "";
              }
            });
            reader.addEventListener("load", () => {
              const result = reader.result;
              if (typeof result === "string") {
                onChange({
                  base64EncodedData: dataURIToBase64EncodedData(result),
                  extension,
                });
                setError(null);
                setIsErrorDialogHidden(true);
                // See Note 1
                if (inputRef.current) {
                  inputRef.current.value = "";
                }
              }
            });
            reader.readAsDataURL(file);
          }
        }
      },
      [onChange, sizeLimitInBytes]
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

    const onDialogDismiss = useCallback(() => {
      // Do not setError(null) to avoid flicking when the dialog is being dismissed.
      setIsErrorDialogHidden(true);
    }, []);

    const onDialogOpenChange = useCallback(
      (open: boolean) => {
        if (!open) {
          onDialogDismiss();
        }
      },
      [onDialogDismiss]
    );

    const onRetry = useCallback(() => {
      // Do not setError(null) to avoid flicking when the dialog is being dismissed.
      setIsErrorDialogHidden(true);
      inputRef.current?.click();
    }, []);

    const errorMessageID = useMemo(() => {
      return error === "size"
        ? "errors.image-too-large"
        : error === "media_type"
        ? "errors.input-file-media-type"
        : "errors.input-file-image-load";
    }, [error]);

    return (
      <div className={className}>
        {/* eslint-disable-next-line react-hooks/refs */}
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
        <Dialog.Root
          open={!isErrorDialogHidden}
          onOpenChange={onDialogOpenChange}
        >
          <Dialog.Content maxWidth="400px" size="3">
            <Dialog.Title>
              <FormattedMessage id="BaseImagePicker.error-dialog.title" />
            </Dialog.Title>
            <Dialog.Description size="2">
              <FormattedMessage id={errorMessageID} />
            </Dialog.Description>
            <div className={styles.actions}>
              <SecondaryButton
                size="2"
                text={
                  <FormattedMessage id="BaseImagePicker.error-dialog.button-cancel-label" />
                }
                onClick={onDialogDismiss}
              />
              <Button type="button" size="2" onClick={onRetry}>
                <FormattedMessage id="BaseImagePicker.error-dialog.button-retry-label" />
              </Button>
            </div>
          </Dialog.Content>
        </Dialog.Root>
      </div>
    );
  };

export default BaseImagePicker;
