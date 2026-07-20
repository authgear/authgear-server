import React from "react";
import { FormattedMessage } from "../../intl";
import { ImageIcon } from "@radix-ui/react-icons";
import { Text } from "@radix-ui/themes";

import cn from "classnames";

import BaseImagePicker, { ImageFileExtension } from "../common/BaseImagePicker";
import { SecondaryButton } from "../v2/Button/SecondaryButton/SecondaryButton";
import { IconButton, IconButtonIcon } from "../v2/IconButton/IconButton";
import { SquareIcon } from "../v2/SquareIcon/SquareIcon";
import { base64EncodedDataToDataURI } from "../../util/uri";

import { AppLogoResource } from "../../graphql/portal/DesignScreen/form";

import styles from "./ImagePicker.module.css";

interface AppLogoPickerProps {
  logo: AppLogoResource;
  descriptionKey?: string;
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
  const {
    logo,
    onChange,
    descriptionKey = "DesignScreen.configuration.logo.description",
  } = props;

  const imagePreviewData =
    logo.base64EncodedData ?? logo.fallbackBase64EncodedData;

  const isShowingFallbackImage =
    logo.base64EncodedData == null && logo.fallbackBase64EncodedData != null;

  return (
    <BaseImagePicker
      sizeLimitInBytes={100 * 1000}
      className={cn(styles.picker)}
      base64EncodedData={logo.base64EncodedData}
      onChange={onChange}
    >
      {({ showFilePicker, clearImage }) => (
        <>
          <button
            type="button"
            className={cn(
              styles.preview,
              imagePreviewData != null && styles.previewFilled
            )}
            onClick={showFilePicker}
          >
            {imagePreviewData == null ? (
              <SquareIcon Icon={ImageIcon} size="7" radius="3" />
            ) : (
              <img
                className={cn(
                  styles.previewImage,
                  isShowingFallbackImage && styles.previewImageFallback
                )}
                src={base64EncodedDataToDataURI(imagePreviewData)}
                alt=""
              />
            )}
          </button>
          <div className={styles.pickerActions}>
            <Text as="p" size="2" color="gray">
              <FormattedMessage id={descriptionKey} />
            </Text>
            <div className={styles.pickerButtons}>
              {logo.base64EncodedData != null ? (
                <>
                  <SecondaryButton
                    type="button"
                    size="2"
                    text={
                      <FormattedMessage id="DesignScreen.configuration.imagePicker.upload" />
                    }
                    onClick={showFilePicker}
                  />
                  <IconButton
                    type="button"
                    size="2"
                    variant="destroy"
                    icon={IconButtonIcon.Trash}
                    onClick={clearImage}
                  />
                </>
              ) : logo.fallbackBase64EncodedData != null ? (
                <SecondaryButton
                  type="button"
                  size="2"
                  text={
                    <FormattedMessage id="DesignScreen.configuration.appLogoPicker.override" />
                  }
                  onClick={showFilePicker}
                />
              ) : (
                <SecondaryButton
                  type="button"
                  size="2"
                  text={
                    <FormattedMessage id="DesignScreen.configuration.imagePicker.upload" />
                  }
                  onClick={showFilePicker}
                />
              )}
            </div>
          </div>
        </>
      )}
    </BaseImagePicker>
  );
};

export default AppLogoPicker;
