/* global JSX */
import React from "react";
import {
  GTMProvider as ReactHookGTMProvider,
  useGTMDispatch as useReactHookGTMDispatch,
} from "@elgorditosalsero/react-gtm-hook";

export function useGTMDispatch(): any {
  try {
    return useReactHookGTMDispatch();
  } catch {
    // if container id is not configured, return no-op function
    return () => {};
  }
}

export interface GTMHookProviderProps {
  containerID?: string;
  children: React.ReactNode;
}

function GTMProvider(pros: GTMHookProviderProps): JSX.Element {
  if (pros.containerID) {
    return (
      <ReactHookGTMProvider state={{ id: pros.containerID }}>
        {pros.children}
      </ReactHookGTMProvider>
    );
  }
  return <>{pros.children}</>;
}

export default GTMProvider;
