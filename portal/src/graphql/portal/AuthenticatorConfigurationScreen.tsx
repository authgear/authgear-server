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
  SecondaryAuthenticatorType,
  secondaryAuthenticatorTypes,
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

function makeAuthenticatorReasonable(state: FormState): FormState {
  return produce(state, (state) => {
    state.primary.forEach((primaryItem) => {
      state.secondary.forEach((secondaryItem) => {
        if (primaryItem.type === secondaryItem.type) {
          if (primaryItem.isChecked) {
            secondaryItem.isChecked = false;
            secondaryItem.isDisabled = true;
          } else {
            secondaryItem.isDisabled = false;
          }
        }
      });
    });
  });
}

interface FormState {
  primary: AuthenticatorTypeFormState<PrimaryAuthenticatorType>[];
  secondary: AuthenticatorTypeFormState<SecondaryAuthenticatorType>[];
}

// eslint-disable-next-line complexity
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
  // Passkey is controlled by the toggle.
  // So we do not show it in the primary authenticator list.
  primary = primary.filter((a) => a.type !== "passkey");

  const secondary: AuthenticatorTypeFormState<SecondaryAuthenticatorType>[] = (
    config.authentication?.secondary_authenticators ?? []
  ).map((t) => ({
    isChecked: true,
    isDisabled: false,
    type: t,
  }));
  for (const type of secondaryAuthenticatorTypes) {
    if (!secondary.some((t) => t.type === type)) {
      secondary.push({
        isChecked: false,
        isDisabled: primary.find((p) => p.type === type)?.isChecked ?? false,
        type,
      });
    }
  }

  return {
    primary,
    secondary,
  };
}

function constructConfig(
  config: PortalAPIAppConfig,
  _initialState: FormState,
  currentState: FormState,
  _effectiveConfig: PortalAPIAppConfig
): PortalAPIAppConfig {
  // eslint-disable-next-line complexity
  return produce(config, (config) => {
    config.authentication ??= {};

    function filterEnabled<T extends string>(
      s: AuthenticatorTypeFormState<T>[]
    ) {
      return s.filter((t) => t.isChecked).map((t) => t.type);
    }

    config.authentication.primary_authenticators = filterEnabled(
      currentState.primary
    );
    config.authentication.secondary_authenticators = filterEnabled(
      currentState.secondary
    );

    clearEmptyObject(config);
  });
}

const primaryAuthenticatorNameIds = {
  oob_otp_email: "AuthenticatorType.primary.oob-otp-email",
  oob_otp_sms: "AuthenticatorType.primary.oob-otp-phone",
  password: "AuthenticatorType.primary.password",
  passkey: "AuthenticatorType.primary.passkey",
};
const secondaryAuthenticatorNameIds = {
  totp: "AuthenticatorType.secondary.totp",
  oob_otp_email: "AuthenticatorType.secondary.oob-otp-email",
  oob_otp_sms: "AuthenticatorType.secondary.oob-otp-phone",
  password: "AuthenticatorType.secondary.password",
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

    const hasSecondaryFeatureDisabled = useMemo(() => {
      for (const key in featureDisabled["secondary"]) {
        if (featureDisabled["secondary"][key]) {
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
    const onSwapSecondaryAuthenticator = useCallback(
      (index1: number, index2: number) => {
        setState((prev) => ({
          ...prev,
          secondary: swap(prev.secondary, index1, index2),
        }));
      },
      [setState]
    );

    const onChangePrimaryAuthenticatorChecked = useCallback(
      (key: string, checked: boolean) => {
        setState((state) =>
          makeAuthenticatorReasonable(
            produce(state, (state) => {
              const t = state.primary.find((t) => t.type === key);
              if (t != null) {
                t.isChecked = checked;
              }
            })
          )
        );
      },
      [setState]
    );

    const onChangeSecondaryAuthenticatorChecked = useCallback(
      (key: string, checked: boolean) => {
        setState((state) =>
          makeAuthenticatorReasonable(
            produce(state, (state) => {
              const t = state.secondary.find((t) => t.type === key);
              if (t != null) {
                t.isChecked = checked;
              }
            })
          )
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

    const secondaryItems: PriorityListItem[] = useMemo(
      () =>
        state.secondary.map(({ type, isChecked, isDisabled }) => ({
          key: type,
          checked: isChecked,
          disabled: isDisabled || featureDisabled.secondary[type],
          content: (
            <div>
              <Text variant="small">
                <FormattedMessage id={secondaryAuthenticatorNameIds[type]} />
              </Text>
            </div>
          ),
        })),
      [state.secondary, featureDisabled.secondary]
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
        <Widget className={styles.widget}>
          <WidgetTitle>
            <FormattedMessage id="AuthenticatorConfigurationScreen.secondary-authenticators.title" />
          </WidgetTitle>
          {hasSecondaryFeatureDisabled ? (
            <FeatureDisabledMessageBar messageID="FeatureConfig.disabled" />
          ) : null}
          <PriorityList
            items={secondaryItems}
            checkedColumnLabel={renderToString(
              "AuthenticatorConfigurationScreen.columns.activate"
            )}
            keyColumnLabel={renderToString(
              "AuthenticatorConfigurationScreen.columns.authenticator"
            )}
            onChangeChecked={onChangeSecondaryAuthenticatorChecked}
            onSwap={onSwapSecondaryAuthenticator}
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
      constructInitialCurrentState: makeAuthenticatorReasonable,
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
