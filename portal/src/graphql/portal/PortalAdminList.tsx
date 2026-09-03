import React, { useContext, useMemo } from "react";
import cn from "classnames";
import { DotsVerticalIcon, TrashIcon } from "@radix-ui/react-icons";
import {
  DropdownMenu,
  IconButton as RadixIconButton,
  Text,
} from "@radix-ui/themes";
import { Context, FormattedMessage } from "../../intl";

import { Collaborator, CollaboratorInvitation } from "./globalTypes.generated";
import { Badge } from "../../components/v2/Badge/Badge";
import styles from "./PortalAdminList.module.css";

interface PortalAdminListProps {
  className?: string;
  collaborators: Collaborator[];
  collaboratorInvitations: CollaboratorInvitation[];
  onRemoveCollaboratorClicked: (id: string) => void;
  onRemoveCollaboratorInvitationClicked: (id: string) => void;
}

interface PortalAdminListCollaboratorItem {
  type: "collaborator";
  id: string;
  createdAt: Date;
  email: string;
  isOwner: boolean;
}

interface PortalAdminListCollaboratorInvitationItem {
  type: "collaboratorInvitation";
  id: string;
  createdAt: Date;
  email: string;
  isOwner: false;
}

type PortalAdminListItem =
  | PortalAdminListCollaboratorItem
  | PortalAdminListCollaboratorInvitationItem;

function isPortalAdminListCollaboratorItem(
  item: PortalAdminListItem
): item is PortalAdminListCollaboratorItem {
  return item.type === "collaborator";
}

const PortalAdminList: React.VFC<PortalAdminListProps> =
  function PortalAdminList(props) {
    const {
      className,
      collaborators,
      collaboratorInvitations,
      onRemoveCollaboratorClicked,
      onRemoveCollaboratorInvitationClicked,
    } = props;

    const { renderToString } = useContext(Context);

    const items: PortalAdminListItem[] = useMemo(() => {
      return [
        ...collaborators.map<PortalAdminListCollaboratorItem>(
          (collaborator) => ({
            type: "collaborator",
            id: collaborator.id,
            createdAt: new Date(collaborator.createdAt),
            email: collaborator.user.email ?? "",
            isOwner: collaborator.role === "OWNER",
          })
        ),
        ...collaboratorInvitations.map<PortalAdminListCollaboratorInvitationItem>(
          (collaboratorInvitation) => ({
            type: "collaboratorInvitation",
            id: collaboratorInvitation.id,
            createdAt: new Date(collaboratorInvitation.createdAt),
            email: collaboratorInvitation.inviteeEmail,
            isOwner: false,
          })
        ),
      ];
    }, [collaboratorInvitations, collaborators]);

    return (
      <div className={cn(styles.tableWrapper, className)}>
        <div className={styles.table}>
          <div className={styles.tableHeader}>
            <div className={styles.headerCellEmail}>
              <FormattedMessage id="PortalAdminList.column.email" />
            </div>
            <div className={styles.headerCellStatus}>
              <FormattedMessage id="PortalAdminList.column.status" />
            </div>
            <div className={styles.headerCellActions} aria-hidden={true} />
          </div>
          {items.map((item) => {
            const isCollaborator = isPortalAdminListCollaboratorItem(item);
            return (
              <div key={item.id} className={styles.tableRow}>
                <div className={styles.cellEmail}>
                  <Text size="2" className={styles.cellEmailText}>
                    {item.isOwner
                      ? `${item.email} (${renderToString(
                          "PortalAdminList.owner"
                        )})`
                      : item.email}
                  </Text>
                </div>
                <div className={styles.cellStatus}>
                  {isCollaborator ? (
                    <Badge
                      size="1"
                      variant="success"
                      text={renderToString("PortalAdminList.status.accepted")}
                    />
                  ) : (
                    <Badge
                      size="1"
                      variant="warning"
                      text={renderToString("PortalAdminList.status.pending")}
                    />
                  )}
                </div>
                <div className={styles.cellActions}>
                  {item.isOwner ? null : (
                    <DropdownMenu.Root>
                      <DropdownMenu.Trigger>
                        <RadixIconButton variant="ghost" color="gray" size="2">
                          <DotsVerticalIcon width="1rem" height="1rem" />
                        </RadixIconButton>
                      </DropdownMenu.Trigger>
                      <DropdownMenu.Content align="end">
                        <DropdownMenu.Item
                          color="red"
                          onSelect={() => {
                            if (isCollaborator) {
                              onRemoveCollaboratorClicked(item.id);
                            } else {
                              onRemoveCollaboratorInvitationClicked(item.id);
                            }
                          }}
                        >
                          <TrashIcon />
                          <FormattedMessage id="PortalAdminList.remove" />
                        </DropdownMenu.Item>
                      </DropdownMenu.Content>
                    </DropdownMenu.Root>
                  )}
                </div>
              </div>
            );
          })}
        </div>
      </div>
    );
  };

export default PortalAdminList;
