import { useCallback, useEffect, useMemo, useState } from "react";
import { SimpleFormModel, useSimpleForm } from "../../../hook/useSimpleForm";
import { produce } from "immer";
import { useLocation, useNavigate } from "react-router-dom";
import {
  projectIDFromCompanyName,
  randomProjectID,
} from "../../../util/projectname";
import { useCreateAppMutation } from "../../../graphql/portal/mutations/createAppMutation";
import { useOptionalAppContext } from "../../../context/AppContext";
import { useSaveProjectWizardDataMutation } from "../../../graphql/portal/mutations/saveProjectWizardDataMutation";
import {
  LoginIDKeyConfig,
  PortalAPIAppConfig,
  PrimaryAuthenticatorType,
} from "../../../types";
import { useAppAndSecretConfigQuery } from "../../../graphql/portal/query/appAndSecretConfigQuery";
import { useUpdateAppAndSecretConfigMutation } from "../../../graphql/portal/mutations/updateAppAndSecretMutation";

export enum ProjectWizardStep {
  "step1" = "step1",
  "step2" = "step2",
  "step3" = "step3",
}

export enum LoginMethod {
  Email = "Email",
  Phone = "Phone",
  Username = "username",
}

export enum AuthMethod {
  Passwordless = "Passwordless",
  Password = "Password",
}

export interface FormState {
  step: ProjectWizardStep;

  // step 1
  projectName: string;
  projectID: string;

  // step 2
  loginMethods: LoginMethod[];
  authMethods: AuthMethod[];

  // step 3
  logoBase64DataURL?: string;
  buttonAndLinkColor: string;
  buttonLabelColor: string;
}

function sanitizeFormState(state: FormState): FormState {
  return produce(state, (newState) => {
    newState.projectID = state.projectID.trim();
    newState.projectName = state.projectName.trim();

    const isEmailSelected = state.loginMethods.includes(LoginMethod.Email);
    const isPhoneSelected = state.loginMethods.includes(LoginMethod.Phone);
    const isUsernameSelected = state.loginMethods.includes(
      LoginMethod.Username
    );

    const selectedAuthMethods = new Set(state.authMethods);

    if (!isEmailSelected && !isPhoneSelected) {
      // Passwordless is not allowed if no email or phone
      selectedAuthMethods.delete(AuthMethod.Passwordless);
    }

    if (isUsernameSelected) {
      // Password is required if username is enabled
      selectedAuthMethods.add(AuthMethod.Password);
    }

    newState.authMethods = Array.from(selectedAuthMethods);
    return newState;
  });
}

export interface ProjectWizardFormModel extends SimpleFormModel<FormState> {
  toPreviousStep: () => void;
  canSave: boolean;
  isInitializing: boolean;
  initializeError: unknown;

  effectiveAuthMethods: AuthMethod[];
  isProjectIDEditable: boolean;
}

function processCompanyName(companyName: string): string {
  return companyName
    .trim()
    .split("")
    .filter((char) => /[a-zA-Z\s]/.exec(char))
    .join("")
    .split(" ")
    .filter((word) => word !== "")
    .join("-")
    .toLowerCase();
}

function makeDefaultState({
  projectID,
  companyName,
}: {
  projectID?: string;
  companyName?: string;
}): FormState {
  const processedCompanyName = companyName
    ? processCompanyName(companyName)
    : null;
  return {
    step: ProjectWizardStep.step1,

    projectName: companyName ? companyName : "",
    projectID: projectID
      ? projectID
      : processedCompanyName
      ? projectIDFromCompanyName(processedCompanyName)
      : randomProjectID(),

    loginMethods: [LoginMethod.Email],
    authMethods: [AuthMethod.Passwordless],

    logoBase64DataURL: undefined,
    buttonAndLinkColor: "#176DF3",
    buttonLabelColor: "#FFFFFF",
  };
}

interface LocationState {
  company_name: string;
}

function deriveLoginIDKeysFromFormState(
  formState: FormState
): LoginIDKeyConfig[] {
  const keys: LoginIDKeyConfig[] = [];
  for (const method of formState.loginMethods) {
    switch (method) {
      case LoginMethod.Email:
        keys.push({ type: "email" });
        break;
      case LoginMethod.Phone:
        keys.push({ type: "phone" });
        break;
      case LoginMethod.Username:
        keys.push({ type: "username" });
        break;
    }
  }
  return keys;
}

function derivePrimaryAuthenticatorsFromFormState(
  formState: FormState
): PrimaryAuthenticatorType[] {
  const authenticators: PrimaryAuthenticatorType[] = [];
  for (const method of formState.authMethods) {
    switch (method) {
      case AuthMethod.Password:
        authenticators.push("password");
        break;
      case AuthMethod.Passwordless:
        if (formState.loginMethods.includes(LoginMethod.Email)) {
          authenticators.push("oob_otp_email");
        }
        if (formState.loginMethods.includes(LoginMethod.Phone)) {
          authenticators.push("oob_otp_sms");
        }
        break;
    }
  }
  return authenticators;
}

function constructConfig(
  config: PortalAPIAppConfig,
  currentState: FormState
): PortalAPIAppConfig {
  return produce(config, (config) => {
    config.identity ??= {};
    config.identity.login_id ??= {};
    config.identity.login_id.keys =
      deriveLoginIDKeysFromFormState(currentState);
    config.authentication ??= {};
    config.authentication.identities = ["oauth", "login_id"];

    config.authentication.primary_authenticators =
      derivePrimaryAuthenticatorsFromFormState(currentState);
  });
}

