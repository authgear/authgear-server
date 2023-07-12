import React from "react";
import cn from "classnames";
import { DefaultEffects, Text } from "@fluentui/react";
import { FormattedMessage } from "@oursky/react-messageformat";
import styles from "./SubscriptionInvoiceErrorSummary.module.css";
import PrimaryButton from "../../PrimaryButton";

export interface SubscriptionInvoiceErrorSummaryProps {
  className?: string;
  invoiceURL?: string;
}

function SubscriptionInvoiceErrorSummary(
  props: SubscriptionInvoiceErrorSummaryProps
): React.ReactElement | null {
  const { className, invoiceURL } = props;
  return (
    <div
      className={cn(className, styles.root)}
      style={{
        boxShadow: DefaultEffects.elevation4,
      }}
    >
      <Text block={true} variant="xLarge">
        <FormattedMessage id="SubscriptionInvoiceErrorSummary.title" />
      </Text>
      <Text block={true} variant="medium">
        <FormattedMessage id="SubscriptionInvoiceErrorSummary.description" />
      </Text>
      <PrimaryButton
        text={<FormattedMessage id="SubscriptionInvoiceErrorSummary.pay" />}
        target="_blank"
        as="a"
        href={invoiceURL}
      />
    </div>
  );
}

export default SubscriptionInvoiceErrorSummary;
