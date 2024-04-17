import React, { useCallback, useMemo, useContext } from "react";
import {
  Routes,
  Navigate,
  Route,
  useParams,
  useNavigate,
} from "react-router-dom";
import {
  ChoiceGroup,
  IChoiceGroupOption,
  TooltipHost,
  Text,
} from "@fluentui/react";
import { FormattedMessage, Context } from "@oursky/react-messageformat";
import { produce } from "immer";
import { FormProvider } from "../../form";
import { FormErrorMessageBar } from "../../FormErrorMessageBar";
import WizardContentLayout, {
  WizardDescription,
  WizardTitle,
} from "../../WizardContentLayout";
import WizardScreenLayout from "../../WizardScreenLayout";
import {
  useAppConfigForm,
  AppConfigFormModel,
} from "../../hook/useAppConfigForm";
import {
  LoginIDKeyConfig,
  PortalAPIAppConfig,
  PrimaryAuthenticatorType,
  SecondaryAuthenticationMode,
  SecondaryAuthenticatorType,
} from "../../types";
import PrimaryButton from "../../PrimaryButton";
import { TooltipIcon, useTooltipTargetElement } from "../../Tooltip";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import styles from "./ProjectWizardScreen.module.css";
import { useCapture } from "../../gtm_v2";
import HorizontalDivider from "../../HorizontalDivider";

const TOOLTIP_HOST_STYLES = {
  root: {
    display: "inline-block",
    position: "relative",
  },
} as const;

type Step1Answer =
  | "email_password"
  | "phone_password"
  | "oob_otp_email"
  | "oob_otp_sms";

type Step2Answer = boolean;

type Step3AuthenticatorType = SecondaryAuthenticatorType;

interface Step3Answer {
  useRecommenededSettings: boolean;
  secondaryAuthenticationMode: SecondaryAuthenticationMode;
  secondaryAuthenticatorType: Step3AuthenticatorType;
}

interface FormState {
  step1Answer: Step1Answer;
  step2Answer: Step2Answer;
  step3Answer: Step3Answer;
}

interface StepProps {
  appID: string;
  rawAppID: string;
  form: AppConfigFormModel<FormState>;
  saveAndThenNavigate: () => void;
}

function constructFormState(_config: PortalAPIAppConfig): FormState {
  return {
    step1Answer: "email_password",
    step2Answer: true,
    step3Answer: {
      useRecommenededSettings: true,
      secondaryAuthenticationMode: "if_exists",
      secondaryAuthenticatorType: "totp",
    },
  };
}

function derivePrimaryAuthenticatorsFromStep1Answer(
  step1Answer: Step1Answer
): PrimaryAuthenticatorType[] {
  switch (step1Answer) {
    case "oob_otp_email":
      return ["oob_otp_email"];
    case "oob_otp_sms":
      return ["oob_otp_sms"];
    case "email_password":
    // fallthrough
    case "phone_password":
      return ["password"];
    default:
      console.error("Unexpected step1Answer", step1Answer);
      break;
  }
  return [];
}

function deriveLoginIDKeysFromStep1Answer(
  step1Answer: Step1Answer
): LoginIDKeyConfig[] {
  switch (step1Answer) {
    case "email_password":
    // fallthrough
    case "oob_otp_email":
      return [{ type: "email" }];
    case "phone_password":
    // fallthrough
    case "oob_otp_sms":
      return [{ type: "phone" }];
    default:
      console.error("Unexpected step1Answer", step1Answer);
      break;
  }
  return [];
}

function constructConfig(
  config: PortalAPIAppConfig,
  _initialState: FormState,
  currentState: FormState
): PortalAPIAppConfig {
  return produce(config, (config) => {
    config.identity ??= {};
    config.identity.login_id ??= {};
    config.identity.login_id.keys = deriveLoginIDKeysFromStep1Answer(
      currentState.step1Answer
    );
    config.authentication ??= {};
    config.authentication.identities = ["oauth", "login_id"];

    config.authentication.primary_authenticators =
      derivePrimaryAuthenticatorsFromStep1Answer(currentState.step1Answer);

    if (currentState.step2Answer) {
      config.authentication.identities.push("passkey");
      config.authentication.primary_authenticators.push("passkey");
    }

    if (currentState.step3Answer.useRecommenededSettings) {
      config.authentication.secondary_authentication_mode = "if_exists";
      config.authentication.secondary_authenticators = ["totp"];
    } else {
      config.authentication.secondary_authentication_mode =
        currentState.step3Answer.secondaryAuthenticationMode;
      config.authentication.secondary_authenticators = [
        currentState.step3Answer.secondaryAuthenticatorType,
      ];
    }
  });
}

