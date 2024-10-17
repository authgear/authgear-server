import { Controller } from "@hotwired/stimulus";

export class ImageInputController extends Controller {
  static targets = [
    // container
    "cameraContainer",

    // states
    "cameraInitial",
    "cameraVideo",

    // buttons
    "openCameraBtn",
    "takePhotoBtn",
    "formSubmitBtn",
    // image capture helper
    "canvas",
    "input",
  ];

  declare readonly cameraContainerTarget: HTMLDivElement;
  declare readonly cameraInitialTarget: HTMLDivElement;
  declare readonly cameraVideoTarget: HTMLVideoElement;

  declare readonly openCameraBtnTarget: HTMLButtonElement;
  declare readonly takePhotoBtnTarget: HTMLButtonElement;

  declare readonly canvasTarget: HTMLCanvasElement;
  declare readonly inputTarget: HTMLInputElement;
  declare readonly formSubmitBtnTarget: HTMLButtonElement;

  onCameraOpen = () => {
    this.cameraContainerTarget.classList.add("open");
    this.cameraInitialTarget.classList.add("hidden");
    this.openCameraBtnTarget.classList.add("hidden");
    this.takePhotoBtnTarget.classList.remove("hidden");
  };

  openCamera = () => {
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
      });
  };

  takePhoto = () => {
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
    this.inputTarget.value = getB64StringFromDataURL(dataURL);
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

    this.canvasTarget.setAttribute("width", w.toString());
    this.canvasTarget.setAttribute("height", h.toString());
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
