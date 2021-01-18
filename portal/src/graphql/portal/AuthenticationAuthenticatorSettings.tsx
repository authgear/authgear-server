import React, { useCallback, useContext, useMemo } from "react";
import {
  Checkbox,
  DefaultEffects,
  Dropdown,
  IColumn,
  SelectionMode,
  Text,
  Toggle,
} from "@fluentui/react";
import produce from "immer";
import deepEqual from "deep-equal";
import { Context, FormattedMessage } from "@oursky/react-messageformat";

import DetailsListWithOrdering from "../../DetailsListWithOrdering";
import { swap } from "../../OrderButtons";
import FormTextField from "../../FormTextField";
import {
  PortalAPIAppConfig,
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
import FormContainer from "../../FormContainer";
import NavBreadcrumb, { BreadcrumbItem } from "../../NavBreadcrumb";

import styles from "./AuthenticationAuthenticatorSettings.module.scss";
import { ErrorParseRule } from "../../error/parse";

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
}

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
  oob_otp: "AuthenticatorType.primary.oob-otp",
  password: "AuthenticatorType.primary.password",
};
const secondaryAuthenticatorNameIds = {
  totp: "AuthenticatorType.secondary.totp",
  oob_otp: "AuthenticatorType.secondary.oob-otp",
  password: "AuthenticatorType.secondary.password",
};

type AuthenticatorColumnItem = (
  | { kind: "primary"; type: PrimaryAuthenticatorType }
  | { kind: "secondary"; type: SecondaryAuthenticatorType }
) & { isEnabled: boolean };

interface AuthenticatorCheckboxProps {
  item: AuthenticatorColumnItem;
  onChange: (item: AuthenticatorColumnItem, checked: boolean) => void;
}

const AuthenticatorCheckbox: React.FC<AuthenticatorCheckboxProps> = function AuthenticatorCheckbox(
  props
) {
  const { item, onChange } = props;
  const onCheckboxChange = useCallback(
    (_event, checked?: boolean) => onChange(item, checked ?? false),
    [item, onChange]
  );

  return <Checkbox checked={item.isEnabled} onChange={onCheckboxChange} />;
};

interface AuthenticationAuthenticatorSettingsContentProps {
  form: AppConfigFormModel<FormState>;
}

