import { z } from "zod";
import { SimpleFormModel, useSimpleForm } from "../../../hook/useSimpleForm";
import { useCallback, useMemo, useState } from "react";
import { useDebouncedEffect } from "../../../hook/useDebouncedEffect";
import { produce } from "immer";
import { useCapture } from "../../../gtm_v2";
import { useSaveOnboardingSurveyMutation } from "../../../graphql/portal/mutations/saveOnboardingSurveyMutation";
import { useViewerQuery } from "../../../graphql/portal/query/viewerQuery";

export enum Role {
  Developer = "Developer",
  ProjectManager = "ProjectManager",
  Business = "Business",
  Other = "Other",
}

export enum TeamOrPersonal {
  Team = "Team",
  Personal = "Personal",
}

export enum CompanySize {
  "1-to-49" = "1-to-49",
  "50-to-199" = "50-to-199",
  "200-to-999" = "200-to-999",
  "1000+" = "1000+",
}

export enum UseCase {
  BuildingNewSoftwareProject = "BuildingNewSoftwareProject",
  SSOSolution = "SSOSolution",
  EnhanceSecurity = "EnhanceSecurity",
  Other = "Other",
}

export enum OnboardingSurveyStep {
  "start" = "start",
  "step1" = "step1",
  "step2" = "step2",
  "step3" = "step3",
  "step4" = "step4",
}

const zFormState = z.object({
  step: z.nativeEnum(OnboardingSurveyStep),

  // step 1
  role: z.nativeEnum(Role).optional(),

  // step 2
  team_or_personal_account: z.nativeEnum(TeamOrPersonal).optional(),

  // step 3
  phone_number: z.string().optional(),
  // step 3 - team
  company_name: z.string().optional(),
  company_size: z.nativeEnum(CompanySize).optional(),
  // step 3 - personal
  project_website: z.string().optional(),

  // step 4
  use_cases: z.nativeEnum(UseCase).array().optional(),
  use_case_other: z.string().optional(),
});

export type FormState = z.infer<typeof zFormState>;

const STORAGE_KEY = "authgear-onboarding-survey-v2";

function readFormStateFromStorage(storage: Storage): FormState | null {
  const raw = storage.getItem(STORAGE_KEY);
  if (raw == null) return null;
  const parseResult = z
    .preprocess((raw) => {
      if (typeof raw !== "string") {
        return null;
      }
      return JSON.parse(raw);
    }, zFormState)
    .safeParse(raw);
  if (parseResult.success) {
    return parseResult.data;
  }
  console.warn("failed to parse saved form state, ignoring", parseResult.error);
  return null;
}

function writeFormStateToStorage(storage: Storage, s: FormState): void {
  storage.setItem(STORAGE_KEY, JSON.stringify(s));
}

function deleteFormStateFromStorage(storage: Storage): void {
  storage.removeItem(STORAGE_KEY);
}

const STORAGE = window.localStorage;

export interface OnboardingSurveyFormModel extends SimpleFormModel<FormState> {
  canNavigateToNextStep: boolean;
  toNextStep: () => void;
  toPreviousStep: () => void;
  canSave: boolean;
}

export function useOnboardingSurveyForm(): OnboardingSurveyFormModel {
  const capture = useCapture();

  const { saveOnboardingSurveyHook: updateCustAttrHook } =
    useSaveOnboardingSurveyMutation();
  const { refetch: refetchViewer } = useViewerQuery();

  const [initialState, setInitialState] = useState<FormState>(() => {
    return (
      readFormStateFromStorage(STORAGE) ?? {
        step: OnboardingSurveyStep.start,
      }
    );
  });

  const submit = useCallback(
    async (formState: FormState) => {
      capture("onboardingSurvey.set-use_cases");
      // TODO(tung): Send the event in server
      // capture("onboardingSurvey.set-completed-survey", {
      //   $set: {
      //     survey_json: formState,
      //   },
      // });
      const stateJsonStr = JSON.stringify(formState);
      await updateCustAttrHook(stateJsonStr);
      await refetchViewer();
      deleteFormStateFromStorage(STORAGE);
      setInitialState(formState);
    },
    [capture, refetchViewer, updateCustAttrHook]
  );

  const form = useSimpleForm<FormState>({
    stateMode: "UpdateInitialStateWithUseEffect",
    defaultState: initialState,
    submit,
  });

  const formState = form.state;
  const isSubmitted = form.isSubmitted;

  useDebouncedEffect(
    useCallback(() => {
      if (!isSubmitted) {
        writeFormStateToStorage(STORAGE, formState);
      } else {
        deleteFormStateFromStorage(STORAGE);
      }
    }, [formState, isSubmitted]),
    1000
  );

  const canNavigateToNextStep = useMemo(() => {
    switch (formState.step) {
      case OnboardingSurveyStep.start:
        return true;
      case OnboardingSurveyStep.step1:
        return formState.role != null;
      case OnboardingSurveyStep.step2:
        return formState.team_or_personal_account != null;
      case OnboardingSurveyStep.step3:
        switch (formState.team_or_personal_account) {
          case TeamOrPersonal.Personal:
            return true;
          case TeamOrPersonal.Team:
            return (
              formState.company_name != null && formState.company_size != null
            );
          default:
            return false;
        }
      case OnboardingSurveyStep.step4:
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
    let nextStep: OnboardingSurveyStep;
    switch (formState.step) {
      case OnboardingSurveyStep.start:
        nextStep = OnboardingSurveyStep.step1;
        break;
      case OnboardingSurveyStep.step1:
        nextStep = OnboardingSurveyStep.step2;
        capture("onboardingSurvey.set-role");
        break;
      case OnboardingSurveyStep.step2:
        capture("onboardingSurvey.set-team_or_personal_account");
        nextStep = OnboardingSurveyStep.step3;
        break;
      case OnboardingSurveyStep.step3:
        switch (form.state.team_or_personal_account) {
          case TeamOrPersonal.Team:
            capture("onboardingSurvey.set-company_details");
            break;
          case TeamOrPersonal.Personal:
            capture("onboardingSurvey.set-project_details");
            break;
          default:
            console.error(
              "unexpected team_or_personal_account",
              form.state.team_or_personal_account
            );
            break;
        }
        nextStep = OnboardingSurveyStep.step4;
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
  }, [canNavigateToNextStep, form, formState.step, capture]);

  const toPreviousStep = useCallback(() => {
    capture("onboardingSurvey.set-clicked-back");
    form.setState((prev) => {
      return produce(prev, (draft) => {
        switch (formState.step) {
          case OnboardingSurveyStep.step2:
            draft.step = OnboardingSurveyStep.step1;
            break;
          case OnboardingSurveyStep.step3:
            draft.step = OnboardingSurveyStep.step2;
            break;
          case OnboardingSurveyStep.step4:
            draft.step = OnboardingSurveyStep.step3;
            break;

          default:
            throw new Error("no previous step is available");
        }
        return draft;
      });
    });
  }, [capture, form, formState.step]);

  const canSave = useMemo(() => {
    if (formState.step !== OnboardingSurveyStep.step4) {
      return false;
    }
    return formState.use_cases !== undefined;
  }, [formState.step, formState.use_cases]);

  return useMemo(
    () => ({
      ...form,
      toNextStep: toNextStep,
      canNavigateToNextStep: canNavigateToNextStep,
      toPreviousStep: toPreviousStep,
      canSave,
    }),
    [canNavigateToNextStep, form, toNextStep, toPreviousStep, canSave]
  );
}
