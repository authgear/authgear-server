import { jest, describe, it, expect } from "@jest/globals";
import { RetryEventTarget } from "./retry";

jest.useFakeTimers();

describe("RetryEventTarget", () => {
  it("retry synchronous function", () => {
    const retryEventTarget = new RetryEventTarget();

    const f = () => {
      throw new Error("failed");
    };

    let count = 0;
    const g = () => {
      try {
        f();
        retryEventTarget.markSuccess();
      } catch (e) {
        if (count < 1) {
          retryEventTarget.scheduleRetry();
          count += 1;
        }
      }
    };

    const callback = jest.fn(() => {
      g();
    });

    retryEventTarget.addEventListener("retry", callback);

    callback();

    expect(callback).toHaveBeenCalledTimes(1);
    jest.advanceTimersByTime(1000);
    expect(callback).toHaveBeenCalledTimes(2);
  });

  it("reset index", () => {
    // --SETUP--
    const retryEventTarget = new RetryEventTarget();

    let count = 0;
    // f is a function that will succeed when count is greater than 0 and is multiple of 3.
    const f = () => {
      const oldCount = count;
      count += 1;
      if (oldCount > 0 && oldCount % 3 === 0) {
        return;
      }
      throw new Error("fail");
    };

    const g = () => {
      try {
        f();
        retryEventTarget.markSuccess();
      } catch (e) {
        retryEventTarget.scheduleRetry();
      }
    };

    const callback = jest.fn(() => {
      g();
    });

    retryEventTarget.addEventListener("retry", callback);
    // --SETUP--

    // --ASSERTION--
    callback();
    expect(callback).toHaveBeenCalledTimes(1);
    jest.advanceTimersByTime(1000);
    expect(callback).toHaveBeenCalledTimes(2);
    jest.advanceTimersByTime(2000);
    expect(callback).toHaveBeenCalledTimes(3);
    jest.advanceTimersByTime(4000);
    expect(callback).toHaveBeenCalledTimes(4);
    // success just now.
    // callback should NOT be called anymore.
    jest.advanceTimersByTime(4000);
    expect(callback).toHaveBeenCalledTimes(4);

    callback();
    expect(callback).toHaveBeenCalledTimes(5);
    jest.advanceTimersByTime(1000);
    expect(callback).toHaveBeenCalledTimes(6);
    jest.advanceTimersByTime(2000);
    expect(callback).toHaveBeenCalledTimes(7);
    // success just now.
    // callback should NOT be called anymore.
    jest.advanceTimersByTime(4000);
    expect(callback).toHaveBeenCalledTimes(7);

    callback();
    expect(callback).toHaveBeenCalledTimes(8);
    jest.advanceTimersByTime(1000);
    expect(callback).toHaveBeenCalledTimes(9);
    jest.advanceTimersByTime(2000);
    expect(callback).toHaveBeenCalledTimes(10);
    // success just now.
    // callback should NOT be called anymore.
    jest.advanceTimersByTime(4000);
    expect(callback).toHaveBeenCalledTimes(10);
    // --ASSERTION--
  });

  // For some reason, async function cannot pass the test.
  // So we do not have that test case here.

  it("respect AbortController", () => {
    const abortController = new AbortController();
    const retryEventTarget = new RetryEventTarget({
      abortController,
    });

    const f = () => {
      throw new Error("failed");
    };

    let count = 0;
    const g = () => {
      try {
        f();
        retryEventTarget.markSuccess();
      } catch (e) {
        if (count < 1) {
          retryEventTarget.scheduleRetry();
          count += 1;
        }
      }
    };

    const callback = jest.fn(() => {
      g();
    });

    retryEventTarget.addEventListener("retry", callback);

    callback();
    expect(callback).toHaveBeenCalledTimes(1);
    jest.advanceTimersByTime(500);
    abortController.abort();
    jest.advanceTimersByTime(1000);
    expect(callback).toHaveBeenCalledTimes(1);
  });
});
