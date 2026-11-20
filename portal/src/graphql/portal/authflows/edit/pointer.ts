// JSON pointers within one flow, as used for node IDs: "/steps/0/one_of/1".
// Flow pointers never contain "~" or "/" inside a token, so no unescaping.

export type PointerToken = string | number;

export function parsePointer(pointer: string): PointerToken[] {
  return pointer
    .split("/")
    .slice(1)
    .map((t) => (/^\d+$/.test(t) ? Number(t) : t));
}

export function getAt(root: unknown, pointer: string): unknown {
  let current: unknown = root;
  for (const token of parsePointer(pointer)) {
    if (current == null || typeof current !== "object") {
      return undefined;
    }
    current = (current as Record<PointerToken, unknown>)[token];
  }
  return current;
}

// A place in a steps array. The array may not exist yet, as for an option
// with no nested steps.
export interface InsertionPoint {
  steps: string;
  index: number;
}

export function parentSteps(stepPointer: string): InsertionPoint {
  const slash = stepPointer.lastIndexOf("/");
  return {
    steps: stepPointer.slice(0, slash),
    index: Number(stepPointer.slice(slash + 1)),
  };
}

export function stepPointer(at: InsertionPoint): string {
  return `${at.steps}/${at.index}`;
}

// Whether pointer is the step at ancestor or nested under it.
export function isWithin(pointer: string, ancestor: string): boolean {
  return pointer === ancestor || pointer.startsWith(`${ancestor}/`);
}

// The option whose steps at is in; null for the flow's top-level steps.
export function optionOf(
  at: InsertionPoint
): { choice: string; optionIndex: number } | null {
  const m = /^(.*)\/one_of\/(\d+)\/steps$/.exec(at.steps);
  return m != null ? { choice: m[1], optionIndex: Number(m[2]) } : null;
}