interface Step1Option extends IChoiceGroupOption {
  key: Step1Answer;
}

interface Step3Question3Option extends IChoiceGroupOption {
  key: Step3AuthenticatorType;
}

function Step1(props: StepProps) {
  const {
    appID,
    rawAppID,
    form: {
      state: { step1Answer },
      setState,
    },
  } = props;
  const navigate = useNavigate();
  const { renderToString } = useContext(Context);

  const capture = useCapture();
  const onClickNext = useCallback(
    (e) => {
      e.preventDefault();
      e.stopPropagation();

      capture("projectWizard.set-primary_auth");

      navigate("./../2");
    },
    [navigate, capture]
  );

  const oobOTPEmailTooltip = useTooltipTargetElement();
  const oobOTPEmailTooltipProps = useMemo(() => {
    return {
      targetElement: oobOTPEmailTooltip.targetElement,
    };
  }, [oobOTPEmailTooltip.targetElement]);

  const oobOTPSMSTooltip = useTooltipTargetElement();
  const oobOTPSMSTooltipProps = useMemo(() => {
    return {
      targetElement: oobOTPSMSTooltip.targetElement,
    };
  }, [oobOTPSMSTooltip.targetElement]);

  const options: Step1Option[] = useMemo(() => {
    return [
      {
        key: "email_password",
        text: renderToString("ProjectWizardScreen.step1.option.email-password"),
      },
      {
        key: "phone_password",
        text: renderToString(
          "ProjectWizardScreen.step1.option.phone-number-password"
        ),
      },
      {
        key: "oob_otp_email",
        text: renderToString("ProjectWizardScreen.step1.option.oob-otp-email"),
        // eslint-disable-next-line react/no-unstable-nested-components
        onRenderField: (props, render) => {
          return (
            <>
              {render!(props)}
              <TooltipHost
                styles={TOOLTIP_HOST_STYLES}
                tooltipProps={oobOTPEmailTooltipProps}
                content={renderToString(
                  "ProjectWizardScreen.step1.tooltip.oob-otp-email"
                )}
              >
                <TooltipIcon
                  id={oobOTPEmailTooltip.id}
                  setRef={oobOTPEmailTooltip.setRef}
                />
              </TooltipHost>
            </>
          );
        },
      },
      {
        key: "oob_otp_sms",
        text: renderToString("ProjectWizardScreen.step1.option.oob-otp-sms"),
        // eslint-disable-next-line react/no-unstable-nested-components
        onRenderField: (props, render) => {
          return (
            <>
              {render!(props)}
              <TooltipHost
                styles={TOOLTIP_HOST_STYLES}
                tooltipProps={oobOTPSMSTooltipProps}
                content={renderToString(
                  "ProjectWizardScreen.step1.tooltip.oob-otp-sms"
                )}
              >
                <TooltipIcon
                  id={oobOTPSMSTooltip.id}
                  setRef={oobOTPSMSTooltip.setRef}
                />
              </TooltipHost>
            </>
          );
        },
      },
    ];
  }, [
    renderToString,
    oobOTPEmailTooltipProps,
    oobOTPEmailTooltip.id,
    oobOTPEmailTooltip.setRef,
    oobOTPSMSTooltipProps,
    oobOTPSMSTooltip.id,
    oobOTPSMSTooltip.setRef,
  ]);

  const onChange = useCallback(
    (_e, option: IChoiceGroupOption | undefined) => {
      if (option != null) {
        capture("projectWizard.clicked-option", {
          label: option.text,
        });

        setState((prev) => ({
          ...prev,
          step1Answer: option.key as any,
        }));
      }
    },
    [setState, capture]
  );

  return (
    <WizardContentLayout
      appID={appID}
      backButtonDisabled={true}
      primaryButton={
        <PrimaryButton
          onClick={onClickNext}
          text={<FormattedMessage id="next" />}
        />
      }
    >
      <WizardTitle>
        <FormattedMessage
          id="ProjectWizardScreen.step.title"
          values={{ appID: rawAppID, currentStep: 1, totalStep: 3 }}
        />
      </WizardTitle>
      <ChoiceGroup
        label={renderToString("ProjectWizardScreen.step1.question")}
        options={options}
        selectedKey={step1Answer}
        onChange={onChange}
      />
    </WizardContentLayout>
  );
}

