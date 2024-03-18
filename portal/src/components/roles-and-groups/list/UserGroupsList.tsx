import React, { useCallback, useContext, useMemo } from "react";
import cn from "classnames";
import {
  ColumnActionsMode,
  DetailsRow,
  IColumn,
  IDetailsRowProps,
} from "@fluentui/react";
import { Context as MessageContext } from "@oursky/react-messageformat";
import { useParams } from "react-router-dom";

import styles from "./UserGroupsList.module.css";
import { Group, User } from "../../../graphql/adminapi/globalTypes.generated";
import Link from "../../../Link";
import ActionButtonCell from "./common/ActionButtonCell";
import TextCell from "./common/TextCell";
import RolesAndGroupsBaseList from "./common/RolesAndGroupsBaseList";

export interface UserGroupsListItem
  extends Pick<Group, "id" | "name" | "key"> {}

export interface UserGroupsListUser
  extends Pick<User, "id" | "formattedName"> {}

export enum UserGroupsListColumnKey {
  Name = "Name",
  Key = "Key",
  Action = "Action",
}

interface UserGroupsListProps {
  user: UserGroupsListUser;
  className?: string;
  groups: UserGroupsListItem[];
}

export const UserGroupsList: React.VFC<UserGroupsListProps> =
  function UserGroupsList({ groups, className }) {
    const { appID } = useParams() as { appID: string };
    const { renderToString } = useContext(MessageContext);

    const columns: IColumn[] = useMemo((): IColumn[] => {
      return [
        {
          key: UserGroupsListColumnKey.Name,
          fieldName: "name",
          name: renderToString("UserGroupsList.column.name"),
          minWidth: 100,
          maxWidth: 300,
          isResizable: true,
          columnActionsMode: ColumnActionsMode.disabled,
        },
        {
          key: UserGroupsListColumnKey.Key,
          fieldName: "key",
          name: renderToString("UserGroupsList.column.key"),
          minWidth: 100,
          maxWidth: 9999,
          isResizable: true,
          columnActionsMode: ColumnActionsMode.disabled,
        },
        {
          key: UserGroupsListColumnKey.Action,
          fieldName: "action",
          name: renderToString("UserGroupsList.column.action"),
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
              (props.item as UserGroupsListItem).id
            }/details`}
          >
            <DetailsRow {...props} />
          </Link>
        );
      },
      [appID]
    );

    const onRenderItemColumn = useCallback(
      (item: UserGroupsListItem, _index?: number, column?: IColumn) => {
        switch (column?.key) {
          case UserGroupsListColumnKey.Action: {
            return (
              <ActionButtonCell
                text={renderToString("UserGroupsList.actions.remove")}
              />
            );
          }
          default:
            return (
              <TextCell>
                {item[column?.fieldName as keyof UserGroupsListItem] ?? ""}
              </TextCell>
            );
        }
      },
      [renderToString]
    );

    const listEmptyText = renderToString("UserGroupsList.empty");

    return (
      <>
        <div className={cn(styles.root, className)}>
          <RolesAndGroupsBaseList
            emptyText={listEmptyText}
            onRenderRow={onRenderRow}
            onRenderItemColumn={onRenderItemColumn}
            items={groups}
            columns={columns}
          />
        </div>
      </>
    );
  };
