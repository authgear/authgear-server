import React, { useCallback, useMemo, useContext } from "react";
import {
  Routes,
  Navigate,
  Route,
  useParams,
  useNavigate,
} from "react-router-dom";
import {
  PrimaryButton,
  ChoiceGroup,
  MessageBar,
  IChoiceGroupOption,
  TooltipHost,
  Text,
} from "@fluentui/react";
import { FormattedMessage, Context } from "@oursky/react-messageformat";
import produce from "immer";
import { FormProvider } from "../../form";
import { FormErrorMessageBar } from "../../FormErrorMessageBar";
import WizardContentLayout from "../../WizardContentLayout";
import WizardScreenLayout from "../../WizardScreenLayout";
import {
  useAppConfigForm,
  AppConfigFormModel,
} from "../../hook/useAppConfigForm";
import {
  PortalAPIAppConfig,
  SecondaryAuthenticationMode,
  SecondaryAuthenticatorType,
} from "../../types";
import { TooltipIcon, useTooltipTargetElement } from "../../Tooltip";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import styles from "./ProjectWizardScreen.module.scss";

const TOOLTIP_HOST_STYLES = {
  root: {
    display: "inline-block",
    position: "relative",
  },
} as const;

type Step1Answer = "password" | "oob_otp_email";
type Step2Answer = SecondaryAuthenticationMode;
type Step3Answer = SecondaryAuthenticatorType;

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

function constructFromState(_config: PortalAPIAppConfig): FormState {
  return {
    step1Answer: "password",
    step2Answer: "if_exists",
    step3Answer: "totp",
  };
}

