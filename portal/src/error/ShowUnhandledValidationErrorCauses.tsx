import React from "react";
import { MessageBar, MessageBarType, Stack, Text } from "@fluentui/react";
import { FormattedMessage } from "@oursky/react-messageformat";

import { ValidationFailedErrorInfoCause } from "./validation";

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
    <Stack key={index}>
      <Text>
        <FormattedMessage
          id="ShowUnhandledValidationErrorCause.kind"
          values={{ kind: cause.kind }}
        />
      </Text>
      <Text>
        <FormattedMessage
          id="ShowUnhandledValidationErrorCause.location"
          values={{ location: cause.location }}
        />
      </Text>
      <Text>
        <FormattedMessage
          id="ShowUnhandledValidationErrorCause.kind"
          values={{ details: String(cause.details) }}
        />
      </Text>
    </Stack>
  ));
  children.unshift(
    <Text>
      <FormattedMessage id="ShowUnhandledValidationErrorCause.title" />
    </Text>
  );

  return (
    <MessageBar messageBarType={MessageBarType.error}>{children}</MessageBar>
  );
};

export default ShowUnhandledValidationErrorCause;
