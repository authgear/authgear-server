import { describe, it, expect } from "@jest/globals";
import { getPreviousPlan, getCTAVariant } from "./plan";

describe("getPreviousPlan", () => {
  it("should work", () => {
    expect(getPreviousPlan("free")).toEqual(null);
    expect(getPreviousPlan("free-approved")).toEqual(null);
    expect(getPreviousPlan("developers")).toEqual(null);

    expect(getPreviousPlan("startups")).toEqual("free");
    expect(getPreviousPlan("business")).toEqual("startups");
    expect(getPreviousPlan("enterprise")).toEqual("business");

    expect(getPreviousPlan("foobar")).toEqual(null);
  });
});

describe("getCTAVariant", () => {
  function f(
    currentPlanName: string,
    cardPlanName: string,
    cancelled: boolean
  ) {
    return getCTAVariant({
      cardPlanName,
      currentPlanName,
      subscriptionCancelled: cancelled,
    });
  }

  it("returns non-applicable for custom plans", () => {
    expect(f("foobar", "free", false)).toEqual("non-applicable");
    expect(f("free", "foobar", false)).toEqual("non-applicable");
  });

  it("returns current or contact-us if card is enterprise", () => {
    expect(f("free", "enterprise", false)).toEqual("contact-us");
    expect(f("enterprise", "enterprise", false)).toEqual("current");
  });

  it("returns current or non-applicable if current is enterprise", () => {
    expect(f("enterprise", "free", false)).toEqual("non-applicable");
    expect(f("enterprise", "developers", false)).toEqual("non-applicable");
    expect(f("enterprise", "enterprise", false)).toEqual("current");
  });

  it("returns reactivate or non-applicable if the plan is cancelled", () => {
    expect(f("free", "free-approved", true)).toEqual("reactivate");
    expect(f("free", "developers", true)).toEqual("non-applicable");
    expect(f("developers", "startups", true)).toEqual("non-applicable");
    expect(f("developers", "developers", true)).toEqual("reactivate");
  });

  it("returns downgrade if the card is less than the current", () => {
    expect(f("developers", "free", false)).toEqual("downgrade");
    expect(f("developers", "free-approved", false)).toEqual("downgrade");
    expect(f("startups", "developers", false)).toEqual("downgrade");
    expect(f("business", "startups", false)).toEqual("downgrade");
  });

  it("returns subscribe or upgrade if the card is greater than the current", () => {
    expect(f("free", "developers", false)).toEqual("subscribe");
    expect(f("free-approved", "developers", false)).toEqual("subscribe");
    expect(f("free", "startups", false)).toEqual("subscribe");
    expect(f("free-approved", "startups", false)).toEqual("subscribe");

    expect(f("startups", "business", false)).toEqual("upgrade");
    expect(f("developers", "startups", false)).toEqual("upgrade");
  });
});
