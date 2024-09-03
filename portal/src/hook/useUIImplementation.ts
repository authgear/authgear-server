import { useSystemConfig } from "../context/SystemConfigContext";
import { UIImplementation } from "../types";

export function useUIImplementation(
  projectValue: UIImplementation | undefined
): UIImplementation {
  const systemConfig = useSystemConfig();
  if (projectValue != null) {
    return projectValue;
  }
  return systemConfig.uiImplementation as UIImplementation;
}
