import React, { useCallback, useMemo } from "react";
import { Text } from "../../../components/onboarding/Text";
import { FormattedMessage } from "@oursky/react-messageformat";
import { PrimaryButton } from "../../../components/v2/PrimaryButton/PrimaryButton";
import { CompanySize, OnboardingSurveyFormModel, TeamOrPersonal } from "./form";
import { useFormContainerBaseContext } from "../../../FormContainerBase";
import { OnboardingSurveyStepper } from "../../../components/onboarding/OnboardingSurveyStepper";
import {
  RadioCardOption,
  RadioCards,
} from "../../../components/v2/RadioCards/RadioCards";
import { produce } from "immer";
import { BackButton } from "../../../components/onboarding/BackButton";
import { TextField } from "../../../components/v2/TextField/TextField";
import { FormField } from "../../../components/v2/FormField/FormField";

export function Step3(): React.ReactElement {
  const { form } = useFormContainerBaseContext<OnboardingSurveyFormModel>();

  return (
    <div className="grid grid-cols-1 gap-16 text-center self-stretch">
      <OnboardingSurveyStepper step={form.state.step} />
      {form.state.team_or_personal_account === TeamOrPersonal.Personal ? (
        <Step3PersonalForm />
      ) : (
        <Step3TeamForm />
      )}
      <div className="flex items-center justify-center gap-8">
        <BackButton onClick={form.toPreviousStep} />
        <PrimaryButton
          size="4"
          highContrast={true}
          text={<FormattedMessage id="OnboardingSurveyScreen.actions.next" />}
          onClick={form.toNextStep}
          disabled={!form.canNavigateToNextStep}
        />
      </div>
    </div>
  );
}

function Step3PersonalForm() {
  const { form } = useFormContainerBaseContext<OnboardingSurveyFormModel>();
  return (
    <div className="grid grid-cols-1 gap-8">
      <div className="grid grid-cols-1 gap-4">
        <Text.Heading>
          <FormattedMessage id="OnboardingSurveyScreen.step3.personal.header" />
        </Text.Heading>
      </div>
      <div className="grid grid-cols-1 gap-8 max-w-full w-110 justify-self-center">
        <TextField
          size="3"
          optional={true}
          label={
            <FormattedMessage id="OnboardingSurveyScreen.step3.personal.fields.projectWebsite" />
          }
          onChange={useCallback(
            (e: React.ChangeEvent<HTMLInputElement>) => {
              const value = e.currentTarget.value;
              form.setState((prev) =>
                produce(prev, (draft) => {
                  draft.project_website = value;
                  return draft;
                })
              );
            },
            [form]
          )}
        />
        {/* TODO(tung): Phone number field */}
      </div>
    </div>
  );
}

function Step3TeamForm() {
  const { form } = useFormContainerBaseContext<OnboardingSurveyFormModel>();
  return (
    <div className="grid grid-cols-1 gap-8">
      <div className="grid grid-cols-1 gap-4">
        <Text.Heading>
          <FormattedMessage id="OnboardingSurveyScreen.step3.team.header" />
        </Text.Heading>
      </div>
      <div className="grid grid-cols-1 gap-8 max-w-full w-110 justify-self-center">
        <TextField
          size="3"
          label={
            <FormattedMessage id="OnboardingSurveyScreen.step3.team.fields.companyName" />
          }
          value={form.state.company_name ?? ""}
          onChange={useCallback(
            (e: React.ChangeEvent<HTMLInputElement>) => {
              const value = e.currentTarget.value;
              form.setState((prev) =>
                produce(prev, (draft) => {
                  draft.company_name = value;
                  return draft;
                })
              );
            },
            [form]
          )}
        />
        <FormField
          size="3"
          label={
            <FormattedMessage id="OnboardingSurveyScreen.step3.team.fields.companySize" />
          }
        >
          <RadioCards
            highContrast={true}
            size="1"
            value={form.state.company_size ?? null}
            itemMinWidth={90}
            itemFillSpaces={true}
            options={useMemo<RadioCardOption<CompanySize>[]>(
              () => [
                {
                  value: CompanySize["1-to-49"],
                  title: (
                    <FormattedMessage id="OnboardingSurveyScreen.step3.team.size.small.title" />
                  ),
                  subtitle: (
                    <FormattedMessage id="OnboardingSurveyScreen.step3.team.size.small.subtitle" />
                  ),
                },
                {
                  value: CompanySize["50-to-199"],
                  title: (
                    <FormattedMessage id="OnboardingSurveyScreen.step3.team.size.growing.title" />
                  ),
                  subtitle: (
                    <FormattedMessage id="OnboardingSurveyScreen.step3.team.size.growing.subtitle" />
                  ),
                },
                {
                  value: CompanySize["200-to-999"],
                  title: (
                    <FormattedMessage id="OnboardingSurveyScreen.step3.team.size.midsize.title" />
                  ),
                  subtitle: (
                    <FormattedMessage id="OnboardingSurveyScreen.step3.team.size.midsize.subtitle" />
                  ),
                },
                {
                  value: CompanySize["1000+"],
                  title: (
                    <FormattedMessage id="OnboardingSurveyScreen.step3.team.size.enterprise.title" />
                  ),
                  subtitle: (
                    <FormattedMessage id="OnboardingSurveyScreen.step3.team.size.enterprise.subtitle" />
                  ),
                },
              ],
              []
            )}
            onValueChange={useCallback(
              (newValue: CompanySize) => {
                form.setState((prev) =>
                  produce(prev, (draft) => {
                    draft.company_size = newValue;
                    return draft;
                  })
                );
              },
              [form]
            )}
          />
        </FormField>
        {/* TODO(tung): Phone number field */}
      </div>
    </div>
  );
}
