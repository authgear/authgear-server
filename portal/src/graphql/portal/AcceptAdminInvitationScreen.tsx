import React, { useCallback, useContext, useMemo } from "react";
import authgear, { PromptOption } from "@authgear/web";
import { Text, DefaultEffects } from "@fluentui/react";
import { Context, FormattedMessage, FormattedMessageProps } from "../../intl";
import { useLocation, useNavigate } from "react-router-dom";

import { useAcceptCollaboratorInvitationMutation } from "./mutations/acceptCollaboratorInvitationMutation";
import ButtonWithLoading from "../../ButtonWithLoading";

import styles from "./AcceptAdminInvitationScreen.module.css";
import ScreenHeader from "../../ScreenHeader";
import {
  makeReasonErrorParseRule,
  parseAPIErrors,
  parseRawError,
} from "../../error/parse";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import { useAuthenticatedForInvitationQuery } from "./query/authenticatedForInvitationQuery";
import PrimaryButton from "../../PrimaryButton";
import DefaultButton from "../../DefaultButton";

function encodeOAuthState(state: Record<string, unknown>): string {
  return btoa(JSON.stringify(state));
}

interface AcceptAdminInvitationWidgetProps {
  title: FormattedMessageProps;
  descriptions: Array<FormattedMessageProps>;
  children?: React.ReactNode;
}

const AcceptAdminInvitationContent: React.VFC<AcceptAdminInvitationWidgetProps> =
  function AcceptAdminInvitationContent({ title, descriptions, children }) {
    return (
      <main className={styles.root}>
        <ScreenHeader showHamburger={false} />
        <div
          className={styles.widget}
          style={{
            boxShadow: DefaultEffects.elevation4,
          }}
        >
          <Text className={styles.title}>
            <FormattedMessage {...title} />
          </Text>
          <Text className={styles.description}>
            {descriptions.map((description, i) => (
              <FormattedMessage key={i} {...description} />
            ))}
          </Text>
          {children}
        </div>
        <div />
        <div />
      </main>
    );
  };

interface AcceptAdminInvitationIsInviteeProps {
  appID: string;
}

const AcceptAdminInvitationIsInvitee: React.VFC<AcceptAdminInvitationIsInviteeProps> =
  function AcceptAdminInvitationIsInvitee({ appID }) {
    const location = useLocation();
    const navigate = useNavigate();

    const invitationCode = useMemo(() => {
      return new URLSearchParams(location.search).get("code");
    }, [location]);

    const { acceptCollaboratorInvitation, loading, error } =
      useAcceptCollaboratorInvitationMutation();

    const errors = useMemo(() => {
      const apiErrors = parseRawError(error);
      const { topErrors } = parseAPIErrors(
        apiErrors,
        [],
        [
          makeReasonErrorParseRule(
            "CollaboratorInvitationInvalidCode",
            "AcceptAdminInvitationScreen.invalid-code-error"
          ),
          makeReasonErrorParseRule(
            "CollaboratorDuplicate",
            "AcceptAdminInvitationScreen.duplicated-collaborator-error"
          ),
          makeReasonErrorParseRule(
            "CollaboratorInvitationInvalidEmail",
            "AcceptAdminInvitationScreen.invalid-email-error"
          ),
        ]
      );
      return topErrors;
    }, [error]);

    const onAccept = useCallback(() => {
      acceptCollaboratorInvitation(invitationCode ?? "")
        .then((appID) => {
          if (appID !== null) {
            navigate(`/project/${appID}`);
          }
        })
        .catch(() => { });
    }, [acceptCollaboratorInvitation, invitationCode, navigate]);

    if (errors.length > 0) {
      return (
        <AcceptAdminInvitationContent
          title={{ id: "AcceptAdminInvitationScreen.accept-error.title" }}
          descriptions={errors
            .filter((err) => !!err.messageID)
            .map((err) => ({ id: err.messageID! }))}
        />
      );
    }

    return (
      <AcceptAdminInvitationContent
        title={{
          id: "AcceptAdminInvitationScreen.is-invitee.title",
          values: { appID },
        }}
        descriptions={[
          {
            id: "AcceptAdminInvitationScreen.is-invitee.description",
          },
        ]}
      >
        <ButtonWithLoading
          type="submit"
          loading={loading}
          labelId="AcceptAdminInvitationScreen.accept.label"
          loadingLabelId="loading"
          onClick={onAccept}
        />
      </AcceptAdminInvitationContent>
    );
  };

