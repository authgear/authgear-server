export function toOptionalText(raw?: string | null): string | undefined {
  if (raw == null) {
    return undefined;
  }
  const trimmedRaw = raw.trim();
  if (!trimmedRaw) {
    return undefined;
  }
  return trimmedRaw;
}
