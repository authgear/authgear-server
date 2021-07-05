import React, { useCallback, useContext, useMemo } from "react";
import {
  Checkbox,
  Dropdown,
  IColumn,
  SelectionMode,
  Toggle,
  MessageBar,
} from "@fluentui/react";
import produce from "immer";
import deepEqual from "deep-equal";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import cn from "classnames";
import DetailsListWithOrdering from "../../DetailsListWithOrdering";
import { swap } from "../../OrderButtons";
import FormTextField from "../../FormTextField";
import {
  PortalAPIAppConfig,
  PortalAPIFeatureConfig,
  PrimaryAuthenticatorType,
  primaryAuthenticatorTypes,
  SecondaryAuthenticationMode,
  secondaryAuthenticationModes,
  SecondaryAuthenticatorType,
  secondaryAuthenticatorTypes,
} from "../../types";
import {
  useCheckbox,
  useDropdown,
  useIntegerTextField,
} from "../../hook/useInput";
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

import styles from "./AuthenticatorConfigurationScreen.module.scss";
import { useAppFeatureConfigQuery } from "./query/appFeatureConfigQuery";

interface AuthenticatorTypeFormState<
  T = PrimaryAuthenticatorType | SecondaryAuthenticatorType
> {
  isEnabled: boolean;
  type: T;
}

interface FormState {
  primary: AuthenticatorTypeFormState<PrimaryAuthenticatorType>[];
  secondary: AuthenticatorTypeFormState<SecondaryAuthenticatorType>[];

  mfaMode: SecondaryAuthenticationMode;
  numRecoveryCode: number;
  allowListRecoveryCode: boolean;
  disableDeviceToken: boolean;
}

// eslint-disable-next-line complexity
function constructFormState(config: PortalAPIAppConfig): FormState {
  const primary: AuthenticatorTypeFormState<PrimaryAuthenticatorType>[] = (
    config.authentication?.primary_authenticators ?? []
  ).map((t) => ({
    isEnabled: true,
    type: t,
  }));
  for (const type of primaryAuthenticatorTypes) {
    if (!primary.some((t) => t.type === type)) {
      primary.push({ isEnabled: false, type });
    }
  }
  const secondary: AuthenticatorTypeFormState<SecondaryAuthenticatorType>[] = (
    config.authentication?.secondary_authenticators ?? []
  ).map((t) => ({
    isEnabled: true,
    type: t,
  }));
  for (const type of secondaryAuthenticatorTypes) {
    if (!secondary.some((t) => t.type === type)) {
      secondary.push({ isEnabled: false, type });
    }
  }

  return {
    primary,
    secondary,
    mfaMode:
      config.authentication?.secondary_authentication_mode ?? "if_exists",
    numRecoveryCode: config.authentication?.recovery_code?.count ?? 16,
    allowListRecoveryCode:
      config.authentication?.recovery_code?.list_enabled ?? false,
    disableDeviceToken: config.authentication?.device_token?.disabled ?? false,
  };
}

function constructConfig(
  config: PortalAPIAppConfig,
  initialState: FormState,
  currentState: FormState
): PortalAPIAppConfig {
  // eslint-disable-next-line complexity
  return produce(config, (config) => {
    config.authentication ??= {};
    config.authentication.recovery_code ??= {};
    config.authentication.device_token ??= {};

    function filterEnabled<T extends string>(
      s: AuthenticatorTypeFormState<T>[]
    ) {
      return s.filter((t) => t.isEnabled).map((t) => t.type);
    }

    if (
      !deepEqual(
        filterEnabled(currentState.primary),
        filterEnabled(initialState.primary),
        { strict: true }
      )
    ) {
      config.authentication.primary_authenticators = filterEnabled(
        currentState.primary
      );
    }
    if (
      !deepEqual(
        filterEnabled(currentState.secondary),
        filterEnabled(initialState.secondary),
        { strict: true }
      )
    ) {
      config.authentication.secondary_authenticators = filterEnabled(
        currentState.secondary
      );
    }

    if (initialState.mfaMode !== currentState.mfaMode) {
      config.authentication.secondary_authentication_mode =
        currentState.mfaMode;
    }
    if (initialState.numRecoveryCode !== currentState.numRecoveryCode) {
      config.authentication.recovery_code.count = currentState.numRecoveryCode;
    }
    if (
      initialState.allowListRecoveryCode !== currentState.allowListRecoveryCode
    ) {
      config.authentication.recovery_code.list_enabled =
        currentState.allowListRecoveryCode;
    }

    if (initialState.disableDeviceToken !== currentState.disableDeviceToken) {
      config.authentication.device_token.disabled =
        currentState.disableDeviceToken;
    }

    clearEmptyObject(config);
  });
}

