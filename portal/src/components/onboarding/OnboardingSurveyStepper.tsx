import React, { useMemo } from "react";
import { OnboardingSurveyStep } from "../../screens/v2/OnboardingSurvey/form";
import { Stepper, Step } from "../v2/Stepper/Stepper";
import styles from "./OnboardingSurveyStepper.module.css";

function stepToNumber(step: OnboardingSurveyStep) {
  switch (step) {
    case OnboardingSurveyStep.start:
      return 0;
    case OnboardingSurveyStep.step1:
      return 1;
    case OnboardingSurveyStep.step2:
      return 2;
    case OnboardingSurveyStep.step3:
      return 3;
    case OnboardingSurveyStep.step4:
      return 4;
  }
}

export function OnboardingSurveyStepper({
  step,
}: {
  step: OnboardingSurveyStep;
}): React.ReactElement {
  const steps = useMemo<Step[]>(() => {
    return [
      { text: "1", checked: stepToNumber(step) > 1 },
      { text: "2", checked: stepToNumber(step) > 2 },
      { text: "3", checked: stepToNumber(step) > 3 },
      { text: "4", checked: false },
    ];
  }, [step]);
  return (
    <div className={styles.onboardingSurveyStepper}>
      <Stepper steps={steps} />
    </div>
  );
}
