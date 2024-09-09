// eslint-disable-next-line no-restricted-imports
import zxcvbn from "zxcvbn";

export type GuessableLevel = 0 | 1 | 2 | 3 | 4 | 5;

export function zxcvbnGuessableLevel(
  password: string | null,
  excludedKeywords?: string[]
): GuessableLevel {
  if (password === "" || password == null) {
    return 0;
  }

  const result = zxcvbn(password, excludedKeywords);
  return Math.floor(
    Math.min(5, Math.max(1, result.score + 1))
  ) as GuessableLevel;
}
