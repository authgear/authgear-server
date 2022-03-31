import axios from "axios";
import {
  disableAllButtons,
  hideProgressBar,
  showProgressBar,
  progressEventHandler,
} from "./loading";
import { handleAxiosError, showErrorMessage, hideErrorMessage } from "./error";

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

function onChange(e: Event) {
  const target = e.currentTarget;
  if (!(target instanceof HTMLInputElement)) {
    return;
  }

  const file = target.files?.[0];
  if (file == null) {
    return;
  }

  const imgCropper = document.getElementById("imagepicker-img-cropper");
  if (!(imgCropper instanceof HTMLImageElement)) {
    return;
  }

  hideErrorMessage("error-message-invalid-selected-image");

  imgCropper?.classList.remove("hidden");
  const buttonFile = document.getElementById("imagepicker-button-file");
  buttonFile?.classList.add("hidden");
  const buttonRemove = document.getElementById("imagepicker-button-remove");
  buttonRemove?.classList.add("hidden");
  const imgPreview = document.getElementById("imagepicker-img-preview");
  imgPreview?.classList.add("hidden");
  const buttonSave = document.getElementById("imagepicker-button-save");
  buttonSave?.classList.remove("hidden");

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

function onError(e: Event) {
  const target = e.currentTarget;
  if (!(target instanceof HTMLImageElement)) {
    return;
  }

  const src = target.src;
  // It is a file from the file system and it does not load.
  // It is probably the file is broken.
  if (/^data:/.test(src)) {
    const buttonFile = document.getElementById("imagepicker-button-file");
    buttonFile?.classList.remove("hidden");
    const buttonSave = document.getElementById("imagepicker-button-save");
    buttonSave?.classList.add("hidden");
    showErrorMessage("error-message-invalid-selected-image");
  }
}

function onClickFile(e: Event) {
  e.preventDefault();
  e.stopPropagation();

  const inputFile = document.getElementById("imagepicker-input-file");
  inputFile?.click();
}

function onClickSave(e: Event) {
  e.preventDefault();
  e.stopPropagation();

  const imgCropper = document.getElementById("imagepicker-img-cropper");
  if (!(imgCropper instanceof HTMLImageElement)) {
    return;
  }

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

    const inputValue = document.getElementById("imagepicker-input-value");
    if (!(inputValue instanceof HTMLInputElement)) {
      return;
    }

    const formUpload = document.getElementById("imagepicker-form-upload");
    if (!(formUpload instanceof HTMLFormElement)) {
      return;
    }

    const revert = disableAllButtons();
    showProgressBar();
    try {
      const resp = await axios("/api/images/upload", {
        method: "POST",
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

export function setupImagePicker(): () => void {
  // The image picker recognizes the following elements:
  // #imagepicker-form-remove
  //   The form that unsets picture.
  // #imagepicker-button-remove
  //   The submit button of #imagepicker-form-remove
  //
  // #imagepicker-form-upload
  //   The form that sets picture.
  // #imagepicker-input-value
  //   The input to hold the authgearimages: URI.
  //
  // #imagepicker-img-cropper
  //   The <img> to inject cropperjs
  //
  // #imagepicker-input-file
  //   The hidden <input type="file"> to let the end-user to select a file.
  // #imageicker-button-file
  //   The button visually represents #imagepicker-input-file
  //
  // #imagepicker-button-save
  //   The button that crops the image, requests signed url, uploads the image, and submit #imagepicker-form-upload.
  const inputFile = document.getElementById("imagepicker-input-file");
  const buttonFile = document.getElementById("imagepicker-button-file");
  const buttonSave = document.getElementById("imagepicker-button-save");
  const img = document.getElementById("imagepicker-img-cropper");
  inputFile?.addEventListener("change", onChange);
  img?.addEventListener("error", onError);
  buttonFile?.addEventListener("click", onClickFile);
  buttonSave?.addEventListener("click", onClickSave);
  return () => {
    inputFile?.removeEventListener("change", onChange);
    img?.addEventListener("error", onError);
    buttonFile?.removeEventListener("click", onClickFile);
    buttonSave?.removeEventListener("click", onClickSave);
  };
}
