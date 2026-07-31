import React, { useCallback, useMemo, useState } from "react";
import { FormattedMessage } from "../../intl";
import { useParams } from "react-router-dom";
import { PlusIcon } from "@radix-ui/react-icons";
import { Text } from "@radix-ui/themes";

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
import InviteAdminDialog from "./InviteAdminDialog";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import ErrorDialog from "../../error/ErrorDialog";

import styles from "./PortalAdminsSettings.module.css";
import ScreenContent from "../../ScreenContent";
import ScreenLayoutScrollView from "../../ScreenLayoutScrollView";
import { Callout } from "../../components/v2/Callout/Callout";
import { PrimaryButton } from "../../components/v2/Button/PrimaryButton/PrimaryButton";
import Link from "../../Link";
import ExternalLink from "../../ExternalLink";
import { getNextPlan } from "../../util/plan";

const PortalAdminsSettings: React.VFC = function PortalAdminsSettings() {
  const { appID } = useParams() as { appID: string };

  const {
    effectiveFeatureConfig,
    planName,
    isLoading: featureConfigLoading,
    loadError: featureConfigError,
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
  const [isInviteAdminDialogVisible, setIsInviteAdminDialogVisible] =
    useState(false);
  const [
    removePortalAdminInvitationConfirmationDialogData,
    setRemovePortalAdminInvitationConfirmationDialogData,
  ] = useState<RemovePortalAdminInvitationConfirmationDialogData | null>(null);

  const retry = useCallback(() => {
    refetchCollaboratorsAndInvitations().finally(() => {});
    featureConfigRefetch().finally(() => {});
  }, [refetchCollaboratorsAndInvitations, featureConfigRefetch]);

  const inviteDisabled = useMemo(() => {
    if (effectiveFeatureConfig?.collaborator?.maximum != null) {
      const maximum = effectiveFeatureConfig.collaborator.maximum;
      const length1 = collaborators?.length ?? 0;
      const length2 = collaboratorInvitations?.length ?? 0;
      if (length1 + length2 >= maximum) {
        return true;
      }
    }
    return false;
  }, [collaborators, collaboratorInvitations, effectiveFeatureConfig]);

  const onInviteClicked = useCallback(() => {
    setIsInviteAdminDialogVisible(true);
  }, []);

  const dismissInviteAdminDialog = useCallback(() => {
    setIsInviteAdminDialogVisible(false);
  }, []);

  const onRemoveCollaboratorClicked = useCallback(
    (id: string) => {
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
    (id: string) => {
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

  const displayedCollaboratorMaximum = useMemo<number | undefined>(() => {
    return (
      effectiveFeatureConfig?.collaborator?.soft_maximum ??
      effectiveFeatureConfig?.collaborator?.maximum
    );
  }, [effectiveFeatureConfig]);

  const canUpgradePlan = useMemo(() => {
    return getNextPlan(planName ?? "") != null;
  }, [planName]);

  const displayMaximumWarning = useMemo(() => {
    if (
      collaborators == null ||
      collaboratorInvitations == null ||
      displayedCollaboratorMaximum == null
    ) {
      return false;
    }
    return (
      collaborators.length + collaboratorInvitations.length >=
      displayedCollaboratorMaximum
    );
  }, [collaborators, collaboratorInvitations, displayedCollaboratorMaximum]);

  const maximumWarningMessageValues = useMemo(() => {
    const planPagePath = `/project/${appID}/billing`;
    const contactUsHref =
      "https://www.authgear.com/schedule-demo?utm_source=portal&utm_medium=link&utm_campaign=additional_order";
    return {
      maximum: displayedCollaboratorMaximum!,
      // eslint-disable-next-line react/no-unstable-nested-components
      ReactRouterLink: (chunks: React.ReactNode) => (
        <Link to={planPagePath}>{chunks}</Link>
      ),
      // eslint-disable-next-line react/no-unstable-nested-components
      ExternalLink: (chunks: React.ReactNode) => (
        <ExternalLink href={contactUsHref}>{chunks}</ExternalLink>
      ),
    };
  }, [appID, displayedCollaboratorMaximum]);

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
      <ScreenLayoutScrollView>
        <ScreenContent layout="list">
          <div className={styles.widget}>
            <div className={styles.header}>
              <Text as="p" size="5" weight="bold" className={styles.pageTitle}>
                <FormattedMessage id="PortalAdminSettings.title" />
              </Text>
              <PrimaryButton
                size="2"
                disabled={inviteDisabled}
                onClick={onInviteClicked}
                text={
                  <span className={styles.inviteButtonContent}>
                    <PlusIcon width="1rem" height="1rem" />
                    <FormattedMessage id="PortalAdminsSettings.invite" />
                  </span>
                }
              />
            </div>
          </div>
          {displayMaximumWarning ? (
            <div className={styles.widget}>
              <Callout
                type="info"
                showCloseButton={false}
                text={
                  <FormattedMessage
                    id={
                      canUpgradePlan
                        ? "FeatureConfig.collaborator.upgrade"
                        : "FeatureConfig.collaborator.contact-us"
                    }
                    values={maximumWarningMessageValues}
                  />
                }
              />
            </div>
          ) : null}
          <PortalAdminList
            className={styles.widget}
            collaborators={collaborators ?? []}
            collaboratorInvitations={collaboratorInvitations ?? []}
            onRemoveCollaboratorClicked={onRemoveCollaboratorClicked}
            onRemoveCollaboratorInvitationClicked={
              onRemoveCollaboratorInvitationClicked
            }
          />
        </ScreenContent>
      </ScreenLayoutScrollView>
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
      <InviteAdminDialog
        open={isInviteAdminDialogVisible}
        onDismiss={dismissInviteAdminDialog}
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
