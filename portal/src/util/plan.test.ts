import { describe, it, expect } from "@jest/globals";
import { getNextPlan, getCTAVariant } from "./plan";

describe("getNextPlan", () => {
  it("should work", () => {
    expect(getNextPlan("free")).toEqual("developers2025");
    expect(getNextPlan("free-approved")).toEqual("developers2025");
    expect(getNextPlan("developers2025")).toEqual("business2025");

    expect(getNextPlan("startups")).toEqual(null);
    expect(getNextPlan("business")).toEqual(null);
    expect(getNextPlan("enterprise")).toEqual(null);

    expect(getNextPlan("foobar")).toEqual(null);
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

  it("returns reactivate, downgrading, reactivate-to-upgrade or reactivate-to-downgrade if the plan is cancelled", () => {
    expect(f("developers", "free-approved", true)).toEqual("downgrading");
    expect(f("free", "developers", true)).toEqual("reactivate-to-upgrade");
    expect(f("business", "developers", true)).toEqual(
      "reactivate-to-downgrade"
    );
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
