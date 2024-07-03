import { useState } from "react";

let globalID = 0;

export function useId(): string {
  const [id] = useState(() => `id-${globalID++}`);
  return id;
}
