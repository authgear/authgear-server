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
