import { fetchUnresolvedCounts } from "./api";
import { setUnresolvedCounts } from "./countsStore";

export async function refreshUnresolvedCounts(): Promise<void> {
  try {
    setUnresolvedCounts(await fetchUnresolvedCounts());
  } catch {
    setUnresolvedCounts({});
  }
}
