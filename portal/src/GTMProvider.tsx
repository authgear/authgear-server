import React, { useMemo } from "react";
import {
  GTMProvider as ReactHookGTMProvider,
  useGTMDispatch as useReactHookGTMDispatch,
} from "@elgorditosalsero/react-gtm-hook";
import { useAppContext } from "./context/AppContext";

export interface AuthgearGTMEvent {
  event: AuthgearGTMEventType;
  appID?: string;
  value1?: string;
}

export enum AuthgearGTMEventType {
  CreateProject = "authgear.createProject",
  ClickGetStarted = "authgear.clickGetStarted",
  CreateApplication = "authgear.createApplication",
}

interface AuthgearGTMEventParams {
  event: AuthgearGTMEventType;
  value1?: string;
}

export function useAuthgearGTMEvent({
  event,
  value1,
}: AuthgearGTMEventParams): AuthgearGTMEvent {
  let appContextID: string | undefined;
  try {
    const appContext = useAppContext();
    appContextID = appContext.appID;
  } catch {}

  return useMemo(() => {
    return {
      event: event,
      appID: appContextID,
      value1: value1,
    };
  }, [event, value1, appContextID]);
}

export function useAuthgearGTMEventDataAttributes(
  params: AuthgearGTMEventParams
): Record<string, string> {
  const event = useAuthgearGTMEvent(params);
  return useMemo(() => {
    const attributes: Record<string, string> = {
      "data-authgear-event": event.event,
    };
    if (event.appID) {
      attributes["data-authgear-event-app-id"] = event.appID;
    }
    if (event.value1) {
      attributes["data-authgear-event-value1"] = event.value1;
    }
    return attributes;
  }, [event]);
}

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
