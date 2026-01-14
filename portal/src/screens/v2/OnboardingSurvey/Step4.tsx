import React, { useCallback, useContext, useMemo } from "react";
import cn from "classnames";
import { Text } from "../../../components/onboarding/Text";
import { Context as MessageContext, FormattedMessage } from "../../../intl";
import { OnboardingSurveyFormModel, UseCase } from "./form";
import { useFormContainerBaseContext } from "../../../FormContainerBase";
import { OnboardingSurveyStepper } from "../../../components/onboarding/OnboardingSurveyStepper";
import {
  MultiSelectRadioCards,
  RadioCardOption,
} from "../../../components/v2/RadioCards/RadioCards";
import { produce } from "immer";
import { OnboardingSurveyBackButton } from "../../../components/onboarding/OnboardingSurveyBackButton";
import { TextArea } from "../../../components/v2/TextArea/TextArea";
import { WhiteButton } from "../../../components/v2/Button/WhiteButton/WhiteButton";

export function Step4(): React.ReactElement {
  const { form } = useFormContainerBaseContext<OnboardingSurveyFormModel>();
  const { renderToString } = useContext(MessageContext);

  return (
    <div className="grid grid-cols-1 gap-16 text-center self-stretch">
      <OnboardingSurveyStepper step={form.state.step} />
      <div className="grid grid-cols-1 gap-8 max-w-full w-110 justify-self-center">
        <div className="grid grid-cols-1 gap-4">
          <Text.Heading>
            <FormattedMessage
              id="OnboardingSurveyScreen.step4.header"
              values={{
                Hint: Text.HeadingHint,
              }}
            />
          </Text.Heading>
        </div>
        <MultiSelectRadioCards
          highContrast={true}
          size="3"
          numberOfColumns={1}
          itemFillSpaces={true}
          values={form.state.use_cases ?? []}
          options={useMemo<RadioCardOption<UseCase>[]>(
            () => [
              {
                value: UseCase.BuildingNewSoftwareProject,
                title: (
                  <FormattedMessage id="OnboardingSurveyScreen.step4.usecases.newProject" />
                ),
              },
              {
                value: UseCase.SSOSolution,
                title: (
                  <FormattedMessage id="OnboardingSurveyScreen.step4.usecases.sso" />
                ),
              },
              {
                value: UseCase.EnhanceSecurity,
                title: (
                  <FormattedMessage id="OnboardingSurveyScreen.step4.usecases.security" />
                ),
              },
              {
                value: UseCase.Other,
                title: (
                  <FormattedMessage id="OnboardingSurveyScreen.step4.usecases.other" />
                ),
              },
            ],
            []
          )}
          onValuesChange={useCallback(
            (newValues: UseCase[]) => {
              form.setState((prev) =>
                produce(prev, (draft) => {
                  draft.use_cases = newValues;
                  return draft;
                })
              );
            },
            [form]
          )}
        />
        <div
          className={cn(
            form.state.use_cases?.includes(UseCase.Other) ? null : "hidden",
            "h-[6.375rem] flex flex-col [&>*]:flex-1"
          )}
        >
          <TextArea
            size="2"
            optional={true}
            label={
              <FormattedMessage id="OnboardingSurveyScreen.step4.usecases.other.label" />
            }
            placeholder={renderToString(
              "OnboardingSurveyScreen.step4.usecases.other.placeholder"
            )}
            value={form.state.use_case_other}
            onChange={useCallback(
              (e: React.ChangeEvent<HTMLTextAreaElement>) => {
                const value = e.currentTarget.value;
                form.setState((prev) =>
                  produce(prev, (draft) => {
                    draft.use_case_other = value;
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
          text={<FormattedMessage id="OnboardingSurveyScreen.actions.finish" />}
          onClick={form.save}
          loading={form.isUpdating}
          disabled={!form.canSave}
        />
      </div>
    </div>
  );
}
