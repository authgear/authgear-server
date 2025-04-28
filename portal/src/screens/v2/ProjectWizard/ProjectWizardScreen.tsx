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
import { Step1 } from "./Step1";
import { Step2 } from "./Step2";
import { Step3 } from "./Step3";

function ProjectWizardScreen(): React.ReactElement {
  const form = useProjectWizardForm();

  const handleFormSubmit = useCallback(
    (e: React.FormEvent) => {
      e.preventDefault();

      if (form.canSave) {
        form.save();
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
      return <Step1 />;
    case ProjectWizardStep.step2:
      return <Step2 />;
    case ProjectWizardStep.step3:
      return <Step3 />;
  }
}

export default ProjectWizardScreen;