function Step2(props: StepProps) {
  const {
    appID,
    rawAppID,
    form: {
      state: { step2Answer },
      setState,
    },
  } = props;
  const { renderToString } = useContext(Context);
  const navigate = useNavigate();

  const options: IChoiceGroupOption[] = useMemo(() => {
    return [
      {
        key: "true",
        text: renderToString("ProjectWizardScreen.step2.option.true"),
      },
      {
        key: "false",
        text: renderToString("ProjectWizardScreen.step2.option.false"),
      },
    ];
  }, [renderToString]);

  const capture = useCapture();
  const onClickNext = useCallback(
    (e) => {
      e.preventDefault();
      e.stopPropagation();

      capture("projectWizard.set-passkey");

      navigate("./../3");
    },
    [navigate, capture]
  );

  const onChange = useCallback(
    (_e, option: IChoiceGroupOption | undefined) => {
      if (option != null) {
        capture("projectWizard.clicked-option", {
          label: option.text,
        });

        setState((prev) => ({
          ...prev,
          step2Answer: option.key === "true" ? true : false,
        }));
      }
    },
    [setState, capture]
  );

  return (
    <WizardContentLayout
      appID={appID}
      primaryButton={
        <PrimaryButton
          onClick={onClickNext}
          text={<FormattedMessage id="next" />}
        />
      }
    >
      <WizardTitle>
        <FormattedMessage
          id="ProjectWizardScreen.step.title"
          values={{ appID: rawAppID, currentStep: 2, totalStep: 3 }}
        />
      </WizardTitle>
      <ChoiceGroup
        label={renderToString("ProjectWizardScreen.step2.question")}
        options={options}
        selectedKey={String(step2Answer)}
        onChange={onChange}
      />
      <WizardDescription>
        <FormattedMessage id="ProjectWizardScreen.step2.description" />
      </WizardDescription>
    </WizardContentLayout>
  );
}

