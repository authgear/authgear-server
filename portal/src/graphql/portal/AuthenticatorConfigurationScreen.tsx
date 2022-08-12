import React, { useCallback, useContext, useMemo } from "react";
import cn from "classnames";
import {
  Checkbox,
  Dropdown,
  IColumn,
  SelectionMode,
  Toggle,
  DetailsList,
  Link,
  Text,
  TooltipHost,
} from "@fluentui/react";
import produce from "immer";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import OrderButtons, { swap } from "../../OrderButtons";
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
import { useCheckbox, useDropdown } from "../../hook/useInput";
import { clearEmptyObject } from "../../util/misc";
import { parseIntegerAllowLeadingZeros } from "../../util/input";
import { useParams } from "react-router-dom";
import {
  AppConfigFormModel,
  useAppConfigForm,
} from "../../hook/useAppConfigForm";
import { useTooltipTargetElement } from "../../Tooltip";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import ScreenContent from "../../ScreenContent";
import ScreenTitle from "../../ScreenTitle";
import ScreenDescription from "../../ScreenDescription";
import WidgetTitle from "../../WidgetTitle";
import WidgetDescription from "../../WidgetDescription";
import Widget from "../../Widget";
import FormContainer from "../../FormContainer";
import FeatureDisabledMessageBar from "./FeatureDisabledMessageBar";

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

  mfaMode: SecondaryAuthenticationMode;
  recoveryCodeEnabled: boolean;
  numRecoveryCode: number | undefined;
  allowListRecoveryCode: boolean;
  disableDeviceToken: boolean;
  passkeyChecked: boolean;
  passkeyDisabled: boolean;
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

  const passkeyIndex =
    config.authentication?.primary_authenticators?.indexOf("passkey");
  const passkeyChecked = passkeyIndex != null && passkeyIndex >= 0;

  const passkeyDisabled = !(
    config.authentication?.identities?.includes("login_id") ?? true
  );

  return {
    primary,
    secondary,
    mfaMode:
      config.authentication?.secondary_authentication_mode ?? "if_exists",
    numRecoveryCode: config.authentication?.recovery_code?.count,
    recoveryCodeEnabled: !(
      config.authentication?.recovery_code?.disabled ?? false
    ),
    allowListRecoveryCode:
      config.authentication?.recovery_code?.list_enabled ?? false,
    disableDeviceToken: config.authentication?.device_token?.disabled ?? false,
    passkeyChecked,
    passkeyDisabled,
  };
}

function constructConfig(
  config: PortalAPIAppConfig,
  _initialState: FormState,
  currentState: FormState,
  effectiveConfig: PortalAPIAppConfig
): PortalAPIAppConfig {
  // eslint-disable-next-line complexity
  return produce(config, (config) => {
    config.authentication ??= {};
    config.authentication.recovery_code ??= {};
    config.authentication.device_token ??= {};

    function filterEnabled<T extends string>(
      s: AuthenticatorTypeFormState<T>[]
    ) {
      return s.filter((t) => t.isChecked).map((t) => t.type);
    }

    function setEnable<T extends string>(
      arr: T[],
      value: T,
      enabled: boolean
    ): T[] {
      const index = arr.indexOf(value);

      if (enabled) {
        if (index >= 0) {
          return arr;
        }
        return [...arr, value];
      }

      if (index < 0) {
        return arr;
      }
      return [...arr.slice(0, index), ...arr.slice(index + 1)];
    }

    config.authentication.primary_authenticators = filterEnabled(
      currentState.primary
    );
    config.authentication.secondary_authenticators = filterEnabled(
      currentState.secondary
    );

    // Construct primary_authenticators and identities
    // We do not offer any control to modify identities in this screen,
    // so we read from effectiveConfig.authentication?.identities
    // On the other hand, we always write config.authentication.primary_authenticators,
    // so we read from it.
    if (currentState.passkeyChecked) {
      config.authentication.primary_authenticators = setEnable(
        config.authentication.primary_authenticators,
        "passkey",
        true
      );
      config.authentication.identities = setEnable(
        effectiveConfig.authentication?.identities ?? [],
        "passkey",
        true
      );
    } else {
      config.authentication.primary_authenticators = setEnable(
        config.authentication.primary_authenticators,
        "passkey",
        false
      );
      config.authentication.identities = setEnable(
        effectiveConfig.authentication?.identities ?? [],
        "passkey",
        false
      );
    }

    config.authentication.secondary_authentication_mode = currentState.mfaMode;
    config.authentication.recovery_code.disabled =
      !currentState.recoveryCodeEnabled;
    config.authentication.recovery_code.count = currentState.numRecoveryCode;
    config.authentication.recovery_code.list_enabled =
      currentState.allowListRecoveryCode;
    config.authentication.device_token.disabled =
      currentState.disableDeviceToken;

    clearEmptyObject(config);
  });
}

