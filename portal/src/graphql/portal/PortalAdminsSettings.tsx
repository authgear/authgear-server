import React, { useCallback, useContext, useMemo, useState } from "react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import { ICommandBarItemProps } from "@fluentui/react";
import { useNavigate, useParams } from "react-router-dom";

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
import CommandBarContainer from "../../CommandBarContainer";
import NavBreadcrumb, { BreadcrumbItem } from "../../NavBreadcrumb";

const PortalAdminsSettings: React.FC = function PortalAdminsSettings() {
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
          navigate("./invite");
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

  const navBreadcrumbItems: BreadcrumbItem[] = useMemo(() => {
    return [
      { to: ".", label: <FormattedMessage id="PortalAdminSettings.title" /> },
    ];
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
    <CommandBarContainer isLoading={false} farItems={commandBarItems}>
      <div className={styles.content}>
        <NavBreadcrumb
          className={styles.breadcrumb}
          items={navBreadcrumbItems}
        />
        <PortalAdminList
          className={styles.list}
          loading={false}
          collaborators={collaborators ?? []}
          collaboratorInvitations={collaboratorInvitations ?? []}
          onRemoveCollaboratorClicked={onRemoveCollaboratorClicked}
          onRemoveCollaboratorInvitationClicked={
            onRemoveCollaboratorInvitationClicked
          }
        />
      </div>
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
    </CommandBarContainer>
  );
};

export default PortalAdminsSettings;
