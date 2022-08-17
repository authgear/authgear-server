import React, { useMemo, useContext } from "react";
import { useParams, useLocation } from "react-router-dom";
import { Text, Label } from "@fluentui/react";
import { FormattedMessage, Context } from "@oursky/react-messageformat";
import { useQuery } from "@apollo/client";
import { CopyBlock, dracula } from "react-code-blocks";
import NavBreadcrumb from "../../NavBreadcrumb";
import CommandBarContainer from "../../CommandBarContainer";
import ShowError from "../../ShowError";
import ScreenContent from "../../ScreenContent";
import Widget from "../../Widget";
import { formatDatetime } from "../../util/formatDatetime";
import { extractRawID } from "../../util/graphql";
import { useSystemConfig } from "../../context/SystemConfigContext";
import {
  AuditLogEntryQueryQuery,
  AuditLogEntryQueryQueryVariables,
  AuditLogEntryQueryDocument,
  AuditLogEntryFragment,
} from "./query/auditLogEntryQuery.generated";

import styles from "./AuditLogEntryScreen.module.css";

function getRawUserIDFromAuditLog(
  node: AuditLogEntryFragment
): string | undefined {
  // The simple case is just use the user.id.
  const userID = node.user?.id ?? null;
  if (userID != null) {
    return extractRawID(userID);
  }
  // Otherwise use the user ID in the payload.
  const rawUserID = node.data?.payload?.user?.id;
  return rawUserID ?? undefined;
}

function SummaryText(props: { children: React.ReactNode; light?: boolean }) {
  const { themes } = useSystemConfig();
  const lightColor = themes.main.palette.neutralTertiary;
  const { children, light } = props;
  return (
    <Text
      as="p"
      block={true}
      style={{
        color: light === true ? lightColor : undefined,
      }}
    >
      {children}
    </Text>
  );
}

// eslint-disable-next-line complexity
const AuditLogEntryScreen: React.FC = function AuditLogEntryScreen() {
  const { logID } = useParams() as { logID: string };
  const location = useLocation();
  const state = location.state as { searchParams?: string };

  const { renderToString, locale } = useContext(Context);

  const navBreadcrumbItems = useMemo(() => {
    return [
      {
        to: `./../..?${state.searchParams ?? ""}`,
        label: <FormattedMessage id="AuditLogScreen.title" />,
      },
      { to: ".", label: <FormattedMessage id="AuditLogEntryScreen.title" /> },
    ];
  }, [state]);

  const { data, loading, error, refetch } = useQuery<
    AuditLogEntryQueryQuery,
    AuditLogEntryQueryQueryVariables
  >(AuditLogEntryQueryDocument, {
    variables: {
      logID,
    },
  });

  const messageBar = useMemo(() => {
    if (error != null) {
      return <ShowError error={error} onRetry={refetch} />;
    }
    return null;
  }, [error, refetch]);

  let activityType: string | undefined;
  let loggedAt: string | undefined;
  let rawUserID: string | undefined;
  let ipAddress: string | undefined;
  let userAgent: string | undefined;
  let clientID: string | undefined;
  let code: string | undefined;
  let deleted = false;
  if (data?.node?.__typename === "AuditLog") {
    activityType = data.node.activityType;
    loggedAt = formatDatetime(locale, data.node.createdAt) ?? undefined;
    rawUserID = getRawUserIDFromAuditLog(data.node);
    deleted = data.node.user?.id == null && rawUserID != null;
    ipAddress = data.node.ipAddress ?? undefined;
    userAgent = data.node.userAgent ?? undefined;
    clientID = data.node.clientID ?? undefined;
    code =
      data.node.data != null
        ? JSON.stringify(data.node.data, null, 2)
        : undefined;
  }

  return (
    <CommandBarContainer isLoading={loading} messageBar={messageBar}>
      <ScreenContent>
        <NavBreadcrumb items={navBreadcrumbItems} />
        <Widget className={styles.widget}>
          {activityType && (
            <SummaryText>
              <FormattedMessage
                id="AuditLogEntryScreen.activity-type"
                values={{
                  type: renderToString("AuditLogActivityType." + activityType),
                }}
              />
            </SummaryText>
          )}
          {loggedAt && (
            <SummaryText light={true}>
              <FormattedMessage
                id="AuditLogEntryScreen.logged-at"
                values={{
                  datetime: loggedAt,
                }}
              />
            </SummaryText>
          )}
          {rawUserID && (
            <SummaryText light={true}>
              <FormattedMessage
                id="AuditLogEntryScreen.user-id"
                values={{
                  id: rawUserID,
                  deleted: String(deleted),
                }}
              />
            </SummaryText>
          )}
          {ipAddress && (
            <SummaryText light={true}>
              <FormattedMessage
                id="AuditLogEntryScreen.ip-address"
                values={{
                  ip: ipAddress,
                }}
              />
            </SummaryText>
          )}
          {userAgent && (
            <SummaryText light={true}>
              <FormattedMessage
                id="AuditLogEntryScreen.user-agent"
                values={{
                  userAgent,
                }}
              />
            </SummaryText>
          )}
          {clientID && (
            <SummaryText light={true}>
              <FormattedMessage
                id="AuditLogEntryScreen.client-id"
                values={{
                  clientID,
                }}
              />
            </SummaryText>
          )}
        </Widget>
        <Widget className={styles.widget}>
          <Label>
            <FormattedMessage id="AuditLogEntryScreen.raw-event-log" />
          </Label>
          {code != null && (
            <div className={styles.codeBlock}>
              <CopyBlock
                text={code}
                language="json"
                codeBlock={true}
                theme={dracula}
              />
            </div>
          )}
        </Widget>
      </ScreenContent>
    </CommandBarContainer>
  );
};

export default AuditLogEntryScreen;
