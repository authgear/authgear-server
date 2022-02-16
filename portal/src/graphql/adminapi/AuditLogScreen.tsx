import React, { useState, useMemo, useCallback, useContext } from "react";
import { useParams } from "react-router-dom";
import {
  ICommandBarItemProps,
  IDropdownOption,
  MessageBar,
  addDays,
} from "@fluentui/react";
import { useConst } from "@fluentui/react-hooks";
import { FormattedMessage, Context } from "@oursky/react-messageformat";
import { gql, useQuery } from "@apollo/client";
import { DateTime } from "luxon";
import NavBreadcrumb from "../../NavBreadcrumb";
import AuditLogList from "./AuditLogList";
import CommandBarDropdown, {
  CommandBarDropdownProps,
} from "../../CommandBarDropdown";
import CommandBarContainer from "../../CommandBarContainer";
import ScreenContent from "../../ScreenContent";
import ShowError from "../../ShowError";
import DateRangeDialog from "../portal/DateRangeDialog";
import { encodeOffsetToCursor } from "../../util/pagination";
import useTransactionalState from "../../hook/useTransactionalState";
import {
  AuditLogListQuery,
  AuditLogListQueryVariables,
} from "./__generated__/AuditLogListQuery";
import { AuditLogActivityType } from "./__generated__/globalTypes";

import styles from "./AuditLogScreen.module.scss";
import { useAppFeatureConfigQuery } from "../portal/query/appFeatureConfigQuery";

const pageSize = 10;

