import React, { useCallback, useContext } from "react";
import { Text } from "../../../components/onboarding/Text";
import {
  Context as MessageContext,
  FormattedMessage,
} from "@oursky/react-messageformat";
import { PrimaryButton } from "../../../components/v2/PrimaryButton/PrimaryButton";
import { useFormContainerBaseContext } from "../../../FormContainerBase";
import { ProjectWizardStepper } from "../../../components/project-wizard/ProjectWizardStepper";
import { ProjectWizardFormModel } from "./form";
import { TextField } from "../../../components/v2/TextField/TextField";
import { produce } from "immer";
import { useSystemConfig } from "../../../context/SystemConfigContext";

export function Step1(): React.ReactElement {
  const { form } = useFormContainerBaseContext<ProjectWizardFormModel>();
  const systemConfig = useSystemConfig();
  const { renderToString } = useContext(MessageContext);

  return (
    <div className="grid grid-cols-1 gap-12 text-left self-stretch">
      <ProjectWizardStepper step={form.state.step} />
      <div className="grid grid-cols-1 gap-6">
        <Text.Heading>
          <FormattedMessage id="ProjectWizardScreen.step1.header" />
        </Text.Heading>
        <TextField
          size="3"
          label={
            <FormattedMessage id="ProjectWizardScreen.step1.fields.projectName.label" />
          }
          placeholder={renderToString(
            "ProjectWizardScreen.step1.fields.projectName.placeholder"
          )}
          value={form.state.projectName}
          onChange={useCallback(
            (e: React.ChangeEvent<HTMLInputElement>) => {
              const value = e.currentTarget.value;
              form.setState((prev) =>
                produce(prev, (draft) => {
                  draft.projectName = value;
                  return draft;
                })
              );
            },
            [form]
          )}
        />
        <TextField
          size="3"
          label={
            <FormattedMessage id="ProjectWizardScreen.step1.fields.projectID.label" />
          }
          hint={
            <FormattedMessage id="ProjectWizardScreen.step1.fields.projectID.hint" />
          }
          placeholder={renderToString(
            "ProjectWizardScreen.step1.fields.projectID.placeholder"
          )}
          suffix={systemConfig.appHostSuffix}
          value={form.state.projectID}
          onChange={useCallback(
            (e: React.ChangeEvent<HTMLInputElement>) => {
              const value = e.currentTarget.value;
              form.setState((prev) =>
                produce(prev, (draft) => {
                  draft.projectID = value;
                  return draft;
                })
              );
            },
            [form]
          )}
        />
      </div>
      <div>
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
