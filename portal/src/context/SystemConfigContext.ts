import { createContext, useContext } from "react";

import { SystemConfig } from "../system-config";
export { SystemConfig };

type SystemConfigContextValue = SystemConfig | null;

export const SystemConfigContext = createContext<SystemConfigContextValue>(
  null
);

export function useSystemConfig(): SystemConfigContextValue {
  return useContext(SystemConfigContext);
}
