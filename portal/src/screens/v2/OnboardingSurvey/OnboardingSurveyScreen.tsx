import React from "react";
import {
  FormContainerBase,
  useFormContainerBaseContext,
} from "../../../FormContainerBase";
import { Start } from "./Start";
import {
  useOnboardingSurveyForm,
  Step,
  OnboardingSurveyFormModel,
} from "./form";
import { OnboardingSurveyLayout } from "../../../components/onboarding/OnboardingSurveyLayout";

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
    case Step.start:
      return <Start />;
    case Step.step1:
      return <></>;
    case Step.step2:
      return <></>;
    case Step.step3:
      return <></>;
    case Step.step4:
      return <></>;
  }
}

export default OnboardingSurveyScreen;
