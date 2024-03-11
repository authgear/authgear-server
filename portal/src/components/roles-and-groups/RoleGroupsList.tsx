import React, { useCallback, useContext, useMemo, useState } from "react";
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

import styles from "./RoleGroupsList.module.css";
import { Group, Role } from "../../graphql/adminapi/globalTypes.generated";
import Link from "../../Link";
import ActionButton from "../../ActionButton";
import { useSystemConfig } from "../../context/SystemConfigContext";
import useDelayedValue from "../../hook/useDelayedValue";
import DeleteRoleGroupDialog, {
  DeleteRoleGroupDialogData,
} from "./DeleteRoleGroupDialog";

export interface RoleGroupsListItem
  extends Pick<Group, "id" | "name" | "key"> {}

export interface RoleGroupsListRole extends Pick<Role, "id" | "name" | "key"> {}

export enum RoleGroupsListColumnKey {
  Name = "Name",
  Key = "Key",
  Action = "Action",
}

interface RoleGroupsListProps {
  role: RoleGroupsListRole;
  className?: string;
  isLoading?: boolean;
  groups: RoleGroupsListItem[];
}

export const RoleGroupsList: React.VFC<RoleGroupsListProps> =
  function RoleGroupsList({ role, groups, isLoading, className }) {
    const { themes } = useSystemConfig();
    const delayedLoading = useDelayedValue(isLoading, 500);
    const { appID } = useParams() as { appID: string };
    const { renderToString } = useContext(MessageContext);

    const [deleteDialogData, setDeleteDialogData] =
      useState<DeleteRoleGroupDialogData | null>(null);
    const onDismissDeleteDialog = useCallback(
      () => setDeleteDialogData(null),
      []
    );
    const onClickDeleteGroup = useCallback(
      (e: React.MouseEvent<unknown>, item: RoleGroupsListItem) => {
        e.preventDefault();
        e.stopPropagation();
        setDeleteDialogData({
          roleID: role.id,
          roleKey: role.key,
          roleName: role.name ?? null,
          groupID: item.id,
          groupKey: item.key,
          groupName: item.name ?? null,
        });
      },
      [role]
    );

    const onRenderTextActionButtonText = useCallback(() => {
      return (
        <Text className={styles.actionButtonText} theme={themes.destructive}>
          <FormattedMessage id="RoleGroupsList.actions.remove" />
        </Text>
      );
    }, [themes.destructive]);

    const columns: IColumn[] = useMemo((): IColumn[] => {
      return [
        {
          key: RoleGroupsListColumnKey.Name,
          fieldName: "name",
          name: renderToString("RoleGroupsList.column.name"),
          minWidth: 100,
          maxWidth: 300,
          targetWidthProportion: 1,
          isResizable: true,
          columnActionsMode: ColumnActionsMode.disabled,
        },
        {
          key: RoleGroupsListColumnKey.Key,
          fieldName: "key",
          name: renderToString("RoleGroupsList.column.key"),
          minWidth: 100,
          maxWidth: 9999,
          isResizable: true,
          columnActionsMode: ColumnActionsMode.disabled,
        },
        {
          key: RoleGroupsListColumnKey.Action,
          fieldName: "action",
          name: renderToString("RoleGroupsList.column.action"),
          minWidth: 67,
          maxWidth: 67,
          columnActionsMode: ColumnActionsMode.disabled,
        },
      ];
    }, [renderToString]);

    const onRenderRow = React.useCallback(
      (props?: IDetailsRowProps) => {
        if (props == null) {
          return null;
        }
        return (
          <Link
            className="contents"
            to={`/project/${appID}/user-management/groups/${
              (props.item as RoleGroupsListItem).id
            }/details`}
          >
            <DetailsRow {...props} />
          </Link>
        );
      },
      [appID]
    );

    const onRenderItemColumn = useCallback(
      (item: RoleGroupsListItem, _index?: number, column?: IColumn) => {
        switch (column?.key) {
          case RoleGroupsListColumnKey.Action: {
            return (
              <div className={styles.cell}>
                <ActionButton
                  onRenderText={onRenderTextActionButtonText}
                  className={styles.actionButton}
                  theme={themes.destructive}
                  onClick={(e) => {
                    onClickDeleteGroup(e, item);
                  }}
                />
              </div>
            );
          }
          default:
            return (
              <div className={styles.cell}>
                <div className={styles.cellText}>
                  {item[column?.fieldName as keyof RoleGroupsListItem] ?? ""}
                </div>
              </div>
            );
        }
      },
      [onClickDeleteGroup, onRenderTextActionButtonText, themes.destructive]
    );

    return (
      <>
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
        <DeleteRoleGroupDialog
          data={deleteDialogData}
          onDismiss={onDismissDeleteDialog}
        />
      </>
    );
  };
