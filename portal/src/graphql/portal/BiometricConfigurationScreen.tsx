import React, { useCallback, useContext } from "react";
import { useParams } from "react-router-dom";
import { produce } from "immer";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import { Toggle } from "@fluentui/react";
import cn from "classnames";
import { PortalAPIAppConfig } from "../../types";
import { clearEmptyObject } from "../../util/misc";
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
import styles from "./BiometricConfigurationScreen.module.scss";

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
  // eslint-disable-next-line complexity
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
}

const BiometricConfigurationContent: React.FC<BiometricConfigurationContentProps> = function BiometricConfigurationContent(
  props
) {
  const { state, setState } = props.form;

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

  return (
    <ScreenContent className={styles.root}>
      <ScreenTitle>
        <FormattedMessage id="BiometricConfigurationScreen.title" />
      </ScreenTitle>
      <ScreenDescription className={styles.widget}>
        <FormattedMessage id="BiometricConfigurationScreen.description" />
      </ScreenDescription>
      <Widget className={cn(styles.widget, styles.controlGroup)}>
        <WidgetTitle>
          <FormattedMessage id="BiometricConfigurationScreen.title" />
        </WidgetTitle>
        <Toggle
          className={styles.control}
          checked={state.enabled}
          onChange={onEnableChange}
          label={renderToString("BiometricConfigurationScreen.enable.label")}
          inlineLabel={true}
        />
        <Toggle
          className={styles.control}
          disabled={!state.enabled}
          checked={state.list_enabled ?? false}
          onChange={onListEnabledChange}
          label={renderToString(
            "BiometricConfigurationScreen.list-enabled.label"
          )}
          inlineLabel={true}
        />
      </Widget>
    </ScreenContent>
  );
};

const BiometricConfigurationScreen: React.FC = function BiometricConfigurationScreen() {
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
      <BiometricConfigurationContent form={form} />
    </FormContainer>
  );
};

export default BiometricConfigurationScreen;
