import React, { useMemo } from "react";
import {
  GTMProvider as ReactHookGTMProvider,
  useGTMDispatch as useReactHookGTMDispatch,
} from "@elgorditosalsero/react-gtm-hook";

export interface AuthgearGTMEvent {
  event: AuthgearGTMEventType;
  value1?: string;
}

export enum AuthgearGTMEventType {}

export function useGTMDispatch(): (event: AuthgearGTMEvent) => void {
  try {
    return useReactHookGTMDispatch();
  } catch {
    // if container id is not configured, return no-op function
    return () => {};
  }
}

export interface GTMProviderProps {
  containerID?: string;
  children: React.ReactNode;
}

const GTMProvider: React.FC<GTMProviderProps> = ({ containerID, children }) => {
  const state = useMemo(() => {
    return { id: containerID ?? "" };
  }, [containerID]);

  if (containerID) {
    return (
      <ReactHookGTMProvider state={state}>{children}</ReactHookGTMProvider>
    );
  }
  return <>{children}</>;
};

export default GTMProvider;
