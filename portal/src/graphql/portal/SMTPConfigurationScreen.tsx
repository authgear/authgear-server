import React, { useCallback, useContext } from "react";
import { useParams } from "react-router-dom";
import produce from "immer";
import { Toggle, TextField, PrimaryButton } from "@fluentui/react";
import { FormattedMessage, Context } from "@oursky/react-messageformat";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import FormContainer from "../../FormContainer";
import {
  AppSecretConfigFormModel,
  useAppSecretConfigForm,
} from "../../hook/useAppSecretConfigForm";
import ScreenContent from "../../ScreenContent";
import ScreenTitle from "../../ScreenTitle";
import ScreenDescription from "../../ScreenDescription";
import Widget from "../../Widget";
import { startReauthentication } from "./Authenticated";
import { PortalAPIAppConfig, PortalAPISecretConfig } from "../../types";
import styles from "./SMTPConfigurationScreen.module.scss";

const MASKED_PASSWORD_VALUE = "****************";

interface FormState {
  enabled: boolean;
  host: string;
  portString: string;
  username: string;
  password: string;
  isPasswordMasked: boolean;
}

function constructFormState(
  _config: PortalAPIAppConfig,
  secrets: PortalAPISecretConfig
): FormState {
  const enabled = secrets.smtpSecret != null;
  const host = secrets.smtpSecret?.host ?? "";
  const portString =
    secrets.smtpSecret?.port != null ? String(secrets.smtpSecret.port) : "";
  const username = secrets.smtpSecret?.username ?? "";
  const password = secrets.smtpSecret?.password ?? "";
  const isPasswordMasked = enabled && secrets.smtpSecret?.password == null;
  return {
    enabled,
    host,
    portString,
    username,
    password,
    isPasswordMasked,
  };
}

function constructConfig(
  config: PortalAPIAppConfig,
  secrets: PortalAPISecretConfig,
  _initialState: FormState,
  currentState: FormState,
  _effectiveConfig: PortalAPIAppConfig
): [PortalAPIAppConfig, PortalAPISecretConfig] {
  const newSecrets = produce(secrets, (secrets) => {
    if (!currentState.enabled) {
      secrets.smtpSecret = null;
    } else if (currentState.isPasswordMasked) {
      secrets.smtpSecret = {
        host: currentState.host,
        port: Number(currentState.portString),
        username: currentState.username,
        password: null,
      };
    } else {
      secrets.smtpSecret = {
        host: currentState.host,
        port: Number(currentState.portString),
        username: currentState.username,
        password: currentState.password,
      };
    }
  });
  return [config, newSecrets];
}

interface SMTPConfigurationScreenContentProps {
  form: AppSecretConfigFormModel<FormState>;
}

const SMTPConfigurationScreenContent: React.FC<SMTPConfigurationScreenContentProps> =
  function SMTPConfigurationScreenContent(props) {
    const { state, setState } = props.form;
    const { renderToString } = useContext(Context);

    const onChangeEnabled = useCallback(
      (_event, checked?: boolean) => {
        if (checked != null) {
          setState((state) => {
            return {
              ...state,
              enabled: checked,
            };
          });
        }
      },
      [setState]
    );

    const onChangeHost = useCallback(
      (_, value?: string) => {
        if (value != null) {
          setState((state) => {
            return {
              ...state,
              host: value,
            };
          });
        }
      },
      [setState]
    );

    const onChangePort = useCallback(
      (_, value?: string) => {
        if (value != null) {
          if (value === "") {
            setState((state) => {
              return {
                ...state,
                portString: "",
              };
            });
          } else {
            const port = Number(value);
            if (!isNaN(port)) {
              setState((state) => {
                return {
                  ...state,
                  portString: value,
                };
              });
            }
          }
        }
      },
      [setState]
    );

    const onChangeUsername = useCallback(
      (_, value?: string) => {
        if (value != null) {
          setState((state) => {
            return {
              ...state,
              username: value,
            };
          });
        }
      },
      [setState]
    );

    const onChangePassword = useCallback(
      (_, value?: string) => {
        if (value != null) {
          setState((state) => {
            return {
              ...state,
              password: value,
            };
          });
        }
      },
      [setState]
    );

    const onClickEdit = useCallback((e: React.MouseEvent<unknown>) => {
      e.preventDefault();
      e.stopPropagation();

      startReauthentication().catch((e) => {
        // Normally there should not be any error.
        console.error(e);
      });
    }, []);

    return (
      <ScreenContent className={styles.root}>
        <ScreenTitle>
          <FormattedMessage id="SMTPConfigurationScreen.title" />
        </ScreenTitle>
        <ScreenDescription className={styles.widget}>
          <FormattedMessage id="SMTPConfigurationScreen.description" />
        </ScreenDescription>

        <Widget className={styles.widget}>
          <Toggle
            className={styles.control}
            checked={state.enabled}
            onChange={onChangeEnabled}
            label={renderToString("SMTPConfigurationScreen.enable.label")}
            inlineLabel={true}
          />
          {state.enabled && (
            <>
              <TextField
                className={styles.control}
                type="text"
                label={renderToString("SMTPConfigurationScreen.host.label")}
                value={state.host}
                disabled={state.isPasswordMasked}
                required={true}
                onChange={onChangeHost}
              />
              <TextField
                className={styles.control}
                type="number"
                min="1"
                step="1"
                max="65535"
                label={renderToString("SMTPConfigurationScreen.port.label")}
                value={state.portString}
                disabled={state.isPasswordMasked}
                required={true}
                onChange={onChangePort}
              />
              <TextField
                className={styles.control}
                type="text"
                label={renderToString("SMTPConfigurationScreen.username.label")}
                value={state.username}
                disabled={state.isPasswordMasked}
                onChange={onChangeUsername}
              />
              <TextField
                className={styles.control}
                type="password"
                label={renderToString("SMTPConfigurationScreen.password.label")}
                value={
                  state.isPasswordMasked
                    ? MASKED_PASSWORD_VALUE
                    : state.password
                }
                disabled={state.isPasswordMasked}
                onChange={onChangePassword}
              />
              {state.isPasswordMasked && (
                <PrimaryButton className={styles.control} onClick={onClickEdit}>
                  <FormattedMessage id="edit" />
                </PrimaryButton>
              )}
            </>
          )}
        </Widget>
      </ScreenContent>
    );
  };

const SMTPConfigurationScreen: React.FC = function SMTPConfigurationScreen() {
  const { appID } = useParams();
  const form = useAppSecretConfigForm(
    appID,
    constructFormState,
    constructConfig
  );

  if (form.isLoading) {
    return <ShowLoading />;
  }

  if (form.loadError) {
    return <ShowError error={form.loadError} onRetry={form.reload} />;
  }

  return (
    <FormContainer form={form}>
      <SMTPConfigurationScreenContent form={form} />
    </FormContainer>
  );
};

export default SMTPConfigurationScreen;
