import React, { useContext, useMemo } from "react";
import cn from "classnames";
import { PartialTheme, Text, ThemeProvider } from "@fluentui/react";
import styles from "./SubscriptionScreenFooter.module.css";
import {
  Context as MessageContext,
  FormattedMessage,
  Values as MessageValues,
} from "@oursky/react-messageformat";
import iconContact from "../../images/subscription-contact.svg";
import iconDoc from "../../images/subscription-doc.svg";
import { formatDatetime } from "../../util/formatDatetime";
import LinkButton from "../../LinkButton";

function FooterRow({
  iconSrc,
  messageId,
  messageValues,
}: {
  iconSrc: string;
  messageId: string;
  messageValues?: MessageValues;
}) {
  return (
    <div className={styles.leftItem}>
      <img className={styles.leftItemIcon} src={iconSrc} />
      <div className={styles.leftItemText}>
        <Text>
          <FormattedMessage id={messageId} values={messageValues} />
        </Text>
      </div>
    </div>
  );
}

const CANCEL_THEME: PartialTheme = {
  palette: {
    themePrimary: "#c8c8c8",
    neutralPrimary: "#c8c8c8",
  },
  semanticColors: {
    linkHovered: "#c8c8c8",
  },
};

interface SubscriptionScreenFooterProps {
  className?: string;
  onClickEnterprisePlan: (e: React.MouseEvent) => void;
  onClickCancel: (e: React.MouseEvent) => void;
  isStripePlan: boolean;
  subscriptionCancelled: boolean;
  subscriptionEndedAt?: string;
}

export const SubscriptionScreenFooter: React.VFC<SubscriptionScreenFooterProps> =
  function SubscriptionScreenFooter({
    className,
    onClickEnterprisePlan,
    onClickCancel,
    isStripePlan,
    subscriptionCancelled,
    subscriptionEndedAt,
  }) {
    const { locale } = useContext(MessageContext);

    const formattedSubscriptionEndedAt = useMemo(() => {
      return subscriptionEndedAt != null
        ? formatDatetime(locale, subscriptionEndedAt)
        : null;
    }, [subscriptionEndedAt, locale]);

    return (
      <div className={cn(styles.container, className)}>
        <div className={cn(styles.column, styles.leftColumn)}>
          <FooterRow
            iconSrc={iconContact}
            messageId="SubscriptionScreen.footer.enterprise-plan"
            messageValues={{
              onClick: onClickEnterprisePlan,
            }}
          />
          <FooterRow
            iconSrc={iconDoc}
            messageId="SubscriptionScreen.footer.pricing-details"
          />
        </div>
        <div className={cn(styles.column, styles.rightColumn)}>
          <Text block={true}>
            <FormattedMessage id="SubscriptionScreen.footer.tax" />
          </Text>
          {isStripePlan ? (
            <>
              <Text block={true}>
                <FormattedMessage id="SubscriptionScreen.footer.usage-delay-disclaimer" />
              </Text>
            </>
          ) : null}
          {subscriptionCancelled ? (
            <Text block={true}>
              <FormattedMessage
                id="SubscriptionScreen.footer.expire"
                values={{
                  date: formattedSubscriptionEndedAt ?? "",
                }}
              />
            </Text>
          ) : (
            <ThemeProvider theme={CANCEL_THEME}>
              <LinkButton onClick={onClickCancel}>
                <Text>
                  <FormattedMessage id="SubscriptionScreen.footer.cancel" />
                </Text>
              </LinkButton>
            </ThemeProvider>
          )}
        </div>
      </div>
    );
  };
