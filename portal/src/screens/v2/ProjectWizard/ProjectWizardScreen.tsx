import React, { useCallback } from "react";
import {
  FormContainerBase,
  useFormContainerBaseContext,
} from "../../../FormContainerBase";
import {
  ProjectWizardFormModel,
  ProjectWizardStep,
  useProjectWizardForm,
} from "./form";
import { ProjectWizardLayout } from "../../../components/project-wizard/ProjectWizardLayout";

function ProjectWizardScreen(): React.ReactElement {
  const form = useProjectWizardForm();

  const handleFormSubmit = useCallback(
    (e: React.FormEvent) => {
      e.preventDefault();

      switch (form.state.step) {
        case ProjectWizardStep.step1:
        case ProjectWizardStep.step2:
          if (form.canNavigateToNextStep) {
            form.toNextStep();
          }
          break;
        case ProjectWizardStep.step3:
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
      <ProjectWizardLayout>
        <form className="contents" onSubmit={handleFormSubmit}>
          <ProjectWizardScreenContent />
        </form>
      </ProjectWizardLayout>
    </FormContainerBase>
  );
}

function ProjectWizardScreenContent() {
  const { form } = useFormContainerBaseContext<ProjectWizardFormModel>();
  switch (form.state.step) {
    case ProjectWizardStep.step1:
      return <></>;
    case ProjectWizardStep.step2:
      return <></>;
    case ProjectWizardStep.step3:
      return <></>;
  }
}

export default ProjectWizardScreen;
