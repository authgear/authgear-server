// This file lazily imports zxcvbn, so that it's not included in the main bundle

import type zxcvbn from "zxcvbn";

type Zxcvbn = (password: string, userInputs?: string[]) => zxcvbn.ZXCVBNResult;

export async function runZxcvbn(value: string): Promise<zxcvbn.ZXCVBNResult> {
  const z = (await import("zxcvbn")) as unknown as Zxcvbn;
  return z(value);
}