function constructConfig(
  config: PortalAPIAppConfig,
  _initialState: FormState,
  currentState: FormState
): PortalAPIAppConfig {
  return produce(config, (config) => {
    config.identity ??= {};
    config.identity.login_id ??= {};
    config.identity.login_id.keys = [{ type: "email" }];

    config.authentication ??= {};
    config.authentication.primary_authenticators = [currentState.step1Answer];
    config.authentication.secondary_authentication_mode =
      currentState.step2Answer;
    config.authentication.secondary_authenticators = [currentState.step3Answer];
  });
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

  const onClickNext = useCallback(
    (e) => {
      e.preventDefault();
      e.stopPropagation();
      navigate("./../2");
    },
    [navigate]
  );

  const { id, setRef, targetElement } = useTooltipTargetElement();
  const tooltipProps = useMemo(() => {
    return {
      targetElement,
    };
  }, [targetElement]);

  const options: IChoiceGroupOption[] = useMemo(() => {
    return [
      {
        key: "password",
        text: renderToString("ProjectWizardScreen.step1.option.email-password"),
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
                tooltipProps={tooltipProps}
                content={renderToString(
                  "ProjectWizardScreen.step1.tooltip.oob-otp-email"
                )}
              >
                <TooltipIcon id={id} setRef={setRef} />
              </TooltipHost>
            </>
          );
        },
      },
      {
        key: "__not_important_1__",
        text: renderToString(
          "ProjectWizardScreen.step1.option.phone-number-password"
        ),
        disabled: true,
      },
      {
        key: "__not_important_2__",
        text: renderToString("ProjectWizardScreen.step1.option.oob-otp-sms"),
        disabled: true,
      },
    ];
  }, [renderToString, id, setRef, tooltipProps]);

  const onChange = useCallback(
    (_e, option) => {
      if (option != null) {
        setState((prev) => ({
          ...prev,
          step1Answer: option.key,
        }));
      }
    },
    [setState]
  );

  return (
    <WizardContentLayout
      appID={appID}
      title={
        <FormattedMessage
          id="ProjectWizardScreen.step.title"
          values={{ appID: rawAppID, currentStep: 1, totalStep: 3 }}
        />
      }
      backButtonDisabled={true}
      primaryButton={
        <PrimaryButton onClick={onClickNext}>
          <FormattedMessage id="next" />
        </PrimaryButton>
      }
    >
      <ChoiceGroup
        label={renderToString("ProjectWizardScreen.step1.question")}
        options={options}
        selectedKey={step1Answer}
        onChange={onChange}
      />
      <MessageBar>
        <FormattedMessage id="ProjectWizardScreen.step1.message" />
      </MessageBar>
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
    saveAndThenNavigate,
  } = props;
  const { renderToString } = useContext(Context);
  const navigate = useNavigate();

  const options: IChoiceGroupOption[] = useMemo(() => {
    return [
      {
        key: "if_exists",
        text: renderToString("ProjectWizardScreen.step2.option.if-exists"),
      },
      {
        key: "required",
        text: renderToString("ProjectWizardScreen.step2.option.required"),
      },
      {
        key: "disabled",
        text: renderToString("ProjectWizardScreen.step2.option.disabled"),
      },
    ];
  }, [renderToString]);

  const onClickNext = useCallback(
    (e) => {
      e.preventDefault();
      e.stopPropagation();
      if (step2Answer === "disabled") {
        saveAndThenNavigate();
      } else {
        navigate("./../3");
      }
    },
    [step2Answer, navigate, saveAndThenNavigate]
  );

  const onChange = useCallback(
    (_e, option) => {
      if (option != null) {
        setState((prev) => ({
          ...prev,
          step2Answer: option.key,
        }));
      }
    },
    [setState]
  );

  return (
    <WizardContentLayout
      appID={appID}
      title={
        <FormattedMessage
          id="ProjectWizardScreen.step.title"
          values={{ appID: rawAppID, currentStep: 2, totalStep: 3 }}
        />
      }
      primaryButton={
        <PrimaryButton onClick={onClickNext}>
          <FormattedMessage id="next" />
        </PrimaryButton>
      }
    >
      <ChoiceGroup
        label={renderToString("ProjectWizardScreen.step2.question")}
        options={options}
        selectedKey={step2Answer}
        onChange={onChange}
      />
      <Text block={true}>
        <FormattedMessage id="ProjectWizardScreen.step2.description" />
      </Text>
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

  const options: IChoiceGroupOption[] = useMemo(() => {
    const options = [
      {
        key: "totp",
        text: (
          <FormattedMessage
            id="ProjectWizardScreen.step3.option.totp"
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
    if (step1Answer === "password") {
      options.push({
        key: "oob_otp_email",
        text: renderToString("ProjectWizardScreen.step3.option.oob-otp-email"),
      });
    } else {
      options.push({
        key: "password",
        text: renderToString("ProjectWizardScreen.step3.option.password"),
      });
    }
    return options;
  }, [step1Answer, renderToString]);

  const onClickNext = useCallback(
    (e) => {
      e.preventDefault();
      e.stopPropagation();
      saveAndThenNavigate();
    },
    [saveAndThenNavigate]
  );

  const onChange = useCallback(
    (_e, option) => {
      if (option != null) {
        setState((prev) => ({
          ...prev,
          step3Answer: option.key,
        }));
      }
    },
    [setState]
  );

  return (
    <WizardContentLayout
      appID={appID}
      title={
        <FormattedMessage
          id="ProjectWizardScreen.step.title"
          values={{ appID: rawAppID, currentStep: 3, totalStep: 3 }}
        />
      }
      primaryButton={
        <PrimaryButton onClick={onClickNext}>
          <FormattedMessage id="next" />
        </PrimaryButton>
      }
    >
      <ChoiceGroup
        label={renderToString("ProjectWizardScreen.step2.question")}
        options={options}
        selectedKey={step3Answer}
        onChange={onChange}
      />
    </WizardContentLayout>
  );
}

export default function ProjectWizardScreen(): React.ReactElement {
  const { appID } = useParams() as { appID: string };
  const navigate = useNavigate();
  const form = useAppConfigForm(appID, constructFromState, constructConfig);
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
        navigate("./done");
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
