import React, { useCallback, useMemo } from "react";
import { MessageBar, MessageBarType, Text } from "@fluentui/react";
import { FormattedMessage } from "@oursky/react-messageformat";
import { useLocation, useNavigate } from "react-router-dom";
import cn from "classnames";

import { useAcceptCollaboratorInvitationMutation } from "./mutations/acceptCollaboratorInvitationMutation";
import ButtonWithLoading from "../../ButtonWithLoading";
import { useGenericError } from "../../error/useGenericError";

import styles from "./AcceptAdminInvitationScreen.module.scss";

const AcceptAdminInvitationScreen: React.FC = function AcceptAdminInvitationScreen() {
  const location = useLocation();
  const navigate = useNavigate();

  const invitationCode = useMemo(() => {
    return new URLSearchParams(location.search).get("code");
  }, [location]);

  const {
    acceptCollaboratorInvitation,
    loading,
    error,
  } = useAcceptCollaboratorInvitationMutation();
  const errorMessage = useGenericError(error, [
    {
      reason: "CollaboratorInvitationInvalidCode",
      errorMessageID: "AcceptAdminInvitationScreen.invalid-code-error",
    },
    {
      reason: "CollaboratorDuplicate",
      errorMessageID:
        "AcceptAdminInvitationScreen.duplicated-collaborator-error",
    },
  ]);

  const onAccept = useCallback(() => {
    acceptCollaboratorInvitation(invitationCode ?? "")
      .then((appID) => {
        if (appID !== null) {
          navigate(`/app/${appID}`);
        }
      })
      .catch(() => {});
  }, [acceptCollaboratorInvitation, invitationCode, navigate]);

  return (
    <main className={cn(styles.root, { [styles.loading]: loading })}>
      {errorMessage && (
        <MessageBar messageBarType={MessageBarType.error}>
          <Text>{errorMessage}</Text>
        </MessageBar>
      )}
      <Text as="h1" className={styles.title}>
        <FormattedMessage id="AcceptAdminInvitationScreen.title" />
      </Text>
      <ButtonWithLoading
        type="submit"
        loading={loading}
        labelId="AcceptAdminInvitationScreen.accept.label"
        loadingLabelId="loading"
        onClick={onAccept}
      />
    </main>
  );
};

export default AcceptAdminInvitationScreen;
