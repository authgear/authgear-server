import React, { useState, useMemo, useCallback, useContext } from "react";
import {
  ICommandBarItemProps,
  IDropdownOption,
  Dialog,
  DialogFooter,
  PrimaryButton,
  DefaultButton,
  DatePicker,
  TextField,
} from "@fluentui/react";
import { FormattedMessage, Context } from "@oursky/react-messageformat";
import { gql, useQuery } from "@apollo/client";
import NavBreadcrumb from "../../NavBreadcrumb";
import AuditLogList from "./AuditLogList";
import CommandBarDropdown, {
  CommandBarDropdownProps,
} from "../../CommandBarDropdown";
import CommandBarContainer from "../../CommandBarContainer";
import ShowError from "../../ShowError";
import { encodeOffsetToCursor } from "../../util/pagination";
import useTranactionalState from "../../hook/useTranactionalState";
import {
  AuditLogListQuery,
  AuditLogListQueryVariables,
} from "./__generated__/AuditLogListQuery";
import { AuditLogActivityType } from "./__generated__/globalTypes";

import styles from "./AuditLogScreen.module.scss";

const pageSize = 10;

const QUERY = gql`
  query AuditLogListQuery(
    $pageSize: Int!
    $cursor: String
    $activityTypes: [AuditLogActivityType!]
  ) {
    auditLogs(first: $pageSize, after: $cursor, activityTypes: $activityTypes) {
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
  } = useTranactionalState<Date | null>(null);

  const {
    committedValue: rangeTo,
    uncommittedValue: uncommittedRangeTo,
    setValue: setRangeTo,
    setCommittedValue: setRangeToImmediately,
    commit: commitRangeTo,
    rollback: rollbackRangeTo,
  } = useTranactionalState<Date | null>(null);

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
    },
    fetchPolicy: "network-only",
  });

  const messageBar = useMemo(() => {
    if (error != null) {
      return <ShowError error={error} onRetry={refetch} />;
    }
    return null;
  }, [error, refetch]);

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

  const dateRangeDialogContentProps = useMemo(() => {
    const title = renderToString("AuditLogScreen.date-range.custom");
    return {
      title,
    };
  }, [renderToString]);

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
        className={styles.root}
        messageBar={messageBar}
        farItems={commandBarFarItems}
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
      <Dialog
        hidden={dateRangeDialogHidden}
        onDismiss={onDismissDateRangeDialog}
        dialogContentProps={dateRangeDialogContentProps}
        /* https://developer.microsoft.com/en-us/fluentui#/controls/web/dialog
         * Best practice says the max width is 340 */
        minWidth={340}
      >
        {/* Dialog is based on Modal, which will focus the first child on open. *
        However, we do not want the date picker to be opened at the same time. *
        So we make the first focusable element a hidden TextField */}
        <TextField className={styles.hidden} />
        <DatePicker
          label={renderToString("AuditLogScreen.date-range.start-date")}
          value={uncommittedRangeFrom ?? undefined}
          onSelectDate={onSelectRangeFrom}
        />
        <DatePicker
          label={renderToString("AuditLogScreen.date-range.end-date")}
          value={uncommittedRangeTo ?? undefined}
          onSelectDate={onSelectRangeTo}
        />
        <DialogFooter>
          <PrimaryButton onClick={commitDateRange}>
            <FormattedMessage id="done" />
          </PrimaryButton>
          <DefaultButton onClick={onDismissDateRangeDialog}>
            <FormattedMessage id="cancel" />
          </DefaultButton>
        </DialogFooter>
      </Dialog>
    </>
  );
};

export default AuditLogScreen;
