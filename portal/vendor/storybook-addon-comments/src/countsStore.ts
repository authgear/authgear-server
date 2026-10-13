type Listener = () => void;

let counts: Record<string, number> = {};
const listeners = new Set<Listener>();

export function getUnresolvedCounts(): Record<string, number> {
  return counts;
}

export function setUnresolvedCounts(next: Record<string, number>) {
  counts = next;
  listeners.forEach((listen) => listen());
}

export function subscribeUnresolvedCounts(listen: Listener): () => void {
  listeners.add(listen);
  return () => listeners.delete(listen);
}
