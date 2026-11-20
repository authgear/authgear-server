import { FlowStepItem } from "./tree";

export function findStep(
  tree: FlowStepItem[],
  id: string
): FlowStepItem | null {
  for (const item of tree) {
    if (item.id === id) {
      return item;
    }
    for (const branch of item.branches ?? []) {
      const found = findStep(branch.sequence, id);
      if (found != null) {
        return found;
      }
    }
  }
  return null;
}
