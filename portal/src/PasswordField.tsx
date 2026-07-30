import React, { useCallback, useContext, useMemo, useState } from "react";
import cn from "classnames";
import {
  CheckIcon,
  Cross2Icon,
  EyeNoneIcon,
  EyeOpenIcon,
  InfoCircledIcon,
} from "@radix-ui/react-icons";
import { Text } from "@radix-ui/themes";
import { Context, FormattedMessage, Values } from "./intl";

import PasswordStrengthMeter from "./PasswordStrengthMeter";
import { PasswordPolicyConfig } from "./types";
import { checkPasswordPolicy } from "./error/password";
import { TextField } from "./components/v2/TextField/TextField";
import { ErrorParseRule } from "./error/parse";
import { GuessableLevel, zxcvbnGuessableLevel } from "./util/zxcvbn";
import { generatePassword } from "./util/passwordGenerator";

import styles from "./PasswordField.module.css";

export type GuessableLevelNames = Record<GuessableLevel, string>;

interface PasswordFieldProps {
  className?: string;
  label?: React.ReactNode;
  disabled?: boolean;
  value?: string;
  onChange?: (value: string) => void;
  passwordPolicy: PasswordPolicyConfig;
  canGeneratePassword?: boolean;
  canRevealPassword?: boolean;
  parentJSONPointer: string | RegExp;
  fieldName: string;
  errorRules?: ErrorParseRule[];
}

interface PasswordPolicyData {
  key: keyof PasswordPolicyConfig;
  messageId: string;
  messageValues?: Values;
}

function renderGuessableLevelNames(
  renderToString: (messageId: string) => string
): GuessableLevelNames {
  return {
    0: renderToString("PasswordField.guessable-level.0"),
    1: renderToString("PasswordField.guessable-level.1"),
    2: renderToString("PasswordField.guessable-level.2"),
    3: renderToString("PasswordField.guessable-level.3"),
    4: renderToString("PasswordField.guessable-level.4"),
    5: renderToString("PasswordField.guessable-level.5"),
  };
}

function makePasswordPolicyData(
  passwordPolicy: PasswordPolicyConfig,
  guessableLevelNames: GuessableLevelNames
) {
  const policyData: PasswordPolicyData[] = [];
  if (passwordPolicy.min_length != null) {
    policyData.push({
      key: "min_length",
      messageId: "PasswordField.min-length",
      messageValues: { minLength: passwordPolicy.min_length },
    });
  }
  if (passwordPolicy.lowercase_required === true) {
    policyData.push({
      key: "lowercase_required",
      messageId: "PasswordField.lowercase-required",
    });
  }
  if (passwordPolicy.uppercase_required === true) {
    policyData.push({
      key: "uppercase_required",
      messageId: "PasswordField.uppercase-required",
    });
  }
  if (passwordPolicy.alphabet_required === true) {
    policyData.push({
      key: "alphabet_required",
      messageId: "PasswordField.alphabet-required",
    });
  }
  if (passwordPolicy.digit_required === true) {
    policyData.push({
      key: "digit_required",
      messageId: "PasswordField.digit-required",
    });
  }
  if (passwordPolicy.symbol_required === true) {
    policyData.push({
      key: "symbol_required",
      messageId: "PasswordField.symbol-required",
    });
  }
  if (passwordPolicy.minimum_guessable_level != null) {
    policyData.push({
      key: "minimum_guessable_level",
      messageId: "PasswordField.minimum-guessable-level",
      messageValues: {
        level: passwordPolicy.minimum_guessable_level,
        levelName: guessableLevelNames[passwordPolicy.minimum_guessable_level],
      },
    });
  }
  if (passwordPolicy.excluded_keywords != null) {
    policyData.push({
      key: "excluded_keywords",
      messageId: "PasswordField.excluded-keywords",
    });
  }
  if (passwordPolicy.history_size != null) {
    policyData.push({
      key: "history_size",
      messageId: "PasswordField.history-size",
      messageValues: { size: passwordPolicy.history_size },
    });
  }
  if (passwordPolicy.history_days != null) {
    policyData.push({
      key: "history_days",
      messageId: "PasswordField.history-days",
      messageValues: { days: passwordPolicy.history_days },
    });
  }
  return policyData;
}

