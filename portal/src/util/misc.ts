export function isArrayEqualInOrder<T>(arr1: T[], arr2: T[]): boolean {
  if (arr1.length !== arr2.length) {
    return false;
  }
  return arr1.every((elem, index) => elem === arr2[index]);
}

export function setFieldIfChanged<
  M extends Record<K, unknown>,
  K extends string
>(map: Partial<M>, key: K, initialValue: M[K], value: M[K]): void {
  if (initialValue !== value) {
    map[key] = value;
  }
}

export function setNumericFieldIfChanged<M, K extends keyof M>(
  map: Partial<M>,
  key: K,
  initialValue: M[K] extends number ? M[K] : never,
  value: M[K] extends number ? M[K] : never
): void {
  if (initialValue !== value) {
    if (value === 0) {
      map[key] = undefined;
    } else {
      map[key] = value;
    }
  }
}

export function setFieldIfListNonEmpty<T, K extends keyof T>(
  map: T,
  field: K,
  list: T[K] extends unknown[] | undefined ? T[K] : never
): void {
  if (list == null || list.length === 0) {
    delete map[field];
  } else {
    map[field] = list;
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

// if string is empty or only contain space, return undefined
// the corresponding entry will be deleted by clearEmptyObject function
// defined above
export function ensureNonEmptyString(value?: string): string | undefined {
  if (value == null) {
    return undefined;
  }
  if (value.trim() === "") {
    return undefined;
  }
  return value;
}
