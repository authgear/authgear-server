import React, { useMemo } from "react";
// eslint-disable-next-line no-restricted-imports
import { GTMProvider as ReactHookGTMProvider } from "@elgorditosalsero/react-gtm-hook";

export interface GTMProviderProps {
  containerID?: string;
  children: React.ReactNode;
}

const GTMProvider: React.VFC<GTMProviderProps> = ({
  containerID,
  children,
}) => {
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
