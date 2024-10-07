export type Plan =
  | "free"
  | "free-approved"
  | "developers"
  | "startups"
  | "business"
  | "enterprise";

export const SUBSCRIPTABLE_PLANS: Plan[] = ["startups", "business"];

export const ENTERPRISE_PLAN: Plan = "enterprise";

export function isPlan(planName: string): planName is Plan {
  switch (planName) {
    case "free":
      return true;
    case "free-approved":
      return true;
    case "developers":
      return true;
    case "startups":
      return true;
    case "business":
      return true;
    case "enterprise":
      return true;
  }
  return false;
}

export function isCustomPlan(planName: string): boolean {
  return !isPlan(planName);
}

export function isLimitedFreePlan(planName: string): planName is Plan {
  return planName === "free";
}

export function isFreePlan(planName: string): planName is Plan {
  return planName === "free" || planName === "free-approved";
}

export function isStripePlan(planName: string): planName is Plan {
  switch (planName) {
    case "developers":
      return true;
    case "startups":
      return true;
    case "business":
      return true;
  }
  return false;
}

export function getPreviousPlan(planName: string): Plan | null {
  const planNameIsPlan = isPlan(planName);
  if (!planNameIsPlan) {
    return null;
  }
  switch (planName) {
    case "free":
      return null;
    case "free-approved":
      return null;
    case "developers":
      return null;
    case "startups":
      return "free";
    case "business":
      return "startups";
    case "enterprise":
      return "business";
  }
}

function isRecommendedPlan(planName: string): planName is Plan {
  return planName === "startups";
}

export function shouldShowRecommendedTag(
  planName: string,
  currentPlanName: string
): boolean {
  const thePlanIsRecommended = isRecommendedPlan(planName);
  const currentPlanIsFree = isFreePlan(currentPlanName);
  return thePlanIsRecommended && currentPlanIsFree;
}

export function getMAULimit(planName: string): number | undefined {
  switch (planName) {
    case "free":
      return 5000;
    case "free-approved":
      return 5000;
    case "startups":
      return 5000;
    case "business":
      return 10000;
  }
  return undefined;
}

function planToNumber(planName: Plan): number {
  switch (planName) {
    case "free":
      return 0;
    case "free-approved":
      return 0;
    case "developers":
      return 1;
    case "startups":
      return 2;
    case "business":
      return 3;
    case "enterprise":
      return 4;
  }
}

function comparePlan(a: Plan, b: Plan): -1 | 0 | 1 {
  const numberA = planToNumber(a);
  const numberB = planToNumber(b);
  if (numberA > numberB) {
    return 1;
  } else if (numberA < numberB) {
    return -1;
  }
  return 0;
}

export type CTAVariant =
  | "non-applicable"
  | "subscribe"
  | "downgrade"
  | "upgrade"
  | "reactivate"
  | "current"
  | "contact-us";

export function getCTAVariant(opts: {
  cardPlanName: string;
  currentPlanName: string;
  subscriptionCancelled: boolean;
}): CTAVariant {
  const { cardPlanName, currentPlanName, subscriptionCancelled } = opts;

  if (!isPlan(cardPlanName) || !isPlan(currentPlanName)) {
    return "non-applicable";
  }
  // Now we know cardPlanName and currentPlanName are Plan.

  // If the card is enterprise, then it can only be either current or contact-us.
  if (cardPlanName === ENTERPRISE_PLAN) {
    return currentPlanName === ENTERPRISE_PLAN ? "current" : "contact-us";
  }

  // If the current plan is enterprise, then the card is either current or non-applicable.
  if (currentPlanName === ENTERPRISE_PLAN) {
    return cardPlanName === ENTERPRISE_PLAN ? "current" : "non-applicable";
  }

  // Now we know cardPlanName and currentPlanName are Plan, and they ARE NOT enterprise.

  const compareResult = comparePlan(currentPlanName, cardPlanName);

  // If the current plan is cancelled, the only sensible CTA is reactivate.
  if (subscriptionCancelled) {
    if (compareResult === 0) {
      return "reactivate";
    }
    return "non-applicable";
  }

  if (compareResult === 1) {
    // currentPlanName is greater than the card. The CTA on the card is thus downgrade.
    return "downgrade";
  } else if (compareResult === -1) {
    // currentPlanName is less than the card. The CTA on the card is either subscribe or upgrade.
    return isFreePlan(currentPlanName) ? "subscribe" : "upgrade";
  }

  // Now we know cardPlanName is equal to currentPlanName.
  return "current";
}
