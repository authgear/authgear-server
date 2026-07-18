import React, { useId } from "react";
import { useIntl, FormattedMessage } from "react-intl";
import ExternalLink from "../../../ExternalLink";
import WidgetSubtitle from "../../../WidgetSubtitle";
import type { AuthMethodChoice as Stage2Choice } from "./frameworks";
import styles from "./AuthMethodChoice.module.css";

export interface AuthMethodChoiceProps {
  value: Stage2Choice | null;
  onChange: (value: Stage2Choice) => void;
  nginxDocsHref: string;
}

export const AuthMethodChoiceComponent: React.FC<AuthMethodChoiceProps> = ({
  value,
  onChange,
  nginxDocsHref,
}) => {
  const groupId = useId();
  const { formatMessage } = useIntl();
  return (
    <div className={styles.root}>
      <WidgetSubtitle>
        <FormattedMessage id="CreateOAuthClientScreen.stage2.question" />
      </WidgetSubtitle>
      <div className={styles.options}>
        <label className={styles.option}>
          <input
            type="radio"
            name={groupId}
            value="token"
            checked={value === "token"}
            onChange={() => onChange("token")}
            className={styles.radio}
          />
          <div className={styles.content}>
            <div className={styles.title}>
              {formatMessage({
                id: "CreateOAuthClientScreen.stage2.option.token",
              })}
            </div>
            <div className={styles.description}>
              <FormattedMessage id="CreateOAuthClientScreen.stage2.option.token.description" />
            </div>
          </div>
        </label>
        <label className={styles.option}>
          <input
            type="radio"
            name={groupId}
            value="cookie"
            checked={value === "cookie"}
            onChange={() => onChange("cookie")}
            className={styles.radio}
          />
          <div className={styles.content}>
            <div className={styles.title}>
              {formatMessage({
                id: "CreateOAuthClientScreen.stage2.option.cookie",
              })}
            </div>
            <div className={styles.description}>
              <FormattedMessage id="CreateOAuthClientScreen.stage2.option.cookie.description" />
            </div>
          </div>
        </label>
      </div>
      <div className={styles.cookieHelp}>
        <ExternalLink href={nginxDocsHref}>
          <FormattedMessage id="CreateOAuthClientScreen.stage2.option.cookie.docs-link" />
        </ExternalLink>
      </div>
    </div>
  );
};
