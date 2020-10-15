import React, { useCallback, useContext, useMemo, useState } from "react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import {
  ActionButton,
  DetailsListLayoutMode,
  IColumn,
  SelectionMode,
  ShimmeredDetailsList,
} from "@fluentui/react";
import cn from "classnames";

import {
  Collaborator,
  CollaboratorInvitation,
} from "./query/collaboratorsAndInvitationsQuery";
import { destructiveTheme } from "../../theme";
import PaginationWidget from "../../PaginationWidget";

import styles from "./PortalAdminList.module.scss";

const pageSize = 10;

interface PortalAdminListProps {
  className?: string;
  loading: boolean;
  collaborators: Collaborator[];
  collaboratorInvitations: CollaboratorInvitation[];
  onRemoveCollaboratorClicked: (
    event: React.MouseEvent<unknown>,
    id: string
  ) => void;
  onRemoveCollaboratorInvitationClicked: (
    event: React.MouseEvent<unknown>,
    id: string
  ) => void;
}

interface PortalAdminListCollaboratorItem {
  type: "collaborator";
  id: string;
  createdAt: Date;
  email: string;
}

interface PortalAdminListCollaboratorInvitationItem {
  type: "collaboratorInvitation";
  id: string;
  createdAt: Date;
  email: string;
}

type PortalAdminListItem =
  | PortalAdminListCollaboratorItem
  | PortalAdminListCollaboratorInvitationItem;

function isPortalAdminListCollaboratorItem(
  item: PortalAdminListItem
): item is PortalAdminListCollaboratorItem {
  return item.type === "collaborator";
}

function isPortalAdminListCollaboratorInvitationItem(
  item: PortalAdminListItem
): item is PortalAdminListCollaboratorInvitationItem {
  return item.type === "collaboratorInvitation";
}

const PortalAdminList: React.FC<PortalAdminListProps> = function PortalAdminList(
  props
) {
  const {
    className,
    loading,
    collaborators,
    collaboratorInvitations,
    onRemoveCollaboratorClicked,
    onRemoveCollaboratorInvitationClicked,
  } = props;

  const { renderToString } = useContext(Context);

  const [offset, setOffset] = useState(0);

  const columns: IColumn[] = useMemo(() => {
    return [
      {
        key: "email",
        fieldName: "email",
        name: renderToString("PortalAdminList.column.email"),
        minWidth: 400,
      },
      {
        key: "status",
        fieldName: "status",
        name: renderToString("PortalAdminList.column.status"),
        minWidth: 150,
      },
      {
        key: "action",
        fieldName: "action",
        name: renderToString("PortalAdminList.column.action"),
        minWidth: 150,
      },
    ];
  }, [renderToString]);

  const items: PortalAdminListItem[] = useMemo(() => {
    return [
      ...collaborators.map<PortalAdminListCollaboratorItem>((collaborator) => ({
        type: "collaborator",
        id: collaborator.id,
        createdAt: new Date(collaborator.createdAt),
        // TODO: obtain admin user email
        email: "dummy@example.com",
      })),
      ...collaboratorInvitations.map<PortalAdminListCollaboratorInvitationItem>(
        (collaboratorInvitation) => ({
          type: "collaboratorInvitation",
          id: collaboratorInvitation.id,
          createdAt: new Date(collaboratorInvitation.createdAt),
          email: collaboratorInvitation.inviteeEmail,
        })
      ),
    ].sort((a, b) => b.createdAt.getTime() - a.createdAt.getTime());
  }, [collaboratorInvitations, collaborators]);

  const paginatedItems: PortalAdminListItem[] = useMemo(() => {
    return items.slice(offset, offset + pageSize);
  }, [items, offset]);

  const onRenderItemColumn = useCallback(
    (item: PortalAdminListItem, _index?: number, column?: IColumn) => {
      switch (column?.key) {
        case "email":
          return <span>{item.email}</span>;
        case "status":
          if (isPortalAdminListCollaboratorItem(item)) {
            return <span className={styles.acceptedStatus}>Accepted</span>;
          }
          return <span className={styles.pendingStatus}>Pending</span>;
        case "action":
          return (
            <ActionButton
              className={styles.actionButton}
              styles={{ flexContainer: { alignItems: "normal" } }}
              theme={destructiveTheme}
              onClick={(event) => {
                if (isPortalAdminListCollaboratorItem(item)) {
                  onRemoveCollaboratorClicked(event, item.id);
                } else if (isPortalAdminListCollaboratorInvitationItem(item)) {
                  onRemoveCollaboratorInvitationClicked(event, item.id);
                }
              }}
            >
              <FormattedMessage id="PortalAdminList.remove" />
            </ActionButton>
          );
        default:
          return null;
      }
    },
    [onRemoveCollaboratorClicked, onRemoveCollaboratorInvitationClicked]
  );

  const onChangeOffset = useCallback((offset: number) => {
    setOffset(offset);
  }, []);

  return (
    <div className={cn(styles.root, className)}>
      <ShimmeredDetailsList
        enableShimmer={loading}
        onRenderItemColumn={onRenderItemColumn}
        selectionMode={SelectionMode.none}
        layoutMode={DetailsListLayoutMode.justified}
        columns={columns}
        items={paginatedItems}
      />
      <PaginationWidget
        className={styles.pagination}
        offset={offset}
        pageSize={pageSize}
        totalCount={collaborators.length + collaboratorInvitations.length}
        onChangeOffset={onChangeOffset}
      />
    </div>
  );
};

export default PortalAdminList;
