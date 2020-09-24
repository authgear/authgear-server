import React, { useContext, useMemo } from "react";
import cn from "classnames";
import zxcvbn from "zxcvbn";
import { ITextFieldProps, Text, TextField } from "@fluentui/react";
import { Context, FormattedMessage, Values } from "@oursky/react-messageformat";

import PasswordStrengthMeter from "./PasswordStrengthMeter";
import { PasswordPolicyConfig } from "./types";

import styles from "./PasswordField.module.scss";

export type GuessableLevel = 0 | 1 | 2 | 3 | 4 | 5;
export type GuessableLevelNames = Record<GuessableLevel, string>;

interface PasswordFieldProps extends ITextFieldProps {
  className?: string;
  textFieldClassName?: string;
  passwordPolicy: PasswordPolicyConfig;
}

interface PasswordPolicyData {
  key: keyof PasswordPolicyConfig;
  messageId: string;
  messageValues?: Values;
}

function checkPasswordPolicy(
  passwordPolicy: PasswordPolicyConfig,
  password: string | undefined,
  level: GuessableLevel
) {
  const isPolicySatisfied: Partial<Record<
    keyof PasswordPolicyConfig,
    boolean
  >> = {};
  if (password == null) {
    return isPolicySatisfied;
  }
  if (passwordPolicy.min_length != null) {
    isPolicySatisfied.min_length = password.length >= passwordPolicy.min_length;
  }
  if (passwordPolicy.lowercase_required === true) {
    isPolicySatisfied.lowercase_required = /[a-z]/.test(password);
  }
  if (passwordPolicy.uppercase_required === true) {
    isPolicySatisfied.uppercase_required = /[A-Z]/.test(password);
  }
  if (passwordPolicy.digit_required === true) {
    isPolicySatisfied.digit_required = /\d/.test(password);
  }
  if (passwordPolicy.symbol_required === true) {
    // treat all character which is not alphanumeric as symbol
    isPolicySatisfied.symbol_required = /[^a-zA-Z0-9]/.test(password);
  }
  if (passwordPolicy.minimum_guessable_level != null) {
    isPolicySatisfied.minimum_guessable_level =
      level >= passwordPolicy.minimum_guessable_level;
  }

  return isPolicySatisfied;
}

export function isPasswordValid(
  passwordPolicy: PasswordPolicyConfig,
  password: string | undefined,
  level: GuessableLevel
): boolean {
  if (password == null) {
    return false;
  }
  const isPolicySatisfied = checkPasswordPolicy(
    passwordPolicy,
    password,
    level
  );
  return Object.values(isPolicySatisfied).every(Boolean);
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

export function extractGuessableLevel(
  result: zxcvbn.ZXCVBNResult | null
): GuessableLevel {
  if (result == null) {
    return 0;
  }
  return Math.floor(
    Math.min(5, Math.max(1, result.score + 1))
  ) as GuessableLevel;
}

const PasswordField: React.FC<PasswordFieldProps> = function PasswordField(
  props: PasswordFieldProps
) {
  const {
    className,
    textFieldClassName,
    value: password,
    passwordPolicy,
    ...rest
  } = props;
  const { renderToString } = useContext(Context);

  const guessableLevelNames = useMemo(
    () => renderGuessableLevelNames(renderToString),
    [renderToString]
  );
  const passwordPolicyData = useMemo(
    () => makePasswordPolicyData(passwordPolicy, guessableLevelNames),
    [guessableLevelNames, passwordPolicy]
  );

  const result = useMemo(() => {
    if (password != null && password !== "") {
      return zxcvbn(password, passwordPolicy.excluded_keywords);
    }
    return null;
  }, [password, passwordPolicy]);
  const guessableLevel = extractGuessableLevel(result);

  const isPasswordPolicySatisfied = useMemo(
    () => checkPasswordPolicy(passwordPolicy, password, guessableLevel),
    [password, passwordPolicy, guessableLevel]
  );
  return (
    <div className={className}>
      <TextField
        {...rest}
        value={password}
        className={textFieldClassName}
        type="password"
      />
      <PasswordStrengthMeter
        level={guessableLevel}
        guessableLevelNames={guessableLevelNames}
      />
      <ul className={styles.passwordPolicy}>
        {passwordPolicyData.map((policy) => (
          <li
            key={policy.messageId}
            className={cn({
              [styles.policySatisfied]: isPasswordPolicySatisfied[policy.key],
            })}
          >
            <Text>
              <FormattedMessage
                id={policy.messageId}
                values={policy.messageValues}
              />
            </Text>
          </li>
        ))}
      </ul>
    </div>
  );
};

export default PasswordField;
