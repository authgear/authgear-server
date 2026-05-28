import React, { useCallback, useContext, useMemo } from "react";
import { useParams } from "react-router-dom";
import {
  IColumn,
  MessageBar,
  SelectionMode,
  ShimmeredDetailsList,
  Text,
} from "@fluentui/react";
import { Context, FormattedMessage } from "../../intl";
import Link from "../../Link";
import ShowError from "../../ShowError";
import { formatDatetime } from "../../util/formatDatetime";
import { extractRawID } from "../../util/graphql";
import {
  useAuditLogListQueryQuery,
  AuditLogEdgesNodeFragment,
} from "./query/auditLogListQuery.generated";
import {
  AuditLogActivityType,
  SortDirection,
} from "./globalTypes.generated";
import { ACTIVITY_TYPE_ALL } from "../../components/audit-log/ActivityTypeFilterDropdown";
import {
  ADMIN_ACTIVITY_TYPES,
  AuditLogKind,
  USER_ACTIVITY_TYPES,
} from "./auditLogActivityTypes";
import styles from "./UserDetailsLogs.module.css";

const LOG_PREVIEW_PAGE_SIZE = 5;

interface UserDetailsLogsProps {
  userID: string;
}

function buildAuditLogListSearchParams(
  kind: AuditLogKind,
  rawUserID: string
): string {
  return new URLSearchParams({
    kind,
    q: rawUserID,
    page: "1",
    order_by: SortDirection.Desc,
    activity_type: ACTIVITY_TYPE_ALL,
    last_updated_at: Date.now().toString(),
    from: "",
    to: "",
  }).toString();
}

function buildAuditLogListHref(
  appID: string,
  kind: AuditLogKind,
  rawUserID: string
): string {
  return `/project/${appID}/audit-log?${buildAuditLogListSearchParams(
    kind,
    rawUserID
  )}`;
}

interface LogTableItem {
  id: string;
  activityType: string;
  createdAt: string;
}

interface UserDetailsLogsActivitySectionProps {
  titleMessageID: string;
  viewAllMessageID: string;
  auditLogKind: AuditLogKind;
  userID: string;
  rawUserID: string;
  activityTypes: AuditLogActivityType[];
}

const UserDetailsLogsActivitySection: React.VFC<
  UserDetailsLogsActivitySectionProps
> = function UserDetailsLogsActivitySection(props) {
  const {
    titleMessageID,
    viewAllMessageID,
    auditLogKind,
    userID,
    rawUserID,
    activityTypes,
  } = props;
  const { appID } = useParams() as { appID: string };
  const { renderToString, locale } = useContext(Context);

  const { data, error, loading, refetch } = useAuditLogListQueryQuery({
    variables: {
      pageSize: LOG_PREVIEW_PAGE_SIZE,
      activityTypes,
      userIDs: [userID],
      sortDirection: SortDirection.Desc,
    },
    fetchPolicy: "network-only",
  });

  const auditLogListHref = useMemo(
    () => buildAuditLogListHref(appID, auditLogKind, rawUserID),
    [appID, auditLogKind, rawUserID]
  );

  const auditLogListSearchParams = useMemo(
    () => buildAuditLogListSearchParams(auditLogKind, rawUserID),
    [auditLogKind, rawUserID]
  );

  const columns: IColumn[] = useMemo(
    () => [
      {
        key: "activityType",
        fieldName: "activityType",
        name: renderToString("UserDetails.logs.column.activity"),
        minWidth: 300,
        maxWidth: 400,
        className: styles.cell,
      },
      {
        key: "createdAt",
        fieldName: "createdAt",
        name: renderToString("UserDetails.logs.column.timestamp"),
        minWidth: 220,
        maxWidth: 220,
        className: styles.cell,
      },
    ],
    [renderToString]
  );

  const totalCount = data?.auditLogs?.totalCount;
  const showViewAllLink =
    !loading &&
    error == null &&
    totalCount != null &&
    totalCount > LOG_PREVIEW_PAGE_SIZE;

  const items: LogTableItem[] = useMemo(() => {
    const edges = data?.auditLogs?.edges;
    const result: LogTableItem[] = [];
    if (edges == null) {
      return result;
    }
    for (const edge of edges) {
      const node: AuditLogEdgesNodeFragment | null | undefined = edge?.node;
      if (node == null) {
        continue;
      }
      result.push({
        id: node.id,
        activityType: renderToString(
          "AuditLogActivityType." + node.activityType
        ),
        createdAt: formatDatetime(locale, node.createdAt) ?? "-",
      });
    }
    return result;
  }, [
    data?.auditLogs?.edges,
    locale,
    renderToString,
  ]);

  const onRenderItemColumn = useCallback(
    (item: LogTableItem, _index?: number, column?: IColumn) => {
      const text = item[column?.key as keyof LogTableItem] ?? "-";
      if (column?.key === "activityType") {
        return (
          <Link
            to={`/project/${appID}/audit-log/${item.id}/details`}
            state={{ searchParams: auditLogListSearchParams }}
          >
            {text}
          </Link>
        );
      }
      return <span>{text}</span>;
    },
    [appID, auditLogListSearchParams]
  );

  const isEmpty = !loading && items.length === 0;

  return (
    <section className={styles.section}>
      <Text as="h2" block={true} className={styles.sectionTitle}>
        <FormattedMessage id={titleMessageID} />
      </Text>
      {error != null ? (
        <ShowError error={error} onRetry={refetch} />
      ) : (
        <div className={styles.tableArea}>
          <ShimmeredDetailsList
            enableShimmer={loading}
            enableUpdateAnimations={false}
            selectionMode={SelectionMode.none}
            columns={columns}
            items={items}
            onRenderItemColumn={onRenderItemColumn}
          />
          {isEmpty ? (
            <MessageBar className={styles.emptyMessageBar}>
              <FormattedMessage id="UserDetails.logs.empty" />
            </MessageBar>
          ) : null}
          {showViewAllLink ? (
            <div className={styles.viewAllRow}>
              <Link
                className={styles.viewAllLink}
                to={auditLogListHref}
              >
                <FormattedMessage id={viewAllMessageID} />
              </Link>
            </div>
          ) : null}
        </div>
      )}
    </section>
  );
};

const UserDetailsLogs: React.VFC<UserDetailsLogsProps> = function UserDetailsLogs(
  props
) {
  const { userID } = props;
  const rawUserID = useMemo(() => extractRawID(userID), [userID]);

  return (
    <div className={styles.root}>
      <UserDetailsLogsActivitySection
        titleMessageID="UserDetails.logs.user-activities.title"
        viewAllMessageID="UserDetails.logs.user-activities.view-all"
        auditLogKind={AuditLogKind.User}
        userID={userID}
        rawUserID={rawUserID}
        activityTypes={USER_ACTIVITY_TYPES}
      />
      <UserDetailsLogsActivitySection
        titleMessageID="UserDetails.logs.admin-api-portal.title"
        viewAllMessageID="UserDetails.logs.admin-api-portal.view-all"
        auditLogKind={AuditLogKind.Admin}
        userID={userID}
        rawUserID={rawUserID}
        activityTypes={ADMIN_ACTIVITY_TYPES}
      />
    </div>
  );
};

export default UserDetailsLogs;
