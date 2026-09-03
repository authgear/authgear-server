import React, { useCallback, useEffect, useMemo, useState } from "react";
import cn from "classnames";
import { useNavigate, useParams } from "react-router-dom";
import * as Collapsible from "@radix-ui/react-collapsible";
import { ChevronDownIcon, ChevronLeftIcon } from "@radix-ui/react-icons";
import {
  Checkbox,
  Flex,
  Heading,
  RadioGroup,
  Separator,
  Text,
} from "@radix-ui/themes";
import { FormattedMessage } from "../../intl";
import { useAppAndSecretConfigQuery } from "../portal/query/appAndSecretConfigQuery";
import { useCreateUserMutation } from "./mutations/createUserMutation";
import ScreenContent from "../../ScreenContent";
import Link from "../../Link";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import PasswordField from "../../PasswordField";
import {
  PrimaryAuthenticatorType,
  LoginIDKeyType,
  loginIDKeyTypes,
  PasswordPolicyConfig,
  PortalAPIAppConfig,
} from "../../types";
import {
  ErrorParseRule,
  makeInvariantViolatedErrorParseRule,
} from "../../error/parse";
import { SimpleFormModel, useSimpleForm } from "../../hook/useSimpleForm";
import FormPhoneTextField from "../../FormPhoneTextField";
import { PhoneTextFieldValues } from "../../PhoneTextField";
import FormContainer from "../../FormContainer";
import { useFormContainerBaseContext } from "../../FormContainerBase";
import { TextField } from "../../components/v2/TextField/TextField";
import { PrimaryButton } from "../../components/v2/Button/PrimaryButton/PrimaryButton";
import { Callout } from "../../components/v2/Callout/Callout";
import { AccountValidPeriodForm } from "./UserDetailsAccountStatus";
import { validatePassword } from "../../error/password";

import styles from "./AddUserScreen.module.css";

enum PasswordCreationType {
  ManualEntry = "manual_entry",
  AutoGenerate = "auto_generate",
  NoPassword = "no_password",
}

interface FormState {
  selectedLoginIDType: LoginIDKeyType | null;
  username: string;
  email: string;
  phone: string;
  passwordCreationType: PasswordCreationType;
  manualEntryPassword: string;
  manualEntrySendPassword: boolean;
  setPasswordExpired: boolean;
  accountValidFrom: Date | null;
  accountValidUntil: Date | null;
}

const loginIdTypeNameIds: Record<LoginIDKeyType, string> = {
  username: "login-id-key.username",
  email: "login-id-key.email",
  phone: "login-id-key.phone",
};

const passwordCreationTypeLabelIds: Record<PasswordCreationType, string> = {
  [PasswordCreationType.ManualEntry]:
    "AddUserScreen.password-creation-type.manual",
  [PasswordCreationType.AutoGenerate]:
    "AddUserScreen.password-creation-type.auto",
  [PasswordCreationType.NoPassword]:
    "AddUserScreen.password-creation-type.no-password",
};

const passwordCreationTypeDescriptionIds: Record<PasswordCreationType, string> =
  {
    [PasswordCreationType.ManualEntry]:
      "AddUserScreen.password-creation-type.manual.description",
    [PasswordCreationType.AutoGenerate]:
      "AddUserScreen.password-creation-type.auto-generate.description",
    [PasswordCreationType.NoPassword]:
      "AddUserScreen.password-creation-type.no-password.description",
  };

function makeDefaultFormState(loginIDTypes: LoginIDKeyType[]): FormState {
  if (loginIDTypes.length === 1) {
    return {
      selectedLoginIDType: loginIDTypes[0],
      username: "",
      email: "",
      phone: "",
      manualEntryPassword: "",
      passwordCreationType: PasswordCreationType.ManualEntry,
      manualEntrySendPassword: false,
      setPasswordExpired: true,
      accountValidFrom: null,
      accountValidUntil: null,
    };
  }

  return {
    selectedLoginIDType: null,
    username: "",
    email: "",
    phone: "",
    manualEntryPassword: "",
    passwordCreationType: PasswordCreationType.ManualEntry,
    manualEntrySendPassword: false,
    setPasswordExpired: true,
    accountValidFrom: null,
    accountValidUntil: null,
  };
}

