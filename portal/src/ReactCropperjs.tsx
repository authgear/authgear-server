import React, { createRef } from "react";
import cn from "classnames";
import Cropperjs from "cropperjs";
import { Image, ImageFit, FontIcon } from "@fluentui/react";

import styles from "./ReactCropperjs.module.css";

export interface ReactCropperjsProps {
  className?: string;
  onClickSelectImage?: () => void;
  editSrc?: string;
  displaySrc?: string;
  onError?: () => void;
  onLoad?: () => void;
}

const maxDimensions = 1024;

function calculateDimensions(
  cropper: Cropperjs
): Cropperjs.GetCroppedCanvasOptions {
  const imageData = cropper.getImageData();
  const cropBoxData = cropper.getCropBoxData();
  // assume the cropped area is square
  if (
    imageData.naturalWidth > 0 &&
    imageData.width > 0 &&
    cropBoxData.width > 0
  ) {
    const imageScale = imageData.naturalWidth / imageData.width;
    const croppedImageWidth = Math.floor(cropBoxData.width * imageScale);
    const resultDimensions = Math.min(croppedImageWidth, maxDimensions);
    return {
      width: resultDimensions,
      height: resultDimensions,
    };
  }

  // last resort when any of the image or crop box data is unavailable
  return {
    maxWidth: maxDimensions,
    maxHeight: maxDimensions,
  };
}

class ReactCropperjs extends React.Component<ReactCropperjsProps> {
  instance: Cropperjs | null = null;
  img: React.RefObject<HTMLImageElement> = createRef();

  componentDidUpdate(prevProps: ReactCropperjsProps): void {
    if (prevProps.editSrc !== this.props.editSrc) {
      this.instance?.destroy();
      if (this.props.editSrc != null) {
        if (this.img.current != null) {
          this.instance = new Cropperjs(this.img.current, {
            // Make crop region not able to move outside the image.
            viewMode: 1,
            // We want to crop a square image.
            aspectRatio: 1,
            movable: false,
            rotatable: false,
            scalable: false,
            zoomable: false,
            zoomOnTouch: false,
            zoomOnWheel: false,
          });
        }
      }
    }
  }

  render(): React.ReactNode {
    const {
      className,
      editSrc,
      displaySrc,
      onError,
      onLoad,
      onClickSelectImage,
    } = this.props;
    return (
      <div className={cn(className, styles.container)}>
        <img
          ref={this.img}
          className={cn(styles.img, editSrc == null && styles.hidden)}
          src={editSrc}
          onError={onError}
          onLoad={onLoad}
        />
        {editSrc == null &&
          (displaySrc == null ? (
            <FontIcon
              role="button"
              className={styles.placeholder}
              iconName="Contact"
              onClick={onClickSelectImage}
            />
          ) : (
            <Image
              className={styles.preview}
              src={displaySrc}
              imageFit={ImageFit.contain}
            />
          ))}
      </div>
    );
  }

  async getBlob(): Promise<Blob> {
    return new Promise((resolve) => {
      const canvas = this.instance?.getCroppedCanvas({
        ...calculateDimensions(this.instance),
        imageSmoothingQuality: "high",
      });
      canvas?.toBlob((blob) => {
        if (blob != null) {
          resolve(blob);
        }
      });
    });
  }
}

export default ReactCropperjs;
