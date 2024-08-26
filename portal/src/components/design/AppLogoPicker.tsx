import React from "react";
import { FormattedMessage } from "@oursky/react-messageformat";
import { Image, ImageFit } from "@fluentui/react";

import cn from "classnames";

import BaseImagePicker, { ImageFileExtension } from "../common/BaseImagePicker";
import PrimaryButton from "../../PrimaryButton";
import DefaultButton from "../../DefaultButton";

import { useSystemConfig } from "../../context/SystemConfigContext";
import { base64EncodedDataToDataURI } from "../../util/uri";

import { AppLogoResource } from "../../graphql/portal/DesignScreen/form";

import styles from "./ImagePicker.module.css";

interface AppLogoPickerProps {
  logo: AppLogoResource;
  onChange: (
    image: {
      base64EncodedData: string;
      extension: ImageFileExtension;
    } | null
  ) => void;
}
const AppLogoPicker: React.VFC<AppLogoPickerProps> = function AppLogoPicker(
  props
) {
  const { logo, onChange } = props;
  const { themes } = useSystemConfig();

  const imagePreviewData =
    logo.base64EncodedData ?? logo.fallbackBase64EncodedData;

  const isShowingFallbackImage =
    logo.base64EncodedData == null && logo.fallbackBase64EncodedData != null;

  return (
    <BaseImagePicker
      className={cn("flex", "items-center", "gap-x-6")}
      base64EncodedData={logo.base64EncodedData}
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
            {imagePreviewData == null ? (
              <span className={styles.icImagePlaceholder}></span>
            ) : (
              <Image
                src={base64EncodedDataToDataURI(imagePreviewData)}
                className={cn("h-full", "w-full")}
                imageFit={ImageFit.centerCover}
                maximizeFrame={true}
                styles={{
                  image: isShowingFallbackImage
                    ? {
                        opacity: 0.3,
                        filter: "grayscale(1)",
                      }
                    : null,
                }}
              />
            )}
          </div>
          {logo.base64EncodedData != null ? (
            <PrimaryButton
              theme={themes.destructive}
              text={<FormattedMessage id={"ImageFilePicker.remove"} />}
              onClick={clearImage}
            />
          ) : logo.fallbackBase64EncodedData != null ? (
            <DefaultButton
              text={
                <FormattedMessage id="DesignScreen.configuration.appLogoPicker.override" />
              }
              onClick={showFilePicker}
            />
          ) : (
            <DefaultButton
              text={
                <FormattedMessage id="DesignScreen.configuration.imagePicker.upload" />
              }
              onClick={showFilePicker}
            />
          )}
        </>
      )}
    </BaseImagePicker>
  );
};

export default AppLogoPicker;
