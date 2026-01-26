import React, { useCallback, useMemo } from "react";
import { Text } from "../../../components/onboarding/Text";
import { FormattedMessage } from "../../../intl";
import { OnboardingSurveyFormModel, TeamOrPersonal } from "./form";
import { useFormContainerBaseContext } from "../../../FormContainerBase";
import { OnboardingSurveyStepper } from "../../../components/onboarding/OnboardingSurveyStepper";
import {
  RadioCardOption,
  RadioCards,
} from "../../../components/v2/RadioCards/RadioCards";
import { produce } from "immer";
import { WhiteButton } from "../../../components/v2/Button/WhiteButton/WhiteButton";
import { OnboardingSurveyBackButton } from "../../../components/onboarding/OnboardingSurveyBackButton";

export function Step2(): React.ReactElement {
  const { form } = useFormContainerBaseContext<OnboardingSurveyFormModel>();

  return (
    <div className="grid grid-cols-1 gap-16 text-center self-stretch">
      <OnboardingSurveyStepper step={form.state.step} />
      <div className="grid grid-cols-1 gap-8">
        <div className="grid grid-cols-1 gap-4">
          <Text.Heading>
            <FormattedMessage id="OnboardingSurveyScreen.step2.header" />
          </Text.Heading>
          <Text.Body>
            <FormattedMessage id="OnboardingSurveyScreen.step2.body" />
          </Text.Body>
        </div>
        <div>
          <RadioCards
            highContrast={true}
            size="3"
            value={form.state.team_or_personal_account ?? null}
            options={useMemo<RadioCardOption<TeamOrPersonal>[]>(
              () => [
                {
                  value: TeamOrPersonal.Team,
                  title: (
                    <FormattedMessage id="OnboardingSurveyScreen.step2.options.team" />
                  ),
                },
                {
                  value: TeamOrPersonal.Personal,
                  title: (
                    <FormattedMessage id="OnboardingSurveyScreen.step2.options.personal" />
                  ),
                },
              ],
              []
            )}
            onValueChange={useCallback(
              (newValue: TeamOrPersonal) => {
                form.setState((prev) =>
                  produce(prev, (draft) => {
                    draft.team_or_personal_account = newValue;
                    return draft;
                  })
                );
              },
              [form]
            )}
          />
        </div>
      </div>
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
