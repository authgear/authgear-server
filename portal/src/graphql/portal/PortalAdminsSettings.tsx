import React, { useMemo, useContext, useCallback, useState } from "react";
import { Context } from "@oursky/react-messageformat";
import { CommandBar, ICommandBarItemProps } from "@fluentui/react";
import cn from "classnames";

import PortalAdminList from "./PortalAdminList";
import { Collaborator, CollaboratorInvitation } from "../../types";
import RemovePortalAdminConfirmationDialog, {
  RemovePortalAdminConfirmationDialogData,
} from "./RemovePortalAdminConfirmationDialog";
import RemovePortalAdminInvitationConfirmationDialog, {
  RemovePortalAdminInvitationConfirmationDialogData,
} from "./RemovePortalAdminInvitationConfirmationDialog";

import styles from "./PortalAdminsSettings.module.scss";

interface PortalAdminsSettingsProps {
  className?: string;
}

const PortalAdminsSettings: React.FC<PortalAdminsSettingsProps> = function PortalAdminsSettings(
  props
) {
  const { className } = props;

  const { renderToString } = useContext(Context);

  const [
    isRemovePortalAdminConfirmationDialogVisible,
    setIsRemovePortalAdminConfirmationDialogVisible,
  ] = useState(false);
  const [
    removePortalAdminConfirmationDialogData,
    setRemovePortalAdminConfirmationDialogData,
  ] = useState<RemovePortalAdminConfirmationDialogData | null>(null);

  const [
    isRemovePortalAdminInvitationConfirmationDialogVisible,
    setIsRemovePortalAdminInvitationConfirmationDialogVisible,
  ] = useState(false);
  const [
    removePortalAdminInvitationConfirmationDialogData,
    setRemovePortalAdminInvitationConfirmationDialogData,
  ] = useState<RemovePortalAdminInvitationConfirmationDialogData | null>(null);

  const commandBarItems: ICommandBarItemProps[] = useMemo(() => {
    return [
      {
        key: "invite",
        text: renderToString("PortalAdminsSettings.invite"),
        iconProps: { iconName: "CirclePlus" },
        onClick: () => {
          // TODO: handle invite admin action
        },
      },
    ];
  }, [renderToString]);

  // TODO: use real data
  const collaborators: Collaborator[] = useMemo(
    () => [
      {
        createdAt: new Date("2020/10/01"),
        id: "1",
        userID: "1",
        email: "user1@gmail.com",
      },
      {
        createdAt: new Date("2020/10/03"),
        id: "3",
        userID: "3",
        email: "user3@gmail.com",
      },
      {
        createdAt: new Date("2020/10/04"),
        id: "4",
        userID: "4",
        email: "user4@gmail.com",
      },
      {
        createdAt: new Date("2020/10/05"),
        id: "5",
        userID: "5",
        email: "user5@gmail.com",
      },
      {
        createdAt: new Date("2020/10/06"),
        id: "6",
        userID: "6",
        email: "user6@gmail.com",
      },
      {
        createdAt: new Date("2020/10/07"),
        id: "7",
        userID: "7",
        email: "user7@gmail.com",
      },
      {
        createdAt: new Date("2020/10/09"),
        id: "9",
        userID: "9",
        email: "user9@gmail.com",
      },
      {
        createdAt: new Date("2020/10/10"),
        id: "10",
        userID: "10",
        email: "user10@gmail.com",
      },
      {
        createdAt: new Date("2020/10/11"),
        id: "11",
        userID: "11",
        email: "user11@gmail.com",
      },
    ],
    []
  );

  // TODO: use real data
  const collaboratorInvitations: CollaboratorInvitation[] = useMemo(
    () => [
      {
        createdAt: new Date("2020/10/02"),
        expireAt: new Date("2021/10/02"),
        id: "2",
        invitedBy: "admin@gmail.com",
        inviteeEmail: "user2@gmail.com",
      },
      {
        createdAt: new Date("2020/10/08"),
        expireAt: new Date("2021/10/08"),
        id: "8",
        invitedBy: "admin@gmail.com",
        inviteeEmail: "user8@gmail.com",
      },
    ],
    []
  );

  const deletingCollaborator = false;
  const deletingCollaboratorInvitation = false;

  const onRemoveCollaboratorClicked = useCallback(
    (_event: React.MouseEvent<unknown>, id: string) => {
      const collaborator = collaborators.find(
        (collaborator) => collaborator.id === id
      );
      if (collaborator) {
        setRemovePortalAdminConfirmationDialogData({
          userID: id,
          email: collaborator.email,
        });
        setIsRemovePortalAdminConfirmationDialogVisible(true);
      }
    },
    [collaborators]
  );

  const onRemoveCollaboratorInvitationClicked = useCallback(
    (_event: React.MouseEvent<unknown>, id: string) => {
      const collaboratorInvitation = collaboratorInvitations.find(
        (collaboratorInvitation) => collaboratorInvitation.id === id
      );
      if (collaboratorInvitation) {
        setRemovePortalAdminInvitationConfirmationDialogData({
          invitationID: id,
          email: collaboratorInvitation.inviteeEmail,
        });
        setIsRemovePortalAdminInvitationConfirmationDialogVisible(true);
      }
    },
    [collaboratorInvitations]
  );

  const dismissRemovePortalAdminConfirmationDialog = useCallback(() => {
    setIsRemovePortalAdminConfirmationDialogVisible(false);
  }, []);

  const dismissRemovePortalAdminInvitationConfirmationDialog = useCallback(() => {
    setIsRemovePortalAdminInvitationConfirmationDialogVisible(false);
  }, []);

  const deleteCollaborator = useCallback((_userId: string) => {
    // TODO: handle delete collaborator mutation
    alert("Not yet implemented");
    setIsRemovePortalAdminConfirmationDialogVisible(false);
  }, []);

  const deleteCollaboratorInvitation = useCallback((_invitationID: string) => {
    // TODO: handle delete collaborator invitation mutation
    alert("Not yet implemented");
    setIsRemovePortalAdminInvitationConfirmationDialogVisible(false);
  }, []);

  return (
    <div className={cn(styles.root, className)}>
      <CommandBar
        className={styles.commandBar}
        items={[]}
        farItems={commandBarItems}
      />
      <PortalAdminList
        loading={false}
        collaborators={collaborators}
        collaboratorInvitations={collaboratorInvitations}
        onRemoveCollaboratorClicked={onRemoveCollaboratorClicked}
        onRemoveCollaboratorInvitationClicked={
          onRemoveCollaboratorInvitationClicked
        }
      />
      <RemovePortalAdminConfirmationDialog
        visible={isRemovePortalAdminConfirmationDialogVisible}
        data={removePortalAdminConfirmationDialogData ?? undefined}
        onDismiss={dismissRemovePortalAdminConfirmationDialog}
        deleteCollaborator={deleteCollaborator}
        deletingCollaborator={deletingCollaborator}
      />
      <RemovePortalAdminInvitationConfirmationDialog
        visible={isRemovePortalAdminInvitationConfirmationDialogVisible}
        data={removePortalAdminInvitationConfirmationDialogData ?? undefined}
        onDismiss={dismissRemovePortalAdminInvitationConfirmationDialog}
        deleteCollaboratorInvitation={deleteCollaboratorInvitation}
        deletingCollaboratorInvitation={deletingCollaboratorInvitation}
      />
    </div>
  );
};

export default PortalAdminsSettings;
