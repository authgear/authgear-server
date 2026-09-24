import {
  jest,
  describe,
  it,
  expect,
  beforeAll,
  afterAll,
  afterEach,
} from "@jest/globals";
import { Application } from "@hotwired/stimulus";
import { visit } from "@hotwired/turbo";
import { ClockSkewController } from "./clockSkew";

jest.mock("@hotwired/turbo", () => ({ visit: jest.fn() }));

const visitMock = jest.mocked(visit);

const SERVER_TIME = Date.UTC(2026, 8, 24);
const MINUTE = 60 * 1000;
const REDIRECT_URL = "/authflow/v2/clock_skew";

let application: Application;

beforeAll(() => {
  application = Application.start();
  application.register("clock-skew", ClockSkewController);
});

afterAll(() => {
  application.stop();
});

afterEach(() => {
  document.body.innerHTML = "";
  document.documentElement.removeAttribute("data-turbo-preview");
  visitMock.mockReset();
  jest.restoreAllMocks();
});

// Stimulus connects controllers from a MutationObserver, so wait a tick.
async function render(extraAttrs = ""): Promise<void> {
  document.body.innerHTML = `
    <div
      hidden
      data-controller="clock-skew"
      data-clock-skew-server-time-value="${SERVER_TIME}"
      data-clock-skew-max-ahead-value="${5 * MINUTE}"
      data-clock-skew-max-behind-value="${5 * MINUTE}"
      data-clock-skew-redirect-url-value="${REDIRECT_URL}"
      ${extraAttrs}
    ></div>
  `;
  await new Promise((resolve) => {
    setTimeout(resolve, 0);
  });
}

function setDeviceClock(offsetMs: number) {
  jest.spyOn(Date, "now").mockReturnValue(SERVER_TIME + offsetMs);
}

describe("ClockSkewController", () => {
  it.each([
    ["6 minutes ahead", 6 * MINUTE],
    ["6 minutes behind", -6 * MINUTE],
  ])("redirects when the device clock is %s", async (_, offset) => {
    setDeviceClock(offset);
    await render();
    expect(visitMock).toHaveBeenCalledTimes(1);
    expect(visitMock).toHaveBeenCalledWith(REDIRECT_URL, { action: "replace" });
  });

  it.each([
    ["in sync", 0],
    ["4 minutes ahead", 4 * MINUTE],
    ["4 minutes behind", -4 * MINUTE],
    ["exactly 5 minutes ahead", 5 * MINUTE],
    ["exactly 5 minutes behind", -5 * MINUTE],
  ])("does not redirect when the device clock is %s", async (_, offset) => {
    setDeviceClock(offset);
    await render();
    expect(visitMock).not.toHaveBeenCalled();
  });

  it("does not redirect on a Turbo preview", async () => {
    setDeviceClock(6 * MINUTE);
    document.documentElement.setAttribute("data-turbo-preview", "");
    await render();
    expect(visitMock).not.toHaveBeenCalled();
  });

  it("does not redirect on a restored Turbo snapshot", async () => {
    setDeviceClock(6 * MINUTE);
    await render("data-clock-skew-checked");
    expect(visitMock).not.toHaveBeenCalled();
  });
});
