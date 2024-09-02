import axios from "axios";
import {
  disableAllButtons,
  hideProgressBar,
  showProgressBar,
  progressEventHandler,
} from "./loading";
import {
  handleAxiosError,
  showErrorMessage,
  hideErrorMessage,
} from "./messageBar";
import Cropper from "cropperjs";
import { Controller } from "@hotwired/stimulus";
import Cropper from "cropperjs";

function destroyCropper(img: HTMLImageElement) {
  // The namespace .cropper is known by reading the source code.
  // It could change anytime!
  const cropper = (img as any).cropper;
  if (cropper instanceof Cropper) {
    cropper.destroy();
  }
}

function initCropper(img: HTMLImageElement) {
  new Cropper(img, {
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

function getCropper(img: HTMLImageElement): Cropper | undefined {
  const cropper = (img as any).cropper;
  return cropper;
}

export class ImagePickerController extends Controller {
  static targets = [
    "inputFile",
    "buttonFile",
    "buttonSave",
    "buttonRemove",
    "imgCropper",
    "imgPreview",
    "inputValue",
    "formUpload",
  ];

  declare inputFileTarget: HTMLInputElement;
  declare buttonFileTarget: HTMLButtonElement;
  declare buttonSaveTarget: HTMLButtonElement;
  declare buttonRemoveTarget: HTMLButtonElement;
  declare imgCropperTarget: HTMLImageElement;
  declare imgPreviewTarget: HTMLButtonElement;
  declare inputValueTarget: HTMLInputElement;
  declare formUploadTarget: HTMLFormElement;

  onChange(e: Event) {
    const target = this.inputFileTarget;

    const file = target.files?.[0];
    if (file == null) {
      return;
    }

    const imgCropper = this.imgCropperTarget;

    hideErrorMessage("error-message-invalid-selected-image");

    imgCropper.classList.remove("hidden");
    const buttonFile = this.buttonFileTarget;
    buttonFile.classList.add("hidden");
    const buttonRemove = this.buttonRemoveTarget;
    buttonRemove.classList.add("hidden");
    const imgPreview = this.imgPreviewTarget;
    imgPreview.classList.add("hidden");
    const buttonSave = this.buttonSaveTarget;
    buttonSave.classList.remove("hidden");

    const reader = new FileReader();
    reader.addEventListener("load", () => {
      if (typeof reader.result === "string") {
        imgCropper.src = reader.result;
        destroyCropper(imgCropper);
        initCropper(imgCropper);
      }
    });
    reader.readAsDataURL(file);
  }

  onError() {
    const target = this.imgCropperTarget;

    const src = target.src;
    // It is a file from the file system and it does not load.
    // It is probably the file is broken.
    if (/^data:/.test(src)) {
      const buttonFile = this.buttonFileTarget;
      buttonFile.classList.remove("hidden");
      const buttonSave = this.buttonSaveTarget;
      buttonSave.classList.add("hidden");
      showErrorMessage("error-message-invalid-selected-image");
    }
  }

  onClickFile(e: Event) {
    e.preventDefault();
    e.stopPropagation();

    const inputFile = this.inputFileTarget;
    inputFile.click();
  }

  onClickSave(e: Event) {
    e.preventDefault();
    e.stopPropagation();

    const imgCropper = this.imgCropperTarget;

    const cropper = getCropper(imgCropper);
    if (cropper == null) {
      return;
    }

    const maxDimensions = 1024;
    const dimensionsOptions = (function () {
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
    })();

    const canvas = cropper.getCroppedCanvas({
      ...dimensionsOptions,
      imageSmoothingQuality: "high",
    });
    canvas.toBlob(async (blob) => {
      if (blob == null) {
        return;
      }

      const inputValue = this.inputValueTarget;
      const formUpload = this.formUploadTarget;

      const revert = disableAllButtons();
      showProgressBar();
      try {
        const resp = await axios("/api/images/upload", {
          method: "POST",
          headers: {
            Accept: "text/vnd.turbo-stream.html, application/json",
          },
          onDownloadProgress: progressEventHandler,
          onUploadProgress: progressEventHandler,
        });
        const body = resp.data;

        const {
          result: { upload_url },
        } = body;

        const formData = new FormData();
        formData.append("file", blob);
        const uploadResp = await axios(upload_url, {
          method: "POST",
          data: formData,
          headers: {
            Accept: "text/vnd.turbo-stream.html, application/json",
          },
          onDownloadProgress: progressEventHandler,
          onUploadProgress: progressEventHandler,
        });
        const uploadRespBody = uploadResp.data;
        const {
          result: { url },
        } = uploadRespBody;

        inputValue.value = url;
        formUpload.submit();
      } catch (e) {
        // revert is only called for error branch because
        // The success branch also loads a new page.
        // Keeping the buttons in disabled state reduce flickering in the UI.
        revert();
        handleAxiosError(e);
      } finally {
        hideProgressBar();
      }
    });
  }
}