const ALL_REQUIRE_MFA_OPTIONS: SecondaryAuthenticationMode[] = [
  ...secondaryAuthenticationModes,
];
const HIDDEN_REQUIRE_MFA_OPTIONS: SecondaryAuthenticationMode[] = [
  "if_requested",
];

const primaryAuthenticatorNameIds = {
  oob_otp_email: "AuthenticatorType.primary.oob-otp-email",
  oob_otp_sms: "AuthenticatorType.primary.oob-otp-sms",
  password: "AuthenticatorType.primary.password",
};
const secondaryAuthenticatorNameIds = {
  totp: "AuthenticatorType.secondary.totp",
  oob_otp_email: "AuthenticatorType.secondary.oob-otp-email",
  oob_otp_sms: "AuthenticatorType.secondary.oob-otp-sms",
  password: "AuthenticatorType.secondary.password",
};

type AuthenticatorColumnItem = (
  | { kind: "primary"; type: PrimaryAuthenticatorType }
  | { kind: "secondary"; type: SecondaryAuthenticatorType }
) & { isEnabled: boolean };

interface AuthenticatorCheckboxProps {
  disabled: boolean;
  item: AuthenticatorColumnItem;
  onChange: (item: AuthenticatorColumnItem, checked: boolean) => void;
}

const AuthenticatorCheckbox: React.FC<AuthenticatorCheckboxProps> =
  function AuthenticatorCheckbox(props) {
    const { disabled, item, onChange } = props;

    const onCheckboxChange = useCallback(
      (_event, checked?: boolean) => onChange(item, checked ?? false),
      [item, onChange]
    );

    return (
      <Checkbox
        checked={item.isEnabled}
        onChange={onCheckboxChange}
        disabled={disabled}
      />
    );
  };

interface AuthenticationAuthenticatorSettingsContentProps {
  form: AppConfigFormModel<FormState>;
  featureConfig?: PortalAPIFeatureConfig;
}

