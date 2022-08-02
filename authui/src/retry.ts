export function exponential(index: number): number {
  return Math.pow(2, index) * 1000;
}

export interface RetryOptions {
  // The default value is 2.
  // So the index would be 0, 1, and 2.
  maxIndex?: number;
  // The default is the exponential function
  indexToMillis?: (index: number) => number;

  // So by default, retry in 1, 2, and 4 seconds.

  abortController?: AbortController;
}

export class RetryEventTarget extends EventTarget {
  options?: RetryOptions;
  index: number;
  handle: ReturnType<typeof setTimeout> | null;

  constructor(options?: RetryOptions) {
    super();
    this.options = options;
    this.index = 0;
    this.handle = null;

    this.options?.abortController?.signal?.addEventListener("abort", () => {
      this._cancelSchedule();
    });
  }

  _cancelSchedule() {
    if (this.handle != null) {
      clearTimeout(this.handle);
      this.handle = null;
    }
  }

  _dispatchRetry() {
    this.dispatchEvent(new CustomEvent("retry"));
  }

  scheduleRetry() {
    const maxIndex = this.options?.maxIndex ?? 2;
    const indexToMillis = this.options?.indexToMillis ?? exponential;

    const oldIndex = this.index;
    const newIndex = oldIndex < maxIndex ? oldIndex + 1 : maxIndex;
    this.index = newIndex;

    this._cancelSchedule();
    const delay = indexToMillis(oldIndex);
    this.handle = setTimeout(() => {
      this._dispatchRetry();
    }, delay);
  }

  markSuccess() {
    this.index = 0;
    this._cancelSchedule();
  }
}
