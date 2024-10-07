import React, { useCallback, useContext, useMemo } from "react";
import { useParams } from "react-router-dom";
import { produce } from "immer";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import { PortalAPIAppConfig, IdentityFeatureConfig } from "../../types";
import { clearEmptyObject } from "../../util/misc";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import ScreenContent from "../../ScreenContent";
import ScreenTitle from "../../ScreenTitle";
import ScreenDescription from "../../ScreenDescription";
import WidgetTitle from "../../WidgetTitle";
import Widget from "../../Widget";
import Toggle from "../../Toggle";
import {
  AppConfigFormModel,
  useAppConfigForm,
} from "../../hook/useAppConfigForm";
import FormContainer from "../../FormContainer";
import { useAppFeatureConfigQuery } from "./query/appFeatureConfigQuery";
import styles from "./BiometricConfigurationScreen.module.css";
import FeatureDisabledMessageBar from "./FeatureDisabledMessageBar";
import ShowOnlyIfSIWEIsDisabled from "./ShowOnlyIfSIWEIsDisabled";

interface FormState {
  enabled: boolean;
  list_enabled?: boolean;
}

function constructFormState(config: PortalAPIAppConfig): FormState {
  const enabled =
    config.authentication?.identities?.includes("biometric") ?? false;
  const list_enabled = config.identity?.biometric?.list_enabled;
  return { enabled, list_enabled };
}

function constructConfig(
  config: PortalAPIAppConfig,
  initialState: FormState,
  currentState: FormState,
  effectiveConfig: PortalAPIAppConfig
): PortalAPIAppConfig {
  return produce(config, (config) => {
    if (initialState.enabled !== currentState.enabled) {
      const identities = (
        effectiveConfig.authentication?.identities ?? []
      ).slice();
      const index = identities.indexOf("biometric");
      if (currentState.enabled && index === -1) {
        identities.push("biometric");
      } else if (!currentState.enabled && index >= 0) {
        identities.splice(index, 1);
      }
      config.authentication ??= {};
      config.authentication.identities = identities;
    }
    if (initialState.list_enabled !== currentState.list_enabled) {
      config.identity ??= {};
      config.identity.biometric ??= {};
      config.identity.biometric.list_enabled = currentState.list_enabled;
    }

    clearEmptyObject(config);
  });
}

interface BiometricConfigurationContentProps {
  form: AppConfigFormModel<FormState>;
  identityFeatureConfig?: IdentityFeatureConfig;
}

const BiometricConfigurationContent: React.VFC<BiometricConfigurationContentProps> =
  function BiometricConfigurationContent(props) {
    const { state, setState } = props.form;

    const { identityFeatureConfig } = props;

    const { renderToString } = useContext(Context);

    const onEnableChange = useCallback(
      (_event, checked?: boolean) =>
        setState((state) => ({
          ...state,
          enabled: checked ?? false,
        })),
      [setState]
    );

    const onListEnabledChange = useCallback(
      (_event, checked?: boolean) =>
        setState((state) => ({
          ...state,
          list_enabled: checked ?? false,
        })),
      [setState]
    );

    const biometricDisabled = useMemo(() => {
      return identityFeatureConfig?.biometric?.disabled ?? false;
    }, [identityFeatureConfig]);

    return (
      <ScreenContent>
        <ScreenTitle className={styles.widget}>
          <FormattedMessage id="BiometricConfigurationScreen.title" />
        </ScreenTitle>
        <ShowOnlyIfSIWEIsDisabled className={styles.widget}>
          <ScreenDescription className={styles.widget}>
            <FormattedMessage id="BiometricConfigurationScreen.description" />
          </ScreenDescription>
          <Widget className={styles.widget}>
            <WidgetTitle>
              <FormattedMessage id="BiometricConfigurationScreen.title" />
            </WidgetTitle>
            {biometricDisabled ? (
              <FeatureDisabledMessageBar messageID="FeatureConfig.disabled" />
            ) : null}
            <Toggle
              disabled={biometricDisabled}
              checked={state.enabled}
              onChange={onEnableChange}
              label={renderToString(
                "BiometricConfigurationScreen.enable.label"
              )}
              inlineLabel={true}
            />
            <Toggle
              disabled={!state.enabled || biometricDisabled}
              checked={state.list_enabled ?? false}
              onChange={onListEnabledChange}
              label={renderToString(
                "BiometricConfigurationScreen.list-enabled.label"
              )}
              inlineLabel={true}
            />
          </Widget>
        </ShowOnlyIfSIWEIsDisabled>
      </ScreenContent>
    );
  };

const BiometricConfigurationScreen: React.VFC =
  function BiometricConfigurationScreen() {
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
        <BiometricConfigurationContent
          form={form}
          identityFeatureConfig={featureConfig.effectiveFeatureConfig?.identity}
        />
      </FormContainer>
    );
  };

export default BiometricConfigurationScreen;
