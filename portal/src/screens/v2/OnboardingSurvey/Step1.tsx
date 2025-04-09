import React from "react";
import { Text } from "../../../components/onboarding/Text";
import { FormattedMessage } from "@oursky/react-messageformat";
import { PrimaryButton } from "../../../components/v2/PrimaryButton/PrimaryButton";
import { OnboardingSurveyFormModel } from "./form";
import { useFormContainerBaseContext } from "../../../FormContainerBase";
import { OnboardingSurveyStepper } from "../../../components/onboarding/OnboardingSurveyStepper";

export function Step1(): React.ReactElement {
  const { form } = useFormContainerBaseContext<OnboardingSurveyFormModel>();

  return (
    <div className="grid grid-cols-1 gap-16 text-center">
      <OnboardingSurveyStepper step={form.state.step} />
      <div className="grid grid-cols-1 gap-8">
        <Text.Heading>
          <FormattedMessage id="OnboardingSurveyScreen.step1.header" />
        </Text.Heading>
      </div>
      <div>
        <PrimaryButton
          size="4"
          highContrast={true}
          text={<FormattedMessage id="OnboardingSurveyScreen.actions.next" />}
          onClick={form.toNextStep}
        />
      </div>
    </div>
  );
}
