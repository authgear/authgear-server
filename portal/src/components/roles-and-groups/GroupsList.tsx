import React, { useCallback, useContext, useMemo } from "react";
import cn from "classnames";
import {
  ColumnActionsMode,
  DetailsListLayoutMode,
  DetailsRow,
  IColumn,
  IDetailsRowProps,
  SelectionMode,
  ShimmeredDetailsList,
  Text,
} from "@fluentui/react";
import {
  FormattedMessage,
  Context as MessageContext,
} from "@oursky/react-messageformat";
import { useParams } from "react-router-dom";

import styles from "./GroupsList.module.css";
import { Group } from "../../graphql/adminapi/globalTypes.generated";
import Link from "../../Link";
import ActionButton from "../../ActionButton";
import { useSystemConfig } from "../../context/SystemConfigContext";
import useDelayedValue from "../../hook/useDelayedValue";

export interface GroupsListItem
  extends Pick<Group, "id" | "name" | "key" | "description"> {}

export enum GroupsListColumnKey {
  Name = "Name",
  Key = "Key",
  Description = "description",
  Action = "Action",
}

interface GroupsListProps {
  className?: string;
  isLoading?: boolean;
  columns?: GroupsListColumnKey[];
  groups: GroupsListItem[];
}

const ALL_COLUMN_KEYS = [
  GroupsListColumnKey.Name,
  GroupsListColumnKey.Key,
  GroupsListColumnKey.Description,
  GroupsListColumnKey.Action,
];

export const GroupsList: React.VFC<GroupsListProps> = function GroupsList({
  groups,
  columns: columnKeys = ALL_COLUMN_KEYS,
  isLoading,
  className,
}) {
  const { themes } = useSystemConfig();
  const delayedLoading = useDelayedValue(isLoading, 500);
  const { appID } = useParams() as { appID: string };
  const { renderToString } = useContext(MessageContext);

  const onRenderTextActionButtonText = useCallback(() => {
    return (
      <Text className={styles.actionButtonText} theme={themes.destructive}>
        <FormattedMessage id="GroupsList.actions.remove" />
      </Text>
    );
  }, [themes.destructive]);

  const columns: IColumn[] = useMemo((): IColumn[] => {
    return [
      {
        key: GroupsListColumnKey.Name,
        fieldName: "name",
        name: renderToString("GroupsList.column.name"),
        minWidth: 100,
        maxWidth: 300,
        targetWidthProportion: 1,
        isResizable: true,
        columnActionsMode: ColumnActionsMode.disabled,
      },
      {
        key: GroupsListColumnKey.Key,
        fieldName: "key",
        name: renderToString("GroupsList.column.key"),
        minWidth: 100,
        maxWidth: 300,
        isResizable: true,
        columnActionsMode: ColumnActionsMode.disabled,
      },
      {
        key: GroupsListColumnKey.Description,
        fieldName: "description",
        name: renderToString("GroupsList.column.description"),
        minWidth: 200,
        maxWidth: 9999,
        isResizable: true,
        columnActionsMode: ColumnActionsMode.disabled,
      },
      {
        key: GroupsListColumnKey.Action,
        fieldName: "action",
        name: renderToString("GroupsList.column.action"),
        minWidth: 67,
        maxWidth: 67,
        columnActionsMode: ColumnActionsMode.disabled,
      },
    ].filter((col) => columnKeys.indexOf(col.key) !== -1);
  }, [columnKeys, renderToString]);

  const onRenderRow = React.useCallback(
    (props?: IDetailsRowProps) => {
      if (props == null) {
        return null;
      }
      return (
        <Link
          className="contents"
          to={`/project/${appID}/user-management/groups/${
            (props.item as GroupsListItem).id
          }/details`}
        >
          <DetailsRow {...props} />
        </Link>
      );
    },
    [appID]
  );

  const onRenderItemColumn = useCallback(
    (item: GroupsListItem, _index?: number, column?: IColumn) => {
      switch (column?.key) {
        case GroupsListColumnKey.Description:
          return (
            <div className={styles.cell}>
              <div className={styles.description}>
                {item[column.fieldName as keyof GroupsListItem] ?? ""}
              </div>
            </div>
          );
        case GroupsListColumnKey.Action: {
          return (
            <div className={styles.cell}>
              <ActionButton
                onRenderText={onRenderTextActionButtonText}
                className={styles.actionButton}
                theme={themes.destructive}
                onClick={(_e) => {
                  // onClickDeleteGroup(e, item);
                }}
              />
            </div>
          );
        }
        default:
          return (
            <div className={styles.cell}>
              <div className={styles.cellText}>
                {item[column?.fieldName as keyof GroupsListItem] ?? ""}
              </div>
            </div>
          );
      }
    },
    [onRenderTextActionButtonText, themes.destructive]
  );

  return (
    <div className={cn(styles.root, className)}>
      <div
        className={styles.listWrapper}
        // For DetailList to correctly know what to display
        // https://developer.microsoft.com/en-us/fluentui#/controls/web/detailslist
        data-is-scrollable="true"
      >
        <ShimmeredDetailsList
          enableShimmer={delayedLoading}
          enableUpdateAnimations={false}
          onRenderRow={onRenderRow}
          onRenderItemColumn={onRenderItemColumn}
          selectionMode={SelectionMode.none}
          layoutMode={DetailsListLayoutMode.justified}
          items={groups}
          columns={columns}
        />
      </div>
    </div>
  );
};
