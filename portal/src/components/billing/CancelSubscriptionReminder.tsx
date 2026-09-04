import React from "react";
import { FormattedMessage } from "../../intl";
import { Callout } from "../v2/Callout/Callout";

interface CancelSubscriptionReminderProps {
  formattedBillingDate: string;
}

export function CancelSubscriptionReminder({
  formattedBillingDate,
}: CancelSubscriptionReminderProps): React.ReactElement {
  return (
    <Callout
      type="info"
      showCloseButton={false}
      text={
        <>
          <FormattedMessage id="DowngradeReminder.title" />
          <br />
          <FormattedMessage
            id="DowngradeReminder.description"
            values={{ date: formattedBillingDate }}
          />
        </>
      }
    />
  );
}
