import React, { useCallback, useMemo } from "react";
import cn from "classnames";
import { Text } from "../../../components/onboarding/Text";
import { FormattedMessage } from "@oursky/react-messageformat";
import { PrimaryButton } from "../../../components/v2/PrimaryButton/PrimaryButton";
import { useFormContainerBaseContext } from "../../../FormContainerBase";
import { ProjectWizardStepper } from "../../../components/project-wizard/ProjectWizardStepper";
import { ProjectWizardFormModel, LoginMethod, AuthMethod } from "./form";
import { FormField } from "../../../components/v2/FormField/FormField";
import { ProjectWizardBackButton } from "../../../components/project-wizard/ProjectWizardBackButton";
import {
  ToggleGroup,
  ToggleGroupOption,
} from "../../../components/v2/ToggleGroup/ToggleGroup";
import loginEmailIcon from "../../../images/login_email.svg";
import loginPhoneIcon from "../../../images/login_phone.svg";
import loginUsernameIcon from "../../../images/login_username.svg";
import passwordlessIcon from "../../../images/passwordless_icon.svg";
import passwordIcon from "../../../images/password_icon.svg";
import { produce } from "immer";
import {
  IconRadioCardOption,
  MultiSelectIconRadioCards,
} from "../../../components/v2/IconRadioCards/IconRadioCards";

export function Step2(): React.ReactElement {
  const { form } = useFormContainerBaseContext<ProjectWizardFormModel>();
  const emailEnabled = useMemo(
    () => form.state.loginMethods.includes(LoginMethod.Email),
    [form.state.loginMethods]
  );
  const phoneEnabled = useMemo(
    () => form.state.loginMethods.includes(LoginMethod.Phone),
    [form.state.loginMethods]
  );
  const usernameEnabled = useMemo(
    () => form.state.loginMethods.includes(LoginMethod.Username),
    [form.state.loginMethods]
  );

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
          error={
            form.state.loginMethods.length === 0 ? (
              <FormattedMessage id="ProjectWizardScreen.errors.loginMethodRequired" />
            ) : null
          }
        >
          <div className="flex flex-col max-h-[316px] max-w-100">
            <ToggleGroup
              items={useMemo<ToggleGroupOption<LoginMethod>[]>(() => {
                return [
                  {
                    value: LoginMethod.Email,
                    text: (
                      <FormattedMessage id="ProjectWizardScreen.loginMethods.email" />
                    ),
                    icon: <img src={loginEmailIcon} width={20} height={20} />,
                  },
                  {
                    value: LoginMethod.Phone,
                    text: (
                      <FormattedMessage id="ProjectWizardScreen.loginMethods.phone" />
                    ),
                    icon: <img src={loginPhoneIcon} width={20} height={20} />,
                  },
                  {
                    value: LoginMethod.Username,
                    text: (
                      <FormattedMessage id="ProjectWizardScreen.loginMethods.username" />
                    ),
                    icon: (
                      <img src={loginUsernameIcon} width={20} height={20} />
                    ),
                  },
                ];
              }, [])}
              values={form.state.loginMethods}
              onValuesChange={useCallback(
                (newValues: LoginMethod[]) => {
                  form.setState((prev) =>
                    produce(prev, (draft) => {
                      draft.loginMethods = newValues;
                      return draft;
                    })
                  );
                },
                [form]
              )}
            />
          </div>
        </FormField>
        <div
          className={cn(
            usernameEnabled || phoneEnabled || emailEnabled ? null : "hidden"
          )}
        >
          <FormField
            size="3"
            label={
              <FormattedMessage
                id={"ProjectWizardScreen.step2.fields.authMethod.label"}
              />
            }
            error={
              form.effectiveAuthMethods.length === 0 ? (
                <FormattedMessage id="ProjectWizardScreen.errors.authMethodRequired" />
              ) : null
            }
          >
            <MultiSelectIconRadioCards
              size="2"
              itemMinWidth={286}
              options={useMemo<IconRadioCardOption<AuthMethod>[]>(() => {
                return [
                  {
                    value: AuthMethod.Passwordless,
                    icon: <img src={passwordlessIcon} width={40} height={40} />,
                    title: (
                      <FormattedMessage id="ProjectWizardScreen.authMethod.passwordless.title" />
                    ),
                    subtitle: (
                      <FormattedMessage id="ProjectWizardScreen.authMethod.passwordless.subtitle" />
                    ),
                    disabled: !emailEnabled && !phoneEnabled,
                    tooltip:
                      !emailEnabled && !phoneEnabled ? (
                        <FormattedMessage id="ProjectWizardScreen.authMethod.passwordless.tooltip" />
                      ) : null,
                  },
                  {
                    value: AuthMethod.Password,
                    icon: <img src={passwordIcon} width={40} height={40} />,
                    title: (
                      <FormattedMessage id="ProjectWizardScreen.authMethod.password.title" />
                    ),
                    subtitle: (
                      <FormattedMessage id="ProjectWizardScreen.authMethod.password.subtitle" />
                    ),
                    disabled: usernameEnabled,
                    tooltip: usernameEnabled ? (
                      <FormattedMessage id="ProjectWizardScreen.authMethod.password.tooltip" />
                    ) : null,
                  },
                ];
              }, [emailEnabled, phoneEnabled, usernameEnabled])}
              values={form.effectiveAuthMethods}
              onValuesChange={useCallback(
                (newValues: AuthMethod[]) => {
                  form.setState((prev) =>
                    produce(prev, (draft) => {
                      draft.authMethods = newValues;
                      return draft;
                    })
                  );
                },
                [form]
              )}
            />
          </FormField>
        </div>
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
