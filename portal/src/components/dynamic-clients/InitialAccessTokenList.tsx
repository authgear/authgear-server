import React, { useContext, useMemo } from "react";
import { Text } from "@radix-ui/themes";
import { Context as MessageContext, FormattedMessage } from "../../intl";
import { Badge } from "../v2/Badge/Badge";
import { SecondaryButton } from "../v2/Button/SecondaryButton/SecondaryButton";
import { InitialAccessTokenType } from "../../graphql/adminapi/globalTypes.generated";
import { formatDatetime } from "../../util/formatDatetime";
import styles from "./InitialAccessTokenList.module.css";

export interface InitialAccessTokenListItem {
  id: string;
  createdAt: string;
  expiresAt: string;
  type: InitialAccessTokenType;
}

export interface InitialAccessTokenListProps {
  tokens: InitialAccessTokenListItem[];
  onRevoke: (token: InitialAccessTokenListItem) => void;
}

export function InitialAccessTokenList({
  tokens,
  onRevoke,
}: InitialAccessTokenListProps): React.ReactElement {
  const { locale } = useContext(MessageContext);

  const rows = useMemo(
    () =>
      tokens.map((token) => ({
        token,
        createdAt: formatDatetime(locale, token.createdAt) ?? "",
        expiresAt: formatDatetime(locale, token.expiresAt) ?? "",
      })),
    [locale, tokens]
  );

  if (tokens.length === 0) {
    return (
      <Text as="p" size="2" color="gray">
        <FormattedMessage id="InitialAccessTokenList.empty" />
      </Text>
    );
  }

  return (
    <div className={styles.table}>
      <div className={styles.headerRow}>
        <Text size="1" color="gray">
          <FormattedMessage id="InitialAccessTokenList.column.type" />
        </Text>
        <Text size="1" color="gray">
          <FormattedMessage id="InitialAccessTokenList.column.created-at" />
        </Text>
        <Text size="1" color="gray">
          <FormattedMessage id="InitialAccessTokenList.column.expires-at" />
        </Text>
        <span aria-hidden={true} />
      </div>
      {rows.map(({ token, createdAt, expiresAt }) => (
        <div key={token.id} className={styles.row}>
          <div>
            {token.type === InitialAccessTokenType.FirstParty ? (
              <Badge
                size="1"
                variant="warning"
                text={
                  <FormattedMessage id="InitialAccessTokenList.type.first-party" />
                }
              />
            ) : (
              <Badge
                size="1"
                variant="neutral"
                text={
                  <FormattedMessage id="InitialAccessTokenList.type.third-party" />
                }
              />
            )}
          </div>
          <Text size="2" className="truncate">
            {createdAt}
          </Text>
          <Text size="2" className="truncate">
            {expiresAt}
          </Text>
          <SecondaryButton
            size="1"
            text={<FormattedMessage id="InitialAccessTokenList.revoke" />}
            onClick={() => {
              onRevoke(token);
            }}
          />
        </div>
      ))}
    </div>
  );
}
