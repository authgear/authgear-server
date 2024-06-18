import { useCallback } from "react";
// eslint-disable-next-line no-restricted-imports
import { useGTMDispatch as useReactHookGTMDispatch } from "@elgorditosalsero/react-gtm-hook";
import { AppContextValue, useAppContext } from "./context/AppContext";

enum AuthgearGTMTriggerTypeV2 {
  Identify = "ag.v2.identify",
  Reset = "ag.v2.reset",
  CustomEvent = "ag.v2.customevent",
}

export type AuthgearGTMEventTypeV2 =
  | "projectWizard.clicked-option"
  | "projectWizard.set-primary_auth"
  | "projectWizard.set-passkey"
  | "projectWizard.set-2fa"
  | "projectWizard.clicked-skip"
  | "projectWizard.clicked-back"
  | "projectWizard.completed"
  | "getStarted.clicked-signup"
  | "getStarted.clicked-customize"
  | "getStarted.clicked-create_app"
  | "getStarted.clicked-social_login"
  | "getStarted.clicked-add_members"
  | "getStarted.clicked-done"
  | "getStarted.clicked-forum"
  | "getStarted.clicked-contact_us"
  | "onboardingSurvey.set-role"
  | "onboardingSurvey.set-team_or_personal_account"
  | "onboardingSurvey.set-company_details"
  | "onboardingSurvey.set-project_details"
  | "onboardingSurvey.set-use_cases"
  | "onboardingSurvey.set-completed-survey"
  | "onboardingSurvey.set-clicked-back"
  | "enteredProject"
  | "header.clicked-contact_us"
  | "header.clicked-docs";

export type AuthgearGTMEventDataValueV2 =
  | string
  | number
  | boolean
  | Record<string, string>
  | undefined;
export type AuthgearGTMEventDataV2 = Record<
  string,
  AuthgearGTMEventDataValueV2
>;

function useDispatch(): (event: object) => void {
  try {
    return useReactHookGTMDispatch();
  } catch {
    // if container id is not configured, return no-op function
    return () => {};
  }
}

export function useIdentify(): (userID: string, email?: string) => void {
  const dispatch = useDispatch();
  const callback = useCallback(
    (userID: string, email?: string) => {
      dispatch({
        // event is a builtin variable.
        // https://support.google.com/tagmanager/answer/7679219?hl=en
        event: AuthgearGTMTriggerTypeV2.Identify,
        // event_data is a user-defined variable.
        event_data: {
          distinct_id: userID,
          email,
        },
        // Prevent GTM recursive merge event data object
        // https://github.com/google/data-layer-helper#preventing-default-recursive-merge
        _clear: true,
      });
    },
    [dispatch]
  );
  return callback;
}

export function useReset(): () => void {
  const dispatch = useDispatch();
  const callback = useCallback(() => {
    dispatch({
      event: AuthgearGTMTriggerTypeV2.Reset,
      // Prevent GTM recursive merge event data object
      // https://github.com/google/data-layer-helper#preventing-default-recursive-merge
      _clear: true,
    });
  }, [dispatch]);
  return callback;
}

export function useCapture(): (
  event: AuthgearGTMEventTypeV2,
  data?: AuthgearGTMEventDataV2,
  // app_context_override is specifically for enteredProject.
  // When enteredProject is tracked, app_context is still old and contains no project_id.
  app_context_override?: { project_id: string }
) => void {
  let appContext: AppContextValue | undefined;
  try {
    const ac = useAppContext();
    appContext = ac;
  } catch {}

  const dispatch = useDispatch();
  const callback = useCallback(
    (
      event: AuthgearGTMEventTypeV2,
      data?: AuthgearGTMEventDataV2,
      app_context_override?: { project_id: string }
    ) => {
      dispatch({
        // event is a builtin variable.
        // https://support.google.com/tagmanager/answer/7679219?hl=en
        event: AuthgearGTMTriggerTypeV2.CustomEvent,
        // app_context is a user-defined variable.
        app_context: {
          project_id: appContext?.appID,
          ...app_context_override,
        },
        // event_data is a user-defined variable.
        event_data: {
          event: event,
          ...data,
        },
        // Prevent GTM recursive merge event data object
        // https://github.com/google/data-layer-helper#preventing-default-recursive-merge
        _clear: true,
      });
    },
    [dispatch, appContext?.appID]
  );
  return callback;
}
