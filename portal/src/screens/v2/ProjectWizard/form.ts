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
          // Set it to null to indicate the flow is finished
          await saveProjectWizardDataMutation(existingAppNodeID!, null);
          setDefaultState(updatedState);
          return `/project/${existingAppNodeID}`;
      }
    },
    [createApp, existingAppNodeID, saveProjectWizardDataMutation]
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
      navigate(nextPath);
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
      effectiveAuthMethods: formStateSanitized.authMethods,
      isProjectIDEditable: existingAppNodeID == null,
    }),
    [
      form,
      toPreviousStep,
      canSave,
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
