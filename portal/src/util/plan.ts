import {
  UsageSmsRegion,
  SubscriptionItemPriceType,
  UsageType,
  UsageWhatsappRegion,
  SubscriptionUsage,
  Usage,
} from "../graphql/portal/globalTypes.generated";

export type Plan =
  | "free"
  | "free-approved"
  | "developers"
  | "developers2025"
  | "startups"
  | "business"
  | "business2025"
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
    case "developers2025":
      return true;
    case "startups":
      return true;
    case "business":
      return true;
    case "business2025":
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
    case "developers2025":
      return true;
    case "startups":
      return true;
    case "business":
      return true;
    case "business2025":
      return true;
  }
  return false;
}

export function getNextPlan(planName: string): Plan | null {
  const planNameIsPlan = isPlan(planName);
  if (!planNameIsPlan) {
    return null;
  }
  switch (planName) {
    case "free":
      return "developers2025";
    case "free-approved":
      return "developers2025";
    case "developers2025":
      return "business2025";
    default:
      return null;
  }
}

export function getMAULimit(planName: string): number | undefined {
  switch (planName) {
    case "startups":
      return 5000;
    case "business":
      return 10000;
    case "business2025":
      return 25000;
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
    case "developers2025":
      return 1;
    case "startups":
      return 2;
    case "business":
      return 3;
    case "business2025":
      return 3;
    case "enterprise":
      return 4;
  }
}

export function comparePlan(a: Plan, b: Plan): -1 | 0 | 1 {
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
  | "reactivate-to-upgrade"
  | "reactivate-to-downgrade"
  | "downgrading"
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

  if (subscriptionCancelled) {
    const isFreeCard = comparePlan(cardPlanName, "free");

    switch (compareResult) {
      case 0:
        return "reactivate";
      case -1:
        return "reactivate-to-upgrade";
      case 1:
        return isFreeCard === 0 ? "downgrading" : "reactivate-to-downgrade";
    }
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

function centToDollar(cents: number) {
  return cents / 100;
}

export interface SMSCost {
  totalCost: number;
  northAmericaUnitCost: number;
  northAmericaCount: number;
  northAmericaTotalCost: number;
  otherRegionsCount: number;
  otherRegionsUnitCost: number;
  otherRegionsTotalCost: number;
}

export function getSMSCost(
  planName: string,
  subscriptionUsage: SubscriptionUsage
): SMSCost | undefined {
  if (!isStripePlan(planName)) {
    return undefined;
  }

  const cost = {
    totalCost: 0,
    northAmericaUnitCost: 0,
    northAmericaCount: 0,
    northAmericaTotalCost: 0,
    otherRegionsCount: 0,
    otherRegionsUnitCost: 0,
    otherRegionsTotalCost: 0,
  } satisfies SMSCost;

  for (const item of subscriptionUsage.items) {
    if (
      item.type === SubscriptionItemPriceType.Usage &&
      item.usageType === UsageType.Sms
    ) {
      cost.totalCost += centToDollar(item.totalAmount ?? 0);
      if (item.smsRegion === UsageSmsRegion.NorthAmerica) {
        cost.northAmericaCount = item.quantity;
        cost.northAmericaUnitCost = centToDollar(item.unitAmount ?? 0);
        cost.northAmericaTotalCost = centToDollar(item.totalAmount ?? 0);
      }
      if (item.smsRegion === UsageSmsRegion.OtherRegions) {
        cost.otherRegionsCount = item.quantity;
        cost.otherRegionsUnitCost = centToDollar(item.unitAmount ?? 0);
        cost.otherRegionsTotalCost = centToDollar(item.totalAmount ?? 0);
      }
    }
  }

  return cost;
}

export interface WhatsappCost {
  totalCost: number;
  northAmericaUnitCost: number;
  northAmericaCount: number;
  northAmericaTotalCost: number;
  otherRegionsCount: number;
  otherRegionsUnitCost: number;
  otherRegionsTotalCost: number;
}

export function getWhatsappCost(
  planName: string,
  subscriptionUsage: SubscriptionUsage
): WhatsappCost | undefined {
  if (!isStripePlan(planName)) {
    return undefined;
  }

  const cost = {
    totalCost: 0,
    northAmericaUnitCost: 0,
    northAmericaCount: 0,
    northAmericaTotalCost: 0,
    otherRegionsCount: 0,
    otherRegionsUnitCost: 0,
    otherRegionsTotalCost: 0,
  } satisfies WhatsappCost;

  for (const item of subscriptionUsage.items) {
    if (
      item.type === SubscriptionItemPriceType.Usage &&
      item.usageType === UsageType.Whatsapp
    ) {
      cost.totalCost += centToDollar(item.totalAmount ?? 0);
      if (item.whatsappRegion === UsageWhatsappRegion.NorthAmerica) {
        cost.northAmericaCount = item.quantity;
        cost.northAmericaUnitCost = centToDollar(item.unitAmount ?? 0);
        cost.northAmericaTotalCost = centToDollar(item.totalAmount ?? 0);
      }
      if (item.whatsappRegion === UsageWhatsappRegion.OtherRegions) {
        cost.otherRegionsCount = item.quantity;
        cost.otherRegionsUnitCost = centToDollar(item.unitAmount ?? 0);
        cost.otherRegionsTotalCost = centToDollar(item.totalAmount ?? 0);
      }
    }
  }

  return cost;
}

export interface SMSUsage {
  northAmericaCount: number;
  otherRegionsCount: number;
}

export function getSMSUsage(usage: Usage): SMSUsage | undefined {
  const result = {
    northAmericaCount: 0,
    otherRegionsCount: 0,
  } satisfies SMSUsage;

  for (const item of usage.items) {
    if (item.usageType === UsageType.Sms) {
      if (item.smsRegion === UsageSmsRegion.NorthAmerica) {
        result.northAmericaCount = item.quantity;
      }
      if (item.smsRegion === UsageSmsRegion.OtherRegions) {
        result.otherRegionsCount = item.quantity;
      }
    }
  }

  return result;
}

export interface WhatsappUsage {
  northAmericaCount: number;
  otherRegionsCount: number;
}

export function getWhatsappUsage(usage: Usage): WhatsappUsage | undefined {
  const result = {
    northAmericaCount: 0,
    otherRegionsCount: 0,
  } satisfies WhatsappUsage;

  for (const item of usage.items) {
    if (item.usageType === UsageType.Whatsapp) {
      if (item.whatsappRegion === UsageWhatsappRegion.NorthAmerica) {
        result.northAmericaCount = item.quantity;
      }
      if (item.whatsappRegion === UsageWhatsappRegion.OtherRegions) {
        result.otherRegionsCount = item.quantity;
      }
    }
  }

  return result;
}
