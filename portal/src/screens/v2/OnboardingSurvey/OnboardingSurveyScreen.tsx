import React, { useCallback } from "react";
import {
  FormContainerBase,
  useFormContainerBaseContext,
} from "../../../FormContainerBase";
import { Start } from "./Start";
import {
  useOnboardingSurveyForm,
  OnboardingSurveyStep,
  OnboardingSurveyFormModel,
} from "./form";
import { OnboardingSurveyLayout } from "../../../components/onboarding/OnboardingSurveyLayout";
import { Step1 } from "./Step1";
import { Step2 } from "./Step2";
import { Step3 } from "./Step3";
import { Step4 } from "./Step4";
import { Completed } from "./Completed";

function OnboardingSurveyScreen(): React.ReactElement {
  const form = useOnboardingSurveyForm();

  const handleFormSubmit = useCallback(
    (e: React.FormEvent) => {
      e.preventDefault();

      switch (form.state.step) {
        case OnboardingSurveyStep.start:
        case OnboardingSurveyStep.step1:
        case OnboardingSurveyStep.step2:
        case OnboardingSurveyStep.step3:
          if (form.canNavigateToNextStep) {
            form.toNextStep();
          }
          break;
        case OnboardingSurveyStep.step4:
          if (form.canSave) {
            form.save();
          }
          break;
      }
    },
    [form]
  );

  return (
    <FormContainerBase form={form}>
      <OnboardingSurveyLayout>
        <form className="contents" onSubmit={handleFormSubmit}>
          <OnboardingSurveyScreenContent />
        </form>
      </OnboardingSurveyLayout>
    </FormContainerBase>
  );
}

function OnboardingSurveyScreenContent() {
  const { form } = useFormContainerBaseContext<OnboardingSurveyFormModel>();

  if (form.isSubmitted) {
    return <Completed />;
  }

  switch (form.state.step) {
    case OnboardingSurveyStep.start:
      return <Start />;
    case OnboardingSurveyStep.step1:
      return <Step1 />;
    case OnboardingSurveyStep.step2:
      return <Step2 />;
    case OnboardingSurveyStep.step3:
      return <Step3 />;
    case OnboardingSurveyStep.step4:
      return <Step4 />;
  }
}

export default OnboardingSurveyScreen;
