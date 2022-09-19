import React, { useCallback, useContext, useMemo } from "react";
import { Text } from "@fluentui/react";
import produce from "immer";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import { swap } from "../../OrderButtons";
import {
  PortalAPIAppConfig,
  PortalAPIFeatureConfig,
  PrimaryAuthenticatorType,
  primaryAuthenticatorTypes,
} from "../../types";
import { clearEmptyObject } from "../../util/misc";
import { useParams } from "react-router-dom";
import {
  AppConfigFormModel,
  useAppConfigForm,
} from "../../hook/useAppConfigForm";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import ScreenContent from "../../ScreenContent";
import ScreenTitle from "../../ScreenTitle";
import ScreenDescription from "../../ScreenDescription";
import WidgetTitle from "../../WidgetTitle";
import Widget from "../../Widget";
import FormContainer from "../../FormContainer";
import FeatureDisabledMessageBar from "./FeatureDisabledMessageBar";
import PriorityList, { PriorityListItem } from "../../PriorityList";
import Link from "../../Link";

import styles from "./AuthenticatorConfigurationScreen.module.css";
import { useAppFeatureConfigQuery } from "./query/appFeatureConfigQuery";

interface AuthenticatorTypeFormState<T> {
  isChecked: boolean;
  isDisabled: boolean;
  type: T;
}

interface FormState {
  primary: AuthenticatorTypeFormState<PrimaryAuthenticatorType>[];
}

function constructFormState(config: PortalAPIAppConfig): FormState {
  let primary: AuthenticatorTypeFormState<PrimaryAuthenticatorType>[] = (
    config.authentication?.primary_authenticators ?? []
  ).map((t) => ({
    isChecked: true,
    isDisabled: false,
    type: t,
  }));
  for (const type of primaryAuthenticatorTypes) {
    if (!primary.some((t) => t.type === type)) {
      primary.push({ isChecked: false, isDisabled: false, type });
    }
  }

  // Passkey is configured in another screen.
  // So we do not show it in the primary authenticator list.
  primary = primary.filter((a) => a.type !== "passkey");

  return {
    primary,
  };
}

function constructConfig(
  config: PortalAPIAppConfig,
  _initialState: FormState,
  currentState: FormState,
  effectiveConfig: PortalAPIAppConfig
): PortalAPIAppConfig {
  return produce(config, (config) => {
    config.authentication ??= {};

    const primary = currentState.primary
      .filter((t) => t.isChecked)
      .map((t) => t.type);

    if (
      effectiveConfig.authentication?.primary_authenticators?.includes(
        "passkey"
      ) === true
    ) {
      primary.push("passkey");
    }

    config.authentication.primary_authenticators = primary;

    clearEmptyObject(config);
  });
}

const primaryAuthenticatorNameIds = {
  oob_otp_email: "AuthenticatorType.primary.oob-otp-email",
  oob_otp_sms: "AuthenticatorType.primary.oob-otp-phone",
  password: "AuthenticatorType.primary.password",
  passkey: "AuthenticatorType.primary.passkey",
};

interface AuthenticationAuthenticatorSettingsContentProps {
  appID: string;
  form: AppConfigFormModel<FormState>;
  featureConfig?: PortalAPIFeatureConfig;
}

const AuthenticationAuthenticatorSettingsContent: React.VFC<AuthenticationAuthenticatorSettingsContentProps> =
  function AuthenticationAuthenticatorSettingsContent(props) {
    const { appID, featureConfig } = props;

    const { state, setState, effectiveConfig } = props.form;

    const { renderToString } = useContext(Context);

    const featureDisabled: Record<
      string,
      Record<string, boolean>
    > = useMemo(() => {
      return {
        primary: {
          oob_otp_sms:
            featureConfig?.identity?.login_id?.types?.phone?.disabled ?? false,
        },
        secondary: {
          oob_otp_sms:
            featureConfig?.authentication?.secondary_authenticators?.oob_otp_sms
              ?.disabled ?? false,
        },
      };
    }, [featureConfig]);

    const hasPrimaryFeatureDisabled = useMemo(() => {
      for (const key in featureDisabled["primary"]) {
        if (featureDisabled["primary"][key]) {
          return true;
        }
      }
      return false;
    }, [featureDisabled]);

    const isPhoneLoginIdDisabled = useMemo(
      () =>
        effectiveConfig.identity?.login_id?.keys?.find(
          (t) => t.type === "phone"
        ) == null,
      [effectiveConfig.identity?.login_id?.keys]
    );

    const onSwapPrimaryAuthenticator = useCallback(
      (index1: number, index2: number) => {
        setState((prev) => ({
          ...prev,
          primary: swap(prev.primary, index1, index2),
        }));
      },
      [setState]
    );

    const onChangePrimaryAuthenticatorChecked = useCallback(
      (key: string, checked: boolean) => {
        setState((state) =>
          produce(state, (state) => {
            const t = state.primary.find((t) => t.type === key);
            if (t != null) {
              t.isChecked = checked;
            }
          })
        );
      },
      [setState]
    );

    const primaryItems: PriorityListItem[] = useMemo(
      () =>
        state.primary.map(({ type, isChecked, isDisabled }) => ({
          key: type,
          checked: isChecked,
          disabled: isDisabled || featureDisabled.primary[type],
          content: (
            <div>
              <Text variant="small" block={true}>
                <FormattedMessage id={primaryAuthenticatorNameIds[type]} />
              </Text>
              {type === "oob_otp_sms" && isChecked && isPhoneLoginIdDisabled ? (
                <Link
                  to={`/project/${appID}/configuration/authentication/login-id`}
                >
                  <FormattedMessage id="AuthenticatorHint.primary.oob-otp-phone" />
                </Link>
              ) : undefined}
            </div>
          ),
        })),
      [state.primary, featureDisabled.primary, isPhoneLoginIdDisabled, appID]
    );

    return (
      <ScreenContent>
        <ScreenTitle className={styles.widget}>
          <FormattedMessage id="AuthenticatorConfigurationScreen.title" />
        </ScreenTitle>
        <ScreenDescription className={styles.widget}>
          <FormattedMessage id="AuthenticatorConfigurationScreen.description" />
        </ScreenDescription>
        <Widget className={styles.widget}>
          <WidgetTitle>
            <FormattedMessage id="AuthenticatorConfigurationScreen.primary-authenticators.title" />
          </WidgetTitle>
          {hasPrimaryFeatureDisabled ? (
            <FeatureDisabledMessageBar messageID="FeatureConfig.disabled" />
          ) : null}
          <PriorityList
            items={primaryItems}
            checkedColumnLabel={renderToString(
              "AuthenticatorConfigurationScreen.columns.activate"
            )}
            keyColumnLabel={renderToString(
              "AuthenticatorConfigurationScreen.columns.authenticator"
            )}
            onChangeChecked={onChangePrimaryAuthenticatorChecked}
            onSwap={onSwapPrimaryAuthenticator}
          />
        </Widget>
      </ScreenContent>
    );
  };

const AuthenticatorConfigurationScreen: React.VFC =
  function AuthenticatorConfigurationScreen() {
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
          error={form.loadError}
          onRetry={() => {
            form.reload();
            featureConfig.refetch().finally(() => {});
          }}
        />
      );
    }

    return (
      <FormContainer form={form}>
        <AuthenticationAuthenticatorSettingsContent
          appID={appID}
          form={form}
          featureConfig={featureConfig.effectiveFeatureConfig ?? undefined}
        />
      </FormContainer>
    );
  };

export default AuthenticatorConfigurationScreen;
