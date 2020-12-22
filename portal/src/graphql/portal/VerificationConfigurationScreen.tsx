import React, { useCallback, useContext, useMemo } from "react";
import { useParams } from "react-router-dom";
import { produce } from "immer";
import cn from "classnames";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import {
  Checkbox,
  DefaultEffects,
  Dropdown,
  IDropdownOption,
  TextField,
  Toggle,
} from "@fluentui/react";
import {
  PortalAPIAppConfig,
  VerificationClaimConfig,
  VerificationCriteria,
  verificationCriteriaList,
} from "../../types";
import { clearEmptyObject } from "../../util/misc";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import {
  AppConfigFormModel,
  useAppConfigForm,
} from "../../hook/useAppConfigForm";
import NavBreadcrumb, { BreadcrumbItem } from "../../NavBreadcrumb";
import FormContainer from "../../FormContainer";

import styles from "./VerificationConfigurationScreen.module.scss";

const cardStyle = {
  boxShadow: DefaultEffects.elevation4,
};

interface FormState {
  codeExpirySeconds: number;
  criteria: VerificationCriteria;
  email: Required<VerificationClaimConfig>;
  phone: Required<VerificationClaimConfig>;
}

function constructFormState(config: PortalAPIAppConfig): FormState {
  return {
    codeExpirySeconds: config.verification?.code_expiry_seconds ?? 3600,
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
}

const VerificationConfigurationContent: React.FC<VerificationConfigurationContentProps> = function VerificationConfigurationContent(
  props
) {
  const { state, setState } = props.form;

  const { renderToString } = useContext(Context);

  const navBreadcrumbItems: BreadcrumbItem[] = useMemo(() => {
    return [
      {
        to: ".",
        label: <FormattedMessage id="VerificationConfigurationScreen.title" />,
      },
    ];
  }, []);

  const onCodeExpirySecondsChange = useCallback(
    (_, value?: string) => {
      setState((state) => ({
        ...state,
        codeExpirySeconds: Number(value),
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

  return (
    <div className={styles.root}>
      <NavBreadcrumb items={navBreadcrumbItems} />
      <TextField
        className={styles.formField}
        type="number"
        min="0"
        step="1"
        label={renderToString(
          "VerificationConfigurationScreen.code-expiry-seconds.label"
        )}
        value={String(state.codeExpirySeconds)}
        onChange={onCodeExpirySecondsChange}
      />
      <Dropdown
        className={styles.formField}
        label={renderToString("VerificationConfigurationScreen.criteria.label")}
        options={criteriaOptions}
        selectedKey={state.criteria}
        onChange={onCriteriaChange}
      />
      <div
        className={cn(styles.formField, styles.claimConfigCard)}
        style={cardStyle}
      >
        <Toggle
          className={styles.claimConfigTitle}
          checked={state.email.enabled}
          onChange={onEmailEnabledChange}
          label={renderToString(
            "VerificationConfigurationScreen.verification.claims.email"
          )}
          inlineLabel={true}
        />
        <Checkbox
          disabled={!state.email.enabled}
          checked={state.email.required}
          onChange={onEmailRequiredChange}
          label={renderToString(
            "VerificationConfigurationScreen.verification.required.label"
          )}
        />
      </div>
      <div
        className={cn(styles.formField, styles.claimConfigCard)}
        style={cardStyle}
      >
        <Toggle
          className={styles.claimConfigTitle}
          checked={state.phone.enabled}
          onChange={onPhoneEnabledChange}
          label={renderToString(
            "VerificationConfigurationScreen.verification.claims.phoneNumber"
          )}
          inlineLabel={true}
        />
        <Checkbox
          disabled={!state.phone.enabled}
          checked={state.phone.required}
          onChange={onPhoneRequiredChange}
          label={renderToString(
            "VerificationConfigurationScreen.verification.required.label"
          )}
        />
      </div>
    </div>
  );
};

const VerificationConfigurationScreen: React.FC = function VerificationConfigurationScreen() {
  const { appID } = useParams();
  const form = useAppConfigForm(appID, constructFormState, constructConfig);

  if (form.isLoading) {
    return <ShowLoading />;
  }

  if (form.loadError) {
    return <ShowError error={form.loadError} onRetry={form.reload} />;
  }

  return (
    <FormContainer form={form}>
      <VerificationConfigurationContent form={form} />
    </FormContainer>
  );
};

export default VerificationConfigurationScreen;
