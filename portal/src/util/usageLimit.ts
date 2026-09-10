import { StandingUsageLimitConfig } from "../types";

// A standing usage limit can carry several entries; only an `action: block`
// entry stops anything, and the smallest of those is the effective cap.
// Returns null when the plan does not cap this usage at all.
export function smallestBlockQuota(
  limits: StandingUsageLimitConfig[] | undefined
): number | null {
  const blockQuotas = (limits ?? [])
    .filter((limit) => limit.action === "block")
    .map((limit) => limit.quota)
    .filter((quota): quota is number => quota != null);
  if (blockQuotas.length === 0) {
    return null;
  }
  return Math.min(...blockQuotas);
}
