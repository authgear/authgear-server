import React, { useCallback, useContext, useMemo, useState } from "react";
import cn from "classnames";
import {
  ColumnActionsMode,
  DetailsListLayoutMode,
  DetailsRow,
  IColumn,
  IDetailsRowProps,
  MessageBar,
  SelectionMode,
  ShimmeredDetailsList,
  Text,
} from "@fluentui/react";
import {
  FormattedMessage,
  Context as MessageContext,
} from "@oursky/react-messageformat";
import { useParams } from "react-router-dom";

import styles from "./GroupRolesList.module.css";
import { Group, Role } from "../../graphql/adminapi/globalTypes.generated";
import Link from "../../Link";
import ActionButton from "../../ActionButton";
import { useSystemConfig } from "../../context/SystemConfigContext";

export interface GroupRolesListItem
  extends Pick<Group, "id" | "name" | "key"> {}

export interface GroupRolesListGroup
  extends Pick<Role, "id" | "name" | "key"> {}

export enum GroupRolesListColumnKey {
  Name = "Name",
  Key = "Key",
  Action = "Action",
}

interface GroupRolesListProps {
  group: GroupRolesListGroup;
  className?: string;
  roles: GroupRolesListItem[];
}

export const GroupRolesList: React.VFC<GroupRolesListProps> =
  function GroupRolesList({ group, roles, className }) {
    const { themes } = useSystemConfig();
    const { appID } = useParams() as { appID: string };
    const { renderToString } = useContext(MessageContext);

    const columns: IColumn[] = useMemo((): IColumn[] => {
      return [
        {
          key: GroupRolesListColumnKey.Name,
          fieldName: "name",
          name: renderToString("GroupRolesList.column.name"),
          minWidth: 100,
          maxWidth: 300,
          isResizable: true,
          columnActionsMode: ColumnActionsMode.disabled,
        },
        {
          key: GroupRolesListColumnKey.Key,
          fieldName: "key",
          name: renderToString("GroupRolesList.column.key"),
          minWidth: 100,
          maxWidth: 9999,
          isResizable: true,
          columnActionsMode: ColumnActionsMode.disabled,
        },
        {
          key: GroupRolesListColumnKey.Action,
          fieldName: "action",
          name: renderToString("GroupRolesList.column.action"),
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
            to={`/project/${appID}/user-management/roles/${
              (props.item as GroupRolesListItem).id
            }/details`}
          >
            <DetailsRow {...props} />
          </Link>
        );
      },
      [appID]
    );

    const onRenderItemColumn = useCallback(
      (item: GroupRolesListItem, _index?: number, column?: IColumn) => {
        switch (column?.key) {
          case GroupRolesListColumnKey.Action: {
            return (
              <div className={styles.cell}>
                <ActionButton
                  text={
                    <Text
                      className={styles.actionButtonText}
                      theme={themes.destructive}
                    >
                      <FormattedMessage id="GroupRolesList.actions.remove" />
                    </Text>
                  }
                  className={styles.actionButton}
                  theme={themes.destructive}
                />
              </div>
            );
          }
          default:
            return (
              <div className={styles.cell} key={item.key}>
                <div className={styles.cellText}>
                  {item[column?.fieldName as keyof GroupRolesListItem] ?? ""}
                </div>
              </div>
            );
        }
      },
      [themes.destructive]
    );

    const isEmpty = roles.length === 0;

    return (
      <>
        <div className={cn(styles.root, className)}>
          {isEmpty ? (
            <MessageBar className={styles.empty}>
              <FormattedMessage id="GroupRolesList.empty" />
            </MessageBar>
          ) : (
            <div
              className={styles.listWrapper}
              // For DetailList to correctly know what to display
              // https://developer.microsoft.com/en-us/fluentui#/controls/web/detailslist
              data-is-scrollable="true"
            >
              <ShimmeredDetailsList
                enableUpdateAnimations={false}
                onRenderRow={onRenderRow}
                onRenderItemColumn={onRenderItemColumn}
                selectionMode={SelectionMode.none}
                layoutMode={DetailsListLayoutMode.justified}
                items={roles}
                columns={columns}
              />
            </div>
          )}
        </div>
      </>
    );
  };
