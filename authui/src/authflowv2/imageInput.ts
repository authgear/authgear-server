import { Controller } from "@hotwired/stimulus";

const CANVAS_WIDTH = 1280;
export class ImageInputController extends Controller {
  static targets = [
    // container
    "cameraContainer",

    // camera interface states
    "cameraInitial",
    "cameraVideo",
    "cameraOutput",

    // buttons
    "openCameraBtn",
    "takePhotoBtn",
    "submitPhotoBtn",
    "formSubmitBtn",

    // image capture helper
    "canvas",
    "input",
  ];

  declare readonly cameraContainerTarget: HTMLDivElement;
  declare readonly cameraInitialTarget: HTMLDivElement;
  declare readonly cameraVideoTarget: HTMLVideoElement;
  declare readonly cameraOutputTarget: HTMLImageElement;

  declare readonly openCameraBtnTarget: HTMLButtonElement;
  declare readonly takePhotoBtnTarget: HTMLButtonElement;
  declare readonly submitPhotoBtnTarget: HTMLButtonElement;

  declare readonly canvasTarget: HTMLCanvasElement;
  declare readonly inputTarget: HTMLInputElement;
  declare readonly formSubmitBtnTarget: HTMLButtonElement;

  onCameraOpen = () => {
    // orders matter here, otherwise UI might flash
    this.cameraVideoTarget.classList.remove("hidden");
    this.cameraInitialTarget.classList.add("hidden");
    this.openCameraBtnTarget.classList.add("hidden");
    this.takePhotoBtnTarget.classList.remove("hidden");
  };

  openCamera = () => {
    this.openCameraBtnTarget.disabled = true;
    const cameraSupported = "mediaDevices" in navigator;
    if (!cameraSupported) {
      //TODO (identity-week-demo): Show error to user
      throw new Error("Camera not supported");
    }

    navigator.mediaDevices
      .getUserMedia({
        video: {
          facingMode: "user", // prefer front camera
        },
        audio: false,
      })
      .then((stream) => {
        this.cameraVideoTarget.srcObject = stream;
        this.cameraVideoTarget
          .play()
          .catch((err: unknown) => console.error(err)); //TODO (identity-week-demo): Handle play error
        this.onCameraOpen();
      })
      .catch((err: unknown) => {
        console.error(err);
        if (isNotAllowedErr(err)) {
          //TODO (identity-week-demo): Show error to user
          alert("Please allow camera access to proceed");
        }
      })
      .finally(() => {
        this.openCameraBtnTarget.disabled = false;
      });
  };

  onPhotoTaken = () => {
    this.cameraOutputTarget.classList.remove("hidden");

    // wait for image process finish, hard-code as 1 second for now
    setTimeout(() => {
      this.cameraVideoTarget.classList.add("hidden");
      this.cameraVideoTarget.pause();
      this.takePhotoBtnTarget.classList.add("hidden");
      this.takePhotoBtnTarget.disabled = false;
      this.submitPhotoBtnTarget.classList.remove("hidden");
    }, 1000);
  };
  takePhoto = () => {
    this.takePhotoBtnTarget.disabled = true;
    const context = this.canvasTarget.getContext("2d");
    if (context == null) {
      console.error("Canvas context not available");
      return;
    }
    context.drawImage(
      this.cameraVideoTarget,
      0,
      0,
      this.canvasTarget.width,
      this.canvasTarget.height
    );
    const dataURL = this.canvasTarget.toDataURL("image/png");
    this.cameraOutputTarget.src = dataURL;
    this.inputTarget.value = getB64StringFromDataURL(dataURL);
    this.onPhotoTaken();
  };

  submitPhoto = () => {
    this.submitForm();
  };

  submitForm = () => {
    this.formSubmitBtnTarget.click();
  };

  handleVideoCanplay = () => {
    const w = this.cameraContainerTarget.offsetWidth;
    const vW = this.cameraVideoTarget.videoWidth;
    const vH = this.cameraVideoTarget.videoHeight;
    const h = (vH / vW) * w;

    this.cameraVideoTarget.setAttribute("width", w.toString());
    this.cameraVideoTarget.setAttribute("height", h.toString());

    const cW = CANVAS_WIDTH;
    const cH = (vH / vW) * cW;
    this.canvasTarget.setAttribute("width", cW.toString());
    this.canvasTarget.setAttribute("height", cH.toString());
  };

  connect(): void {
    this.cameraVideoTarget.addEventListener("canplay", this.handleVideoCanplay);
  }

  disconnect(): void {
    this.cameraVideoTarget.removeEventListener(
      "canplay",
      this.handleVideoCanplay
    );
  }
}

/**
 * per MDN,
 * > To retrieve only the Base64 encoded string, first remove `data:\*\/\*;base64,` from the result
 *
 * [MDN Reference](https://developer.mozilla.org/en-US/docs/Web/API/FileReader/readAsDataURL)
 */
function getB64StringFromDataURL(dataURL: string): string {
  const prefixPattern = /data:[a-z]{1,10}\/[a-z]{1,10};base64,/;
  const b64 = dataURL.replace(prefixPattern, "");
  return b64;
}

function isNotAllowedErr(err: unknown): boolean {
  return err instanceof DOMException && err.name === "NotAllowedError";
}
