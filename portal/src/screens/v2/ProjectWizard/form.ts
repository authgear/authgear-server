import { useCallback, useMemo } from "react";
import { SimpleFormModel, useSimpleForm } from "../../../hook/useSimpleForm";
import { produce } from "immer";

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

  // step 3 (TODO)
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
  canNavigateToNextStep: boolean;
  toNextStep: () => void;
  toPreviousStep: () => void;
  canSave: boolean;

  effectiveAuthMethods: AuthMethod[];
}

const initialState: FormState = {
  step: ProjectWizardStep.step1,

  projectName: "",
  projectID: "",

  loginMethods: [LoginMethod.Email, LoginMethod.Phone],
  authMethods: [AuthMethod.Passwordless],
};

export function useProjectWizardForm(): ProjectWizardFormModel {
  const submit = useCallback(async (_formState: FormState) => {
    // TODO
  }, []);

  const form = useSimpleForm<FormState>({
    stateMode: "UpdateInitialStateWithUseEffect",
    defaultState: initialState,
    submit,
  });

  const formState = form.state;

  const formStateSanitized = useMemo(
    () => sanitizeFormState(formState),
    [formState]
  );

  const canNavigateToNextStep = useMemo(() => {
    switch (formStateSanitized.step) {
      case ProjectWizardStep.step1:
        return (
          formStateSanitized.projectID.trim() !== "" &&
          formStateSanitized.projectName.trim() !== ""
        );
      case ProjectWizardStep.step2:
        return formStateSanitized.loginMethods.length > 0;
      case ProjectWizardStep.step3:
        // No next step
        return false;
    }
  }, [formStateSanitized]);

  const toNextStep = useCallback(() => {
    if (!canNavigateToNextStep) {
      throw new Error(
        "Cannot navigate to next step, check canNavigateToNextStep"
      );
    }
    let nextStep: ProjectWizardStep;
    switch (formState.step) {
      case ProjectWizardStep.step1:
        nextStep = ProjectWizardStep.step2;
        break;
      case ProjectWizardStep.step2:
        nextStep = ProjectWizardStep.step3;
        break;
      default:
        throw new Error("no next step is available");
    }
    form.setState((prev) => {
      return produce(prev, (draft) => {
        draft.step = nextStep;
        return draft;
      });
    });
  }, [canNavigateToNextStep, form, formState.step]);

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

  const canSave = useMemo(() => {
    if (formState.step !== ProjectWizardStep.step3) {
      return false;
    }
    // TODO
    return true;
  }, [formState.step]);

  return useMemo(
    () => ({
      ...form,
      toNextStep: toNextStep,
      canNavigateToNextStep: canNavigateToNextStep,
      toPreviousStep: toPreviousStep,
      canSave,
      effectiveAuthMethods: formStateSanitized.authMethods,
    }),
    [
      form,
      toNextStep,
      canNavigateToNextStep,
      toPreviousStep,
      canSave,
      formStateSanitized.authMethods,
    ]
  );
}
