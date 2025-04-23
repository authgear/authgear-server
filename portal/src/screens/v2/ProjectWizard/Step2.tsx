import React from "react";
import { Text } from "../../../components/onboarding/Text";
import { FormattedMessage } from "@oursky/react-messageformat";
import { PrimaryButton } from "../../../components/v2/PrimaryButton/PrimaryButton";
import { useFormContainerBaseContext } from "../../../FormContainerBase";
import { ProjectWizardStepper } from "../../../components/project-wizard/ProjectWizardStepper";
import { ProjectWizardFormModel } from "./form";
import { FormField } from "../../../components/v2/FormField/FormField";
import { ProjectWizardBackButton } from "../../../components/project-wizard/ProjectWizardBackButton";

export function Step2(): React.ReactElement {
  const { form } = useFormContainerBaseContext<ProjectWizardFormModel>();

  return (
    <div className="grid grid-cols-1 gap-12 text-left self-stretch">
      <ProjectWizardStepper step={form.state.step} />
      <div className="grid grid-cols-1 gap-6">
        <Text.Heading>
          <FormattedMessage
            id="ProjectWizardScreen.step2.header"
            values={{ projectName: form.state.projectName }}
          />
        </Text.Heading>
        <FormField
          size="3"
          label={
            <FormattedMessage id="ProjectWizardScreen.step2.fields.loginMethods.label" />
          }
        >
          {/* TODO */}
        </FormField>
        <FormField
          size="3"
          label={
            <FormattedMessage id="ProjectWizardScreen.step2.fields.isPasswordEnabled.label" />
          }
        >
          {/* TODO */}
        </FormField>
      </div>
      <div className="grid grid-flow-col grid-rows-1 gap-8 items-center justify-start">
        <ProjectWizardBackButton onClick={form.toPreviousStep} />
        <PrimaryButton
          type="submit"
          size="3"
          text={<FormattedMessage id="ProjectWizardScreen.actions.next" />}
          onClick={form.toNextStep}
          disabled={!form.canNavigateToNextStep}
        />
      </div>
    </div>
  );
}
