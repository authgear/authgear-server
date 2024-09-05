import React, { useCallback, useContext, useMemo, useState } from "react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import { useNavigate, useParams } from "react-router-dom";

import { makeReasonErrorParseRule } from "../../error/parse";
import { useCollaboratorsAndInvitationsQuery } from "./query/collaboratorsAndInvitationsQuery";
import { useDeleteCollaboratorInvitationMutation } from "./mutations/deleteCollaboratorInvitationMutation";
import { useDeleteCollaboratorMutation } from "./mutations/deleteCollaboratorMutation";
import { useAppFeatureConfigQuery } from "./query/appFeatureConfigQuery";
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

import styles from "./PortalAdminsSettings.module.css";
import ScreenContent from "../../ScreenContent";
import NavBreadcrumb, { BreadcrumbItem } from "../../NavBreadcrumb";
import FeatureDisabledMessageBar from "./FeatureDisabledMessageBar";
import PrimaryButton from "../../PrimaryButton";

const PortalAdminsSettings: React.VFC = function PortalAdminsSettings() {
  const { renderToString } = useContext(Context);
  const { appID } = useParams() as { appID: string };
  const navigate = useNavigate();

  const {
    effectiveFeatureConfig,
    loading: featureConfigLoading,
    error: featureConfigError,
    refetch: featureConfigRefetch,
  } = useAppFeatureConfigQuery(appID);

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

  const retry = useCallback(() => {
    refetchCollaboratorsAndInvitations().finally(() => {});
    featureConfigRefetch().finally(() => {});
  }, [refetchCollaboratorsAndInvitations, featureConfigRefetch]);

  const inviteButtonProps = useMemo(() => {
    let disabled = false;
    if (effectiveFeatureConfig?.collaborator.maximum != null) {
      const maximum = effectiveFeatureConfig?.collaborator.maximum;
      const length1 = collaborators?.length ?? 0;
      const length2 = collaboratorInvitations?.length ?? 0;
      if (length1 + length2 >= maximum) {
        disabled = true;
      }
    }
    return {
      text: renderToString("PortalAdminsSettings.invite"),
      iconProps: { iconName: "Add" },
      disabled,
      onClick: () => {
        navigate("./invite");
      },
    };
  }, [
    navigate,
    renderToString,
    collaborators,
    collaboratorInvitations,
    effectiveFeatureConfig,
  ]);

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

  const dismissRemovePortalAdminInvitationConfirmationDialog =
    useCallback(() => {
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

  if (loadingCollaboratorsAndInvitations || featureConfigLoading) {
    return <ShowLoading />;
  }

  if (collaboratorsAndInvitationsError != null || featureConfigError != null) {
    return (
      <ShowError error={collaboratorsAndInvitationsError} onRetry={retry} />
    );
  }

  return (
    <>
      <ScreenContent layout="list">
        <div className={styles.widget}>
          <div className={styles.header}>
            <NavBreadcrumb
              className={styles.widget}
              items={navBreadcrumbItems}
            />
            <PrimaryButton {...inviteButtonProps} />
          </div>
          {effectiveFeatureConfig?.collaborator.maximum != null ? (
            <FeatureDisabledMessageBar
              className={styles.messageBar}
              messageID="FeatureConfig.collaborator"
              messageValues={{
                maximum: effectiveFeatureConfig?.collaborator.maximum,
              }}
            />
          ) : null}
        </div>
        <PortalAdminList
          className={styles.widget}
          loading={false}
          collaborators={collaborators ?? []}
          collaboratorInvitations={collaboratorInvitations ?? []}
          onRemoveCollaboratorClicked={onRemoveCollaboratorClicked}
          onRemoveCollaboratorInvitationClicked={
            onRemoveCollaboratorInvitationClicked
          }
        />
      </ScreenContent>
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
          makeReasonErrorParseRule(
            "CollaboratorSelfDeletion",
            "PortalAdminList.error.self-deletion"
          ),
        ]}
        fallbackErrorMessageID="PortalAdminsSettings.delete-collaborator-dialog.generic-error"
      />
      <ErrorDialog
        error={deleteCollaboratorInvitationError}
        rules={[]}
        fallbackErrorMessageID="PortalAdminsSettings.delete-collaborator-invitation-dialog.generic-error"
      />
    </>
  );
};

export default PortalAdminsSettings;
