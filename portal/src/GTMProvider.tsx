import React, { useCallback, useMemo } from "react";
import {
  GTMProvider as ReactHookGTMProvider,
  useGTMDispatch as useReactHookGTMDispatch,
} from "@elgorditosalsero/react-gtm-hook";
import { useAppContext } from "./context/AppContext";

export type TrackValue = string | string[] | boolean | undefined;
export type EventData = Record<string, TrackValue>;

export interface AuthgearGTMEventBase {
  app_id?: string;
  _clear: boolean;
}

export interface AuthgearGTMEvent extends AuthgearGTMEventBase {
  event: AuthgearGTMEventType;
  event_data?: EventData;
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
  Identified = "ag.lifecycle.identified",
}

export function useAuthgearGTMEventBase(): AuthgearGTMEventBase {
  let appContextID: string | undefined;
  try {
    const appContext = useAppContext();
    appContextID = appContext.appID;
  } catch {}

  return useMemo(() => {
    return {
      app_id: appContextID,
      // Prevent GTM recursive merge event data object
      // https://github.com/google/data-layer-helper#preventing-default-recursive-merge
      _clear: true,
    };
  }, [appContextID]);
}

interface AuthgearGTMEventDataAttributesParams {
  event: AuthgearGTMEventType;
  eventDataAttributes?: Record<string, string>;
}

export type EventDataAttributes = Record<string, string>;

export function useMakeAuthgearGTMEventDataAttributes(): (
  params: AuthgearGTMEventDataAttributesParams
) => EventDataAttributes {
  let appContextID: string | undefined;
  try {
    const appContext = useAppContext();
    appContextID = appContext.appID;
  } catch {}

  const makeGTMEventDataAttributes = useCallback(
    ({ event, eventDataAttributes }: AuthgearGTMEventDataAttributesParams) => {
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
    },
    [appContextID]
  );

  return makeGTMEventDataAttributes;
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
