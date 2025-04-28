import React, { useCallback, useRef } from "react";
import styles from "./ImageInput.module.css";

import { FormattedMessage } from "@oursky/react-messageformat";
import { SecondaryButton } from "../SecondaryButton/SecondaryButton";
import { IconButton, IconButtonIcon } from "../IconButton/IconButton";
import { SquareIcon } from "../SquareIcon/SquareIcon";
import { ImageIcon } from "@radix-ui/react-icons";

export interface ImageInputProps {
  sizeLimitKB?: number;

  value: string | null; // Must be a base64 data url of an image if not null
  onValueChange?: (imageBase64DataURL: string | null) => void;
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
  onValueChange,
}: ImageInputProps): React.ReactElement {
  const inputRef = useRef<HTMLInputElement>(null);

  const handleUpload = useCallback(() => {
    inputRef.current?.click();
  }, []);

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
      fileToBase64DataURL(file)
        .then((url) => {
          onValueChange?.(url);
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

  return (
    <div className={styles.imageInput}>
      <button
        type="button"
        className={styles.imageInput__imageContainer}
        onClick={handleUpload}
      >
        {value == null ? (
          <SquareIcon
            className={styles.imageInput__placeholder}
            Icon={ImageIcon}
            size="7"
            radius="3"
          />
        ) : (
          <img className={styles.imageInput__preview} src={value} />
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

async function fileToBase64DataURL(file: File) {
  return new Promise<string>((resolve, reject) => {
    const reader = new FileReader();
    reader.readAsDataURL(file);
    reader.onload = () => resolve(reader.result as string);
    reader.onerror = reject;
  });
}
