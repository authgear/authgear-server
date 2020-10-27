import React from "react";
import cn from "classnames";
import { MessageBar, MessageBarType, Stack, Text } from "@fluentui/react";
import { FormattedMessage } from "@oursky/react-messageformat";

import { ValidationFailedErrorInfoCause } from "./validation";

import styles from "./ShowUnhandledValidationErrorCauses.module.scss";

interface ShowUnhandledValidationErrorCauseProps {
  causes?: ValidationFailedErrorInfoCause[];
}

const ShowUnhandledValidationErrorCause: React.FC<ShowUnhandledValidationErrorCauseProps> = function ShowUnhandledValidationErrorCause(
  props: ShowUnhandledValidationErrorCauseProps
) {
  const { causes } = props;
  if (causes == null || causes.length === 0) {
    return null;
  }

  const children = causes.map((cause, index) => (
    <Stack key={index} className={styles.cause}>
      <Text className={styles.text}>
        <FormattedMessage
          id="ShowUnhandledValidationErrorCause.kind"
          values={{ kind: cause.kind }}
        />
      </Text>
      <Text className={styles.text}>
        <FormattedMessage
          id="ShowUnhandledValidationErrorCause.location"
          values={{ location: cause.location }}
        />
      </Text>
      <Text className={styles.text}>
        <FormattedMessage
          id="ShowUnhandledValidationErrorCause.details"
          values={{ details: JSON.stringify(cause.details, null, 2) }}
        />
      </Text>
    </Stack>
  ));
  children.unshift(
    <Text className={cn(styles.text, styles.title)}>
      <FormattedMessage id="ShowUnhandledValidationErrorCause.title" />
    </Text>
  );

  return (
    <MessageBar messageBarType={MessageBarType.error}>{children}</MessageBar>
  );
};

export default ShowUnhandledValidationErrorCause;
