export function parseIntegerAllowLeadingZeros(
  value?: string
): number | undefined {
  if (value != null && value !== "") {
    try {
      // Remove leading zeros.
      const num = parseInt(value, 10);
      if (!isNaN(num)) {
        return num;
      }
    } catch {}
  }
  return undefined;
}

export function checkNumberInput(value: string): boolean {
  // Allow empty.
  if (value === "") {
    return true;
  }
  // Allow "-"
  if (value === "-") {
    return true;
  }
  return /^-?(0|[1-9][0-9]*)(\.[0-9]*)?$/.test(value);
}

export function checkIntegerInput(value: string): boolean {
  // Allow "", "-", or valid integer.
  return /^-?(|0|[1-9][0-9]*)$/.test(value);
}

export function parseNumber(value: string | undefined): number | undefined {
  if (value == null || value === "") {
    return undefined;
  }
  const num = Number(value.trim());
  if (Number.isNaN(num)) {
    throw new Error("Value is not a number");
  }
  return num;
}

export function ensureInteger(value: number): number {
  if (!Number.isInteger(value)) {
    throw new Error("Number is non-integer");
  }
  return value;
}

export function ensurePositiveNumber(value: number): number {
  if (Number.isNaN(value) || value <= 0) {
    throw new Error("Number is non-positive");
  }
  return value;
}

export function tryProduce<T>(fallback: T, produce: () => T): T {
  try {
    return produce();
  } catch {
    return fallback;
  }
}
