export function isArrayEqualInOrder<T>(arr1: T[], arr2: T[]): boolean {
  if (arr1.length !== arr2.length) {
    return false;
  }
  return arr1.every((elem, index) => elem === arr2[index]);
}

export function setFieldIfChanged<K extends string>(
  map: Partial<Record<K, unknown>>,
  key: K,
  initialValue: unknown,
  value: unknown
): void {
  if (initialValue !== value) {
    map[key] = value;
  }
}

// This function is used to clear empty objects in new app config constructed from
// state. We create all object needed, then mutate the object, remove all empty
// object afterwards, to avoid the need of conditionally constructing new object
export function clearEmptyObject(obj: Record<string, any>): void {
  for (const key in obj) {
    if (!Object.prototype.hasOwnProperty.call(obj, key)) {
      continue;
    }
    // undefined is not supported by js-yaml library
    if (obj[key] === undefined) {
      delete obj[key];
      continue;
    }
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
  }
}
