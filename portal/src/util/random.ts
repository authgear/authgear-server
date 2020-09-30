export function genRandomHexadecimalString(length: number = 16): string {
  let result = "";
  if (length < 0) {
    return result;
  }
  for (let i = 0; i < length; i++) {
    result += Math.floor(Math.random() * 16).toString(16);
  }
  return result;
}