export function useProjectWizardForm(
  initialState: FormState | null
): ProjectWizardFormModel {
  const appContext = useOptionalAppContext();
  const existingAppNodeID = appContext?.appNodeID;
  const navigate = useNavigate();
  const { state } = useLocation();

  const { createApp } = useCreateAppMutation();
  const { mutate: saveProjectWizardDataMutation } =
    useSaveProjectWizardDataMutation();

  const skip = existingAppNodeID == null;
  const {
    error: appConfigLoadError,
    rawAppConfig,
    rawAppConfigChecksum,
    refetch: reloadAppConfig,
  } = useAppAndSecretConfigQuery(existingAppNodeID!, null, skip);
  const { updateAppAndSecretConfig: updateAppConfig } =
    useUpdateAppAndSecretConfigMutation(existingAppNodeID!);

  const [defaultState, setDefaultState] = useState(() => {
    if (initialState != null) {
      return initialState;
    }
    const typedState: LocationState | null = state as LocationState | null;
    const defaultState = makeDefaultState({
      projectID: appContext?.appID,
      companyName: typedState?.company_name,
    });
    return defaultState;
  });

  const submit = useCallback(
    async (formState: FormState): Promise<string | null> => {
      const sanitizedFormState = sanitizeFormState(formState);
      if (!computeCanSave(sanitizedFormState)) {
        throw new Error(
          "Cannot navigate to next step, check canNavigateToNextStep"
        );
      }
      const updatedState = produce(formState, (draft) => {
        switch (draft.step) {
          case ProjectWizardStep.step1: {
            draft.step = ProjectWizardStep.step2;
            break;
          }
          case ProjectWizardStep.step2:
            draft.step = ProjectWizardStep.step3;
            break;
          case ProjectWizardStep.step3:
            break;
        }
        return draft;
      });
      switch (formState.step) {
        case ProjectWizardStep.step1: {
          if (!existingAppNodeID) {
            const appID = await createApp(
              sanitizedFormState.projectID,
              updatedState
            );
            setDefaultState(updatedState);
            return `/project/${encodeURIComponent(appID!)}/wizard`;
            // eslint-disable-next-line no-else-return
          } else {
            await saveProjectWizardDataMutation(
              existingAppNodeID,
              updatedState
            );
            setDefaultState(updatedState);
            return null;
          }
        }
        case ProjectWizardStep.step2:
          await saveProjectWizardDataMutation(existingAppNodeID!, updatedState);
          setDefaultState(updatedState);
          return null;
        case ProjectWizardStep.step3:
          if (rawAppConfig == null) {
            throw new Error("unexpected error: rawAppConfig is null");
          }
          await updateAppConfig({
            appConfig: constructConfig(rawAppConfig, updatedState),
            appConfigChecksum: rawAppConfigChecksum,
            ignoreConflict: true,
          });
          // Set it to null to indicate the flow is finished
          await saveProjectWizardDataMutation(existingAppNodeID!, null);
          await reloadAppConfig();
          setDefaultState(updatedState);

          return `/project/${existingAppNodeID}`;
      }
    },
    [
      createApp,
      existingAppNodeID,
      rawAppConfig,
      rawAppConfigChecksum,
      reloadAppConfig,
      saveProjectWizardDataMutation,
      updateAppConfig,
    ]
  );

  const form = useSimpleForm<FormState>({
    stateMode: "UpdateInitialStateWithUseEffect",
    defaultState: defaultState,
    submit,
  });

  const formState = form.state;
  const nextPath = form.submissionResult;

  useEffect(() => {
    if (nextPath) {
      navigate(nextPath, { replace: true });
    }
  }, [navigate, nextPath]);

  const formStateSanitized = useMemo(
    () => sanitizeFormState(formState),
    [formState]
  );

  const toPreviousStep = useCallback(() => {
    form.setState((prev) => {
      return produce(prev, (draft) => {
        switch (formState.step) {
          case ProjectWizardStep.step2:
            draft.step = ProjectWizardStep.step1;
            break;
          case ProjectWizardStep.step3:
            draft.step = ProjectWizardStep.step2;
            break;

          default:
            throw new Error("no previous step is available");
        }
        return draft;
      });
    });
  }, [form, formState.step]);

  const canSave = useMemo(
    () => computeCanSave(formStateSanitized),
    [formStateSanitized]
  );

  return useMemo(
    () => ({
      ...form,
      toPreviousStep: toPreviousStep,
      canSave,
      isInitializing: rawAppConfig == null,
      initializeError: appConfigLoadError,
      effectiveAuthMethods: formStateSanitized.authMethods,
      isProjectIDEditable: existingAppNodeID == null,
    }),
    [
      form,
      toPreviousStep,
      canSave,
      rawAppConfig,
      appConfigLoadError,
      formStateSanitized.authMethods,
      existingAppNodeID,
    ]
  );
}

function computeCanSave(formStateSanitized: FormState): boolean {
  switch (formStateSanitized.step) {
    case ProjectWizardStep.step1:
      return (
        formStateSanitized.projectID.trim() !== "" &&
        formStateSanitized.projectName.trim() !== ""
      );
    case ProjectWizardStep.step2: {
      if (formStateSanitized.loginMethods.length === 0) {
        return false;
      }
      const loginMethods = new Set(formStateSanitized.loginMethods);
      if (
        formStateSanitized.authMethods.length === 0 &&
        [LoginMethod.Email, LoginMethod.Phone, LoginMethod.Username].some(
          (method) => loginMethods.has(method)
        )
      ) {
        return false;
      }
      return true;
    }
    case ProjectWizardStep.step3:
      return true;
  }
}
