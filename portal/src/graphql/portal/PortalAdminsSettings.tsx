import React, { useMemo, useContext, useCallback, useState } from "react";
import { Context } from "@oursky/react-messageformat";
import { CommandBar, ICommandBarItemProps } from "@fluentui/react";
import { useNavigate, useParams } from "react-router-dom";
import cn from "classnames";

import { useCollaboratorsAndInvitationsQuery } from "./query/collaboratorsAndInvitationsQuery";
import PortalAdminList from "./PortalAdminList";
import RemovePortalAdminConfirmationDialog, {
  RemovePortalAdminConfirmationDialogData,
} from "./RemovePortalAdminConfirmationDialog";
import RemovePortalAdminInvitationConfirmationDialog, {
  RemovePortalAdminInvitationConfirmationDialogData,
} from "./RemovePortalAdminInvitationConfirmationDialog";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";

import styles from "./PortalAdminsSettings.module.scss";

interface PortalAdminsSettingsProps {
  className?: string;
}

const PortalAdminsSettings: React.FC<PortalAdminsSettingsProps> = function PortalAdminsSettings(
  props
) {
  const { className } = props;

  const { renderToString } = useContext(Context);
  const { appID } = useParams();
  const navigate = useNavigate();

  const {
    collaborators,
    collaboratorInvitations,
    loading: loadingCollaboratorsAndInvitations,
    error: collaboratorsAndInvitationsError,
    refetch: refetchCollaboratorsAndInvitations,
  } = useCollaboratorsAndInvitationsQuery(appID);

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
          navigate("./invite-admin");
        },
      },
    ];
  }, [navigate, renderToString]);

  const deletingCollaborator = false;
  const deletingCollaboratorInvitation = false;

  const onRemoveCollaboratorClicked = useCallback(
    (_event: React.MouseEvent<unknown>, id: string) => {
      if (!collaborators) {
        return;
      }
      const collaborator = collaborators.find(
        (collaborator) => collaborator.id === id
      );
      if (collaborator) {
        setRemovePortalAdminConfirmationDialogData({
          userID: id,
          // TODO: obtain admin user email
          email: "dummy@example.com",
        });
        setIsRemovePortalAdminConfirmationDialogVisible(true);
      }
    },
    [collaborators]
  );

  const onRemoveCollaboratorInvitationClicked = useCallback(
    (_event: React.MouseEvent<unknown>, id: string) => {
      if (!collaboratorInvitations) {
        return;
      }
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

  if (loadingCollaboratorsAndInvitations) {
    return <ShowLoading />;
  }

  if (collaboratorsAndInvitationsError != null) {
    return (
      <ShowError
        error={collaboratorsAndInvitationsError}
        onRetry={refetchCollaboratorsAndInvitations}
      />
    );
  }

  return (
    <div className={cn(styles.root, className)}>
      <CommandBar
        className={styles.commandBar}
        items={[]}
        farItems={commandBarItems}
      />
      <PortalAdminList
        loading={false}
        collaborators={collaborators ?? []}
        collaboratorInvitations={collaboratorInvitations ?? []}
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
