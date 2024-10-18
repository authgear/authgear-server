import { Controller } from "@hotwired/stimulus";
import Toastify, { ToastifyInstance, ToastifyOptions } from "toastify-js";

const CANVAS_WIDTH = 1280;
const TOAST_DISPLAY_INTERVAL = 3500;
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

  declare toasts: ToastifyInstance[] | undefined;
  declare toastTimers: NodeJS.Timeout[] | undefined;

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
      this.unsetPhotoLoading();
      this.submitPhotoBtnTarget.classList.remove("hidden");
    }, 1000);
  };

  setPhotoLoading = () => {
    this.takePhotoBtnTarget.disabled = true;
  };

  unsetPhotoLoading = () => {
    this.takePhotoBtnTarget.disabled = false;
  };

  takePhoto = () => {
    this.setPhotoLoading();
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

  setSubmitLoading = () => {
    this.displayToasts();
  };

  submitPhoto = () => {
    this.setSubmitLoading();
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

  buildToasts = (): ToastifyInstance[] => {
    const commonOpts: Partial<ToastifyOptions> = {
      duration: TOAST_DISPLAY_INTERVAL,
      close: false,
      gravity: "bottom",
      position: "center",
      stopOnFocus: true,
      selector: this.cameraContainerTarget,
    };

    const t1 = Toastify({
      text: "Uploading... ", // TODO handle translation
      ...commonOpts,
    });

    const t2 = Toastify({
      text: "Processing image... ", // TODO handle translation
      ...commonOpts,
    });

    const t3 = Toastify({
      text: "Analyzing results... ", // TODO handle translation
      ...commonOpts,
    });

    const t4 = Toastify({
      text: "Finalizing... ", // TODO handle translation
      ...commonOpts,
      duration: 10000,
    });

    return [t1, t2, t3, t4];
  };

  cleanupToasts = () => {
    this.cleanupToastTimer();
    this.toasts?.forEach((t) => {
      try {
        t?.hideToast();
      } catch (_: unknown) {
        // slience expected error - toast elements might not be in DOM already
      }
    });
    this.toasts = undefined;
  };

  displayToasts = () => {
    if (this.toasts == null) {
      console.error("toasts not initialized");
      return;
    }
    this.toasts.forEach((t, i) => {
      const timer = setTimeout(() => {
        t?.showToast();
      }, TOAST_DISPLAY_INTERVAL * i);
      this.pushToastTimer(timer);
    });
  };

  pushToastTimer = (timer: NodeJS.Timeout) => {
    if (this.toastTimers == null) {
      this.toastTimers = [];
    }

    this.toastTimers.push(timer);
  };

  cleanupToastTimer = () => {
    if (this.toastTimers == null) {
      return;
    }
    this.toastTimers.forEach((t) => clearTimeout(t));
    this.toastTimers = [];
  };

  connect(): void {
    this.cameraVideoTarget.addEventListener("canplay", this.handleVideoCanplay);
    this.toasts = this.buildToasts();
  }

  disconnect(): void {
    this.cameraVideoTarget.removeEventListener(
      "canplay",
      this.handleVideoCanplay
    );
    this.cleanupToasts();
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