function Step3(props: StepProps) {
  const {
    appID,
    rawAppID,
    form: {
      state: { step1Answer, step3Answer },
      setState,
    },
    saveAndThenNavigate,
  } = props;
  const { renderToString } = useContext(Context);

  const q1Options: IChoiceGroupOption[] = useMemo(() => {
    return [
      {
        key: "true",
        text: renderToString("ProjectWizardScreen.step3.q1.option.true"),
      },
      {
        key: "false",
        text: renderToString("ProjectWizardScreen.step3.q1.option.false"),
      },
    ];
  }, [renderToString]);

  const q2Options: IChoiceGroupOption[] = useMemo(() => {
    return [
      {
        key: "if_exists",
        text: renderToString("ProjectWizardScreen.step3.q2.option.if_exists"),
      },
      {
        key: "required",
        text: renderToString("ProjectWizardScreen.step3.q2.option.required"),
      },
      {
        key: "disabled",
        text: renderToString("ProjectWizardScreen.step3.q2.option.disabled"),
      },
    ];
  }, [renderToString]);

  const q3Options: Step3Question3Option[] = useMemo(() => {
    const options: Step3Question3Option[] = [
      {
        key: "totp",
        text: (
          <FormattedMessage
            id="ProjectWizardScreen.step3.q3.option.totp"
            components={{
              Text,
            }}
            values={{
              className: styles.totpExamples,
            }}
          />
        ) as any,
      },
    ];
    switch (step1Answer) {
      case "email_password":
      // fallthrough
      case "phone_password":
        options.push({
          key: "oob_otp_email",
          text: renderToString(
            "ProjectWizardScreen.step3.q3.option.oob_otp_email"
          ),
        });
        break;
      case "oob_otp_email":
      // fallthrough
      case "oob_otp_sms":
        options.push({
          key: "password",
          text: renderToString("ProjectWizardScreen.step3.q3.option.password"),
        });
        break;
    }
    return options;
  }, [step1Answer, renderToString]);

  const capture = useCapture();
  const onClickNext = useCallback(
    (e) => {
      e.preventDefault();
      e.stopPropagation();

      capture("projectWizard.set-2fa");
      capture("projectWizard.completed");

      saveAndThenNavigate();
    },
    [saveAndThenNavigate, capture]
  );

  const onChangeQ1 = useCallback(
    (_e, option: IChoiceGroupOption | undefined) => {
      if (option != null) {
        capture("projectWizard.clicked-option", {
          label: option.text,
        });

        setState((prev) => ({
          ...prev,
          step3Answer: {
            ...prev.step3Answer,
            useRecommenededSettings: option.key === "true" ? true : false,
          },
        }));
      }
    },
    [setState, capture]
  );

  const onChangeQ2 = useCallback(
    (_e, option: IChoiceGroupOption | undefined) => {
      if (option != null) {
        capture("projectWizard.clicked-option", {
          label: option.text,
        });

        setState((prev) => ({
          ...prev,
          step3Answer: {
            ...prev.step3Answer,
            secondaryAuthenticationMode: option.key as any,
          },
        }));
      }
    },
    [setState, capture]
  );

  const onChangeQ3 = useCallback(
    (_e, option: IChoiceGroupOption | undefined) => {
      if (option != null) {
        capture("projectWizard.clicked-option", {
          label: option.text,
        });

        setState((prev) => ({
          ...prev,
          step3Answer: {
            ...prev.step3Answer,
            secondaryAuthenticatorType: option.key as any,
          },
        }));
      }
    },
    [setState, capture]
  );

  return (
    <WizardContentLayout
      appID={appID}
      primaryButton={
        <PrimaryButton
          onClick={onClickNext}
          text={<FormattedMessage id="next" />}
        />
      }
    >
      <WizardTitle>
        <FormattedMessage
          id="ProjectWizardScreen.step.title"
          values={{ appID: rawAppID, currentStep: 3, totalStep: 3 }}
        />
      </WizardTitle>
      <ChoiceGroup
        label={renderToString("ProjectWizardScreen.step3.q1")}
        options={q1Options}
        selectedKey={String(step3Answer.useRecommenededSettings)}
        onChange={onChangeQ1}
      />
      <HorizontalDivider />
      {step3Answer.useRecommenededSettings ? (
        <>
          <WizardTitle>
            <FormattedMessage id="ProjectWizardScreen.step3.recommended-settings.title" />
          </WizardTitle>
          <WizardDescription>
            <FormattedMessage id="ProjectWizardScreen.step3.recommended-settings.description" />
          </WizardDescription>
        </>
      ) : (
        <>
          <WizardTitle>
            <FormattedMessage id="ProjectWizardScreen.step3.custom-settings.title" />
          </WizardTitle>
          <ChoiceGroup
            label={renderToString("ProjectWizardScreen.step3.q2")}
            options={q2Options}
            selectedKey={String(step3Answer.secondaryAuthenticationMode)}
            onChange={onChangeQ2}
          />
          <ChoiceGroup
            label={renderToString("ProjectWizardScreen.step3.q3")}
            options={q3Options}
            selectedKey={String(step3Answer.secondaryAuthenticatorType)}
            onChange={onChangeQ3}
          />
        </>
      )}
    </WizardContentLayout>
  );
}

export default function ProjectWizardScreen(): React.ReactElement {
  const { appID } = useParams() as { appID: string };
  const navigate = useNavigate();
  const form = useAppConfigForm({
    appID,
    constructFormState,
    constructConfig,
    initialCanSave: true,
  });
  const {
    isLoading,
    loadError,
    reload,
    updateError,
    isUpdating,
    effectiveConfig,
    save,
  } = form;

  const rawAppID = effectiveConfig.id;

  const saveAndThenNavigate = useCallback(() => {
    save().then(
      () => {
        navigate("./../");
      },
      () => {}
    );
  }, [save, navigate]);

  if (isLoading) {
    return <ShowLoading />;
  }

  if (loadError != null) {
    return <ShowError error={loadError} onRetry={reload} />;
  }

  return (
    <FormProvider loading={isUpdating} error={updateError}>
      <WizardScreenLayout>
        <FormErrorMessageBar />
        <Routes>
          <Route
            path="/1"
            element={
              <Step1
                appID={appID}
                rawAppID={rawAppID}
                form={form}
                saveAndThenNavigate={saveAndThenNavigate}
              />
            }
          />
          <Route
            path="/2"
            element={
              <Step2
                appID={appID}
                rawAppID={rawAppID}
                form={form}
                saveAndThenNavigate={saveAndThenNavigate}
              />
            }
          />
          <Route
            path="/3"
            element={
              <Step3
                appID={appID}
                rawAppID={rawAppID}
                form={form}
                saveAndThenNavigate={saveAndThenNavigate}
              />
            }
          />
          <Route path="*" element={<Navigate to="1" replace={true} />} />
        </Routes>
      </WizardScreenLayout>
    </FormProvider>
  );
}
