import React, { useCallback } from "react";
import { useNavigate } from "react-router-dom";
import { Text } from "@radix-ui/themes";
import { FormattedMessage } from "../../intl";
import { SettingsSectionCard } from "../v2/SettingsSectionCard/SettingsSectionCard";
import { SecondaryButton } from "../v2/Button/SecondaryButton/SecondaryButton";
import { useDynamicClientsQueryQuery } from "../../graphql/adminapi/query/dynamicClientsQuery.generated";
import { OAuthClientSource } from "../../graphql/adminapi/globalTypes.generated";
import styles from "./ClientCountCard.module.css";

export interface ClientCountSource {
  source: OAuthClientSource;
  // The smallest block-action quota configured for this source's usage limit,
  // or null when the plan does not cap it. CIMD and DCR clients are counted
  // against separate limits (docs/specs/cimd.md § Client Limit), so neither
  // number can stand in for the other -- hence one entry per source rather
  // than one combined total.
  quota: number | null;
}

export interface ClientCountCardProps {
  sources: ClientCountSource[];
  // Where "View all" goes. Required rather than defaulted: the listing is a
  // sibling of the page this card sits on, not a child, so a relative
  // default would resolve against the caller's own route and land nowhere.
  listPath: string;
}

// One source's count. It is its own component because the count comes from a
// hook, which cannot be called in a loop over `sources`.
const ClientCountStat: React.VFC<ClientCountSource> = function ClientCountStat({
  source,
  quota,
}) {
  // first: 1 because only totalCount is read; the source argument is what
  // makes the count this mechanism's rather than every dynamic client's.
  const { data } = useDynamicClientsQueryQuery({
    variables: { first: 1, source },
    fetchPolicy: "cache-and-network",
  });
  // null only while the query is in flight; the stat then renders an em
  // dash rather than a misleading 0.
  const count = data?.dynamicClients?.totalCount ?? null;

  // Its own strings, not DynamicClientSource.*: those are the list's Type
  // column and filter options, where "Via CIMD" would read wrong.
  const labelID =
    source === OAuthClientSource.Cimd
      ? "ClientCountCard.source.cimd"
      : "ClientCountCard.source.dcr";

  return (
    <div className={styles.stat}>
      <Text as="p" size="1" color="gray" className={styles.statLabel}>
        <FormattedMessage id={labelID} />
      </Text>
      <Text as="p" size="6" weight="bold" className={styles.statValue}>
        {count == null ? (
          <FormattedMessage id="ClientCountCard.count.unknown" />
        ) : quota != null ? (
          <FormattedMessage
            id="ClientCountCard.count.quota"
            values={{ count, quota }}
          />
        ) : (
          <FormattedMessage id="ClientCountCard.count" values={{ count }} />
        )}
      </Text>
    </div>
  );
};

/**
 * How many clients each onboarding mechanism has produced so far, and a way
 * into the list holding them.
 *
 * Every count is always shown, including at 0: turning a mechanism off
 * neither deletes nor disables the clients it already resolved or registered,
 * so the card has to keep reporting them either way.
 */
export const ClientCountCard: React.VFC<ClientCountCardProps> =
  function ClientCountCard({ sources, listPath }) {
    const navigate = useNavigate();

    const onViewAllClick = useCallback(() => {
      // No ?source: this card covers every mechanism, so the list opens
      // unfiltered rather than on one of the sources it reports.
      navigate(listPath);
    }, [navigate, listPath]);

    return (
      <SettingsSectionCard
        contentClassName="gap-4"
        title={<FormattedMessage id="ClientCountCard.title" />}
      >
        <div className={styles.stats}>
          {sources.map((entry) => (
            <ClientCountStat
              key={entry.source}
              source={entry.source}
              quota={entry.quota}
            />
          ))}
        </div>
        <div className="self-start">
          <SecondaryButton
            size="2"
            text={<FormattedMessage id="ClientCountCard.view-all" />}
            onClick={onViewAllClick}
          />
        </div>
      </SettingsSectionCard>
    );
  };
