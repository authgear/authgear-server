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

import styles from "./UserRolesList.module.css";
import { Role, User } from "../../../graphql/adminapi/globalTypes.generated";
import Link from "../../../Link";
import ActionButtonCell from "./common/ActionButtonCell";
import TextCell from "./common/TextCell";
import RolesAndGroupsBaseList from "./common/RolesAndGroupsBaseList";

export interface UserRolesListItem extends Pick<Role, "id" | "name" | "key"> {}

export interface UserRolesListUser
  extends Pick<User, "id" | "formattedName" | "endUserAccountID"> {}

export enum UserRolesListColumnKey {
  Name = "Name",
  Key = "Key",
  Action = "Action",
}

interface UserRolesListProps {
  user: UserRolesListUser;
  className?: string;
  roles: UserRolesListItem[];
  isSearch: boolean;
  offset: number;
  pageSize: number;
  totalCount?: number;
  onChangeOffset?: (offset: number) => void;
}

export const UserRolesList: React.VFC<UserRolesListProps> =
  function UserRolesList({
    roles,
    className,
    isSearch,
    offset,
    pageSize,
    totalCount,
    onChangeOffset,
  }) {
    const { appID } = useParams() as { appID: string };
    const { renderToString } = useContext(MessageContext);

    const columns: IColumn[] = useMemo((): IColumn[] => {
      return [
        {
          key: UserRolesListColumnKey.Name,
          fieldName: "name",
          name: renderToString("UserRolesList.column.name"),
          minWidth: 100,
          maxWidth: 300,
          isResizable: true,
          columnActionsMode: ColumnActionsMode.disabled,
        },
        {
          key: UserRolesListColumnKey.Key,
          fieldName: "key",
          name: renderToString("UserRolesList.column.key"),
          minWidth: 100,
          maxWidth: 9999,
          isResizable: true,
          columnActionsMode: ColumnActionsMode.disabled,
        },
        {
          key: UserRolesListColumnKey.Action,
          fieldName: "action",
          name: renderToString("UserRolesList.column.action"),
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
              (props.item as UserRolesListItem).id
            }/details`}
          >
            <DetailsRow {...props} />
          </Link>
        );
      },
      [appID]
    );

    const onRenderItemColumn = useCallback(
      (item: UserRolesListItem, _index?: number, column?: IColumn) => {
        switch (column?.key) {
          case UserRolesListColumnKey.Action: {
            return (
              <ActionButtonCell
                text={renderToString("UserRolesList.actions.remove")}
              />
            );
          }
          default:
            return (
              <TextCell>
                {item[column?.fieldName as keyof UserRolesListItem] ?? ""}
              </TextCell>
            );
        }
      },
      [renderToString]
    );

    const paginationProps = useMemo(
      () => ({
        isSearch,
        offset,
        pageSize,
        totalCount,
        onChangeOffset,
      }),
      [isSearch, offset, pageSize, totalCount, onChangeOffset]
    );

    const listEmptyText = renderToString("UserRolesList.empty");

    return (
      <>
        <div className={cn(styles.root, className)}>
          <RolesAndGroupsBaseList
            emptyText={listEmptyText}
            onRenderRow={onRenderRow}
            onRenderItemColumn={onRenderItemColumn}
            items={roles}
            columns={columns}
            pagination={paginationProps}
          />
        </div>
      </>
    );
  };
