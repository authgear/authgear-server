import { createContext, useContext } from "react";

export interface RuntimeConfig {
  authgear_client_id: string;
  authgear_endpoint: string;
}

type RuntimeConfigContextValue = RuntimeConfig | null;

export const RuntimeConfigContext = createContext<RuntimeConfigContextValue>(
  null
);

export function useRuntimeConfig(): RuntimeConfigContextValue {
  return useContext(RuntimeConfigContext);
}
