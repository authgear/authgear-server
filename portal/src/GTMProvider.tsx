import React, { useMemo } from "react";
import {
  GTMProvider as ReactHookGTMProvider,
  useGTMDispatch as useReactHookGTMDispatch,
} from "@elgorditosalsero/react-gtm-hook";
import { useAppContext } from "./context/AppContext";

export type TrackValue = string | string[] | undefined;
export type EventData = Record<string, TrackValue>;

export interface AuthgearGTMEvent {
  event: AuthgearGTMEventType;
  appId?: string;
  eventData?: EventData;
}

export enum AuthgearGTMEventType {
  CreateProject = "authgear.createProject",
  ClickGetStarted = "authgear.clickGetStarted",
  CreateApplication = "authgear.createApplication",
  InviteAdmin = "authgear.inviteAdmin",
  AddSSOProviders = "authgear.addSSOProviders",
}

interface AuthgearGTMEventParams {
  event: AuthgearGTMEventType;
  eventData?: EventData;
}

export function useAuthgearGTMEvent({
  event,
  eventData,
}: AuthgearGTMEventParams): AuthgearGTMEvent {
  let appContextID: string | undefined;
  try {
    const appContext = useAppContext();
    appContextID = appContext.appID;
  } catch {}

  return useMemo(() => {
    return {
      event: event,
      appId: appContextID,
      eventData,
    };
  }, [event, eventData, appContextID]);
}

export function useAuthgearGTMEventDataAttributes(
  params: AuthgearGTMEventParams
): Record<string, string> {
  const event = useAuthgearGTMEvent(params);
  return useMemo(() => {
    const attributes: Record<string, string> = {
      "data-authgear-event": event.event,
    };
    if (event.appId) {
      attributes["data-authgear-event-data-app-id"] = event.appId;
    }
    if (event.eventData) {
      for (const k in event.eventData) {
        if (!Object.prototype.hasOwnProperty.call(event.eventData, k)) {
          continue;
        }
        // only support string for data attributes
        const v = event.eventData[k];
        if (typeof v === "string") {
          attributes[`data-authgear-event-data-${k}`] = v;
        }
      }
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
