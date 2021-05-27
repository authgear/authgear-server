import React, { useMemo } from "react";
import { Text } from "@fluentui/react";
import { FormattedMessage } from "@oursky/react-messageformat";

import { GuessableLevel, GuessableLevelNames } from "./PasswordField";

import styles from "./PasswordStrengthMeter.module.scss";

interface PasswordStrengthMeterProps {
  className?: string;
  level: GuessableLevel;
  guessableLevelNames: GuessableLevelNames;
}

const PasswordStrengthMeter: React.FC<PasswordStrengthMeterProps> =
  function PasswordStrengthMeter(props: PasswordStrengthMeterProps) {
    const { className, level, guessableLevelNames } = props;
    const descriptionClassName = useMemo(
      () => styles[`passwordStrengthMeterDescriptionLevel${level}`],
      [level]
    );
    return (
      <div className={className}>
        <meter className={styles.passwordStrengthMeter} value={level} />
        <div className={styles.passwordStrengthMeterDescriptionContainer}>
          <Text>
            <FormattedMessage id="PasswordStrengthMeter.password-strength" />
            {": "}
          </Text>
          <Text className={descriptionClassName}>
            {guessableLevelNames[level]}
          </Text>
        </div>
      </div>
    );
  };

export default PasswordStrengthMeter;
