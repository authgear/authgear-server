import React, { useMemo } from "react";
import {
  GTMProvider as ReactHookGTMProvider,
  useGTMDispatch as useReactHookGTMDispatch,
} from "@elgorditosalsero/react-gtm-hook";
import { useAppContext } from "./context/AppContext";

export type TrackValue = string | string[] | boolean | undefined;
export type EventData = Record<string, TrackValue>;

export interface AuthgearGTMEvent {
  event: AuthgearGTMEventType;
  app_id?: string;
  event_data?: EventData;
  _clear: boolean;
}

export enum AuthgearGTMEventType {
  CreatedProject = "ag.event.createdProject",
  ClickedGetStarted = "ag.event.clickedGetStarted",
  CreatedApplication = "ag.event.createdApplication",
  InvitedAdmin = "ag.event.invitedAdmin",
  AddedSSOProviders = "ag.event.addedSSOProviders",
  ClickedDocLink = "ag.event.clickedDocLink",
  ClickedNextInProjectWizard = "ag.event.clickedNextInProjectWizard",
  ClickedSkipInProjectWizard = "ag.event.clickedSkipInProjectWizard",
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
      app_id: appContextID,
      event_data: eventData,
      // Prevent GTM recursive merge event data object
      // https://github.com/google/data-layer-helper#preventing-default-recursive-merge
      _clear: true,
    };
  }, [event, eventData, appContextID]);
}

interface AuthgearGTMEventDataAttributesParams {
  event: AuthgearGTMEventType;
  eventDataAttributes?: Record<string, string>;
}

export type EventDataAttributes = Record<string, string>;

export function useAuthgearGTMEventDataAttributes({
  event,
  eventDataAttributes,
}: AuthgearGTMEventDataAttributesParams): EventDataAttributes {
  let appContextID: string | undefined;
  try {
    const appContext = useAppContext();
    appContextID = appContext.appID;
  } catch {}

  return useMemo(() => {
    const attributes: Record<string, string> = {
      "data-authgear-event": event,
    };
    if (appContextID) {
      attributes["data-authgear-event-data-app-id"] = appContextID;
    }
    if (eventDataAttributes) {
      for (const k in eventDataAttributes) {
        if (!Object.prototype.hasOwnProperty.call(eventDataAttributes, k)) {
          continue;
        }
        // only support string for data attributes
        const v = eventDataAttributes[k];
        attributes[`data-authgear-event-data-${k}`] = v;
      }
    }
    return attributes;
  }, [event, eventDataAttributes, appContextID]);
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
