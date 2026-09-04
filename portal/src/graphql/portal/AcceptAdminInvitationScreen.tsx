import React, { useCallback, useMemo } from "react";
import cn from "classnames";
import authgear, { PromptOption } from "@authgear/web";
import { Text } from "@radix-ui/themes";
import {
  EnvelopeClosedIcon,
  ExclamationTriangleIcon,
} from "@radix-ui/react-icons";
import { FormattedMessage, FormattedMessageProps } from "../../intl";
import { useLocation, useNavigate } from "react-router-dom";

import { useAcceptCollaboratorInvitationMutation } from "./mutations/acceptCollaboratorInvitationMutation";

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
import { PrimaryButton } from "../../components/v2/Button/PrimaryButton/PrimaryButton";
import { SecondaryButton } from "../../components/v2/Button/SecondaryButton/SecondaryButton";

function encodeOAuthState(state: Record<string, unknown>): string {
  return btoa(JSON.stringify(state));
}

type InvitationTone = "info" | "warning";

interface AcceptAdminInvitationWidgetProps {
  tone?: InvitationTone;
  icon?: React.ReactNode;
  title: FormattedMessageProps;
  descriptions: Array<FormattedMessageProps>;
  children?: React.ReactNode;
}

const AcceptAdminInvitationContent: React.VFC<AcceptAdminInvitationWidgetProps> =
  function AcceptAdminInvitationContent({
    tone = "info",
    icon,
    title,
    descriptions,
    children,
  }) {
    return (
      <main className={styles.root}>
        <ScreenHeader showHamburger={false} />
        <div className={styles.body}>
          <section className={styles.widget}>
            {icon != null ? (
              <div
                className={cn(
                  styles.icon,
                  tone === "warning" ? styles.iconWarning : styles.iconInfo
                )}
              >
                {icon}
              </div>
            ) : null}
            <div className={styles.header}>
              <Text as="p" size="5" weight="bold" className={styles.title}>
                <FormattedMessage {...title} />
              </Text>
              <Text as="p" size="2" color="gray" className={styles.description}>
                {descriptions.map((description, i) => (
                  <FormattedMessage key={i} {...description} />
                ))}
              </Text>
            </div>
            {children != null ? (
              <div className={styles.actions}>{children}</div>
            ) : null}
          </section>
        </div>
      </main>
    );
  };

const InfoIcon = <EnvelopeClosedIcon width={22} height={22} />;
const WarningIcon = <ExclamationTriangleIcon width={22} height={22} />;

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
        .catch(() => {});
    }, [acceptCollaboratorInvitation, invitationCode, navigate]);

    if (errors.length > 0) {
      return (
        <AcceptAdminInvitationContent
          tone="warning"
          icon={WarningIcon}
          title={{ id: "AcceptAdminInvitationScreen.accept-error.title" }}
          descriptions={errors
            .filter((err) => !!err.messageID)
            .map((err) => ({ id: err.messageID! }))}
        />
      );
    }

    return (
      <AcceptAdminInvitationContent
        icon={InfoIcon}
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
        <PrimaryButton
          size="3"
          type="submit"
          loading={loading}
          onClick={onAccept}
          text={
            <FormattedMessage id="AcceptAdminInvitationScreen.accept.label" />
          }
        />
      </AcceptAdminInvitationContent>
    );
  };

const AcceptAdminInvitationScreen: React.VFC =
  function AcceptAdminInvitationScreen() {
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

    const goToHome = useCallback(() => {
      navigate("/");
    }, [navigate]);

    if (loading) {
      return <ShowLoading />;
    }

    if (error != null) {
      // eslint-disable-next-line @typescript-eslint/strict-void-return
      return <ShowError error={error} onRetry={refetch} />;
    }

    if (!isCodeValid) {
      return (
        <AcceptAdminInvitationContent
          tone="warning"
          icon={WarningIcon}
          title={{ id: "AcceptAdminInvitationScreen.invalid-code.title" }}
          descriptions={[
            { id: "AcceptAdminInvitationScreen.invalid-code.description" },
          ]}
        >
          <PrimaryButton
            size="3"
            onClick={goToHome}
            text={
              <FormattedMessage id="AcceptAdminInvitationScreen.continue-to-authgear.label" />
            }
          />
        </AcceptAdminInvitationContent>
      );
    }

    if (!isAuthenticated) {
      return (
        <AcceptAdminInvitationContent
          icon={InfoIcon}
          title={{
            id: "AcceptAdminInvitationScreen.not-authenticated.title",
            values: {
              appID: appID!,
              // eslint-disable-next-line react/no-unstable-nested-components
              b: (chunks: React.ReactNode) => <b>{chunks}</b>,
            },
          }}
          descriptions={[
            { id: "AcceptAdminInvitationScreen.not-authenticated.description" },
          ]}
        >
          <PrimaryButton
            size="3"
            onClick={() => goToAuth("login")}
            text={
              <FormattedMessage id="AcceptAdminInvitationScreen.login.label" />
            }
          />
          <SecondaryButton
            size="3"
            onClick={() => goToAuth("signup")}
            text={
              <FormattedMessage id="AcceptAdminInvitationScreen.create-new-account.label" />
            }
          />
        </AcceptAdminInvitationContent>
      );
    }

    if (!isInvitee) {
      return (
        <AcceptAdminInvitationContent
          tone="warning"
          icon={WarningIcon}
          title={{ id: "AcceptAdminInvitationScreen.not-invitee.title" }}
          descriptions={[
            { id: "AcceptAdminInvitationScreen.not-invitee.description" },
          ]}
        >
          <PrimaryButton
            size="3"
            onClick={() => goToAuth("login")}
            text={
              <FormattedMessage id="AcceptAdminInvitationScreen.login-with-another-user.label" />
            }
          />
        </AcceptAdminInvitationContent>
      );
    }

    return <AcceptAdminInvitationIsInvitee appID={appID!} />;
  };

export default AcceptAdminInvitationScreen;