function isPasswordFieldDisplayed(
  primaryAuthenticators: PrimaryAuthenticatorType[],
  loginIdKeySelected: LoginIDKeyType | null
) {
  // Unknown yet.
  if (loginIdKeySelected == null) {
    return false;
  }

  const filterAuthenticators = (allowedTypes: PrimaryAuthenticatorType[]) => {
    return primaryAuthenticators.filter((authenticator) =>
      allowedTypes.includes(authenticator)
    );
  };

  let relatedAuthenticators: PrimaryAuthenticatorType[];
  switch (loginIdKeySelected) {
    case "email":
      relatedAuthenticators = filterAuthenticators([
        "oob_otp_email",
        "password",
      ]);
      break;
    case "phone":
      relatedAuthenticators = filterAuthenticators(["oob_otp_sms", "password"]);
      break;
    case "username":
      relatedAuthenticators = filterAuthenticators(["password"]);
      break;
    default:
      relatedAuthenticators = filterAuthenticators([]);
  }

  return (
    relatedAuthenticators.length > 0 &&
    relatedAuthenticators.includes("password")
  );
}

function getEnabledLoginIDTypes(
  appConfig: PortalAPIAppConfig | null
): LoginIDKeyType[] {
  const enabledIdentities = appConfig?.authentication?.identities ?? [];
  const isLoginIDIdentityEnabled = enabledIdentities.includes("login_id");
  // if login ID identity is disabled
  // we cannot add login ID identity to new user
  if (!isLoginIDIdentityEnabled) {
    return [];
  }

  const loginIdKeys = appConfig?.identity?.login_id?.keys ?? [];
  const loginIdKeyOptions = new Set<LoginIDKeyType>();
  for (const key of loginIdKeys) {
    switch (key.type) {
      case "username": {
        loginIdKeyOptions.add("username");
        break;
      }
      case "email":
        loginIdKeyOptions.add("email");
        break;
      case "phone":
        loginIdKeyOptions.add("phone");
        break;
      default:
        break;
    }
  }
  return Array.from(loginIdKeyOptions);
}

const errorRules: ErrorParseRule[] = [
  makeInvariantViolatedErrorParseRule(
    "DuplicatedIdentity",
    "AddUserScreen.error.duplicated-identity"
  ),
];

interface AddUserContentProps {
  primaryAuthenticators: PrimaryAuthenticatorType[];
  passwordPolicy: PasswordPolicyConfig;
  loginIDTypes: LoginIDKeyType[];
  form: SimpleFormModel<FormState>;
  isPasskeyOnly: boolean;
  phoneInputAllowlist?: string[];
  phoneInputPinnedList?: string[];
}

interface PhoneFieldProps {
  className?: string;
  allowlist?: string[];
  pinnedList?: string[];
  onChange: (e164: string) => void;
}

const PhoneField: React.VFC<PhoneFieldProps> = function PhoneField(props) {
  const { className, allowlist, pinnedList, onChange } = props;
  const [inputValue, setInputValue] = useState("");
  const onChangeValues = useCallback(
    (values: PhoneTextFieldValues) => {
      onChange(values.e164 ?? "");
      setInputValue(values.rawInputValue);
    },
    [onChange]
  );
  return (
    <FormPhoneTextField
      className={className}
      parentJSONPointer=""
      fieldName="login_id"
      errorRules={errorRules}
      allowlist={allowlist}
      pinnedList={pinnedList}
      initialInputValue={inputValue}
      onChange={onChangeValues}
    />
  );
};

const ManualEntryPasswordField: React.VFC<{
  disabled: boolean;
  passwordPolicy: PasswordPolicyConfig;
  password: string;
  onPasswordChange: (value: string) => void;
  selectedLoginIDType: LoginIDKeyType | null;
  sendPassword: boolean;
  onChangeSendPassword: (checked: boolean | "indeterminate") => void;
}> = function ManualEntryPasswordField(props) {
  const {
    disabled,
    passwordPolicy,
    password,
    onPasswordChange,
    selectedLoginIDType,
    sendPassword,
    onChangeSendPassword,
  } = props;

  return (
    <div className={styles.manualPasswordFields}>
      <PasswordField
        label={<FormattedMessage id="AddUserScreen.password.label" />}
        disabled={disabled}
        value={password}
        canRevealPassword={true}
        canGeneratePassword={true}
        onChange={onPasswordChange}
        passwordPolicy={passwordPolicy}
        parentJSONPointer=""
        fieldName="password"
      />
      {selectedLoginIDType === "email" ? (
        <label className={styles.checkboxRow}>
          <Checkbox
            disabled={disabled}
            checked={sendPassword}
            onCheckedChange={onChangeSendPassword}
          />
          <Text size="2">
            <FormattedMessage id="AddUserScreen.send-password" />
          </Text>
        </label>
      ) : null}
    </div>
  );
};

