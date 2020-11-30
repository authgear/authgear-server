import React, { useMemo, useContext, useCallback, useState } from "react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import { CommandBar, ICommandBarItemProps, Text } from "@fluentui/react";
import { useNavigate, useParams } from "react-router-dom";
import cn from "classnames";

import { useCollaboratorsAndInvitationsQuery } from "./query/collaboratorsAndInvitationsQuery";
import { useDeleteCollaboratorInvitationMutation } from "./mutations/deleteCollaboratorInvitationMutation";
import { useDeleteCollaboratorMutation } from "./mutations/deleteCollaboratorMutation";
import PortalAdminList from "./PortalAdminList";
import RemovePortalAdminConfirmationDialog, {
  RemovePortalAdminConfirmationDialogData,
} from "./RemovePortalAdminConfirmationDialog";
import RemovePortalAdminInvitationConfirmationDialog, {
  RemovePortalAdminInvitationConfirmationDialogData,
} from "./RemovePortalAdminInvitationConfirmationDialog";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import ErrorDialog from "../../error/ErrorDialog";

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
  const {
    deleteCollaborator,
    loading: deletingCollaborator,
    error: deleteCollaboratorError,
  } = useDeleteCollaboratorMutation();
  const {
    deleteCollaboratorInvitation,
    loading: deletingCollaboratorInvitation,
    error: deleteCollaboratorInvitationError,
  } = useDeleteCollaboratorInvitationMutation();

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
          email: collaborator.user.email ?? "",
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

  const onDeleteCollaborator = useCallback(
    (userID: string) => {
      deleteCollaborator(userID)
        .catch(() => {})
        .finally(() => {
          setIsRemovePortalAdminConfirmationDialogVisible(false);
        });
    },
    [deleteCollaborator]
  );

  const OnDeleteCollaboratorInvitation = useCallback(
    (invitationID: string) => {
      deleteCollaboratorInvitation(invitationID)
        .catch(() => {})
        .finally(() => {
          setIsRemovePortalAdminInvitationConfirmationDialogVisible(false);
        });
    },
    [deleteCollaboratorInvitation]
  );

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
      <Text as="h1" variant="xLarge" block={true}>
        <FormattedMessage id="PortalAdminSettings.title" />
      </Text>
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
        deleteCollaborator={onDeleteCollaborator}
        deletingCollaborator={deletingCollaborator}
      />
      <RemovePortalAdminInvitationConfirmationDialog
        visible={isRemovePortalAdminInvitationConfirmationDialogVisible}
        data={removePortalAdminInvitationConfirmationDialogData ?? undefined}
        onDismiss={dismissRemovePortalAdminInvitationConfirmationDialog}
        deleteCollaboratorInvitation={OnDeleteCollaboratorInvitation}
        deletingCollaboratorInvitation={deletingCollaboratorInvitation}
      />
      <ErrorDialog
        error={deleteCollaboratorError}
        rules={[
          {
            reason: "CollaboratorSelfDeletion",
            errorMessageID: "PortalAdminList.error.self-deletion",
          },
        ]}
        fallbackErrorMessageID="PortalAdminsSettings.delete-collaborator-dialog.generic-error"
      />
      <ErrorDialog
        error={deleteCollaboratorInvitationError}
        rules={[]}
        fallbackErrorMessageID="PortalAdminsSettings.delete-collaborator-invitation-dialog.generic-error"
      />
    </div>
  );
};

export default PortalAdminsSettings;
