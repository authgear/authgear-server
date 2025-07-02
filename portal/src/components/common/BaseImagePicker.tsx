import React, {
  useCallback,
  useRef,
  useState,
  useMemo,
  useContext,
} from "react";
import cn from "classnames";
import { Dialog, DialogFooter } from "@fluentui/react";
import { Context } from "@oursky/react-messageformat";
import { dataURIToBase64EncodedData } from "../../util/uri";
import PrimaryButton from "../../PrimaryButton";
import DefaultButton from "../../DefaultButton";

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

function formatSize(size: number): string {
  if (size < 1e3) {
    return `${size} B`;
  } else if (size >= 1e3 && size < 1e6) {
    return `${(size / 1e3).toFixed(0)} KB`;
  }
  return `${(size / 1e6).toFixed(0)} MB`;
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
const BaseImagePicker: React.VFC<BaseImagePickerProps> =
  function BaseImagePicker(props) {
    const { renderToString } = useContext(Context);
    const { sizeLimitInBytes, className, base64EncodedData, onChange } = props;
    const [isSizeLimitErrorDialogHidden, setIsSizeLimitErrorDialogHidden] =
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
              setIsSizeLimitErrorDialogHidden(false);
              // See Note 1
              if (inputRef.current) {
                inputRef.current.value = "";
              }
              return;
            }

            const extension = mediaTypeToExtension(file.type);
            const reader = new FileReader();
            reader.addEventListener("load", function () {
              const result = reader.result;
              if (typeof result === "string") {
                onChange({
                  base64EncodedData: dataURIToBase64EncodedData(result),
                  extension,
                });
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
      setIsSizeLimitErrorDialogHidden(true);
    }, []);

    const onRetry = useCallback(() => {
      setIsSizeLimitErrorDialogHidden(true);
      inputRef.current?.click();
    }, []);

    const dialogContentProps = useMemo(() => {
      return {
        title: renderToString("BaseImagePicker.size-dialog.title"),
        subText: renderToString("BaseImagePicker.size-dialog.description", {
          size: formatSize(sizeLimitInBytes),
        }),
      };
    }, [renderToString, sizeLimitInBytes]);

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
        <Dialog
          hidden={isSizeLimitErrorDialogHidden}
          onDismiss={onDialogDismiss}
          dialogContentProps={dialogContentProps}
        >
          <DialogFooter>
            <PrimaryButton
              onClick={onRetry}
              text={renderToString(
                "BaseImagePicker.size-dialog.button-retry-label"
              )}
            />
            <DefaultButton
              onClick={onDialogDismiss}
              text={renderToString(
                "BaseImagePicker.size-dialog.button-cancel-label"
              )}
            />
          </DialogFooter>
        </Dialog>
      </div>
    );
  };

export default BaseImagePicker;
