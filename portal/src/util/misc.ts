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

export function clearEmptyObject(obj: Record<string, any>): void {
  for (const key in obj) {
    // undefined is not supported by js-yaml library
    if (obj[key] !== undefined) {
      // null is valid value in yaml
      if (typeof obj[key] !== "object" || obj[key] === null) {
        continue;
      }
      // must call function on child first to handle the case where obj[key]
      // becomes empty after removing it's empty children object
      clearEmptyObject(obj[key]);
      // array has type "object", allow []
      if (Object.keys(obj[key]).length === 0 && !Array.isArray(obj[key])) {
        delete obj[key];
      }
    } else {
      delete obj[key];
    }
  }
}
