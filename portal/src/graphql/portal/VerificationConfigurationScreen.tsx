import React, { useCallback, useContext, useMemo } from "react";
import { useParams } from "react-router-dom";
import { produce } from "immer";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import {
  Checkbox,
  Dropdown,
  IDropdownOption,
  DirectionalHint,
} from "@fluentui/react";
import {
  IdentityFeatureConfig,
  authenticatorPhoneOTPModeList,
  AuthenticatorPhoneOTPMode,
  PortalAPIAppConfig,
  VerificationClaimConfig,
  VerificationCriteria,
  verificationCriteriaList,
} from "../../types";
import { clearEmptyObject } from "../../util/misc";
import { parseIntegerAllowLeadingZeros } from "../../util/input";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import ScreenContent from "../../ScreenContent";
import ScreenTitle from "../../ScreenTitle";
import ScreenDescription from "../../ScreenDescription";
import WidgetTitle from "../../WidgetTitle";
import Widget from "../../Widget";
import {
  AppConfigFormModel,
  useAppConfigForm,
} from "../../hook/useAppConfigForm";
import FormContainer from "../../FormContainer";

import styles from "./VerificationConfigurationScreen.module.css";
import { useAppFeatureConfigQuery } from "./query/appFeatureConfigQuery";
import LabelWithTooltip from "../../LabelWithTooltip";
import TextField from "../../TextField";
import FeatureDisabledMessageBar from "./FeatureDisabledMessageBar";

const DEFAULT_PHONE_OTP_MODE: AuthenticatorPhoneOTPMode = "whatsapp_sms";

interface FormState {
  codeExpirySeconds: number | undefined;
  criteria: VerificationCriteria;
  email: Partial<VerificationClaimConfig>;
  phone: Partial<VerificationClaimConfig>;
  phoneOTPMode: AuthenticatorPhoneOTPMode;
}

function onRenderCriteriaLabel() {
  return (
    <LabelWithTooltip
      labelId="VerificationConfigurationScreen.criteria.label"
      tooltipMessageId="VerificationConfigurationScreen.criteria.tooltip"
      directionalHint={DirectionalHint.topCenter}
    />
  );
}

function constructFormState(config: PortalAPIAppConfig): FormState {
  return {
    codeExpirySeconds: config.verification?.code_expiry_seconds,
    criteria: config.verification?.criteria ?? "any",
    email: {
      required: config.verification?.claims?.email?.required ?? true,
      enabled: config.verification?.claims?.email?.enabled ?? true,
    },
    phone: {
      required: config.verification?.claims?.phone_number?.required ?? true,
      enabled: config.verification?.claims?.phone_number?.enabled ?? true,
    },
    phoneOTPMode:
      config.authenticator?.oob_otp?.sms?.phone_otp_mode ??
      DEFAULT_PHONE_OTP_MODE,
  };
}

function constructConfig(
  config: PortalAPIAppConfig,
  _initialState: FormState,
  currentState: FormState
): PortalAPIAppConfig {
  // eslint-disable-next-line complexity
  return produce(config, (config) => {
    config.verification ??= {};
    config.authenticator ??= {};
    const v = config.verification;
    const a = config.authenticator;
    v.claims ??= {};
    v.claims.email ??= {};
    v.claims.phone_number ??= {};
    a.oob_otp ??= {};
    a.oob_otp.sms ??= {};

    v.code_expiry_seconds = currentState.codeExpirySeconds;
    v.criteria = currentState.criteria;
    v.claims.email.required = currentState.email.required;
    v.claims.email.enabled = currentState.email.enabled;
    v.claims.phone_number.required = currentState.phone.required;
    v.claims.phone_number.enabled = currentState.phone.enabled;
    a.oob_otp.sms.phone_otp_mode = currentState.phoneOTPMode;

    clearEmptyObject(config);
  });
}

const criteriaMessageIds: Record<VerificationCriteria, string> = {
  any: "VerificationConfigurationScreen.criteria.any",
  all: "VerificationConfigurationScreen.criteria.all",
};

const phoneOTPModeMessageIds: Record<AuthenticatorPhoneOTPMode, string> = {
  whatsapp_sms:
    "VerificationConfigurationScreen.verification.phoneNumber.verify-by.whatsapp-or-sms",
  whatsapp:
    "VerificationConfigurationScreen.verification.phoneNumber.verify-by.whatsapp-only",
  sms: "VerificationConfigurationScreen.verification.phoneNumber.verify-by.sms-only",
};

interface VerificationConfigurationContentProps {
  form: AppConfigFormModel<FormState>;
  identityFeatureConfig?: IdentityFeatureConfig;
}

