export function getActiveCountryCallingCode(
  pinnedCallingCodes: string[],
  selectedCallingCodes: string[]
): string[] {
  const activeCallingCodes = [...pinnedCallingCodes];
  const pinnedCallingCodeSet = new Set(pinnedCallingCodes);
  for (const code of selectedCallingCodes) {
    if (!pinnedCallingCodeSet.has(code)) {
      activeCallingCodes.push(code);
    }
  }
  return activeCallingCodes;
}