const QUERY = gql`
  query AuditLogListQuery(
    $pageSize: Int!
    $cursor: String
    $activityTypes: [AuditLogActivityType!]
    $rangeFrom: DateTime
    $rangeTo: DateTime
  ) {
    auditLogs(
      first: $pageSize
      after: $cursor
      activityTypes: $activityTypes
      rangeFrom: $rangeFrom
      rangeTo: $rangeTo
    ) {
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

function CommandBarDropdownWrapper(props: ICommandBarItemProps) {
  const { dropdownProps } = props;
  return <CommandBarDropdown {...dropdownProps} />;
}

const AuditLogScreen: React.FC = function AuditLogScreen() {
  const [offset, setOffset] = useState(0);
  const [selectedKey, setSelectedKey] = useState("ALL");
  const [dateRangeDialogHidden, setDateRangeDialogHidden] = useState(true);

  const {
    committedValue: rangeFrom,
    uncommittedValue: uncommittedRangeFrom,
    setValue: setRangeFrom,
    setCommittedValue: setRangeFromImmediately,
    commit: commitRangeFrom,
    rollback: rollbackRangeFrom,
  } = useTransactionalState<Date | null>(null);

  const {
    committedValue: rangeTo,
    uncommittedValue: uncommittedRangeTo,
    setValue: setRangeTo,
    setCommittedValue: setRangeToImmediately,
    commit: commitRangeTo,
    rollback: rollbackRangeTo,
  } = useTransactionalState<Date | null>(null);

  const { appID } = useParams();
  const featureConfig = useAppFeatureConfigQuery(appID);

  const logRetrievalDays = useMemo(() => {
    if (featureConfig.loading) {
      return -1;
    }
    return (
      featureConfig.effectiveFeatureConfig?.audit_log?.retrieval_days ?? -1
    );
  }, [
    featureConfig.loading,
    featureConfig.effectiveFeatureConfig?.audit_log?.retrieval_days,
  ]);

  const today = useConst(new Date(Date.now()));

  const datePickerMinDate = useMemo(() => {
    if (logRetrievalDays === -1) {
      return undefined;
    }
    const minDate = addDays(today, -logRetrievalDays + 1);
    minDate.setHours(0, 0, 0, 0);
    return minDate;
  }, [today, logRetrievalDays]);

  const queryRangeFrom = useMemo(() => {
    if (rangeFrom != null) {
      return rangeFrom.toISOString();
    }
    if (datePickerMinDate != null) {
      return datePickerMinDate.toISOString();
    }
    return null;
  }, [rangeFrom, datePickerMinDate]);

  const queryRangeTo = useMemo(() => {
    if (rangeTo != null) {
      return DateTime.fromJSDate(rangeTo)
        .plus({ days: 1 })
        .toJSDate()
        .toISOString();
    }
    return null;
  }, [rangeTo]);

  const isCustomDateRange = rangeFrom != null || rangeTo != null;

  const { renderToString } = useContext(Context);

  const activityTypeOptions = useMemo(() => {
    const options = [
      {
        key: "ALL",
        text: renderToString("AuditLogActivityType.ALL"),
      },
    ];
    for (const key of Object.keys(AuditLogActivityType)) {
      options.push({
        key: key,
        text: renderToString("AuditLogActivityType." + key),
      });
    }
    return options;
  }, [renderToString]);

  const activityTypes: AuditLogActivityType[] | null = useMemo(() => {
    if (selectedKey === "ALL") {
      return null;
    }
    return [selectedKey] as AuditLogActivityType[];
  }, [selectedKey]);

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
      activityTypes,
      rangeFrom: queryRangeFrom,
      rangeTo: queryRangeTo,
    },
    fetchPolicy: "network-only",
    skip: featureConfig.loading,
  });

  const messageBar = useMemo(() => {
    if (error != null) {
      return <ShowError error={error} onRetry={refetch} />;
    }
    if (featureConfig.error != null) {
      return (
        <ShowError
          error={featureConfig.error}
          onRetry={() => {
            featureConfig.refetch().finally(() => {});
          }}
        />
      );
    }
    return null;
  }, [error, refetch, featureConfig]);

  const onChangeSelectedKey = useCallback(
    (_e: React.FormEvent<HTMLDivElement>, item?: IDropdownOption) => {
      if (item != null && typeof item.key === "string") {
        setOffset(0);
        setSelectedKey(item.key);
      }
    },
    []
  );

  const dropdownProps: CommandBarDropdownProps = useMemo(() => {
    return {
      selectedKey,
      placeholder: "",
      label: "",
      options: activityTypeOptions,
      iconProps: {
        iconName: "PC1",
      },
      onChange: onChangeSelectedKey,
    };
  }, [selectedKey, onChangeSelectedKey, activityTypeOptions]);

  const onClickAllDateRange = useCallback(
    (e?: React.MouseEvent<unknown> | React.KeyboardEvent<unknown>) => {
      e?.stopPropagation();
      setRangeFromImmediately(null);
      setRangeToImmediately(null);
    },
    [setRangeFromImmediately, setRangeToImmediately]
  );

  const onClickCustomDateRange = useCallback(
    (e?: React.MouseEvent<unknown> | React.KeyboardEvent<unknown>) => {
      e?.stopPropagation();
      setDateRangeDialogHidden(false);
    },
    []
  );

  const commandBarFarItems: ICommandBarItemProps[] = useMemo(() => {
    const allDateRangeLabel = renderToString("AuditLogScreen.date-range.all");
    const customDateRangeLabel = renderToString(
      "AuditLogScreen.date-range.custom"
    );
    return [
      {
        key: "dateRange",
        text: isCustomDateRange ? customDateRangeLabel : allDateRangeLabel,
        iconProps: { iconName: "Calendar" },
        subMenuProps: {
          items: [
            {
              key: "allDateRange",
              text: allDateRangeLabel,
              onClick: onClickAllDateRange,
            },
            {
              key: "customDateRange",
              text: customDateRangeLabel,
              onClick: onClickCustomDateRange,
            },
          ],
        },
      },
      {
        key: "activityTypes",
        commandBarButtonAs: CommandBarDropdownWrapper,
        dropdownProps,
      },
    ];
  }, [
    dropdownProps,
    renderToString,
    isCustomDateRange,
    onClickAllDateRange,
    onClickCustomDateRange,
  ]);

  const onDismissDateRangeDialog = useCallback(
    (e?: React.MouseEvent<unknown>) => {
      e?.stopPropagation();
      setDateRangeDialogHidden(true);
      rollbackRangeFrom();
      rollbackRangeTo();
    },
    [rollbackRangeFrom, rollbackRangeTo]
  );

  const commitDateRange = useCallback(
    (e?: React.MouseEvent<unknown>) => {
      e?.preventDefault();
      e?.stopPropagation();
      setDateRangeDialogHidden(true);
      commitRangeFrom();
      commitRangeTo();
      setOffset(0);
    },
    [commitRangeFrom, commitRangeTo]
  );

  const onSelectRangeFrom = useCallback(
    (value: Date | null | undefined) => {
      if (value == null) {
        setRangeFrom(null);
      } else {
        if (uncommittedRangeTo != null && value > uncommittedRangeTo) {
          setRangeTo(value);
          setRangeFrom(uncommittedRangeTo);
        } else {
          setRangeFrom(value);
        }
      }
    },
    [setRangeFrom, setRangeTo, uncommittedRangeTo]
  );

  const onSelectRangeTo = useCallback(
    (value: Date | null | undefined) => {
      if (value == null) {
        setRangeTo(null);
      } else {
        if (uncommittedRangeFrom != null && value < uncommittedRangeFrom) {
          setRangeFrom(value);
          setRangeTo(uncommittedRangeFrom);
        } else {
          setRangeTo(value);
        }
      }
    },
    [setRangeTo, setRangeFrom, uncommittedRangeFrom]
  );

  return (
    <>
      <CommandBarContainer
        isLoading={loading}
        messageBar={messageBar}
        primaryItems={commandBarFarItems}
        className={styles.root}
      >
        <ScreenContent className={styles.content} layout="list">
          <div className={styles.widget}>
            <NavBreadcrumb items={items} />
            {logRetrievalDays !== -1 && (
              <MessageBar className={styles.messageBar}>
                <FormattedMessage
                  id="FeatureConfig.audit-log.retrieval-days"
                  values={{
                    planPagePath: "../billing",
                    logRetrievalDays: logRetrievalDays,
                  }}
                />
              </MessageBar>
            )}
          </div>
          <AuditLogList
            className={styles.widget}
            loading={loading}
            auditLogs={data?.auditLogs ?? null}
            offset={offset}
            pageSize={pageSize}
            totalCount={data?.auditLogs?.totalCount ?? undefined}
            onChangeOffset={onChangeOffset}
          />
        </ScreenContent>
      </CommandBarContainer>
      <DateRangeDialog
        hidden={dateRangeDialogHidden}
        title={renderToString("AuditLogScreen.date-range.custom")}
        fromDatePickerLabel={renderToString(
          "AuditLogScreen.date-range.start-date"
        )}
        toDatePickerLabel={renderToString("AuditLogScreen.date-range.end-date")}
        rangeFrom={uncommittedRangeFrom ?? undefined}
        rangeTo={uncommittedRangeTo ?? undefined}
        fromDatePickerMinDate={datePickerMinDate}
        fromDatePickerMaxDate={today}
        toDatePickerMinDate={datePickerMinDate}
        toDatePickerMaxDate={today}
        onSelectRangeFrom={onSelectRangeFrom}
        onSelectRangeTo={onSelectRangeTo}
        onCommitDateRange={commitDateRange}
        onDismiss={onDismissDateRangeDialog}
      />
    </>
  );
};

export default AuditLogScreen;
