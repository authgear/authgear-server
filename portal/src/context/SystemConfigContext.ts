import { createContext, useContext } from "react";
import { SystemConfig } from "../system-config";

type SystemConfigContextValue = SystemConfig;

export const SystemConfigContext =
  createContext<SystemConfigContextValue | null>(null);

export function useSystemConfig(): SystemConfigContextValue {
  const value = useContext(SystemConfigContext);
  if (!value) {
    // Should not happen at runtime
    throw Error();
  }
  return value;
}