const AddUserContent: React.VFC<AddUserContentProps> = function AddUserContent(
  props: AddUserContentProps
) {
  const {
    primaryAuthenticators,
    passwordPolicy,
    loginIDTypes,
    form: { state, setState },
    isPasskeyOnly,
    phoneInputAllowlist,
    phoneInputPinnedList,
  } = props;
  const { canSave, isUpdating, onSave } = useFormContainerBaseContext();

  const [advancedOpen, setAdvancedOpen] = useState(false);

  const {
    username,
    email,
    manualEntryPassword: password,
    selectedLoginIDType,
  } = state;

  const onUsernameChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const value = e.currentTarget.value;
      setState((prev) => ({ ...prev, username: value }));
    },
    [setState]
  );
  const onEmailChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const value = e.currentTarget.value;
      setState((prev) => ({ ...prev, email: value }));
    },
    [setState]
  );
  const onPhoneChange = useCallback(
    (value: string) => {
      setState((prev) => ({ ...prev, phone: value }));
    },
    [setState]
  );
  const onPasswordChange = useCallback(
    (value: string) => {
      setState((prev) => ({ ...prev, manualEntryPassword: value }));
    },
    [setState]
  );
  const onChangeSendPassword = useCallback(
    (checked: boolean | "indeterminate") => {
      if (typeof checked === "boolean") {
        setState((prev) => ({ ...prev, manualEntrySendPassword: checked }));
      }
    },
    [setState]
  );
  const onChangeForceChangeOnLogin = useCallback(
    (checked: boolean | "indeterminate") => {
      if (typeof checked === "boolean") {
        setState((prev) => ({ ...prev, setPasswordExpired: checked }));
      }
    },
    [setState]
  );
  const onPickAccountValidFrom = useCallback(
    (date: Date | null) => {
      setState((prev) => ({ ...prev, accountValidFrom: date }));
    },
    [setState]
  );
  const onPickAccountValidUntil = useCallback(
    (date: Date | null) => {
      setState((prev) => ({ ...prev, accountValidUntil: date }));
    },
    [setState]
  );

  const passwordCreationTypes = useMemo((): PasswordCreationType[] => {
    switch (selectedLoginIDType) {
      case "email":
        return [
          PasswordCreationType.ManualEntry,
          PasswordCreationType.AutoGenerate,
          PasswordCreationType.NoPassword,
        ];
      case "phone":
      case "username":
        return [
          PasswordCreationType.ManualEntry,
          PasswordCreationType.NoPassword,
        ];
      default:
        return [];
    }
  }, [selectedLoginIDType]);

  const onChangePasswordCreationType = useCallback(
    (value: string) => {
      setState((prev) => {
        const newPasswordCreationType = value as PasswordCreationType;
        if (prev.passwordCreationType === newPasswordCreationType) {
          return prev;
        }
        return {
          ...prev,
          passwordCreationType: newPasswordCreationType,
        };
      });
    },
    [setState]
  );

  const passwordFieldNeeded = useMemo(() => {
    return isPasswordFieldDisplayed(primaryAuthenticators, selectedLoginIDType);
  }, [primaryAuthenticators, selectedLoginIDType]);

  const onSelectLoginIdType = useCallback(
    (value: string) => {
      const loginIdType = value as LoginIDKeyType;
      if (!loginIDKeyTypes.includes(loginIdType)) {
        return;
      }
      setState(() => ({
        ...makeDefaultFormState([...loginIDKeyTypes]),
        selectedLoginIDType: loginIdType,
      }));
    },
    [setState]
  );

  // NOTE: cannot add user identity if none of three field is available
  const canAddUser = loginIDTypes.length > 0;

  // TODO: improve empty state
  if (!canAddUser) {
    return (
      <Text as="p" size="2">
        <FormattedMessage id="AddUserScreen.cannot-add-user" />
      </Text>
    );
  }

  return (
    <ScreenContent>
      <div className={styles.widget}>
        <Link to=".." className={styles.backLink}>
          <ChevronLeftIcon className={styles.backLinkIcon} />
          <span>
            <FormattedMessage id="UsersScreen.title" />
          </span>
        </Link>
        <Heading as="h1" size="5" weight="bold" className={styles.pageTitle}>
          <FormattedMessage id="AddUserScreen.title" />
        </Heading>
      </div>
      <div className={styles.verticalForm}>
        {isPasskeyOnly ? (
          <div className={styles.widget}>
            <Callout
              type="info"
              showCloseButton={false}
              text={
                <FormattedMessage id="AddUserScreen.passkey-only.message" />
              }
            />
          </div>
        ) : (
          <>
            <div className={styles.widget}>
              <Text
                as="p"
                size="2"
                weight="medium"
                className={styles.fieldLabel}
              >
                <FormattedMessage id="AddUserScreen.select-sign-in-method.label" />
              </Text>
              <RadioGroup.Root
                value={selectedLoginIDType ?? undefined}
                onValueChange={onSelectLoginIdType}
                className={styles.loginIdRadioGroup}
              >
                <Flex gap="4" wrap="wrap">
                  {loginIDTypes.map((loginIdType) => (
                    <Text
                      key={loginIdType}
                      as="label"
                      size="2"
                      className={styles.radioOptionLabel}
                    >
                      <Flex gap="2" align="center">
                        <RadioGroup.Item value={loginIdType} />
                        <FormattedMessage
                          id={loginIdTypeNameIds[loginIdType]}
                        />
                      </Flex>
                    </Text>
                  ))}
                </Flex>
              </RadioGroup.Root>
            </div>

            {selectedLoginIDType === "username" ? (
              <div className={styles.widget}>
                <TextField
                  size="2"
                  label={<FormattedMessage id={loginIdTypeNameIds.username} />}
                  value={username}
                  onChange={onUsernameChange}
                  parentJSONPointer=""
                  fieldName="login_id"
                  errorRules={errorRules}
                />
              </div>
            ) : null}

            {selectedLoginIDType === "email" ? (
              <div className={styles.widget}>
                <TextField
                  size="2"
                  label={<FormattedMessage id={loginIdTypeNameIds.email} />}
                  value={email}
                  onChange={onEmailChange}
                  parentJSONPointer=""
                  fieldName="login_id"
                  errorRules={errorRules}
                />
              </div>
            ) : null}

            {selectedLoginIDType === "phone" ? (
              <div className={styles.widget}>
                <Text
                  as="label"
                  size="2"
                  weight="medium"
                  className={styles.fieldLabel}
                >
                  <FormattedMessage id={loginIdTypeNameIds.phone} />
                </Text>
                <PhoneField
                  allowlist={phoneInputAllowlist}
                  pinnedList={phoneInputPinnedList}
                  onChange={onPhoneChange}
                />
              </div>
            ) : null}

            {passwordFieldNeeded ? (
              <>
                <div className={styles.widget}>
                  <Text
                    as="p"
                    size="2"
                    weight="medium"
                    className={styles.fieldLabel}
                  >
                    <FormattedMessage id="AddUserScreen.password-setup.label" />
                  </Text>
                  <RadioGroup.Root
                    value={state.passwordCreationType}
                    onValueChange={onChangePasswordCreationType}
                    className={styles.passwordRadioGroup}
                  >
                    <Flex direction="column" gap="3">
                      {passwordCreationTypes.map((option) => (
                        <div key={option} className={styles.passwordRadioBlock}>
                          <Text
                            as="label"
                            size="2"
                            className={styles.radioOptionLabel}
                          >
                            <Flex gap="2" align="start">
                              <RadioGroup.Item
                                value={option}
                                className={styles.passwordRadioItem}
                              />
                              <div className={styles.passwordRadioContent}>
                                <Text as="span" size="2" weight="medium">
                                  <FormattedMessage
                                    id={passwordCreationTypeLabelIds[option]}
                                  />
                                </Text>
                                <Text as="p" size="1" color="gray">
                                  <FormattedMessage
                                    id={
                                      passwordCreationTypeDescriptionIds[option]
                                    }
                                  />
                                </Text>
                              </div>
                            </Flex>
                          </Text>
                          {option === PasswordCreationType.ManualEntry &&
                          state.passwordCreationType ===
                            PasswordCreationType.ManualEntry ? (
                            <div className={styles.passwordOptionExtra}>
                              <ManualEntryPasswordField
                                disabled={false}
                                passwordPolicy={passwordPolicy}
                                password={password}
                                onPasswordChange={onPasswordChange}
                                selectedLoginIDType={selectedLoginIDType}
                                sendPassword={state.manualEntrySendPassword}
                                onChangeSendPassword={onChangeSendPassword}
                              />
                            </div>
                          ) : null}
                        </div>
                      ))}
                    </Flex>
                  </RadioGroup.Root>
                </div>
                <Separator size="4" className={styles.widget} />
                <div className={styles.widget}>
                  <Text
                    as="p"
                    size="2"
                    weight="medium"
                    className={styles.fieldLabel}
                  >
                    <FormattedMessage id="AddUserScreen.additional-option.label" />
                  </Text>
                  <label className={styles.checkboxRow}>
                    <Checkbox
                      checked={
                        state.passwordCreationType ===
                        PasswordCreationType.NoPassword
                          ? false
                          : state.setPasswordExpired
                      }
                      onCheckedChange={onChangeForceChangeOnLogin}
                      disabled={
                        state.passwordCreationType ===
                        PasswordCreationType.NoPassword
                      }
                    />
                    <Text size="2">
                      <FormattedMessage id="AddUserScreen.force-change-on-login" />
                    </Text>
                  </label>
                </div>
              </>
            ) : null}
            {selectedLoginIDType ? (
              <Collapsible.Root
                className={cn(styles.widget, styles.advancedSection)}
                open={advancedOpen}
                onOpenChange={setAdvancedOpen}
              >
                <Collapsible.Trigger className={styles.advancedTrigger}>
                  <Text as="span" size="2" weight="medium">
                    <FormattedMessage id="AdduserScreen.advanced" />
                  </Text>
                  <ChevronDownIcon
                    className={styles.advancedChevron}
                    aria-hidden={true}
                  />
                </Collapsible.Trigger>
                <Collapsible.Content className={styles.advancedContent}>
                  <div className={styles.accountValidPeriodSection}>
                    <Text
                      as="p"
                      size="3"
                      weight="medium"
                      className={styles.sectionTitle}
                    >
                      <FormattedMessage id="AddUserScreen.valid-period.title" />
                    </Text>
                    <Text
                      as="p"
                      size="2"
                      color="gray"
                      className={styles.sectionDescription}
                    >
                      <FormattedMessage id="AddUserScreen.valid-period.description" />
                    </Text>
                    <AccountValidPeriodForm
                      className={styles.accountValidPeriodForm}
                      accountValidFrom={state.accountValidFrom}
                      accountValidUntil={state.accountValidUntil}
                      onPickAccountValidFrom={onPickAccountValidFrom}
                      onPickAccountValidUntil={onPickAccountValidUntil}
                    />
                  </div>
                </Collapsible.Content>
              </Collapsible.Root>
            ) : null}
          </>
        )}
      </div>
      <div className={cn(styles.widget, styles.saveButtonRow)}>
        <PrimaryButton
          size="2"
          disabled={!canSave}
          loading={isUpdating}
          text={<FormattedMessage id="AddUserScreen.add-user.label" />}
          onClick={onSave}
        />
      </div>
    </ScreenContent>
  );
};

