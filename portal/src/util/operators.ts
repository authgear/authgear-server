export function or_(
  expr0: boolean,
  expr1: boolean,
  ...exprs: boolean[]
): boolean {
  return [expr0, expr1]
    .concat(...exprs)
    .reduce((sum, current) => sum || current, false);
}

export function nullishCoalesce<T>(
  maybe0: T | null,
  maybe1: T | null,
  ...maybes: (T | null)[]
): T | null {
  return [maybe0, maybe1]
    .concat(...maybes)
    .reduce((sum, current) => sum ?? current, null);
}
