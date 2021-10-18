import React, { useCallback, useContext, useMemo } from "react";
import { useParams } from "react-router-dom";
import { produce } from "immer";
import cn from "classnames";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import {
  Checkbox,
  Dropdown,
  IDropdownOption,
  TextField,
  Toggle,
  MessageBar,
} from "@fluentui/react";
import {
  IdentityFeatureConfig,
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

import styles from "./VerificationConfigurationScreen.module.scss";
import { useAppFeatureConfigQuery } from "./query/appFeatureConfigQuery";

interface FormState {
  codeExpirySeconds: number | undefined;
  criteria: VerificationCriteria;
  email: Required<VerificationClaimConfig>;
  phone: Required<VerificationClaimConfig>;
}

function constructFormState(config: PortalAPIAppConfig): FormState {
  return {
    codeExpirySeconds: config.verification?.code_expiry_seconds,
    criteria: config.verification?.criteria ?? "any",
    email: {
      enabled: config.verification?.claims?.email?.enabled ?? true,
      required: config.verification?.claims?.email?.required ?? true,
    },
    phone: {
      enabled: config.verification?.claims?.phone_number?.enabled ?? true,
      required: config.verification?.claims?.phone_number?.required ?? true,
    },
  };
}

function constructConfig(
  config: PortalAPIAppConfig,
  initialState: FormState,
  currentState: FormState
): PortalAPIAppConfig {
  // eslint-disable-next-line complexity
  return produce(config, (config) => {
    config.verification ??= {};
    const v = config.verification;
    v.claims ??= {};
    v.claims.email ??= {};
    v.claims.phone_number ??= {};

    if (initialState.codeExpirySeconds !== currentState.codeExpirySeconds) {
      v.code_expiry_seconds = currentState.codeExpirySeconds;
    }
    if (initialState.criteria !== currentState.criteria) {
      v.criteria = currentState.criteria;
    }
    if (initialState.email.enabled !== currentState.email.enabled) {
      v.claims.email.enabled = currentState.email.enabled;
    }
    if (initialState.email.required !== currentState.email.required) {
      v.claims.email.required = currentState.email.required;
    }
    if (initialState.phone.enabled !== currentState.phone.enabled) {
      v.claims.phone_number.enabled = currentState.phone.enabled;
    }
    if (initialState.phone.required !== currentState.phone.required) {
      v.claims.phone_number.required = currentState.phone.required;
    }

    clearEmptyObject(config);
  });
}

const criteriaMessageIds: Record<VerificationCriteria, string> = {
  any: "VerificationConfigurationScreen.criteria.any",
  all: "VerificationConfigurationScreen.criteria.all",
};

interface VerificationConfigurationContentProps {
  form: AppConfigFormModel<FormState>;
  identityFeatureConfig?: IdentityFeatureConfig;
}

const VerificationConfigurationContent: React.FC<VerificationConfigurationContentProps> =
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
            isSelected: criteria === state.criteria,
          };
        }),
      [state, renderToString]
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

    const onEmailEnabledChange = useCallback(
      (_, value?: boolean) => {
        setState((s) => ({
          ...s,
          email: { ...s.email, enabled: value ?? false },
        }));
      },
      [setState]
    );

    const onEmailRequiredChange = useCallback(
      (_, value?: boolean) => {
        setState((s) => ({
          ...s,
          email: { ...s.email, required: value ?? false },
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

    const onPhoneRequiredChange = useCallback(
      (_, value?: boolean) => {
        setState((s) => ({
          ...s,
          phone: { ...s.phone, required: value ?? false },
        }));
      },
      [setState]
    );

    const loginIDPhoneDisabled = useMemo(() => {
      return identityFeatureConfig?.login_id?.types?.phone?.disabled ?? false;
    }, [identityFeatureConfig]);

    return (
      <ScreenContent className={styles.root}>
        <ScreenTitle>
          <FormattedMessage id="VerificationConfigurationScreen.title" />
        </ScreenTitle>
        <ScreenDescription className={styles.widget}>
          <FormattedMessage id="VerificationConfigurationScreen.description" />
        </ScreenDescription>
        <Widget className={cn(styles.controlGroup, styles.widget)}>
          <WidgetTitle>
            <FormattedMessage id="VerificationConfigurationScreen.basic-settings" />
          </WidgetTitle>
          <TextField
            className={styles.control}
            type="text"
            label={renderToString(
              "VerificationConfigurationScreen.code-expiry-seconds.label"
            )}
            value={state.codeExpirySeconds?.toFixed(0) ?? ""}
            onChange={onCodeExpirySecondsChange}
          />
          <Dropdown
            className={styles.control}
            label={renderToString(
              "VerificationConfigurationScreen.criteria.label"
            )}
            options={criteriaOptions}
            selectedKey={state.criteria}
            onChange={onCriteriaChange}
          />
        </Widget>
        <Widget className={cn(styles.controlGroup, styles.widget)}>
          <Toggle
            className={styles.control}
            checked={state.email.enabled}
            onChange={onEmailEnabledChange}
            label={renderToString(
              "VerificationConfigurationScreen.verification.claims.email"
            )}
            inlineLabel={true}
          />
          <Checkbox
            className={styles.control}
            disabled={!state.email.enabled}
            checked={state.email.required}
            onChange={onEmailRequiredChange}
            label={renderToString(
              "VerificationConfigurationScreen.verification.required.label"
            )}
          />
        </Widget>
        <Widget className={cn(styles.widget)}>
          {loginIDPhoneDisabled && (
            <MessageBar>
              <FormattedMessage
                id="FeatureConfig.disabled"
                values={{
                  planPagePath: "../../../billing",
                }}
              />
            </MessageBar>
          )}
          <div
            className={cn(styles.controlGroup, {
              [styles.readOnly]: loginIDPhoneDisabled,
            })}
          >
            <Toggle
              className={styles.control}
              checked={state.phone.enabled}
              disabled={loginIDPhoneDisabled}
              onChange={onPhoneEnabledChange}
              label={renderToString(
                "VerificationConfigurationScreen.verification.claims.phoneNumber"
              )}
              inlineLabel={true}
            />
            <Checkbox
              className={styles.control}
              disabled={!state.phone.enabled || loginIDPhoneDisabled}
              checked={state.phone.required}
              onChange={onPhoneRequiredChange}
              label={renderToString(
                "VerificationConfigurationScreen.verification.required.label"
              )}
            />
          </div>
        </Widget>
      </ScreenContent>
    );
  };

const VerificationConfigurationScreen: React.FC =
  function VerificationConfigurationScreen() {
    const { appID } = useParams();
    const form = useAppConfigForm(appID, constructFormState, constructConfig);

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