const AuthenticationAuthenticatorSettingsContent: React.FC<AuthenticationAuthenticatorSettingsContentProps> = function AuthenticationAuthenticatorSettingsContent(
  props
) {
  const { state, setState } = props.form;

  const { renderToString } = useContext(Context);

  const navBreadcrumbItems: BreadcrumbItem[] = useMemo(() => {
    return [
      {
        to: ".",
        label: (
          <FormattedMessage id="AuthenticationAuthenticatorSettingsScreen.title" />
        ),
      },
    ];
  }, []);

  const authenticatorColumns: IColumn[] = [
    {
      key: "activated",
      fieldName: "activated",
      name: renderToString(
        "AuthenticationAuthenticatorSettingsScreen.columns.activate"
      ),
      className: styles.authenticatorColumn,
      minWidth: 120,
      maxWidth: 120,
    },
    {
      key: "key",
      fieldName: "key",
      name: renderToString(
        "AuthenticationAuthenticatorSettingsScreen.columns.authenticator"
      ),
      className: styles.authenticatorColumn,
      minWidth: 300,
      maxWidth: 300,
    },
  ];

  const renderSecondaryAuthenticatorMode = useCallback(
    (key: SecondaryAuthenticationMode) => {
      const messageIdMap: Record<SecondaryAuthenticationMode, string> = {
        required:
          "AuthenticationAuthenticatorSettingsScreen.policy.require-mfa.required",
        if_exists:
          "AuthenticationAuthenticatorSettingsScreen.policy.require-mfa.if-exists",
        if_requested:
          "AuthenticationAuthenticatorSettingsScreen.policy.require-mfa.if-requested",
      };

      return renderToString(messageIdMap[key]);
    },
    [renderToString]
  );

  const {
    options: requireMFAOptions,
    onChange: onRequireMFAOptionChange,
  } = useDropdown(
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

  const onRenderColumn = useCallback(
    (item: AuthenticatorColumnItem, _index?: number, column?: IColumn) => {
      switch (column?.key) {
        case "activated":
          return (
            <AuthenticatorCheckbox
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
    [onAuthenticatorEnabledChange]
  );

  const renderPrimaryAriaLabel = React.useCallback(
    (index?: number): string => {
      return index != null
        ? renderToString(primaryAuthenticatorNameIds[state.primary[index].type])
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
    <div className={styles.root}>
      <NavBreadcrumb items={navBreadcrumbItems} />

      <div
        className={styles.widget}
        style={{ boxShadow: DefaultEffects.elevation4 }}
      >
        <Text as="h2" className={styles.widgetHeader}>
          <FormattedMessage id="AuthenticationAuthenticatorSettingsScreen.primary-authenticators.title" />
        </Text>
        <DetailsListWithOrdering
          items={primaryItems}
          columns={authenticatorColumns}
          onRenderItemColumn={onRenderColumn}
          onSwapClicked={onPrimarySwapClicked}
          selectionMode={SelectionMode.none}
          renderAriaLabel={renderPrimaryAriaLabel}
        />
      </div>

      <div
        className={styles.widget}
        style={{ boxShadow: DefaultEffects.elevation4 }}
      >
        <Text as="h2" className={styles.widgetHeader}>
          <FormattedMessage id="AuthenticationAuthenticatorSettingsScreen.secondary-authenticators.title" />
        </Text>
        <DetailsListWithOrdering
          items={secondaryItems}
          columns={authenticatorColumns}
          onRenderItemColumn={onRenderColumn}
          onSwapClicked={onSecondarySwapClicked}
          selectionMode={SelectionMode.none}
          renderAriaLabel={renderSecondaryAriaLabel}
        />
      </div>

      <section className={styles.policy}>
        <Text className={styles.policyHeader} as="h2">
          <FormattedMessage id="AuthenticationAuthenticatorSettingsScreen.policy.title" />
        </Text>
        <Dropdown
          className={styles.requireMFADropdown}
          label={renderToString(
            "AuthenticationAuthenticatorSettingsScreen.policy.require-mfa"
          )}
          options={requireMFAOptions}
          selectedKey={state.mfaMode}
          onChange={onRequireMFAOptionChange}
        />
        <FormTextField
          parentJSONPointer="/authentication/recovery_code"
          fieldName="count"
          fieldNameMessageID="AuthenticationAuthenticatorSettingsScreen.policy.recovery-code-number"
          className={styles.recoveryCodeNumber}
          value={String(state.numRecoveryCode)}
          onChange={onRecoveryCodeNumberChange}
        />
        <Toggle
          className={styles.allowRetrieveRecoveryCode}
          inlineLabel={true}
          label={
            <FormattedMessage id="AuthenticationAuthenticatorSettingsScreen.policy.allow-retrieve-recovery-code" />
          }
          checked={state.allowListRecoveryCode}
          onChange={onAllowRetrieveRecoveryCodeChange}
        />
      </section>
    </div>
  );
};

const AuthenticationAuthenticatorSettingsScreen: React.FC = function AuthenticationAuthenticatorSettingsScreen() {
  const { appID } = useParams();
  const form = useAppConfigForm(appID, constructFormState, constructConfig);

  const errorRules: ErrorParseRule[] = useMemo(
    () => [
      {
        reason: "ValidationFailed",
        kind: "general",
        location: "/authentication/identities/\\d+",
        errorMessageID:
          "AuthenticationAuthenticatorSettingsScreen.error.no-primary-authenticator",
      },
    ],
    []
  );

  if (form.isLoading) {
    return <ShowLoading />;
  }

  if (form.loadError) {
    return <ShowError error={form.loadError} onRetry={form.reload} />;
  }

  return (
    <FormContainer form={form} errorRules={errorRules}>
      <AuthenticationAuthenticatorSettingsContent form={form} />
    </FormContainer>
  );
};

export default AuthenticationAuthenticatorSettingsScreen;
