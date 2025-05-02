import React, { useCallback, useMemo, useRef } from "react";
import styles from "./ImageInput.module.css";

import { FormattedMessage } from "@oursky/react-messageformat";
import { SecondaryButton } from "../SecondaryButton/SecondaryButton";
import { IconButton, IconButtonIcon } from "../IconButton/IconButton";
import { SquareIcon } from "../SquareIcon/SquareIcon";
import { ImageIcon } from "@radix-ui/react-icons";
import {
  base64EncodedDataToDataURI,
  dataURIToBase64EncodedData,
} from "../../../util/uri";

export type ImageFileExtension = ".jpeg" | ".png" | ".gif";

export interface ImageValue {
  base64EncodedData: string;
  extension: ImageFileExtension;
}

export interface ImageInputProps {
  sizeLimitKB?: number;

  value: ImageValue | null;
  onClickUpload?: () => void;
  onValueChange?: (value: ImageValue | null) => void;
  onError?: (error: ImageInputError) => void;
}

export enum ImageInputErrorCode {
  UNKNOWN = "UNKNOWN",
  FILE_TOO_LARGE = "FILE_TOO_LARGE",
}

export class ImageInputError extends Error {
  code: ImageInputErrorCode;
  internalError?: unknown;

  constructor(code: ImageInputErrorCode, internalError?: unknown) {
    super(`image input error: ${code}`);
    this.code = code;
    this.internalError = internalError;
  }
}

export function ImageInput({
  value,
  sizeLimitKB = 100,
  onError,
  onClickUpload,
  onValueChange,
}: ImageInputProps): React.ReactElement {
  const inputRef = useRef<HTMLInputElement>(null);

  const handleUpload = useCallback(() => {
    onClickUpload?.();
    inputRef.current?.click();
  }, [onClickUpload]);

  const clearValue = useCallback(() => {
    onValueChange?.(null);
  }, [onValueChange]);

  const handleFileChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const el = e.currentTarget;
      if (el.files == null || el.files.length === 0) {
        return;
      }

      const file = el.files[0];
      if (file.size / 1024 > sizeLimitKB) {
        onError?.(new ImageInputError(ImageInputErrorCode.FILE_TOO_LARGE));
        return;
      }
      fileToImageValue(file)
        .then((value) => {
          onValueChange?.(value);
        })
        .catch((e) => {
          onError?.(new ImageInputError(ImageInputErrorCode.UNKNOWN, e));
          console.error("unexpected error in image input:", e);
        })
        .finally(() => {
          // Reset the input so the same file can be selected again
          el.value = "";
        });
    },
    [onError, onValueChange, sizeLimitKB]
  );

  const valuesrc = useMemo(() => {
    if (value != null) {
      return base64EncodedDataToDataURI(value.base64EncodedData);
    }
    return undefined;
  }, [value]);

  return (
    <div className={styles.imageInput}>
      <button
        type="button"
        className={styles.imageInput__imageContainer}
        onClick={handleUpload}
      >
        {valuesrc == null ? (
          <SquareIcon
            className={styles.imageInput__placeholder}
            Icon={ImageIcon}
            size="7"
            radius="3"
          />
        ) : (
          <img className={styles.imageInput__preview} src={valuesrc} />
        )}
      </button>
      <div className={styles.imageInput__rightColumn}>
        <p className={styles.imageInput__hint}>
          <FormattedMessage id="ImageInput.hint" />
        </p>
        <div className={styles.imageInput__buttonContainer}>
          <SecondaryButton
            type="button"
            size="2"
            text={<FormattedMessage id="ImageInput.upload" />}
            onClick={handleUpload}
          />
          {value != null ? (
            <IconButton
              type="button"
              size="2"
              variant="destroy"
              icon={IconButtonIcon.Trash}
              onClick={clearValue}
            />
          ) : null}
        </div>
      </div>
      <input
        ref={inputRef}
        className="hidden"
        type="file"
        accept="image/*"
        onChange={handleFileChange}
      />
    </div>
  );
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

async function fileToImageValue(file: File) {
  return new Promise<ImageValue>((resolve, reject) => {
    const reader = new FileReader();
    reader.readAsDataURL(file);
    const extension = mediaTypeToExtension(file.type);
    reader.onload = () =>
      resolve({
        base64EncodedData: dataURIToBase64EncodedData(reader.result as string),
        extension,
      });
    reader.onerror = reject;
  });
}
