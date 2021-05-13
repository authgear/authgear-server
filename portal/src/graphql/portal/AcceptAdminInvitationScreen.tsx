import React, { useCallback, useContext, useMemo } from "react";
import { MessageBar, MessageBarType, Text } from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import { useLocation, useNavigate } from "react-router-dom";
import cn from "classnames";

import { useAcceptCollaboratorInvitationMutation } from "./mutations/acceptCollaboratorInvitationMutation";
import ButtonWithLoading from "../../ButtonWithLoading";

import styles from "./AcceptAdminInvitationScreen.module.scss";
import ScreenHeader from "../../ScreenHeader";
import { parseAPIErrors, parseRawError, renderErrors } from "../../error/parse";

const AcceptAdminInvitationScreen: React.FC = function AcceptAdminInvitationScreen() {
  const location = useLocation();
  const navigate = useNavigate();
  const { renderToString } = useContext(Context);

  const invitationCode = useMemo(() => {
    return new URLSearchParams(location.search).get("code");
  }, [location]);

  const {
    acceptCollaboratorInvitation,
    loading,
    error,
  } = useAcceptCollaboratorInvitationMutation();

  const errorMessage = useMemo(() => {
    const apiErrors = parseRawError(error);
    const { topErrors } = parseAPIErrors(
      apiErrors,
      [],
      [
        {
          reason: "CollaboratorInvitationInvalidCode",
          errorMessageID: "AcceptAdminInvitationScreen.invalid-code-error",
        },
        {
          reason: "CollaboratorDuplicate",
          errorMessageID:
            "AcceptAdminInvitationScreen.duplicated-collaborator-error",
        },
        {
          reason: "CollaboratorInvitationInvalidEmail",
          errorMessageID: "AcceptAdminInvitationScreen.invalid-email-error",
        },
      ]
    );
    return renderErrors(null, topErrors, renderToString);
  }, [error, renderToString]);

  const onAccept = useCallback(() => {
    acceptCollaboratorInvitation(invitationCode ?? "")
      .then((appID) => {
        if (appID !== null) {
          navigate(`/project/${appID}`);
        }
      })
      .catch(() => {});
  }, [acceptCollaboratorInvitation, invitationCode, navigate]);

  return (
    <main className={styles.root}>
      <ScreenHeader />
      <section className={cn(styles.body, { [styles.loading]: loading })}>
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
      </section>
    </main>
  );
};

export default AcceptAdminInvitationScreen;
