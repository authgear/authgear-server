import React, { useState, useMemo, useCallback } from "react";
import { FormattedMessage } from "@oursky/react-messageformat";
import { gql, useQuery } from "@apollo/client";
import NavBreadcrumb from "../../NavBreadcrumb";
import AuditLogList from "./AuditLogList";
import CommandBarContainer from "../../CommandBarContainer";
import ShowError from "../../ShowError";
import { encodeOffsetToCursor } from "../../util/pagination";
import {
  AuditLogListQuery,
  AuditLogListQueryVariables,
} from "./__generated__/AuditLogListQuery";

import styles from "./AuditLogScreen.module.scss";

const pageSize = 10;

const QUERY = gql`
  query AuditLogListQuery($pageSize: Int!, $cursor: String) {
    auditLogs(first: $pageSize, after: $cursor) {
      edges {
        node {
          id
          createdAt
          activityType
          user {
            id
          }
        }
      }
      totalCount
    }
  }
`;

const AuditLogScreen: React.FC = function AuditLogScreen() {
  const [offset, setOffset] = useState(0);

  const items = useMemo(() => {
    return [{ to: ".", label: <FormattedMessage id="AuditLogScreen.title" /> }];
  }, []);

  const cursor = useMemo(() => {
    if (offset === 0) {
      return null;
    }
    return encodeOffsetToCursor(offset - 1);
  }, [offset]);

  const onChangeOffset = useCallback((offset) => {
    setOffset(offset);
  }, []);

  const { data, error, loading, refetch } = useQuery<
    AuditLogListQuery,
    AuditLogListQueryVariables
  >(QUERY, {
    variables: {
      pageSize,
      cursor,
    },
    fetchPolicy: "network-only",
  });

  const messageBar = useMemo(() => {
    if (error != null) {
      return <ShowError error={error} onRetry={refetch} />;
    }
    return null;
  }, [error, refetch]);

  return (
    <CommandBarContainer
      isLoading={loading}
      className={styles.root}
      messageBar={messageBar}
    >
      <main className={styles.content}>
        <NavBreadcrumb items={items} />
        <AuditLogList
          className={styles.list}
          loading={loading}
          auditLogs={data?.auditLogs ?? null}
          offset={offset}
          pageSize={pageSize}
          totalCount={data?.auditLogs?.totalCount ?? undefined}
          onChangeOffset={onChangeOffset}
        />
      </main>
    </CommandBarContainer>
  );
};

export default AuditLogScreen;
