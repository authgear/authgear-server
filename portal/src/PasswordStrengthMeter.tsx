import React, { useMemo } from "react";
import { Text } from "@radix-ui/themes";
import { FormattedMessage } from "./intl";

import styles from "./PasswordStrengthMeter.module.css";
import { GuessableLevel } from "./util/zxcvbn";
import { GuessableLevelNames } from "./PasswordField";

interface PasswordStrengthMeterProps {
  className?: string;
  level: GuessableLevel;
  guessableLevelNames: GuessableLevelNames;
}

const PasswordStrengthMeter: React.VFC<PasswordStrengthMeterProps> =
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
          <Text as="span" size="1">
            <FormattedMessage id="PasswordStrengthMeter.password-strength" />
            {": "}
          </Text>
          <Text as="span" size="1" className={descriptionClassName}>
            {guessableLevelNames[level]}
          </Text>
        </div>
      </div>
    );
  };

export default PasswordStrengthMeter;
