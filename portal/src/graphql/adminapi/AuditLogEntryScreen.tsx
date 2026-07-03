import React, { useMemo, useContext } from "react";
import { useParams, useLocation } from "react-router-dom";
import { Text } from "@radix-ui/themes";
import { ChevronLeftIcon } from "@radix-ui/react-icons";
import { FormattedMessage, Context } from "../../intl";
import { useQuery } from "@apollo/client";
import Link from "../../Link";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import ScreenLayoutScrollView from "../../ScreenLayoutScrollView";
import ScreenContent from "../../ScreenContent";
import { formatDatetime } from "../../util/formatDatetime";
import { extractRawID } from "../../util/graphql";
import {
  AuditLogEntryQueryQuery,
  AuditLogEntryQueryQueryVariables,
  AuditLogEntryQueryDocument,
  AuditLogEntryFragment,
} from "./query/auditLogEntryQuery.generated";

import styles from "./AuditLogEntryScreen.module.css";
import CodeEditor from "../../CodeEditor";

const CODE_EDITOR_OPTIONS = {
  readOnly: true,
  minimap: { enabled: false },
  wordWrap: "on" as const,
  wrappingIndent: "deepIndent" as const,
  renderLineHighlight: "none" as const,
  scrollBeyondLastLine: false,
};

function getRawUserIDFromAuditLog(
  node: AuditLogEntryFragment
): string | undefined {
  const userID = node.user?.id ?? null;
  if (userID != null) {
    return extractRawID(userID);
  }
  const rawUserID = node.data?.payload?.user?.id;
  return rawUserID ?? undefined;
}

interface TableRowProps {
  label: React.ReactNode;
  value: React.ReactNode;
}

function TableRow({ label, value }: TableRowProps) {
  return (
    <div className={styles.tableRow}>
      <div className={styles.cellLabel}>{label}</div>
      <div className={styles.cellValue}>{value}</div>
    </div>
  );
}

const AuditLogEntryScreen: React.VFC = function AuditLogEntryScreen() {
  const { logID, appID } = useParams() as { logID: string; appID: string };
  const location = useLocation();
  const state = location.state as { searchParams?: string } | undefined;

  const { renderToString, locale } = useContext(Context);

  const backURL = `/project/${appID}/audit-log?${state?.searchParams ?? ""}`;

  const { data, loading, error, refetch } = useQuery<
    AuditLogEntryQueryQuery,
    AuditLogEntryQueryQueryVariables
  >(AuditLogEntryQueryDocument, {
    variables: {
      logID,
    },
  });

  const auditLog = useMemo(() => {
    if (data?.node?.__typename === "AuditLog") {
      return data.node;
    }
    return null;
  }, [data]);

  const activityType = auditLog?.activityType;
  const loggedAt =
    auditLog != null
      ? (formatDatetime(locale, auditLog.createdAt) ?? undefined)
      : undefined;
  const rawUserID = auditLog != null ? getRawUserIDFromAuditLog(auditLog) : undefined;
  const deleted = auditLog != null && auditLog.user?.id == null && rawUserID != null;
  const ipAddress = auditLog?.ipAddress ?? undefined;
  const userAgent = auditLog?.userAgent ?? undefined;
  const clientID = auditLog?.clientID ?? undefined;
  const code =
    auditLog?.data != null
      ? JSON.stringify(auditLog.data, null, 2)
      : undefined;

  if (loading) {
    return <ShowLoading />;
  }

  return (
    <ScreenLayoutScrollView>
      <ScreenContent layout="list">
        {/* Header: back link + title */}
        <div className={styles.widget}>
          <Link to={backURL} className={styles.backLink}>
            <ChevronLeftIcon className={styles.backLinkIcon} />
            <span>
              <FormattedMessage id="AuditLogScreen.title" />
            </span>
          </Link>
          <Text as="p" size="5" weight="bold" className={styles.pageTitle}>
            <FormattedMessage id="AuditLogEntryScreen.title" />
          </Text>
        </div>

        {error != null ? (
          <div className={styles.widget}>
            <ShowError error={error} onRetry={refetch} />
          </div>
        ) : null}

        {/* Event details table */}
        {auditLog != null ? (
          <div className={styles.widget}>
            <div className={styles.tableWrapper}>
              <div className={styles.table}>
                <div className={styles.tableHeader}>
                  <div className={styles.headerCellLabel}>
                    <FormattedMessage id="AuditLogEntryScreen.table.event-description" />
                  </div>
                  <div className={styles.headerCellValue}>
                    <FormattedMessage id="AuditLogEntryScreen.table.events-details" />
                  </div>
                </div>

                {activityType != null ? (
                  <TableRow
                    label={
                      <FormattedMessage id="AuditLogEntryScreen.field.activity-type" />
                    }
                    value={renderToString("AuditLogActivityType." + activityType)}
                  />
                ) : null}

                {loggedAt != null ? (
                  <TableRow
                    label={
                      <FormattedMessage id="AuditLogEntryScreen.field.logged-at" />
                    }
                    value={loggedAt}
                  />
                ) : null}

                {rawUserID != null ? (
                  <TableRow
                    label={
                      <FormattedMessage id="AuditLogEntryScreen.field.user-id" />
                    }
                    value={
                      deleted ? (
                        <>
                          {rawUserID}
                          <span className={styles.deletedSuffix}>
                            {" "}
                            <FormattedMessage id="AuditLogEntryScreen.field.user-id.deleted" />
                          </span>
                        </>
                      ) : (
                        rawUserID
                      )
                    }
                  />
                ) : null}

                {ipAddress != null ? (
                  <TableRow
                    label={
                      <FormattedMessage id="AuditLogEntryScreen.field.ip-address" />
                    }
                    value={ipAddress}
                  />
                ) : null}

                {userAgent != null ? (
                  <TableRow
                    label={
                      <FormattedMessage id="AuditLogEntryScreen.field.user-agent" />
                    }
                    value={userAgent}
                  />
                ) : null}

                {clientID != null ? (
                  <TableRow
                    label={
                      <FormattedMessage id="AuditLogEntryScreen.field.client-id" />
                    }
                    value={clientID}
                  />
                ) : null}
              </div>
            </div>
          </div>
        ) : null}

        {/* Raw Event Log */}
        {code != null ? (
          <div className={styles.widget}>
            <div className={styles.editorCard}>
              <div className={styles.editorCardHeader}>
                <Text as="p" size="3" weight="medium">
                  <FormattedMessage id="AuditLogEntryScreen.raw-event-log" />
                </Text>
              </div>
              <CodeEditor
                className={styles.codeEditor}
                language="json"
                value={code}
                options={CODE_EDITOR_OPTIONS}
              />
            </div>
          </div>
        ) : null}
      </ScreenContent>
    </ScreenLayoutScrollView>
  );
};

export default AuditLogEntryScreen;