const PasswordField: React.VFC<PasswordFieldProps> = function PasswordField(
  props: PasswordFieldProps
) {
  const {
    className,
    label,
    disabled,
    value: password,
    onChange,
    passwordPolicy,
    canGeneratePassword,
    canRevealPassword,
    parentJSONPointer,
    fieldName,
    errorRules,
  } = props;
  const { renderToString } = useContext(Context);
  const [showPassword, setShowPassword] = useState(false);

  const guessableLevelNames = useMemo(
    () => renderGuessableLevelNames(renderToString),
    [renderToString]
  );
  const passwordPolicyData = useMemo(
    () => makePasswordPolicyData(passwordPolicy, guessableLevelNames),
    [guessableLevelNames, passwordPolicy]
  );

  const guessableLevel = useMemo(() => {
    if (password != null && password !== "") {
      return zxcvbnGuessableLevel(password, passwordPolicy.excluded_keywords);
    }
    return 0;
  }, [password, passwordPolicy]);

  const isPasswordPolicySatisfied = useMemo(
    () => checkPasswordPolicy(passwordPolicy, password ?? "", guessableLevel),
    [password, passwordPolicy, guessableLevel]
  );

  const onPasswordChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      onChange?.(e.currentTarget.value);
    },
    [onChange]
  );

  const onClickGeneratePassword = useCallback(() => {
    const newPassword = generatePassword(passwordPolicy);
    if (newPassword != null) {
      onChange?.(newPassword);
    }
  }, [passwordPolicy, onChange]);

  const onToggleShowPassword = useCallback(() => {
    setShowPassword((prev) => !prev);
  }, []);

  const suffix =
    canGeneratePassword || canRevealPassword ? (
      <div className={styles.suffixActions}>
        {canRevealPassword && !disabled ? (
          <button
            type="button"
            onClick={onToggleShowPassword}
            className={styles.revealPasswordButton}
            aria-label={showPassword ? "Hide password" : "Show password"}
          >
            {showPassword ? (
              <EyeNoneIcon width="1rem" height="1rem" />
            ) : (
              <EyeOpenIcon width="1rem" height="1rem" />
            )}
          </button>
        ) : null}
        {canGeneratePassword ? (
          <button
            type="button"
            disabled={disabled}
            onClick={onClickGeneratePassword}
            className={styles.generatePasswordButton}
          >
            <FormattedMessage id="PasswordField.generate-password" />
          </button>
        ) : null}
      </div>
    ) : undefined;

  return (
    <div className={className}>
      <TextField
        size="2"
        label={label}
        disabled={disabled}
        value={password}
        onChange={onPasswordChange}
        type={showPassword ? "text" : "password"}
        parentJSONPointer={parentJSONPointer}
        fieldName={fieldName}
        errorRules={errorRules}
        suffixPlain={true}
        suffix={suffix}
      />
      <PasswordStrengthMeter
        level={guessableLevel}
        guessableLevelNames={guessableLevelNames}
      />
      {passwordPolicyData.length > 0 ? (
        <div className={styles.passwordPolicyBox}>
          <div className={styles.passwordPolicyHeader}>
            <InfoCircledIcon
              className={styles.passwordPolicyHeaderIcon}
              width="1rem"
              height="1rem"
            />
            <Text as="span" size="2" weight="medium">
              <FormattedMessage id="PasswordField.password-requirements" />
            </Text>
          </div>
          <ul className={styles.passwordPolicy}>
            {passwordPolicyData.map((policy) => {
              const satisfied =
                isPasswordPolicySatisfied[policy.key] === true;
              return (
                <li
                  key={policy.messageId}
                  className={cn(styles.passwordPolicyItem, {
                    [styles.policySatisfied]: satisfied,
                    [styles.policyUnsatisfied]: !satisfied,
                  })}
                >
                  {satisfied ? (
                    <CheckIcon
                      className={styles.passwordPolicyIcon}
                      width="1rem"
                      height="1rem"
                      aria-hidden={true}
                    />
                  ) : (
                    <Cross2Icon
                      className={styles.passwordPolicyIcon}
                      width="1rem"
                      height="1rem"
                      aria-hidden={true}
                    />
                  )}
                  <Text as="span" size="2">
                    <FormattedMessage
                      id={policy.messageId}
                      values={policy.messageValues}
                    />
                  </Text>
                </li>
              );
            })}
          </ul>
        </div>
      ) : null}
    </div>
  );
};

export default PasswordField;
