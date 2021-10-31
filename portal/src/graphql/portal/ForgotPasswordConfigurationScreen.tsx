import React, { useCallback, useContext } from "react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import { TextField } from "@fluentui/react";
import produce from "immer";
import { clearEmptyObject } from "../../util/misc";
import { parseIntegerAllowLeadingZeros } from "../../util/input";
import { PortalAPIAppConfig } from "../../types";
import {
  AppConfigFormModel,
  useAppConfigForm,
} from "../../hook/useAppConfigForm";
import { useParams } from "react-router-dom";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import ScreenContent from "../../ScreenContent";
import ScreenTitle from "../../ScreenTitle";
import ScreenDescription from "../../ScreenDescription";
import WidgetTitle from "../../WidgetTitle";
import Widget from "../../Widget";
import FormContainer from "../../FormContainer";
import styles from "./ForgotPasswordConfigurationScreen.module.scss";

interface FormState {
  codeExpirySeconds: number | undefined;
}

function constructFormState(config: PortalAPIAppConfig): FormState {
  return {
    codeExpirySeconds: config.forgot_password?.reset_code_expiry_seconds,
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

interface ForgotPasswordConfigurationScreenContentProps {
  form: AppConfigFormModel<FormState>;
}

const ForgotPasswordConfigurationScreenContent: React.FC<ForgotPasswordConfigurationScreenContentProps> =
  function ForgotPasswordConfigurationScreenContent(props) {
    const { state, setState } = props.form;

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

    return (
      <ScreenContent>
        <ScreenTitle className={styles.widget}>
          <FormattedMessage id="ForgotPasswordConfigurationScreen.title" />
        </ScreenTitle>
        <ScreenDescription className={styles.widget}>
          <FormattedMessage id="ForgotPasswordConfigurationScreen.description" />
        </ScreenDescription>
        <Widget className={styles.widget}>
          <WidgetTitle>
            <FormattedMessage id="ForgotPasswordConfigurationScreen.code-settings" />
          </WidgetTitle>
          <TextField
            type="text"
            label={renderToString(
              "ForgotPasswordConfigurationScreen.reset-code-valid-duration.label"
            )}
            value={state.codeExpirySeconds?.toFixed(0) ?? ""}
            onChange={onCodeExpirySecondsChange}
          />
        </Widget>
      </ScreenContent>
    );
  };

const ForgotPasswordConfigurationScreenScreen: React.FC =
  function ForgotPasswordConfigurationScreenScreen() {
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
        <ForgotPasswordConfigurationScreenContent form={form} />
      </FormContainer>
    );
  };

export default ForgotPasswordConfigurationScreenScreen;
