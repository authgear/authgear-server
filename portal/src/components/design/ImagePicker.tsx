import React, { useCallback, useRef } from "react";
import { FormattedMessage } from "@oursky/react-messageformat";
import { Image, ImageFit } from "@fluentui/react";

import cn from "classnames";

import PrimaryButton from "../../PrimaryButton";
import DefaultButton from "../../DefaultButton";

import { useSystemConfig } from "../../context/SystemConfigContext";
import {
  base64EncodedDataToDataURI,
  dataURIToBase64EncodedData,
} from "../../util/uri";

import styles from "./ImagePicker.module.css";

type ImageFileExtension = ".jpeg" | ".png" | ".gif";
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

interface ImagePickerProps {
  base64EncodedData: string | null;
  onChange: (
    image: {
      base64EncodedData: string;
      extension: ImageFileExtension;
    } | null
  ) => void;
}
export const ImagePicker: React.VFC<ImagePickerProps> = function ImagePicker(
  props
) {
  const { base64EncodedData, onChange } = props;
  const inputRef = useRef<HTMLInputElement | null>(null);

  const { themes } = useSystemConfig();

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
            }
          });
          reader.readAsDataURL(file);
        }
      }
    },
    [onChange]
  );

  const onClickUploadButton = useCallback(
    (e: React.MouseEvent<HTMLElement>) => {
      e.preventDefault();
      e.stopPropagation();
      inputRef.current?.click();
    },
    []
  );

  const onClickRemoveButton = useCallback(
    (e: React.MouseEvent<HTMLElement>) => {
      e.preventDefault();
      e.stopPropagation();
      if (inputRef.current) {
        inputRef.current.value = "";
      }
      onChange(null);
    },
    [onChange]
  );

  const onClickImage = useCallback(
    (e: React.MouseEvent<HTMLDivElement>) => {
      e.preventDefault();
      e.stopPropagation();
      if (base64EncodedData) {
        return;
      }
      inputRef.current?.click();
    },
    [base64EncodedData]
  );

  return (
    <div className={cn("flex", "items-center", "gap-x-6")}>
      <div
        className={cn(
          "flex",
          "items-center",
          "justify-center",
          "h-30",
          "w-30",
          "bg-neutral-light",
          "rounded",
          "border",
          "border-solid",
          "border-neutral-tertiaryAlt"
        )}
        onClick={onClickImage}
      >
        {base64EncodedData == null ? (
          <span className={styles.icImagePlaceholder}></span>
        ) : (
          <Image
            src={base64EncodedDataToDataURI(base64EncodedData)}
            className={cn("h-full", "w-full")}
            imageFit={ImageFit.centerCover}
            maximizeFrame={true}
          />
        )}
      </div>
      {base64EncodedData == null ? (
        <DefaultButton
          text={
            <FormattedMessage id="DesignScreen.configuration.imagePicker.upload" />
          }
          onClick={onClickUploadButton}
        />
      ) : (
        <PrimaryButton
          theme={themes.destructive}
          text={<FormattedMessage id={"ImageFilePicker.remove"} />}
          onClick={onClickRemoveButton}
        />
      )}
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
