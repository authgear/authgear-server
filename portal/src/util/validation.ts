export function defaultFormatErrorMessageList(
  errorMessages: string[]
): string | undefined {
  return errorMessages.length === 0 ? undefined : errorMessages.join("\n");
}
