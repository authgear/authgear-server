import React, { useMemo } from "react";
import cn from "classnames";
import { Image, ImageFit } from "@fluentui/react";
import { FormattedMessage } from "@oursky/react-messageformat";
import { useSystemConfig } from "./context/SystemConfigContext";
import { base64EncodedDataToDataURI } from "./util/uri";
import PrimaryButton from "./PrimaryButton";
import DefaultButton from "./DefaultButton";

import styles from "./ImageFilePicker.module.css";
import BaseImagePicker from "./components/common/BaseImagePicker";

export type ImageFileExtension = ".jpeg" | ".png" | ".gif";

export interface ImageFilePickerProps {
  disabled?: boolean;
  className?: string;
  base64EncodedData?: string;
  onChange: (
    image: {
      base64EncodedData: string;
      extension: ImageFileExtension;
    } | null
  ) => void;
}

const ImageFilePicker: React.VFC<ImageFilePickerProps> =
  function ImageFilePicker(props: ImageFilePickerProps) {
    const { disabled, className, base64EncodedData, onChange } = props;

    const hasImage = base64EncodedData != null;

    const { themes } = useSystemConfig();

    const src = useMemo(() => {
      if (base64EncodedData != null) {
        return base64EncodedDataToDataURI(base64EncodedData);
      }
      return undefined;
    }, [base64EncodedData]);

    const borderColor = themes.main.semanticColors.inputBorder;

    return (
      <BaseImagePicker
        className={cn(className, styles.root)}
        base64EncodedData={base64EncodedData ?? null}
        onChange={onChange}
      >
        {({ showFilePicker, clearImage }) => (
          <>
            <Image
              src={src}
              className={styles.image}
              styles={{
                root: {
                  borderColor,
                  borderWidth: "1px",
                  borderStyle: "solid",
                },
              }}
              imageFit={ImageFit.centerContain}
              maximizeFrame={true}
            />
            {hasImage ? (
              <PrimaryButton
                className={styles.button}
                onClick={clearImage}
                theme={themes.destructive}
                disabled={disabled}
                text={<FormattedMessage id={"ImageFilePicker.remove"} />}
              />
            ) : (
              <DefaultButton
                className={styles.button}
                onClick={showFilePicker}
                disabled={disabled}
                text={<FormattedMessage id="ImageFilePicker.upload" />}
              />
            )}
          </>
        )}
      </BaseImagePicker>
    );
  };

export default ImageFilePicker;
