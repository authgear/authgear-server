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
  authMethod: AuthMethod;

  // step 3 (TODO)
}

export interface ProjectWizardFormModel extends SimpleFormModel<FormState> {
  canNavigateToNextStep: boolean;
  toNextStep: () => void;
  toPreviousStep: () => void;
  canSave: boolean;
}

const initialState: FormState = {
  step: ProjectWizardStep.step1,

  projectName: "",
  projectID: "",

  loginMethods: [LoginMethod.Email, LoginMethod.Phone],
  authMethod: AuthMethod.Passwordless,
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

  const canNavigateToNextStep = useMemo(() => {
    switch (formState.step) {
      case ProjectWizardStep.step1:
        return (
          formState.projectID.trim() !== "" &&
          formState.projectName.trim() !== ""
        );
      case ProjectWizardStep.step2:
        return formState.loginMethods.length > 0;
      case ProjectWizardStep.step3:
        // No next step
        return false;
    }
  }, [formState]);

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
    }),
    [form, toNextStep, canNavigateToNextStep, toPreviousStep, canSave]
  );
}
