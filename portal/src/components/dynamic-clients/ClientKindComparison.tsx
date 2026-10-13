import React from "react";
import { Text } from "@radix-ui/themes";
import { FormattedMessage } from "../../intl";
import { SettingsSectionCard } from "../v2/SettingsSectionCard/SettingsSectionCard";
import { CardTable } from "../v2/CardTable/CardTable";
import { ClientKindBadge } from "./ClientKindBadge";
import styles from "./ClientKindComparison.module.css";

interface ComparisonRow {
  labelID: string;
  thirdPartyID: string;
  firstPartyID: string;
}

const ROWS: ComparisonRow[] = [
  {
    labelID: "ClientKindComparison.use-for.label",
    thirdPartyID: "ClientKindComparison.use-for.third-party",
    firstPartyID: "ClientKindComparison.use-for.first-party",
  },
  {
    labelID: "ClientKindComparison.register.label",
    thirdPartyID: "ClientKindComparison.register.third-party",
    firstPartyID: "ClientKindComparison.register.first-party",
  },
  {
    labelID: "ClientKindComparison.consent.label",
    thirdPartyID: "ClientKindComparison.consent.third-party",
    firstPartyID: "ClientKindComparison.consent.first-party",
  },
  {
    labelID: "ClientKindComparison.access-token.label",
    thirdPartyID: "ClientKindComparison.access-token.third-party",
    firstPartyID: "ClientKindComparison.access-token.first-party",
  },
];

export function ClientKindComparison(): React.ReactElement {
  return (
    <SettingsSectionCard
      layout="stacked"
      title={<FormattedMessage id="ClientKindComparison.title" />}
      description={<FormattedMessage id="ClientKindComparison.description" />}
    >
      <CardTable className={styles.table}>
        <CardTable.Header>
          <CardTable.HeaderCell className={styles.colLabel} />
          <CardTable.HeaderCell className={styles.colKind}>
            <ClientKindBadge firstParty={false} />
          </CardTable.HeaderCell>
          <CardTable.HeaderCell className={styles.colKind}>
            <ClientKindBadge firstParty={true} />
          </CardTable.HeaderCell>
        </CardTable.Header>
        {ROWS.map((row) => (
          <CardTable.Row key={row.labelID}>
            <CardTable.Cell className={styles.colLabel}>
              <Text size="2" weight="medium">
                <FormattedMessage id={row.labelID} />
              </Text>
            </CardTable.Cell>
            <CardTable.Cell className={styles.colKind}>
              <Text size="2">
                <FormattedMessage id={row.thirdPartyID} />
              </Text>
            </CardTable.Cell>
            <CardTable.Cell className={styles.colKind}>
              <Text size="2">
                <FormattedMessage id={row.firstPartyID} />
              </Text>
            </CardTable.Cell>
          </CardTable.Row>
        ))}
      </CardTable>
    </SettingsSectionCard>
  );
}
