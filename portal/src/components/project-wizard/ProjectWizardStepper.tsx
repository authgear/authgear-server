import React, { useMemo } from "react";
import { ProjectWizardStep } from "../../screens/v2/ProjectWizard/form";
import { Stepper, Step } from "../v2/Stepper/Stepper";
import styles from "./ProjectWizardStepper.module.css";

function stepToNumber(step: ProjectWizardStep) {
  switch (step) {
    case ProjectWizardStep.step1:
      return 1;
    case ProjectWizardStep.step2:
      return 2;
    case ProjectWizardStep.step3:
      return 3;
  }
}

export function ProjectWizardStepper({
  step,
}: {
  step: ProjectWizardStep;
}): React.ReactElement {
  const steps = useMemo<Step[]>(() => {
    return [
      { text: "1", checked: stepToNumber(step) > 1 },
      { text: "2", checked: stepToNumber(step) > 2 },
      { text: "3", checked: stepToNumber(step) > 3 },
    ];
  }, [step]);
  return (
    <div className={styles.projectWizardStepper}>
      <Stepper steps={steps} />
    </div>
  );
}
