import React, { useCallback, useMemo } from "react";
import { Text } from "../../../components/onboarding/Text";
import { FormattedMessage } from "../../../intl";
import { CompanySize, OnboardingSurveyFormModel, TeamOrPersonal } from "./form";
import { useFormContainerBaseContext } from "../../../FormContainerBase";
import { OnboardingSurveyStepper } from "../../../components/onboarding/OnboardingSurveyStepper";
import {
  RadioCardOption,
  RadioCards,
} from "../../../components/v2/RadioCards/RadioCards";
import { produce } from "immer";
import { OnboardingSurveyBackButton } from "../../../components/onboarding/OnboardingSurveyBackButton";
import { TextField } from "../../../components/v2/TextField/TextField";
import { FormField } from "../../../components/v2/FormField/FormField";
import { OnboardingSurveyPhoneInput } from "../../../components/onboarding/OnboardingSurveyPhoneInput";
import { useViewerQuery } from "../../../graphql/portal/query/viewerQuery";
import ShowLoading from "../../../ShowLoading";
import { PhoneTextFieldValues } from "../../../PhoneTextField";
import { WhiteButton } from "../../../components/v2/Button/WhiteButton/WhiteButton";

export function Step3(): React.ReactElement {
  const { form } = useFormContainerBaseContext<OnboardingSurveyFormModel>();
  const { viewer, loading: viewerLoading } = useViewerQuery({
    fetchPolicy: "cache-first",
  });

  if (viewerLoading || viewer === undefined) {
    return <ShowLoading />;
  }

  return (
    <div className="grid grid-cols-1 gap-16 text-center self-stretch">
      <OnboardingSurveyStepper step={form.state.step} />
      {form.state.team_or_personal_account === TeamOrPersonal.Personal ? (
        <Step3PersonalForm
          geoIPCountryCode={viewer?.geoIPCountryCode ?? undefined}
        />
      ) : (
        <Step3TeamForm
          geoIPCountryCode={viewer?.geoIPCountryCode ?? undefined}
        />
      )}
      <div className="flex items-center justify-center gap-8">
        <OnboardingSurveyBackButton onClick={form.toPreviousStep} />
        <WhiteButton
          type="submit"
          size="4"
          text={<FormattedMessage id="OnboardingSurveyScreen.actions.next" />}
          onClick={form.toNextStep}
          disabled={!form.canNavigateToNextStep}
        />
      </div>
    </div>
  );
}

interface FormProps {
  geoIPCountryCode: string | undefined;
}

function Step3PersonalForm({ geoIPCountryCode }: FormProps) {
  const { form } = useFormContainerBaseContext<OnboardingSurveyFormModel>();
  const [phoneNumber, onPhoneNumberChange] = usePhoneNumberState();

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
          value={form.state.project_website}
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
        <FormField
          size="3"
          label={
            <FormattedMessage id="OnboardingSurveyScreen.step3.personal.fields.phone" />
          }
          optional={true}
        >
          <OnboardingSurveyPhoneInput
            initialInputValue={phoneNumber}
            onChange={onPhoneNumberChange}
            initialCountry={geoIPCountryCode}
          />
        </FormField>
      </div>
    </div>
  );
}

function Step3TeamForm({ geoIPCountryCode }: FormProps) {
  const { form } = useFormContainerBaseContext<OnboardingSurveyFormModel>();
  const [phoneNumber, onPhoneNumberChange] = usePhoneNumberState();
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
        <FormField
          size="3"
          label={
            <FormattedMessage id="OnboardingSurveyScreen.step3.team.fields.phone" />
          }
          optional={true}
        >
          <OnboardingSurveyPhoneInput
            initialInputValue={phoneNumber}
            onChange={onPhoneNumberChange}
            initialCountry={geoIPCountryCode}
          />
        </FormField>
      </div>
    </div>
  );
}

function usePhoneNumberState(): [string, (v: PhoneTextFieldValues) => void] {
  const { form } = useFormContainerBaseContext<OnboardingSurveyFormModel>();
  const onChange = useCallback(
    (v: PhoneTextFieldValues) => {
      form.setState((prev) =>
        produce(prev, (draft) => {
          if (v.e164) {
            draft.phone_number = v.e164;
          } else {
            draft.phone_number = v.partialValue;
          }
          return draft;
        })
      );
    },
    [form]
  );

  return [form.state.phone_number ?? "", onChange];
}
