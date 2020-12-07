import React, { useCallback, useContext, useMemo } from "react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import { TextField } from "@fluentui/react";
import produce from "immer";
import { clearEmptyObject } from "../../util/misc";
import { PortalAPIAppConfig } from "../../types";
import {
  AppConfigFormModel,
  useAppConfigForm,
} from "../../hook/useAppConfigForm";
import NavBreadcrumb, { BreadcrumbItem } from "../../NavBreadcrumb";
import { useParams } from "react-router-dom";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import FormContainer from "../../FormContainer";

import styles from "./ForgotPasswordSettings.module.scss";

interface FormState {
  codeExpirySeconds: number;
}

function constructFormState(config: PortalAPIAppConfig): FormState {
  return {
    codeExpirySeconds:
      config.forgot_password?.reset_code_expiry_seconds ?? 1200,
  };
}

function constructConfig(
  config: PortalAPIAppConfig,
  initialState: FormState,
  currentState: FormState
): PortalAPIAppConfig {
  return produce(config, (config) => {
    config.forgot_password = config.forgot_password ?? {};
    if (initialState.codeExpirySeconds !== currentState.codeExpirySeconds) {
      config.forgot_password.reset_code_expiry_seconds =
        currentState.codeExpirySeconds;
    }
    clearEmptyObject(config);
  });
}

interface ForgotPasswordSettingsContentProps {
  form: AppConfigFormModel<FormState>;
}

const ForgotPasswordSettingsContent: React.FC<ForgotPasswordSettingsContentProps> = function ForgotPasswordSettingsContent(
  props
) {
  const { state, setState } = props.form;

  const { renderToString } = useContext(Context);

  const navBreadcrumbItems: BreadcrumbItem[] = useMemo(() => {
    return [
      {
        to: ".",
        label: <FormattedMessage id="ForgotPasswordSettingsScreen.title" />,
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

  return (
    <div className={styles.root}>
      <NavBreadcrumb items={navBreadcrumbItems} />
      <TextField
        className={styles.textField}
        type="number"
        min="0"
        step="1"
        label={renderToString(
          "ForgotPasswordSettingsScreen.reset-code-valid-duration.label"
        )}
        value={String(state.codeExpirySeconds)}
        onChange={onCodeExpirySecondsChange}
      />
    </div>
  );
};

const ForgotPasswordSettingsScreen: React.FC = function ForgotPasswordSettingsScreen() {
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
      <ForgotPasswordSettingsContent form={form} />
    </FormContainer>
  );
};

export default ForgotPasswordSettingsScreen;
