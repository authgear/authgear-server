import { z } from "zod";
import { SimpleFormModel, useSimpleForm } from "../../../hook/useSimpleForm";
import { useCallback, useMemo, useState } from "react";
import { useDebouncedEffect } from "../../../hook/useDebouncedEffect";
import { produce } from "immer";

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

export enum UseCases {
  BuildingNewSoftwareProject = "BuildingNewSoftwareProject",
  SSOSolution = "SSOSolution",
  EnhanceSecurity = "EnhanceSecurity",
  Other = "Other",
}

export enum Step {
  "start" = "start",
  "step1" = "step1",
  "step2" = "step2",
  "step3" = "step3",
  "step4" = "step4",
}

const zFormState = z.object({
  step: z.nativeEnum(Step),

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
  use_cases: z.nativeEnum(UseCases).array().optional(),
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
  toNextStep: () => void;
}

export function useOnboardingSurveyForm(): OnboardingSurveyFormModel {
  const [defaultState] = useState<FormState>(() => {
    return (
      readFormStateFromStorage(STORAGE) ?? {
        step: Step.start,
      }
    );
  });

  const submit = useCallback(async () => {
    // TODO
    deleteFormStateFromStorage(STORAGE);
  }, []);

  const form = useSimpleForm<FormState>({
    stateMode:
      "ConstantInitialStateAndResetCurrentStatetoInitialStateAfterSave",
    defaultState,
    submit,
  });

  const formState = form.state;

  useDebouncedEffect(
    useCallback(() => {
      writeFormStateToStorage(STORAGE, formState);
    }, [formState]),
    1000
  );

  const toNextStep = useCallback(() => {
    switch (formState.step) {
      case Step.start:
        form.setState((prev) => {
          return produce(prev, (draft) => {
            draft.step = Step.step1;
            return draft;
          });
        });

        break;

      default:
        throw new Error("TODO");
    }
  }, [form, formState]);

  return useMemo(
    () => ({
      ...form,
      toNextStep: toNextStep,
    }),
    [form, toNextStep]
  );
}
