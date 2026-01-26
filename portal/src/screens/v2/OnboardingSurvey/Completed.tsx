import React, { useCallback } from "react";
import { EmojiIcon } from "../../../components/onboarding/EmojiIcon";
import { Text } from "../../../components/onboarding/Text";
import { FormattedMessage } from "../../../intl";
import { PrimaryButton } from "../../../components/v2/Button/PrimaryButton/PrimaryButton";
import { useNavigate } from "react-router-dom";
import { useFormContainerBaseContext } from "../../../FormContainerBase";
import { OnboardingSurveyFormModel } from "./form";

export function Completed(): React.ReactElement {
  const navigate = useNavigate();
  const { form } = useFormContainerBaseContext<OnboardingSurveyFormModel>();
  return (
    <div className="grid grid-cols-1 gap-16 text-center">
      <div className="grid grid-cols-1 gap-8 ">
        <EmojiIcon>✨</EmojiIcon>
        <div className="grid grid-cols-1 gap-4 max-w-[600px]">
          <Text.Heading>
            <FormattedMessage id="OnboardingSurveyScreen.completed.header" />
          </Text.Heading>
          <Text.Body>
            <FormattedMessage id="OnboardingSurveyScreen.completed.body" />
          </Text.Body>
        </div>
      </div>
      <div>
        <PrimaryButton
          size="4"
          highContrast={true}
          text={
            <FormattedMessage id="OnboardingSurveyScreen.actions.createProject" />
          }
          onClick={useCallback(() => {
            const companyName = form.state.company_name;
            if (companyName !== undefined)
              navigate("./../projects/create", {
                state: { company_name: companyName },
              });
            else {
              navigate("./../projects/create");
            }
          }, [form.state.company_name, navigate])}
        />
      </div>
    </div>
  );
}
