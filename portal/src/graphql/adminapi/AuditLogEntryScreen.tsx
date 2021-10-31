import React, { useMemo, useContext } from "react";
import { useParams } from "react-router-dom";
import { Text, Label } from "@fluentui/react";
import { FormattedMessage, Context } from "@oursky/react-messageformat";
import { gql, useQuery } from "@apollo/client";
import { CopyBlock, dracula } from "react-code-blocks";
import NavBreadcrumb from "../../NavBreadcrumb";
import CommandBarContainer from "../../CommandBarContainer";
import ShowError from "../../ShowError";
import ScreenContent from "../../ScreenContent";
import Widget from "../../Widget";
import { formatDatetime } from "../../util/formatDatetime";
import { useSystemConfig } from "../../context/SystemConfigContext";
import {
  AuditLogEntryQuery,
  AuditLogEntryQueryVariables,
} from "./__generated__/AuditLogEntryQuery";

import styles from "./AuditLogEntryScreen.module.scss";

const QUERY = gql`
  query AuditLogEntryQuery($logID: ID!) {
    node(id: $logID) {
      __typename
      ... on AuditLog {
        id
        createdAt
        activityType
        user {
          id
        }
        ipAddress
        userAgent
        clientID
        data
      }
    }
  }
`;

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
  const { logID } = useParams();

  const { renderToString, locale } = useContext(Context);

  const navBreadcrumbItems = useMemo(() => {
    return [
      { to: "../../", label: <FormattedMessage id="AuditLogScreen.title" /> },
      { to: "./", label: <FormattedMessage id="AuditLogEntryScreen.title" /> },
    ];
  }, []);

  const { data, loading, error, refetch } = useQuery<
    AuditLogEntryQuery,
    AuditLogEntryQueryVariables
  >(QUERY, {
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
  let userID: string | undefined;
  let ipAddress: string | undefined;
  let userAgent: string | undefined;
  let clientID: string | undefined;
  let code: string | undefined;
  if (data?.node?.__typename === "AuditLog") {
    activityType = data.node.activityType;
    loggedAt = formatDatetime(locale, data.node.createdAt) ?? undefined;
    userID = data.node.user?.id;
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
        <div className={styles.widget}>
          <NavBreadcrumb items={navBreadcrumbItems} />
        </div>
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
          {userID && (
            <SummaryText light={true}>
              <FormattedMessage
                id="AuditLogEntryScreen.user-id"
                values={{
                  id: userID,
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
