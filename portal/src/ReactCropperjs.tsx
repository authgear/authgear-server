import React, { createRef } from "react";
import cn from "classnames";
import Cropperjs from "cropperjs";
import { Image, ImageFit } from "@fluentui/react";

import styles from "./ReactCropperjs.module.scss";

export interface ReactCropperjsProps {
  className?: string;
  editSrc?: string;
  displaySrc?: string;
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
    const { className, editSrc, displaySrc } = this.props;
    return (
      <div className={cn(className, styles.container)}>
        <img ref={this.img} className={styles.img} src={editSrc} />
        {editSrc == null ? (
          <Image
            className={styles.img}
            src={displaySrc}
            imageFit={ImageFit.contain}
          />
        ) : null}
      </div>
    );
  }

  async getBlob(): Promise<Blob> {
    return new Promise((resolve) => {
      const canvas = this.instance?.getCroppedCanvas({
        width: 240,
        height: 240,
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
