import { Icon, Text } from "@fluentui/react";
import { FormattedMessage } from "../../intl";
import React from "react";

interface CancelSubscriptionReminderProps {
  formattedBillingDate: string;
}

export function CancelSubscriptionReminder({
  formattedBillingDate,
}: CancelSubscriptionReminderProps): React.ReactElement {
  return (
    <section className="rounded-sm py-4 px-6 bg-brand-50">
      <div className="flex items-center gap-x-1">
        <Icon className="text-theme-primary text-[1rem]" iconName="Info" />
        <Text variant="mediumPlus" className="font-semibold">
          <FormattedMessage id="DowngradeReminder.title" />
        </Text>
      </div>
      <Text block={true} variant="medium">
        <FormattedMessage
          id="DowngradeReminder.description"
          values={{ date: formattedBillingDate }}
        />
      </Text>
    </section>
  );
}
