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

  imgCropper?.classList.remove("hidden");
  const buttonFile = document.getElementById("imagepicker-button-file");
  buttonFile?.classList.add("hidden");
  const buttonRemove = document.getElementById("imagepicker-button-remove");
  buttonRemove?.classList.add("hidden");
  const saveButton = document.getElementById("save-button");
  saveButton?.classList.remove("hidden");
  const imgPreview = document.getElementById("imagepicker-img-preview");
  imgPreview?.classList.add("hidden");

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

function onClickFile(e: Event) {
  e.preventDefault();
  e.stopPropagation();

  const inputFile = document.getElementById("imagepicker-input-file");
  inputFile?.click();
}

function onSubmit(e: Event) {
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

  const canvas = cropper.getCroppedCanvas({
    width: 240,
    height: 240,
    imageSmoothingQuality: "high",
  });
  canvas.toBlob((_blob) => {
    // TODO(images): Include the blob in the FormData.
  });
}

export function setupImagePicker(): () => void {
  // The image picker recognizes the following elements:
  // #imagepicker-input-file
  //   The hidden <input type="file"> to let the end-user to select a file.
  // #imagepicker-input-value
  //   The hidden <input type="hidden"> to store the value.
  // #imagepicker-img-cropper
  //   The <img> to inject cropperjs
  // #imageicker-button-file
  //   The button visually represents #imagepicker-input-file
  // #imagepicker-button-remove
  //   The button removes the picture and save.
  // #save-button
  //   Show the save button in edit mode.
  // #form
  //   The form that saves the standard attributes.
  const inputFile = document.getElementById("imagepicker-input-file");
  const buttonFile = document.getElementById("imagepicker-button-file");
  const form = document.getElementById("form");
  inputFile?.addEventListener("change", onChange);
  buttonFile?.addEventListener("click", onClickFile);
  form?.addEventListener("submit", onSubmit);
  return () => {
    inputFile?.removeEventListener("change", onChange);
    buttonFile?.removeEventListener("click", onClickFile);
    form?.removeEventListener("submit", onSubmit);
  };
}
