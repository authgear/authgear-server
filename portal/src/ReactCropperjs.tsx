import React, { createRef } from "react";
import cn from "classnames";
import Cropperjs from "cropperjs";
import { PersonIcon } from "@radix-ui/react-icons";
import { FormattedMessage } from "./intl";

import styles from "./ReactCropperjs.module.css";

export interface ReactCropperjsProps {
  className?: string;
  height?: number;
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

  private initCropper(): void {
    if (this.props.editSrc == null || this.img.current == null) {
      return;
    }
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

  private destroyCropper(): void {
    this.instance?.destroy();
    this.instance = null;
  }

  componentDidMount(): void {
    // Needed when the cropper mounts with editSrc already set (e.g. dialog).
    this.initCropper();
  }

  componentDidUpdate(prevProps: ReactCropperjsProps): void {
    if (prevProps.editSrc !== this.props.editSrc) {
      this.destroyCropper();
      this.initCropper();
    }
  }

  componentWillUnmount(): void {
    this.destroyCropper();
  }

  render(): React.ReactNode {
    const {
      className,
      height,
      editSrc,
      displaySrc,
      onError,
      onLoad,
      onClickSelectImage,
    } = this.props;
    return (
      <div
        className={cn(className, styles.container)}
        style={height != null ? { height } : undefined}
      >
        <img
          ref={this.img}
          className={cn(styles.img, editSrc == null && styles.hidden)}
          src={editSrc}
          onError={onError}
          onLoad={onLoad}
        />
        {editSrc == null ? (
          displaySrc == null ? (
            <button
              type="button"
              className={styles.placeholder}
              onClick={onClickSelectImage}
            >
              <span className={styles.visuallyHidden}>
                <FormattedMessage id="EditPictureScreen.upload-new-picture.label" />
              </span>
              <PersonIcon className={styles.placeholderIcon} aria-hidden={true} />
            </button>
          ) : (
            <img className={styles.preview} src={displaySrc} alt="" />
          )
        ) : null}
      </div>
    );
  }

  async getBlob(): Promise<Blob | null> {
    return new Promise((resolve) => {
      if (this.instance == null) {
        resolve(null);
        return;
      }
      const canvas = this.instance.getCroppedCanvas({
        ...calculateDimensions(this.instance),
        imageSmoothingQuality: "high",
      });
      if (canvas == null) {
        resolve(null);
        return;
      }
      canvas.toBlob((blob) => {
        resolve(blob);
      });
    });
  }
}

export default ReactCropperjs;
