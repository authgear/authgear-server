import React, { useCallback, useMemo } from "react";
import { Text } from "../../../components/onboarding/Text";
import { FormattedMessage } from "../../../intl";
import { OnboardingSurveyFormModel, Role } from "./form";
import { useFormContainerBaseContext } from "../../../FormContainerBase";
import { OnboardingSurveyStepper } from "../../../components/onboarding/OnboardingSurveyStepper";
import {
  RadioCardOption,
  RadioCards,
} from "../../../components/v2/RadioCards/RadioCards";
import { produce } from "immer";
import { WhiteButton } from "../../../components/v2/Button/WhiteButton/WhiteButton";

export function Step1(): React.ReactElement {
  const { form } = useFormContainerBaseContext<OnboardingSurveyFormModel>();

  return (
    <div className="grid grid-cols-1 gap-16 text-center self-stretch">
      <OnboardingSurveyStepper step={form.state.step} />
      <div className="grid grid-cols-1 gap-8">
        <Text.Heading>
          <FormattedMessage id="OnboardingSurveyScreen.step1.header" />
        </Text.Heading>
        <div>
          <RadioCards
            highContrast={true}
            size="3"
            value={form.state.role ?? null}
            options={useMemo<RadioCardOption<Role>[]>(
              () => [
                {
                  value: Role.Developer,
                  title: (
                    <FormattedMessage id="OnboardingSurveyScreen.step1.roles.developer" />
                  ),
                },
                {
                  value: Role.ProjectManager,
                  title: (
                    <FormattedMessage id="OnboardingSurveyScreen.step1.roles.pm" />
                  ),
                },
                {
                  value: Role.Business,
                  title: (
                    <FormattedMessage id="OnboardingSurveyScreen.step1.roles.business" />
                  ),
                },
                {
                  value: Role.Other,
                  title: (
                    <FormattedMessage id="OnboardingSurveyScreen.step1.roles.other" />
                  ),
                },
              ],
              []
            )}
            onValueChange={useCallback(
              (newRole: Role) => {
                form.setState((prev) =>
                  produce(prev, (draft) => {
                    draft.role = newRole;
                    return draft;
                  })
                );
              },
              [form]
            )}
          />
        </div>
      </div>
      <div>
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
