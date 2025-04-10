import React from "react";
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

function OnboardingSurveyScreen(): React.ReactElement {
  const form = useOnboardingSurveyForm();

  return (
    <FormContainerBase form={form}>
      <OnboardingSurveyLayout>
        <OnboardingSurveyScreenContent />
      </OnboardingSurveyLayout>
    </FormContainerBase>
  );
}

function OnboardingSurveyScreenContent() {
  const { form } = useFormContainerBaseContext<OnboardingSurveyFormModel>();

  switch (form.state.step) {
    case OnboardingSurveyStep.start:
      return <Start />;
    case OnboardingSurveyStep.step1:
      return <Step1 />;
    case OnboardingSurveyStep.step2:
      return <Step2 />;
    case OnboardingSurveyStep.step3:
      return <></>;
    case OnboardingSurveyStep.step4:
      return <></>;
  }
}

export default OnboardingSurveyScreen;