const ALL_REQUIRE_MFA_OPTIONS: SecondaryAuthenticationMode[] = [
  ...secondaryAuthenticationModes,
];

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

type AuthenticatorColumnItem = (
  | { kind: "primary"; type: PrimaryAuthenticatorType }
  | { kind: "secondary"; type: SecondaryAuthenticatorType }
) & { isChecked: boolean; isDisabled: boolean };

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
        checked={item.isChecked}
        onChange={onCheckboxChange}
        disabled={disabled && !item.isChecked}
      />
    );
  };

interface AuthenticationAuthenticatorSettingsContentProps {
  form: AppConfigFormModel<FormState>;
  featureConfig?: PortalAPIFeatureConfig;
}

const AuthenticationAuthenticatorSettingsContent: React.FC<AuthenticationAuthenticatorSettingsContentProps> =
  function AuthenticationAuthenticatorSettingsContent(props) {
    const { state, setState, effectiveConfig } = props.form;

    const { featureConfig } = props;

    const tooltipResult = useTooltipTargetElement();
    const passkeyTooltipProps = useMemo(() => {
      return {
        targetElement: tooltipResult.targetElement,
      };
    }, [tooltipResult.targetElement]);

    const { renderToString } = useContext(Context);

    const authenticatorColumns: IColumn[] = [
      {
        key: "activated",
        fieldName: "activated",
        name: renderToString(
          "AuthenticatorConfigurationScreen.columns.activate"
        ),
        className: styles.authenticatorColumn,
        minWidth: 80,
        maxWidth: 80,
      },
      {
        key: "key",
        fieldName: "key",
        name: renderToString(
          "AuthenticatorConfigurationScreen.columns.authenticator"
        ),
        className: styles.authenticatorColumn,
        minWidth: 280,
        maxWidth: 280,
      },
      {
        key: "order",
        fieldName: "order",
        name: renderToString("DetailsListWithOrdering.order"),
        minWidth: 100,
        maxWidth: 100,
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

    const isSecondaryAuthenticatorDisabled = useMemo(
      () => state.mfaMode === "disabled",
      [state.mfaMode]
    );

    const isPhoneLoginIdDisabled = useMemo(
      () =>
        effectiveConfig.identity?.login_id?.keys?.find(
          (t) => t.type === "phone"
        ) == null,
      [effectiveConfig.identity?.login_id?.keys]
    );

    const renderSecondaryAuthenticatorMode = useCallback(
      (key: SecondaryAuthenticationMode) => {
        const messageIdMap: Record<SecondaryAuthenticationMode, string> = {
          disabled:
            "AuthenticatorConfigurationScreen.secondary-authenticators.mode.disabled",
          required:
            "AuthenticatorConfigurationScreen.secondary-authenticators.mode.required",
          if_exists:
            "AuthenticatorConfigurationScreen.secondary-authenticators.mode.if-exists",
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
        renderSecondaryAuthenticatorMode
      );

    const onRecoveryCodeNumberChange = useCallback(
      (_, value?: string) => {
        setState((prev) => ({
          ...prev,
          numRecoveryCode: parseIntegerAllowLeadingZeros(value),
        }));
      },
      [setState]
    );

    const { onChange: onChangeRecoveryCodeEnabled } = useCheckbox(
      (checked: boolean) => {
        setState((prev) => ({
          ...prev,
          recoveryCodeEnabled: checked,
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

    const { onChange: onChangePasskeyChecked } = useCheckbox(
      (checked: boolean) => {
        setState((prev) => ({
          ...prev,
          passkeyChecked: checked,
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
        setState((state) => {
          const s = produce(state, (state) => {
            let t:
              | AuthenticatorTypeFormState<
                  PrimaryAuthenticatorType | SecondaryAuthenticatorType
                >
              | undefined;
            switch (item.kind) {
              case "primary":
                t = state.primary.find((t) => t.type === item.type);
                break;
              case "secondary":
                t = state.secondary.find((t) => t.type === item.type);
                break;
            }
            if (t) {
              t.isChecked = checked;
            }
          });

          return makeAuthenticatorReasonable(s);
        }),
      [setState]
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
        state.primary.map(({ type, isChecked, isDisabled }) => ({
          kind: "primary",
          type,
          isChecked,
          isDisabled,
        })),
      [state.primary]
    );

    const secondaryItems: AuthenticatorColumnItem[] = useMemo(
      () =>
        state.secondary.map(({ type, isChecked, isDisabled }) => ({
          kind: "secondary",
          type,
          isChecked,
          isDisabled,
        })),
      [state.secondary]
    );

    const onRenderPrimaryColumn = useCallback(
      (item: AuthenticatorColumnItem, index?: number, column?: IColumn) => {
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
              <div className={styles.authenticatorColumnSpanRow}>
                <span>
                  <FormattedMessage id={nameId} />
                </span>
                {item.type === "oob_otp_sms" &&
                item.isChecked &&
                isPhoneLoginIdDisabled ? (
                  <span>
                    <Link href="./login-id">
                      <FormattedMessage id="AuthenticatorHint.primary.oob-otp-phone" />
                    </Link>
                  </span>
                ) : null}
              </div>
            );
          }

          case "order": {
            return (
              <OrderButtons
                disabled={disabled}
                index={index}
                itemCount={primaryItems.length}
                onSwapClicked={onPrimarySwapClicked}
                renderAriaLabel={renderPrimaryAriaLabel}
              />
            );
          }

          default:
            return null;
        }
      },
      [
        featureDisabled,
        onAuthenticatorEnabledChange,
        isPhoneLoginIdDisabled,
        primaryItems.length,
        onPrimarySwapClicked,
        renderPrimaryAriaLabel,
      ]
    );

    const onRenderSecondaryColumn = useCallback(
      (item: AuthenticatorColumnItem, index?: number, column?: IColumn) => {
        const disabled =
          featureDisabled[item.kind][item.type] || item.isDisabled;
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

          case "order": {
            return (
              <OrderButtons
                disabled={disabled}
                index={index}
                itemCount={secondaryItems.length}
                onSwapClicked={onSecondarySwapClicked}
                renderAriaLabel={renderSecondaryAriaLabel}
              />
            );
          }

          default:
            return null;
        }
      },
      [
        onAuthenticatorEnabledChange,
        featureDisabled,
        onSecondarySwapClicked,
        secondaryItems.length,
        renderSecondaryAriaLabel,
      ]
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
          {hasPrimaryFeatureDisabled && (
            <FeatureDisabledMessageBar>
              <FormattedMessage
                id="FeatureConfig.disabled"
                values={{
                  planPagePath: "./../../../billing",
                }}
              />
            </FeatureDisabledMessageBar>
          )}
          {state.passkeyDisabled ? (
            <TooltipHost
              content={
                <FormattedMessage id="AuthenticatorConfigurationScreen.passkey.tooltip" />
              }
              tooltipProps={passkeyTooltipProps}
            >
              <Toggle
                id={tooltipResult.id}
                ref={tooltipResult.setRef}
                label={renderToString(
                  "AuthenticatorConfigurationScreen.passkey.title"
                )}
                onText={renderToString(
                  "AuthenticatorConfigurationScreen.passkey.on"
                )}
                offText={renderToString(
                  "AuthenticatorConfigurationScreen.passkey.off"
                )}
                disabled={state.passkeyDisabled}
                checked={state.passkeyChecked}
                onChange={onChangePasskeyChecked}
              />
              <Text as="p" variant="medium" block={true}>
                <FormattedMessage id="AuthenticatorConfigurationScreen.passkey.description" />
              </Text>
            </TooltipHost>
          ) : (
            <div>
              <Toggle
                label={renderToString(
                  "AuthenticatorConfigurationScreen.passkey.title"
                )}
                onText={renderToString(
                  "AuthenticatorConfigurationScreen.passkey.on"
                )}
                offText={renderToString(
                  "AuthenticatorConfigurationScreen.passkey.off"
                )}
                disabled={state.passkeyDisabled}
                checked={state.passkeyChecked}
                onChange={onChangePasskeyChecked}
              />
              <Text as="p" variant="medium" block={true}>
                <FormattedMessage id="AuthenticatorConfigurationScreen.passkey.description" />
              </Text>
            </div>
          )}
          <DetailsList
            items={primaryItems}
            columns={authenticatorColumns}
            onRenderItemColumn={onRenderPrimaryColumn}
            selectionMode={SelectionMode.none}
          />
        </Widget>
        <Widget className={styles.widget}>
          <WidgetTitle>
            <FormattedMessage id="AuthenticatorConfigurationScreen.secondary-authenticators.title" />
          </WidgetTitle>
          <Dropdown
            label={renderToString(
              "AuthenticatorConfigurationScreen.secondary-authenticators.mode.label"
            )}
            options={requireMFAOptions}
            selectedKey={state.mfaMode}
            onChange={onRequireMFAOptionChange}
          />
          <div
            className={cn({
              [styles.readOnly]: isSecondaryAuthenticatorDisabled,
            })}
          >
            {hasSecondaryFeatureDisabled && (
              <FeatureDisabledMessageBar>
                <FormattedMessage
                  id="FeatureConfig.disabled"
                  values={{
                    planPagePath: "./../../../billing",
                  }}
                />
              </FeatureDisabledMessageBar>
            )}
            <DetailsList
              items={secondaryItems}
              columns={authenticatorColumns}
              onRenderItemColumn={onRenderSecondaryColumn}
              selectionMode={SelectionMode.none}
            />
          </div>
          <Toggle
            className={cn({
              [styles.readOnly]: isSecondaryAuthenticatorDisabled,
            })}
            inlineLabel={true}
            label={
              <FormattedMessage id="AuthenticatorConfigurationScreen.secondary-authenticators.disable-device-token.label" />
            }
            checked={state.disableDeviceToken}
            onChange={onDisableDeviceTokenChange}
          />
        </Widget>
        <Widget
          className={cn(styles.widget, {
            [styles.readOnly]: isSecondaryAuthenticatorDisabled,
          })}
        >
          <WidgetTitle>
            <FormattedMessage id="AuthenticatorConfigurationScreen.recovery-code.title" />
          </WidgetTitle>
          <WidgetDescription>
            <FormattedMessage id="AuthenticatorConfigurationScreen.recovery-code.description" />
          </WidgetDescription>
          <Toggle
            inlineLabel={true}
            label={
              <FormattedMessage id="AuthenticatorConfigurationScreen.recovery-code.enable-recovery-code" />
            }
            checked={state.recoveryCodeEnabled}
            onChange={onChangeRecoveryCodeEnabled}
          />
          <FormTextField
            disabled={!state.recoveryCodeEnabled}
            parentJSONPointer="/authentication/recovery_code"
            fieldName="count"
            label={renderToString(
              "AuthenticatorConfigurationScreen.recovery-code.recovery-code-number"
            )}
            value={state.numRecoveryCode?.toFixed(0) ?? ""}
            onChange={onRecoveryCodeNumberChange}
          />
          <Toggle
            disabled={!state.recoveryCodeEnabled}
            inlineLabel={true}
            label={
              <FormattedMessage id="AuthenticatorConfigurationScreen.recovery-code.allow-retrieve-recovery-code" />
            }
            checked={state.allowListRecoveryCode}
            onChange={onAllowRetrieveRecoveryCodeChange}
          />
        </Widget>
      </ScreenContent>
    );
  };

const AuthenticatorConfigurationScreen: React.FC =
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
          form={form}
          featureConfig={featureConfig.effectiveFeatureConfig ?? undefined}
        />
      </FormContainer>
    );
  };

export default AuthenticatorConfigurationScreen;