const AcceptAdminInvitationScreen: React.VFC =
  function AcceptAdminInvitationScreen() {
    const { renderToString } = useContext(Context);
    const navigate = useNavigate();
    const location = useLocation();
    const invitationCode = useMemo(() => {
      return new URLSearchParams(location.search).get("code") ?? "";
    }, [location]);

    const {
      loading,
      error,
      isCodeValid,
      isAuthenticated,
      isInvitee,
      appID,
      refetch,
    } = useAuthenticatedForInvitationQuery(invitationCode);

    const redirectURI = window.location.origin + "/oauth-redirect";
    const originalPath = `${window.location.pathname}${window.location.search}`;

    const goToAuth = useCallback(
      (page: "login" | "signup") => {
        // Normally we should call endAuthorization after being redirected back to here.
        // But we know that we are first party app and are using response_type=none so
        // we can skip that.
        authgear
          .startAuthentication({
            redirectURI,
            prompt: PromptOption.Login,
            state: encodeOAuthState({
              originalPath,
            }),
            page,
          })
          .catch((err) => {
            console.error(err);
          });
      },
      [redirectURI, originalPath]
    );

    const goToHome = useCallback(() => navigate("/"), [navigate]);

    if (loading) {
      return <ShowLoading />;
    }

    if (error != null) {
      return <ShowError error={error} onRetry={refetch} />;
    }

    if (!isCodeValid) {
      return (
        <AcceptAdminInvitationContent
          title={{ id: "AcceptAdminInvitationScreen.invalid-code.title" }}
          descriptions={[
            { id: "AcceptAdminInvitationScreen.invalid-code.description" },
          ]}
        >
          <PrimaryButton
            className={styles.loginButton}
            onClick={goToHome}
            text={renderToString(
              "AcceptAdminInvitationScreen.continue-to-authgear.label"
            )}
          />
        </AcceptAdminInvitationContent>
      );
    }

    if (!isAuthenticated) {
      return (
        <AcceptAdminInvitationContent
          title={{
            id: "AcceptAdminInvitationScreen.not-authenticaed.title",
            values: {
              appID: appID!,
              b: (chunks: React.ReactNode) => <b>{chunks}</b>,
            },
          }}
          descriptions={[
            { id: "AcceptAdminInvitationScreen.not-authenticaed.description" },
          ]}
        >
          <PrimaryButton
            className={styles.loginButton}
            onClick={() => goToAuth("login")}
            text={renderToString("AcceptAdminInvitationScreen.login.label")}
          />
          <DefaultButton
            className={styles.createAccountButton}
            onClick={() => goToAuth("signup")}
            text={renderToString(
              "AcceptAdminInvitationScreen.create-new-account.label"
            )}
          />
        </AcceptAdminInvitationContent>
      );
    }

    if (!isInvitee) {
      return (
        <AcceptAdminInvitationContent
          title={{ id: "AcceptAdminInvitationScreen.not-invitee.title" }}
          descriptions={[
            { id: "AcceptAdminInvitationScreen.not-invitee.description" },
          ]}
        >
          <PrimaryButton
            className={styles.loginButton}
            onClick={() => goToAuth("login")}
            text={renderToString(
              "AcceptAdminInvitationScreen.login-with-another-user.label"
            )}
          />
        </AcceptAdminInvitationContent>
      );
    }

    return <AcceptAdminInvitationIsInvitee appID={appID!} />;
  };

export default AcceptAdminInvitationScreen;
