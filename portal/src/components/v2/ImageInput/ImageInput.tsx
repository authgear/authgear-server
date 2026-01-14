import React, { useCallback, useMemo, useRef, useState } from "react";
import cn from "classnames";
import styles from "./ImageInput.module.css";

import { FormattedMessage } from "../../../intl";
import { SecondaryButton } from "../Button/SecondaryButton/SecondaryButton";
import { IconButton, IconButtonIcon } from "../IconButton/IconButton";
import { SquareIcon } from "../SquareIcon/SquareIcon";
import { ImageIcon } from "@radix-ui/react-icons";
import { Text } from "@radix-ui/themes";
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
  sizeLimitInBytes: number;
  value: ImageValue | null;
  onClickUpload?: () => void;
  onValueChange?: (value: ImageValue | null) => void;
}

type ImageInputError = "size" | "load" | "media_type";

export function ImageInput({
  value,
  sizeLimitInBytes,
  onClickUpload,
  onValueChange,
}: ImageInputProps): React.ReactElement {
  const [error, setError] = useState<ImageInputError | null>(null);
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
      if (file.size > sizeLimitInBytes) {
        setError("size");
        // Reset the input so the same file can be selected again
        el.value = "";
        return;
      }
      const extension = mediaTypeToExtension(file.type);
      if (extension == null) {
        setError("media_type");
        // Reset the input so the same file can be selected again
        el.value = "";
        return;
      }

      fileToImageValue(file, extension)
        .then((value) => {
          setError(null);
          onValueChange?.(value);
        })
        .catch((e) => {
          setError("load");
          console.error("unexpected error in image input:", e);
        })
        .finally(() => {
          // Reset the input so the same file can be selected again
          el.value = "";
        });
    },
    [onValueChange, sizeLimitInBytes]
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
        className={cn(
          styles.imageInput__imageContainer,
          valuesrc == null ? styles["imageInput__imageContainer--hover"] : null
        )}
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
        <Text
          as="p"
          size={"2"}
          weight={"regular"}
          className={styles.imageInput__hint}
        >
          <FormattedMessage id="ImageInput.hint" />
        </Text>
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
        accept="image/png, image/jpeg, image/gif"
        onChange={handleFileChange}
      />
      {error != null ? (
        <Text
          className={styles.imageInput__errorMessage}
          as="p"
          size={"1"}
          weight={"regular"}
        >
          {error === "size" ? (
            <FormattedMessage id="errors.image-too-large" />
          ) : error === "media_type" ? (
            <FormattedMessage id="errors.input-file-media-type" />
          ) : (
            <FormattedMessage id="errors.input-file-image-load" />
          )}
        </Text>
      ) : null}
    </div>
  );
}

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

async function fileToImageValue(file: File, extension: ImageFileExtension) {
  return new Promise<ImageValue>((resolve, reject) => {
    const reader = new FileReader();
    reader.readAsDataURL(file);
    reader.onload = () =>
      resolve({
        base64EncodedData: dataURIToBase64EncodedData(reader.result as string),
        extension,
      });
    reader.onerror = reject;
  });
}
