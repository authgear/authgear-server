export function parseEmail(inp: string): string | null {
  const lookLike = /^\s*\S+@\S+\s*$/.test(inp);
  if (lookLike) {
    return inp.trim();
  }
  return null;
}