const VerificationConfigurationContent: React.VFC<VerificationConfigurationContentProps> =
  function VerificationConfigurationContent(props) {
    const { state, setState } = props.form;

    const { identityFeatureConfig } = props;

    const { renderToString } = useContext(Context);

    const onCodeExpirySecondsChange = useCallback(
      (_, value?: string) => {
        setState((state) => ({
          ...state,
          codeExpirySeconds: parseIntegerAllowLeadingZeros(value),
        }));
      },
      [setState]
    );

    const criteriaOptions = useMemo(
      () =>
        verificationCriteriaList.map((criteria) => {
          return {
            key: criteria,
            text: renderToString(criteriaMessageIds[criteria]),
          };
        }),
      [renderToString]
    );

    const phoneOTPModes = useMemo(
      () =>
        authenticatorPhoneOTPModeList.map((mode) => {
          return {
            key: mode,
            text: renderToString(phoneOTPModeMessageIds[mode]),
          };
        }),
      [renderToString]
    );

    const onCriteriaChange = useCallback(
      (_, option?: IDropdownOption) => {
        const key = option?.key as VerificationCriteria | undefined;
        if (key) {
          setState((state) => ({
            ...state,
            criteria: key,
          }));
        }
      },
      [setState]
    );

    const onPhoneOTPModeChange = useCallback(
      (_, option?: IDropdownOption) => {
        const key = option?.key as AuthenticatorPhoneOTPMode | undefined;
        if (key) {
          setState((state) => ({
            ...state,
            phoneOTPMode: key,
          }));
        }
      },
      [setState]
    );

    const onEmailRequiredChange = useCallback(
      (_, value?: boolean) => {
        setState((s) => ({
          ...s,
          email: {
            ...s.email,
            required: value ?? false,
            enabled: value ? true : s.email.enabled,
          },
        }));
      },
      [setState]
    );

    const onEmailEnabledChange = useCallback(
      (_, value?: boolean) => {
        setState((s) => ({
          ...s,
          email: { ...s.email, enabled: value ?? false },
        }));
      },
      [setState]
    );

    const onPhoneRequiredChange = useCallback(
      (_, value?: boolean) => {
        setState((s) => ({
          ...s,
          phone: {
            ...s.phone,
            required: value ?? false,
            enabled: value ? true : s.phone.enabled,
          },
        }));
      },
      [setState]
    );

    const onPhoneEnabledChange = useCallback(
      (_, value?: boolean) => {
        setState((s) => ({
          ...s,
          phone: { ...s.phone, enabled: value ?? false },
        }));
      },
      [setState]
    );

    const loginIDPhoneDisabled = useMemo(() => {
      return identityFeatureConfig?.login_id?.types?.phone?.disabled ?? false;
    }, [identityFeatureConfig]);

    return (
      <ScreenContent>
        <ScreenTitle className={styles.widget}>
          <FormattedMessage id="VerificationConfigurationScreen.title" />
        </ScreenTitle>
        <ScreenDescription className={styles.widget}>
          <FormattedMessage id="VerificationConfigurationScreen.description" />
        </ScreenDescription>
        <Widget className={styles.widget}>
          <WidgetTitle>
            <FormattedMessage id="VerificationConfigurationScreen.basic-settings" />
          </WidgetTitle>
          <TextField
            type="text"
            label={renderToString(
              "VerificationConfigurationScreen.code-expiry-seconds.label"
            )}
            value={state.codeExpirySeconds?.toFixed(0) ?? ""}
            onChange={onCodeExpirySecondsChange}
          />
          <Dropdown
            options={criteriaOptions}
            selectedKey={state.criteria}
            onChange={onCriteriaChange}
            onRenderLabel={onRenderCriteriaLabel}
          />
        </Widget>
        <Widget className={styles.widget}>
          <WidgetTitle>
            <FormattedMessage id="VerificationConfigurationScreen.verification.claims.email" />
          </WidgetTitle>
          <Checkbox
            checked={state.email.required}
            onChange={onEmailRequiredChange}
            label={renderToString(
              "VerificationConfigurationScreen.verification.email.required.label"
            )}
          />
          <Checkbox
            disabled={state.email.required}
            checked={state.email.enabled}
            onChange={onEmailEnabledChange}
            label={renderToString(
              "VerificationConfigurationScreen.verification.email.allowed.label"
            )}
          />
        </Widget>
        <Widget className={styles.widget}>
          <WidgetTitle>
            <FormattedMessage id="VerificationConfigurationScreen.verification.claims.phoneNumber" />
          </WidgetTitle>
          {loginIDPhoneDisabled ? (
            <FeatureDisabledMessageBar messageID="FeatureConfig.disabled" />
          ) : null}
          <Dropdown
            disabled={loginIDPhoneDisabled}
            label={renderToString(
              "VerificationConfigurationScreen.verification.phoneNumber.verify-by.label"
            )}
            options={phoneOTPModes}
            selectedKey={state.phoneOTPMode}
            onChange={onPhoneOTPModeChange}
          />
          <Checkbox
            disabled={loginIDPhoneDisabled}
            checked={state.phone.required}
            onChange={onPhoneRequiredChange}
            label={renderToString(
              "VerificationConfigurationScreen.verification.phone.required.label"
            )}
          />
          <Checkbox
            disabled={state.phone.required ?? loginIDPhoneDisabled}
            checked={state.phone.enabled}
            onChange={onPhoneEnabledChange}
            label={renderToString(
              "VerificationConfigurationScreen.verification.phone.allowed.label"
            )}
          />
        </Widget>
      </ScreenContent>
    );
  };

const VerificationConfigurationScreen: React.VFC =
  function VerificationConfigurationScreen() {
    const { appID } = useParams() as { appID: string };
    const form = useAppConfigForm({
      appID,
      constructFormState,
      constructConfig,
    });

    const featureConfig = useAppFeatureConfigQuery(appID);

    if (form.isLoading || featureConfig.loading) {
      return <ShowLoading />;
    }

    if (form.loadError ?? featureConfig.error) {
      return (
        <ShowError
          error={form.loadError ?? featureConfig.error}
          onRetry={() => {
            form.reload();
            featureConfig.refetch().finally(() => {});
          }}
        />
      );
    }

    return (
      <FormContainer form={form}>
        <VerificationConfigurationContent
          form={form}
          identityFeatureConfig={featureConfig.effectiveFeatureConfig?.identity}
        />
      </FormContainer>
    );
  };

export default VerificationConfigurationScreen;
