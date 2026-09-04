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

import styles from "./ImagePicker.module.css";

interface ImagePickerProps {
  sizeLimitInBytes: number;
  base64EncodedData: string | null;
  descriptionKey?: string;
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
  const {
    sizeLimitInBytes,
    base64EncodedData,
    onChange,
    descriptionKey = "DesignScreen.configuration.favicon.description",
  } = props;
  return (
    <BaseImagePicker
      sizeLimitInBytes={sizeLimitInBytes}
      className={cn(styles.picker)}
      base64EncodedData={base64EncodedData}
      onChange={onChange}
    >
      {({ showFilePicker, clearImage }) => (
        <>
          <button
            type="button"
            className={cn(
              styles.preview,
              base64EncodedData != null && styles.previewFilled
            )}
            onClick={showFilePicker}
          >
            {base64EncodedData == null ? (
              <SquareIcon Icon={ImageIcon} size="7" radius="3" />
            ) : (
              <img
                className={styles.previewImage}
                src={base64EncodedDataToDataURI(base64EncodedData)}
                alt=""
              />
            )}
          </button>
          <div className={styles.pickerActions}>
            <Text as="p" size="2" color="gray">
              <FormattedMessage id={descriptionKey} />
            </Text>
            <div className={styles.pickerButtons}>
              <SecondaryButton
                type="button"
                size="2"
                text={
                  <FormattedMessage id="DesignScreen.configuration.imagePicker.upload" />
                }
                onClick={showFilePicker}
              />
              {base64EncodedData != null ? (
                <IconButton
                  type="button"
                  size="2"
                  variant="destroy"
                  icon={IconButtonIcon.Trash}
                  onClick={clearImage}
                />
              ) : null}
            </div>
          </div>
        </>
      )}
    </BaseImagePicker>
  );
};
