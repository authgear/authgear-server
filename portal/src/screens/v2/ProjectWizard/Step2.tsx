import React, { useCallback, useMemo } from "react";
import cn from "classnames";
import { Text } from "../../../components/project-wizard/Text";
import { FormattedMessage } from "../../../intl";
import { PrimaryButton } from "../../../components/v2/Button/PrimaryButton/PrimaryButton";
import { useFormContainerBaseContext } from "../../../FormContainerBase";
import { ProjectWizardStepper } from "../../../components/project-wizard/ProjectWizardStepper";
import { ProjectWizardFormModel, LoginMethod, AuthMethod } from "./form";
import { FormField } from "../../../components/v2/FormField/FormField";
import { ProjectWizardBackButton } from "../../../components/project-wizard/ProjectWizardBackButton";
import {
  ToggleGroup,
  ToggleGroupOption,
} from "../../../components/v2/ToggleGroup/ToggleGroup";
import { produce } from "immer";
import {
  IconRadioCardOption,
  MultiSelectIconRadioCards,
} from "../../../components/v2/IconRadioCards/IconRadioCards";
import { SquareIcon } from "../../../components/v2/SquareIcon/SquareIcon";
import { ButtonIcon, InputIcon } from "@radix-ui/react-icons";
import { useCapture } from "../../../gtm_v2";

function LoginMethodIcon({ method }: { method: LoginMethod }) {
  let iconClassName = "";
  switch (method) {
    case LoginMethod.Email:
      iconClassName = "fa fa-envelope";
      break;
    case LoginMethod.Phone:
      iconClassName = "fa fa-mobile";
      break;
    case LoginMethod.Username:
      iconClassName = "fa fa-user";
      break;
    case LoginMethod.Google:
      iconClassName = "fab fa-google";
      break;
    case LoginMethod.Apple:
      iconClassName = "fab fa-apple";
      break;
    case LoginMethod.Facebook:
      iconClassName = "fab fa-facebook";
      break;
    case LoginMethod.Github:
      iconClassName = "fab fa-github";
      break;
    case LoginMethod.LinkedIn:
      iconClassName = "fab fa-linkedin";
      break;
    case LoginMethod.MicrosoftEntraID:
      iconClassName = "fab fa-microsoft";
      break;
    case LoginMethod.MicrosoftADFS:
      iconClassName = "fab fa-microsoft";
      break;
    case LoginMethod.MicrosoftAzureADB2C:
      iconClassName = "fab fa-microsoft";
      break;
    case LoginMethod.WechatWeb:
      iconClassName = "fab fa-weixin";
      break;
    case LoginMethod.WechatMobile:
      iconClassName = "fab fa-weixin";
      break;
  }
  return (
    <i className={cn(iconClassName, "text-xl text-center h-[1em] w-[1em]")} />
  );
}

export function Step2(): React.ReactElement {
  const capture = useCapture();
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
                    icon: <LoginMethodIcon method={LoginMethod.Email} />,
                  },
                  {
                    value: LoginMethod.Phone,
                    text: (
                      <FormattedMessage id="ProjectWizardScreen.loginMethods.phone" />
                    ),
                    icon: <LoginMethodIcon method={LoginMethod.Phone} />,
                  },
                  {
                    value: LoginMethod.Username,
                    text: (
                      <FormattedMessage id="ProjectWizardScreen.loginMethods.username" />
                    ),
                    icon: <LoginMethodIcon method={LoginMethod.Username} />,
                  },
                  {
                    value: LoginMethod.Google,
                    text: (
                      <FormattedMessage id="ProjectWizardScreen.loginMethods.google" />
                    ),
                    icon: <LoginMethodIcon method={LoginMethod.Google} />,
                  },
                  {
                    value: LoginMethod.Apple,
                    text: (
                      <FormattedMessage id="ProjectWizardScreen.loginMethods.apple" />
                    ),
                    icon: <LoginMethodIcon method={LoginMethod.Apple} />,
                  },
                  {
                    value: LoginMethod.Facebook,
                    text: (
                      <FormattedMessage id="ProjectWizardScreen.loginMethods.facebook" />
                    ),
                    icon: <LoginMethodIcon method={LoginMethod.Facebook} />,
                  },
                  {
                    value: LoginMethod.Github,
                    text: (
                      <FormattedMessage id="ProjectWizardScreen.loginMethods.github" />
                    ),
                    icon: <LoginMethodIcon method={LoginMethod.Github} />,
                  },
                  {
                    value: LoginMethod.LinkedIn,
                    text: (
                      <FormattedMessage id="ProjectWizardScreen.loginMethods.linkedin" />
                    ),
                    icon: <LoginMethodIcon method={LoginMethod.LinkedIn} />,
                  },
                  {
                    value: LoginMethod.MicrosoftEntraID,
                    text: (
                      <FormattedMessage id="ProjectWizardScreen.loginMethods.microsoftEntraID" />
                    ),
                    icon: (
                      <LoginMethodIcon method={LoginMethod.MicrosoftEntraID} />
                    ),
                  },
                  {
                    value: LoginMethod.MicrosoftADFS,
                    text: (
                      <FormattedMessage id="ProjectWizardScreen.loginMethods.microsoftADFS" />
                    ),
                    icon: (
                      <LoginMethodIcon method={LoginMethod.MicrosoftADFS} />
                    ),
                  },
                  {
                    value: LoginMethod.MicrosoftAzureADB2C,
                    text: (
                      <FormattedMessage id="ProjectWizardScreen.loginMethods.microsoftAzureADB2C" />
                    ),
                    icon: (
                      <LoginMethodIcon
                        method={LoginMethod.MicrosoftAzureADB2C}
                      />
                    ),
                  },
                  {
                    value: LoginMethod.WechatWeb,
                    text: (
                      <FormattedMessage id="ProjectWizardScreen.loginMethods.wechatWeb" />
                    ),
                    supportingText: (
                      <FormattedMessage id="ProjectWizardScreen.loginMethods.wechatWeb.supportingText" />
                    ),
                    icon: <LoginMethodIcon method={LoginMethod.WechatWeb} />,
                  },
                  {
                    value: LoginMethod.WechatMobile,
                    text: (
                      <FormattedMessage id="ProjectWizardScreen.loginMethods.wechatMobile" />
                    ),
                    supportingText: (
                      <FormattedMessage id="ProjectWizardScreen.loginMethods.wechatMobile.supportingText" />
                    ),
                    icon: <LoginMethodIcon method={LoginMethod.WechatMobile} />,
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
              onToggle={useCallback(
                (checked: boolean, method: LoginMethod) => {
                  capture("projectWizard.clicked-auth", {
                    method: method,
                    checked: checked,
                  });
                },
                [capture]
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
                    icon: (
                      <SquareIcon
                        className="text-[var(--accent-9)]"
                        Icon={ButtonIcon}
                        size="7"
                        radius="4"
                        iconSize="1.375rem"
                      />
                    ),
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
                    icon: (
                      <SquareIcon
                        className="text-[var(--accent-9)]"
                        Icon={InputIcon}
                        size="7"
                        radius="4"
                        iconSize="1.375rem"
                      />
                    ),
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
          loading={form.isUpdating}
          onClick={form.save}
          disabled={!form.canSave}
        />
      </div>
    </div>
  );
}
