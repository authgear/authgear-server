export function isArrayEqualInOrder(arr1: unknown[], arr2: unknown[]): boolean {
  if (arr1.length !== arr2.length) {
    return false;
  }
  return arr1.every((elem, index) => elem === arr2[index]);
}

export function setFieldIfChanged(
  map: Record<string, unknown>,
  field: string,
  initialValue: unknown,
  value: unknown
): void {
  if (initialValue !== value) {
    map[field] = value;
  }
}
