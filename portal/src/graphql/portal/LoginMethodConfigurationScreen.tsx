import React from "react";
import { useParams } from "react-router-dom";
import { produce } from "immer";
import { FormattedMessage } from "@oursky/react-messageformat";
import { PortalAPIAppConfig, IdentityFeatureConfig } from "../../types";
import { clearEmptyObject } from "../../util/misc";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import ScreenContent from "../../ScreenContent";
import ScreenTitle from "../../ScreenTitle";
import ScreenDescription from "../../ScreenDescription";
import {
  AppConfigFormModel,
  useAppConfigForm,
} from "../../hook/useAppConfigForm";
import FormContainer from "../../FormContainer";
import { useAppFeatureConfigQuery } from "./query/appFeatureConfigQuery";
import styles from "./LoginMethodConfigurationScreen.module.css";

interface FormState {}

function constructFormState(_config: PortalAPIAppConfig): FormState {
  return {};
}

function constructConfig(
  config: PortalAPIAppConfig,
  _initialState: FormState,
  _currentState: FormState,
  _effectiveConfig: PortalAPIAppConfig
): PortalAPIAppConfig {
  return produce(config, (config) => {
    clearEmptyObject(config);
  });
}

interface LoginMethodConfigurationContentProps {
  form: AppConfigFormModel<FormState>;
  identityFeatureConfig?: IdentityFeatureConfig;
}

const LoginMethodConfigurationContent: React.VFC<LoginMethodConfigurationContentProps> =
  function LoginMethodConfigurationContent(_props) {
    return (
      <ScreenContent>
        <ScreenTitle className={styles.widget}>
          <FormattedMessage id="LoginMethodConfigurationScreen.title" />
        </ScreenTitle>
        <ScreenDescription className={styles.widget}>
          <FormattedMessage id="LoginMethodConfigurationScreen.description" />
        </ScreenDescription>
      </ScreenContent>
    );
  };

const LoginMethodConfigurationScreen: React.VFC =
  function LoginMethodConfigurationScreen() {
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

    if (form.loadError) {
      return <ShowError error={form.loadError} onRetry={form.reload} />;
    }

    if (featureConfig.error) {
      return (
        <ShowError
          error={featureConfig.error}
          onRetry={featureConfig.refetch}
        />
      );
    }

    return (
      <FormContainer form={form}>
        <LoginMethodConfigurationContent
          form={form}
          identityFeatureConfig={featureConfig.effectiveFeatureConfig?.identity}
        />
      </FormContainer>
    );
  };

export default LoginMethodConfigurationScreen;
