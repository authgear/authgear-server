import React from "react";
import { Text } from "../../../components/onboarding/Text";
import { EmojiIcon } from "../../../components/onboarding/EmojiIcon";
import { FormattedMessage } from "@oursky/react-messageformat";
import { PrimaryButton } from "../../../components/v2/PrimaryButton/PrimaryButton";
import { OnboardingSurveyFormModel } from "./form";
import { useFormContainerBaseContext } from "../../../FormContainerBase";

export function Start(): React.ReactElement {
  const { form } = useFormContainerBaseContext<OnboardingSurveyFormModel>();

  return (
    <div className="grid grid-cols-1 gap-16 text-center">
      <div className="grid grid-cols-1 gap-8 ">
        <EmojiIcon>🪄</EmojiIcon>
        <div className="grid grid-cols-1 gap-4 ">
          <Text.Heading>
            <FormattedMessage id="OnboardingSurveyScreen.start.header" />
          </Text.Heading>
          <Text.Body>
            <FormattedMessage id="OnboardingSurveyScreen.start.body" />
          </Text.Body>
        </div>
      </div>
      <div>
        <PrimaryButton
          size="4"
          highContrast={true}
          text={<FormattedMessage id="OnboardingSurveyScreen.start.start" />}
          onClick={form.toNextStep}
        />
      </div>
    </div>
  );
}
