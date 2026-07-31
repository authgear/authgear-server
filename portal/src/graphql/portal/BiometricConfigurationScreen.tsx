import React, { useCallback, useMemo, useRef } from "react";
import { useParams } from "react-router-dom";
import cn from "classnames";
import { produce } from "immer";
import { Text } from "@radix-ui/themes";
import { FormattedMessage } from "../../intl";
import { PortalAPIAppConfig, IdentityFeatureConfig } from "../../types";
import { clearEmptyObject } from "../../util/misc";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import ScreenContent from "../../ScreenContent";
import { Toggle } from "../../components/v2/Toggle/Toggle";
import { SettingsSectionCard } from "../../components/v2/SettingsSectionCard/SettingsSectionCard";
import { SaveFunctionBar } from "../../components/v2/SaveFunctionBar/SaveFunctionBar";
import {
  AppConfigFormModel,
  useAppConfigForm,
} from "../../hook/useAppConfigForm";
import FormContainer from "../../FormContainer";
import { useFormContainerBaseContext } from "../../FormContainerBase";
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
    const { getIsDirty } = useFormContainerBaseContext();
    const isDirty = useMemo(() => getIsDirty(), [getIsDirty]);
    const contentWidthAnchorRef = useRef<HTMLDivElement>(null);

    const onEnableChange = useCallback(
      (checked: boolean) =>
        setState((state) => ({
          ...state,
          enabled: checked,
        })),
      [setState]
    );

    const onListEnabledChange = useCallback(
      (checked: boolean) =>
        setState((state) => ({
          ...state,
          list_enabled: checked,
        })),
      [setState]
    );

    const biometricDisabled = useMemo(() => {
      return identityFeatureConfig?.biometric?.disabled ?? false;
    }, [identityFeatureConfig]);

    return (
      <ScreenContent className={cn(isDirty ? styles.contentWithSaveBar : null)}>
        <div
          ref={contentWidthAnchorRef}
          className={cn(styles.widget, styles.pageHeader)}
        >
          <Text as="p" size="5" weight="bold" className={styles.pageTitle}>
            <FormattedMessage id="BiometricConfigurationScreen.title" />
          </Text>
          <Text as="p" size="2" color="gray" className={styles.pageDescription}>
            <FormattedMessage id="BiometricConfigurationScreen.description" />
          </Text>
        </div>
        <ShowOnlyIfSIWEIsDisabled className={styles.widget}>
          <SettingsSectionCard
            className={cn(
              styles.widget,
              isDirty && styles.settingsCardSaveBarClearance
            )}
            contentClassName="gap-4"
            title={
              <FormattedMessage id="BiometricConfigurationScreen.settings.label" />
            }
          >
            {biometricDisabled ? (
              <FeatureDisabledMessageBar messageID="FeatureConfig.disabled" />
            ) : null}
            <Toggle
              disabled={biometricDisabled}
              checked={state.enabled}
              onCheckedChange={onEnableChange}
              textWeight="medium"
              text={
                <FormattedMessage id="BiometricConfigurationScreen.enable.label" />
              }
            />
            {state.enabled ? (
              <Toggle
                disabled={biometricDisabled}
                checked={state.list_enabled ?? false}
                onCheckedChange={onListEnabledChange}
                textWeight="medium"
                text={
                  <FormattedMessage id="BiometricConfigurationScreen.list-enabled.label" />
                }
              />
            ) : null}
          </SettingsSectionCard>
        </ShowOnlyIfSIWEIsDisabled>
        <SaveFunctionBar anchorRef={contentWidthAnchorRef} />
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

    if (form.isLoading || featureConfig.isLoading) {
      return <ShowLoading />;
    }

    if (form.loadError) {
      return <ShowError error={form.loadError} onRetry={form.reload} />;
    }

    if (featureConfig.loadError) {
      return (
        <ShowError
          error={featureConfig.loadError}
          // eslint-disable-next-line @typescript-eslint/strict-void-return
          onRetry={featureConfig.refetch}
        />
      );
    }

    return (
      <FormContainer form={form} hideFooterComponent={true}>
        <BiometricConfigurationContent
          form={form}
          identityFeatureConfig={featureConfig.effectiveFeatureConfig?.identity}
        />
      </FormContainer>
    );
  };

export default BiometricConfigurationScreen;
