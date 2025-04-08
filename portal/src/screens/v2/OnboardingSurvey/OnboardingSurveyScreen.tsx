import React from "react";
import { FormattedMessage } from "@oursky/react-messageformat";
import { OnboardingSurveyLayout } from "../../../components/onboarding/OnboardingSurveyLayout";
import { EmojiIcon } from "../../../components/onboarding/EmojiIcon";
import { Text } from "../../../components/onboarding/Text";
import { PrimaryButton } from "../../../components/v2/PrimaryButton/PrimaryButton";

function OnboardingSurveyScreen(): React.ReactElement {
  return (
    <OnboardingSurveyLayout>
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
          />
        </div>
      </div>
    </OnboardingSurveyLayout>
  );
}

export default OnboardingSurveyScreen;