const AddUserScreen: React.VFC = function AddUserScreen() {
  const { appID } = useParams() as { appID: string };
  const navigate = useNavigate();

  const {
    effectiveAppConfig,
    isLoading: loading,
    loadError: error,
    refetch,
  } = useAppAndSecretConfigQuery(appID);

  const primaryAuthenticators = useMemo(
    () => effectiveAppConfig?.authentication?.primary_authenticators ?? [],
    [effectiveAppConfig]
  );

  const loginIDTypes = useMemo(
    () => getEnabledLoginIDTypes(effectiveAppConfig),
    [effectiveAppConfig]
  );

  const passwordPolicy = useMemo(
    () => effectiveAppConfig?.authenticator?.password?.policy ?? {},
    [effectiveAppConfig]
  );

  const isPasskeyOnly = useMemo(() => {
    const primaryAuthenticators =
      effectiveAppConfig?.authentication?.primary_authenticators ?? [];
    return (
      primaryAuthenticators.length === 1 &&
      primaryAuthenticators[0] === "passkey"
    );
  }, [effectiveAppConfig]);

  const defaultState = useMemo(() => {
    return makeDefaultFormState(loginIDTypes);
  }, [loginIDTypes]);

  const { createUser } = useCreateUserMutation();

  const validate = useCallback(
    (state: FormState) => {
      if (
        !isPasswordFieldDisplayed(
          primaryAuthenticators,
          state.selectedLoginIDType
        ) ||
        state.passwordCreationType === PasswordCreationType.NoPassword
      ) {
        return null;
      }
      return validatePassword(state.manualEntryPassword, passwordPolicy);
    },
    [primaryAuthenticators, passwordPolicy]
  );

  const submit = useCallback(
    async (state: FormState) => {
      const loginIDType = state.selectedLoginIDType;
      if (!loginIDType) {
        return;
      }

      const hasPasswordField = isPasswordFieldDisplayed(
        primaryAuthenticators,
        state.selectedLoginIDType
      );
      const identityValue = state[loginIDType];
      let password: string | undefined;
      let sendPassword: boolean | undefined;
      let setPasswordExpired: boolean | undefined;
      if (hasPasswordField) {
        switch (state.passwordCreationType) {
          case PasswordCreationType.AutoGenerate:
            password = "";
            sendPassword = loginIDType === "email" ? true : undefined;
            setPasswordExpired = state.setPasswordExpired;
            break;
          case PasswordCreationType.ManualEntry:
            password = state.manualEntryPassword;
            sendPassword =
              loginIDType === "email"
                ? state.manualEntrySendPassword
                : undefined;
            setPasswordExpired = state.setPasswordExpired;
            break;
          case PasswordCreationType.NoPassword:
            password = undefined;
            sendPassword = false;
            setPasswordExpired = false;
            break;
          default:
            break;
        }
      } else {
        sendPassword = false;
        setPasswordExpired = false;
      }

      await createUser({
        identity: { key: loginIDType, value: identityValue },
        password,
        sendPassword,
        setPasswordExpired,
        accountValidFrom: state.accountValidFrom,
        accountValidUntil: state.accountValidUntil,
      });
    },
    [createUser, primaryAuthenticators]
  );

  const form = useSimpleForm({
    defaultState,
    submit,
    validate,
  });

  const canSave = useMemo(() => {
    if (form.state.selectedLoginIDType == null) {
      return false;
    }
    if (!form.state[form.state.selectedLoginIDType]) {
      return false;
    }

    if (
      isPasswordFieldDisplayed(
        primaryAuthenticators,
        form.state.selectedLoginIDType
      )
    ) {
      switch (form.state.passwordCreationType) {
        case PasswordCreationType.ManualEntry:
          return form.state.manualEntryPassword.length > 0;
        case PasswordCreationType.AutoGenerate:
          return true;
        case PasswordCreationType.NoPassword:
          return true;
        default:
          throw new Error("unknown passwordCreationType");
      }
    }

    return true;
  }, [form.state, primaryAuthenticators]);

  useEffect(() => {
    if (form.isSubmitted) {
      navigate("./..");
    }
  }, [form.isSubmitted, navigate]);

  if (loading) {
    return <ShowLoading />;
  }

  if (error != null) {
    // eslint-disable-next-line @typescript-eslint/strict-void-return
    return <ShowError error={error} onRetry={refetch} />;
  }

  return (
    <FormContainer form={form} canSave={canSave}>
      <AddUserContent
        form={form}
        primaryAuthenticators={primaryAuthenticators}
        loginIDTypes={loginIDTypes}
        passwordPolicy={passwordPolicy}
        isPasskeyOnly={isPasskeyOnly}
        phoneInputAllowlist={effectiveAppConfig?.ui?.phone_input?.allowlist}
        phoneInputPinnedList={effectiveAppConfig?.ui?.phone_input?.pinned_list}
      />
    </FormContainer>
  );
};

export default AddUserScreen;