const AuthenticationAuthenticatorSettingsContent: React.FC<AuthenticationAuthenticatorSettingsContentProps> =
  function AuthenticationAuthenticatorSettingsContent(props) {
    const { state, setState } = props.form;

    const { featureConfig } = props;

    const { renderToString } = useContext(Context);

    const authenticatorColumns: IColumn[] = [
      {
        key: "activated",
        fieldName: "activated",
        name: renderToString(
          "AuthenticatorConfigurationScreen.columns.activate"
        ),
        className: styles.authenticatorColumn,
        minWidth: 120,
        maxWidth: 120,
      },
      {
        key: "key",
        fieldName: "key",
        name: renderToString(
          "AuthenticatorConfigurationScreen.columns.authenticator"
        ),
        className: styles.authenticatorColumn,
        minWidth: 300,
        maxWidth: 300,
      },
    ];

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

    const renderSecondaryAuthenticatorMode = useCallback(
      (key: SecondaryAuthenticationMode) => {
        const messageIdMap: Record<SecondaryAuthenticationMode, string> = {
          required:
            "AuthenticatorConfigurationScreen.policy.require-mfa.required",
          if_exists:
            "AuthenticatorConfigurationScreen.policy.require-mfa.if-exists",
          if_requested:
            "AuthenticatorConfigurationScreen.policy.require-mfa.if-requested",
        };

        return renderToString(messageIdMap[key]);
      },
      [renderToString]
    );

    const { options: requireMFAOptions, onChange: onRequireMFAOptionChange } =
      useDropdown(
        ALL_REQUIRE_MFA_OPTIONS,
        (option) => {
          setState((prev) => ({
            ...prev,
            mfaMode: option,
          }));
        },
        state.mfaMode,
        renderSecondaryAuthenticatorMode,
        // NOTE: not supported yet
        new Set(HIDDEN_REQUIRE_MFA_OPTIONS)
      );

    const { onChange: onRecoveryCodeNumberChange } = useIntegerTextField(
      (value) => {
        setState((prev) => ({
          ...prev,
          numRecoveryCode: Number(value),
        }));
      }
    );

    const { onChange: onAllowRetrieveRecoveryCodeChange } = useCheckbox(
      (checked: boolean) => {
        setState((prev) => ({
          ...prev,
          allowListRecoveryCode: checked,
        }));
      }
    );

    const { onChange: onDisableDeviceTokenChange } = useCheckbox(
      (checked: boolean) => {
        setState((prev) => ({
          ...prev,
          disableDeviceToken: checked,
        }));
      }
    );

    const onPrimarySwapClicked = useCallback(
      (index1: number, index2: number) => {
        setState((prev) => ({
          ...prev,
          primary: swap(prev.primary, index1, index2),
        }));
      },
      [setState]
    );
    const onSecondarySwapClicked = useCallback(
      (index1: number, index2: number) => {
        setState((prev) => ({
          ...prev,
          secondary: swap(prev.secondary, index1, index2),
        }));
      },
      [setState]
    );

    const onAuthenticatorEnabledChange = useCallback(
      (item: AuthenticatorColumnItem, checked: boolean) =>
        setState((state) =>
          produce(state, (state) => {
            let t: AuthenticatorTypeFormState | undefined;
            switch (item.kind) {
              case "primary":
                t = state.primary.find((t) => t.type === item.type);
                break;
              case "secondary":
                t = state.secondary.find((t) => t.type === item.type);
                break;
            }
            if (t) {
              t.isEnabled = checked;
            }
          })
        ),
      [setState]
    );

    const onRenderColumnClassName = useCallback(
      (item: AuthenticatorColumnItem, _index?: number, _?: IColumn) => {
        const disabled = featureDisabled[item.kind][item.type];
        return cn({ [styles.readOnly]: disabled });
      },
      [featureDisabled]
    );

    const onRenderColumn = useCallback(
      (item: AuthenticatorColumnItem, _index?: number, column?: IColumn) => {
        const disabled = featureDisabled[item.kind][item.type];
        switch (column?.key) {
          case "activated":
            return (
              <AuthenticatorCheckbox
                disabled={disabled}
                item={item}
                onChange={onAuthenticatorEnabledChange}
              />
            );

          case "key": {
            let nameId: string;
            switch (item.kind) {
              case "primary":
                nameId = primaryAuthenticatorNameIds[item.type];
                break;
              case "secondary":
                nameId = secondaryAuthenticatorNameIds[item.type];
                break;
            }
            return (
              <span>
                <FormattedMessage id={nameId} />
              </span>
            );
          }

          default:
            return null;
        }
      },
      [onAuthenticatorEnabledChange, featureDisabled]
    );

    const renderPrimaryAriaLabel = React.useCallback(
      (index?: number): string => {
        return index != null
          ? renderToString(
              primaryAuthenticatorNameIds[state.primary[index].type]
            )
          : "";
      },
      [state.primary, renderToString]
    );
    const renderSecondaryAriaLabel = React.useCallback(
      (index?: number): string => {
        return index != null
          ? renderToString(
              secondaryAuthenticatorNameIds[state.secondary[index].type]
            )
          : "";
      },
      [state.secondary, renderToString]
    );

    const primaryItems: AuthenticatorColumnItem[] = useMemo(
      () =>
        state.primary.map(({ type, isEnabled }) => ({
          kind: "primary",
          type,
          isEnabled,
        })),
      [state.primary]
    );

    const secondaryItems: AuthenticatorColumnItem[] = useMemo(
      () =>
        state.secondary.map(({ type, isEnabled }) => ({
          kind: "secondary",
          type,
          isEnabled,
        })),
      [state.secondary]
    );

    return (
      <ScreenContent className={styles.root}>
        <ScreenTitle>
          <FormattedMessage id="AuthenticatorConfigurationScreen.title" />
        </ScreenTitle>
        <ScreenDescription className={styles.widget}>
          <FormattedMessage id="AuthenticatorConfigurationScreen.description" />
        </ScreenDescription>
        <Widget className={styles.widget}>
          <WidgetTitle>
            <FormattedMessage id="AuthenticatorConfigurationScreen.primary-authenticators.title" />
          </WidgetTitle>
          {hasPrimaryFeatureDisabled && (
            <MessageBar>
              <FormattedMessage
                id="FeatureConfig.disabled"
                values={{
                  HREF: "../settings/subscription",
                }}
              />
            </MessageBar>
          )}
          <DetailsListWithOrdering
            items={primaryItems}
            columns={authenticatorColumns}
            onRenderItemColumnClassName={onRenderColumnClassName}
            onRenderItemColumn={onRenderColumn}
            onSwapClicked={onPrimarySwapClicked}
            selectionMode={SelectionMode.none}
            renderAriaLabel={renderPrimaryAriaLabel}
          />
        </Widget>
        <Widget className={styles.widget}>
          <WidgetTitle>
            <FormattedMessage id="AuthenticatorConfigurationScreen.secondary-authenticators.title" />
          </WidgetTitle>
          {hasSecondaryFeatureDisabled && (
            <MessageBar>
              <FormattedMessage
                id="FeatureConfig.disabled"
                values={{
                  HREF: "../settings/subscription",
                }}
              />
            </MessageBar>
          )}
          <DetailsListWithOrdering
            items={secondaryItems}
            columns={authenticatorColumns}
            onRenderItemColumnClassName={onRenderColumnClassName}
            onRenderItemColumn={onRenderColumn}
            onSwapClicked={onSecondarySwapClicked}
            selectionMode={SelectionMode.none}
            renderAriaLabel={renderSecondaryAriaLabel}
          />
        </Widget>
        <Widget className={cn(styles.widget, styles.controlGroup)}>
          <WidgetTitle>
            <FormattedMessage id="AuthenticatorConfigurationScreen.policy.title" />
          </WidgetTitle>
          <Dropdown
            className={styles.control}
            label={renderToString(
              "AuthenticatorConfigurationScreen.policy.require-mfa"
            )}
            options={requireMFAOptions}
            selectedKey={state.mfaMode}
            onChange={onRequireMFAOptionChange}
          />
          <FormTextField
            parentJSONPointer="/authentication/recovery_code"
            fieldName="count"
            fieldNameMessageID="AuthenticatorConfigurationScreen.policy.recovery-code-number"
            className={styles.control}
            value={String(state.numRecoveryCode)}
            onChange={onRecoveryCodeNumberChange}
          />
          <Toggle
            className={styles.control}
            inlineLabel={true}
            label={
              <FormattedMessage id="AuthenticatorConfigurationScreen.policy.allow-retrieve-recovery-code" />
            }
            checked={state.allowListRecoveryCode}
            onChange={onAllowRetrieveRecoveryCodeChange}
          />
          <Toggle
            className={styles.control}
            inlineLabel={true}
            label={
              <FormattedMessage id="AuthenticatorConfigurationScreen.policy.disable-device-token" />
            }
            checked={state.disableDeviceToken}
            onChange={onDisableDeviceTokenChange}
          />
        </Widget>
      </ScreenContent>
    );
  };

const AuthenticatorConfigurationScreen: React.FC =
  function AuthenticatorConfigurationScreen() {
    const { appID } = useParams();
    const form = useAppConfigForm(appID, constructFormState, constructConfig);

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
          form={form}
          featureConfig={featureConfig.effectiveFeatureConfig ?? undefined}
        />
      </FormContainer>
    );
  };

export default AuthenticatorConfigurationScreen;
