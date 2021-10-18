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
