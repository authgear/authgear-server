import React from "react";
import { FormattedMessage } from "@oursky/react-messageformat";
import { Image, ImageFit } from "@fluentui/react";

import cn from "classnames";

import BaseImagePicker, { ImageFileExtension } from "../common/BaseImagePicker";
import PrimaryButton from "../../PrimaryButton";
import DefaultButton from "../../DefaultButton";

import { useSystemConfig } from "../../context/SystemConfigContext";
import { base64EncodedDataToDataURI } from "../../util/uri";

import styles from "./ImagePicker.module.css";

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
  const { themes } = useSystemConfig();
  return (
    <BaseImagePicker
      className={cn("flex", "items-center", "gap-x-6")}
      base64EncodedData={base64EncodedData}
      onChange={onChange}
    >
      {({ showFilePicker, clearImage }) => (
        <>
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
              "border-neutral-tertiaryAlt",
              "overflow-hidden"
            )}
            onClick={showFilePicker}
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
              onClick={showFilePicker}
            />
          ) : (
            <PrimaryButton
              theme={themes.destructive}
              text={<FormattedMessage id={"ImageFilePicker.remove"} />}
              onClick={clearImage}
            />
          )}
        </>
      )}
    </BaseImagePicker>
  );
};
