import React, { useCallback, useContext, useMemo, useState } from "react";
import cn from "classnames";
import {
  Dialog,
  DialogFooter,
  IDialogContentProps,
  IconButton,
  Text,
} from "@fluentui/react";
import { useCopyFeedback } from "../../hook/useCopyFeedback";
import { produce } from "immer";
import { Context, FormattedMessage } from "../../intl";
import ExternalLink from "../../ExternalLink";
import PrimaryButton from "../../PrimaryButton";
import DefaultButton from "../../DefaultButton";
import { AppSecretConfigFormModel } from "../../hook/useAppSecretConfigForm";
import {
  findFramework,
  frameworksForType,
  getQuickStartGuide,
  type FrameworkEntry,
} from "./CreateOAuthClientScreen/frameworks";
import { FrameworkCard } from "./CreateOAuthClientScreen/FrameworkCard";
import type {
  ApplicationType,
  Framework,
  OAuthClientConfig,
  OAuthClientSecretKey,
} from "../../types";
import { useEndpoints } from "../../hook/useEndpoints";
import TextFieldWithCopyButton from "../../TextFieldWithCopyButton";
import { useStartReauthentication } from "../../graphql/portal/Authenticated";
import { useNavigate } from "react-router-dom";
import type { LocationState } from "./EditOAuthClientScreen";
import styles from "./EditOAuthClientFormFrameworkQuickStart.module.css";

const MASKED_SECRET = "***************";
const OIDC_RECOMMENDED_SCOPE =
  "openid offline_access https://authgear.com/scopes/full-userinfo";
const OIDC_DOCS_URL = "https://docs.authgear.com/get-started/oidc-provider";

const titleStyles = { root: { fontWeight: 600 as const } };
const tutorialDurationStyles = { root: { fontWeight: 600 as const } };

interface FormStateShape {
  clients: OAuthClientConfig[];
  editedClient: OAuthClientConfig | null;
  publicOrigin: string;
}

export interface EditOAuthClientFormFrameworkQuickStartProps<
  S extends FormStateShape
> {
  className?: string;
  client: OAuthClientConfig;
  applicationType: ApplicationType;
  form: AppSecretConfigFormModel<S>;
  clientSecrets?: OAuthClientSecretKey[];
}

export function EditOAuthClientFormFrameworkQuickStart<
  S extends FormStateShape
>({
  className,
  client,
  applicationType,
  form,
  clientSecrets,
}: EditOAuthClientFormFrameworkQuickStartProps<S>): React.ReactElement {
  const [dialogVisible, setDialogVisible] = useState(false);
  const [applying, setApplying] = useState(false);

  const framework = findFramework(client.x_framework);

  const openDialog = useCallback(() => setDialogVisible(true), []);
  const closeDialog = useCallback(() => {
    if (!applying) setDialogVisible(false);
  }, [applying]);

  const applyFramework = useCallback(
    async (newFrameworkId: Framework) => {
      setApplying(true);
      try {
        const newState = produce(form.state, (draft) => {
          draft.clients = draft.clients.map((c) =>
            c.client_id === client.client_id
              ? { ...c, x_framework: newFrameworkId }
              : c
          );
          if (draft.editedClient?.client_id === client.client_id) {
            draft.editedClient.x_framework = newFrameworkId;
          }
        });
        form.setState(() => newState);
        await form.saveWithState(newState);
        setDialogVisible(false);
      } finally {
        setApplying(false);
      }
    },
    [client.client_id, form]
  );

  if (framework == null) {
    return (
      <div className={cn(styles.root, className)}>
        <div className={styles.emptyState}>
          <div className={styles.emptyIconWrap}>
            <i
              className={cn("ti", "ti-app-window", styles.emptyIcon)}
              aria-hidden={true}
            />
          </div>
          <Text variant="xLarge" block={true} className={styles.emptyTitle}>
            <FormattedMessage id="EditOAuthClientFormFrameworkQuickStart.no-framework.title" />
          </Text>
          <Text block={true} className={styles.emptyBody}>
            <FormattedMessage id="EditOAuthClientFormFrameworkQuickStart.no-framework.body" />
          </Text>
          <PrimaryButton
            className={styles.emptyButton}
            onClick={openDialog}
            text={
              <FormattedMessage id="EditOAuthClientFormFrameworkQuickStart.choose-framework" />
            }
          />
        </div>
        <ChangeFrameworkDialog
          visible={dialogVisible}
          applicationType={applicationType}
          currentFrameworkId={null}
          applying={applying}
          onApply={applyFramework}
          onDismiss={closeDialog}
        />
      </div>
    );
  }

  return (
    <div className={cn(styles.root, className)}>
      <div className={styles.frameworkRow}>
        <div className={styles.iconWrap}>
          <i
            className={cn(
              "ti",
              `ti-${framework.iconName}`,
              styles.frameworkIcon
            )}
            aria-hidden={true}
          />
        </div>
        <div className={styles.frameworkText}>
          <Text variant="large" block={true} styles={titleStyles}>
            {framework.displayName}
          </Text>
          <Text block={true} className={styles.helperText}>
            {framework.helperText}
          </Text>
        </div>
        <DefaultButton
          className={styles.changeButtonInline}
          onClick={openDialog}
          text={
            <FormattedMessage id="EditOAuthClientFormFrameworkQuickStart.change-button" />
          }
        />
      </div>

      {framework.id === "other-oidc" ? (
        <OIDCProviderSection
          client={client}
          publicOrigin={form.state.publicOrigin}
          clientSecrets={clientSecrets}
        />
      ) : (
        <>
          <Text variant="xLarge" block={true} className={styles.sectionHeading}>
            <FormattedMessage id="EditOAuthClientFormFrameworkQuickStart.step-by-step.title" />
          </Text>
          <div className={styles.tutorialCard}>
            <div className={styles.tutorialHeader}>
              <i
                className={cn("ti", "ti-clock", styles.tutorialIcon)}
                aria-hidden={true}
              />
              <Text styles={tutorialDurationStyles}>
                <FormattedMessage id="EditOAuthClientFormFrameworkQuickStart.tutorial.duration" />
              </Text>
            </div>
            <Text block={true} className={styles.tutorialBody}>
              <FormattedMessage
                id={
                  getQuickStartGuide({
                    x_application_type: applicationType,
                    x_framework: framework.id,
                  }).bodyMessageId
                }
                values={{
                  displayName: framework.displayName,
                  // eslint-disable-next-line react/no-unstable-nested-components
                  docLink: (chunks: React.ReactNode) => (
                    <ExternalLink
                      href={
                        getQuickStartGuide({
                          x_application_type: applicationType,
                          x_framework: framework.id,
                        }).docLink
                      }
                    >
                      {chunks}
                    </ExternalLink>
                  ),
                }}
              />
            </Text>
          </div>
        </>
      )}

      {applicationType === "traditional_webapp" && framework.cookieSnippet ? (
        <CookieSnippetSection snippet={framework.cookieSnippet} />
      ) : null}

      <ChangeFrameworkDialog
        visible={dialogVisible}
        applicationType={applicationType}
        currentFrameworkId={framework.id}
        applying={applying}
        onApply={applyFramework}
        onDismiss={closeDialog}
      />
    </div>
  );
}

interface ChangeFrameworkDialogProps {
  visible: boolean;
  applicationType: ApplicationType;
  currentFrameworkId: Framework | null;
  applying: boolean;
  onApply: (frameworkId: Framework) => Promise<void>;
  onDismiss: () => void;
}

function ChangeFrameworkDialog(props: ChangeFrameworkDialogProps) {
  const {
    visible,
    applicationType,
    currentFrameworkId,
    applying,
    onApply,
    onDismiss,
  } = props;
  const { renderToString } = useContext(Context);
  const [selected, setSelected] = useState<Framework | null>(
    currentFrameworkId
  );

  // Reset selection when dialog opens with a different current framework.
  React.useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    if (visible) setSelected(currentFrameworkId);
  }, [visible, currentFrameworkId]);

  const options = useMemo<FrameworkEntry[]>(
    () => frameworksForType(applicationType),
    [applicationType]
  );

  const dialogContent: IDialogContentProps = useMemo(
    () => ({
      title: renderToString(
        "EditOAuthClientFormFrameworkQuickStart.change-dialog.title"
      ),
    }),
    [renderToString]
  );

  const onApplyClick = useCallback(() => {
    if (selected == null) return;
    void onApply(selected);
  }, [onApply, selected]);

  const canApply =
    selected != null && selected !== currentFrameworkId && !applying;

  return (
    <Dialog
      hidden={!visible}
      dialogContentProps={dialogContent}
      modalProps={{ isBlocking: applying }}
      onDismiss={onDismiss}
      maxWidth={620}
    >
      <div className={styles.dialogGrid}>
        {options.map((f) => (
          <FrameworkCard
            key={f.id}
            framework={f}
            selected={selected === f.id}
            onSelect={() => setSelected(f.id)}
          />
        ))}
      </div>
      <DialogFooter>
        <PrimaryButton
          onClick={onApplyClick}
          disabled={!canApply}
          text={
            <FormattedMessage id="EditOAuthClientFormFrameworkQuickStart.change-dialog.apply" />
          }
        />
        <DefaultButton
          onClick={onDismiss}
          disabled={applying}
          text={<FormattedMessage id="cancel" />}
        />
      </DialogFooter>
    </Dialog>
  );
}

interface CookieSnippetSectionProps {
  snippet: { language: string; code: string };
}

function CookieSnippetSection({ snippet }: CookieSnippetSectionProps) {
  const { copyButtonProps, Feedback } = useCopyFeedback({
    textToCopy: snippet.code,
  });
  return (
    <>
      <Text variant="xLarge" block={true} className={styles.sectionHeading}>
        <FormattedMessage id="EditOAuthClientFormFrameworkQuickStart.snippet.title" />
      </Text>
      <Text block={true} className={styles.snippetDescription}>
        <FormattedMessage id="EditOAuthClientFormFrameworkQuickStart.snippet.description" />
      </Text>
      <div className={styles.snippetCard}>
        <div className={styles.snippetHeader}>
          <span className={styles.snippetLanguage}>{snippet.language}</span>
          <div className={styles.snippetCopyWrap}>
            <IconButton {...copyButtonProps} />
            <Feedback />
          </div>
        </div>
        <pre className={styles.snippetCode}>
          <code>{snippet.code}</code>
        </pre>
      </div>
    </>
  );
}

interface OIDCProviderSectionProps {
  client: OAuthClientConfig;
  publicOrigin: string;
  clientSecrets?: OAuthClientSecretKey[];
}

function OIDCProviderSection({
  client,
  publicOrigin,
  clientSecrets,
}: OIDCProviderSectionProps) {
  const { renderToString } = useContext(Context);
  const navigate = useNavigate();
  const { startReauthentication, isRevealing } =
    useStartReauthentication<LocationState>();
  const endpoints = useEndpoints(publicOrigin, client.x_application_type);
  const firstSecret = clientSecrets?.[0];
  const showSecret = firstSecret != null;
  const isRevealed = !!firstSecret?.key;
  const secretValue = isRevealed ? firstSecret.key : MASKED_SECRET;

  const onRevealClick = useCallback(() => {
    startReauthentication(navigate, { isClientSecretRevealed: true });
  }, [startReauthentication, navigate]);

  const secretAdditionalButtons = useMemo(
    () =>
      isRevealed
        ? undefined
        : [
            {
              iconProps: { iconName: "RedEye" },
              title: renderToString("reveal"),
              ariaLabel: renderToString("reveal"),
              onClick: onRevealClick,
              disabled: isRevealing,
            },
          ],
    [isRevealed, onRevealClick, isRevealing, renderToString]
  );

  return (
    <div className={styles.oidcSection}>
      <Text variant="xLarge" block={true} className={styles.sectionHeading}>
        <FormattedMessage id="EditOAuthClientFormFrameworkQuickStart.oidc.title" />
      </Text>
      <Text block={true} className={styles.oidcDescription}>
        <FormattedMessage
          id="EditOAuthClientFormFrameworkQuickStart.oidc.description"
          values={{
            // eslint-disable-next-line react/no-unstable-nested-components
            docLink: (chunks: React.ReactNode) => (
              <ExternalLink href={OIDC_DOCS_URL}>{chunks}</ExternalLink>
            ),
          }}
        />
      </Text>
      <TextFieldWithCopyButton
        label={renderToString(
          "EditOAuthClientFormFrameworkQuickStart.oidc.client-id"
        )}
        value={client.client_id}
        readOnly={true}
      />
      {showSecret ? (
        <TextFieldWithCopyButton
          label={renderToString(
            "EditOAuthClientFormFrameworkQuickStart.oidc.client-secret"
          )}
          value={secretValue}
          readOnly={true}
          hideCopyButton={!isRevealed}
          additionalIconButtons={secretAdditionalButtons}
        />
      ) : null}
      <TextFieldWithCopyButton
        label={renderToString(
          "EditOAuthClientFormFrameworkQuickStart.oidc.scope"
        )}
        value={OIDC_RECOMMENDED_SCOPE}
        readOnly={true}
      />
      {endpoints.authorize != null ? (
        <TextFieldWithCopyButton
          label={renderToString(
            "EditOAuthClientFormFrameworkQuickStart.oidc.login-endpoint"
          )}
          value={endpoints.authorize}
          readOnly={true}
        />
      ) : null}
      {endpoints.userinfo != null ? (
        <TextFieldWithCopyButton
          label={renderToString(
            "EditOAuthClientFormFrameworkQuickStart.oidc.userinfo-endpoint"
          )}
          value={endpoints.userinfo}
          readOnly={true}
        />
      ) : null}
      <TextFieldWithCopyButton
        label={renderToString(
          "EditOAuthClientFormFrameworkQuickStart.oidc.token-endpoint"
        )}
        value={endpoints.token}
        readOnly={true}
      />
      {endpoints.endSession != null ? (
        <TextFieldWithCopyButton
          label={renderToString(
            "EditOAuthClientFormFrameworkQuickStart.oidc.end-session-endpoint"
          )}
          value={endpoints.endSession}
          readOnly={true}
        />
      ) : null}
      <TextFieldWithCopyButton
        label={renderToString(
          "EditOAuthClientFormFrameworkQuickStart.oidc.jwks-uri"
        )}
        value={endpoints.jwksUri}
        readOnly={true}
      />
    </div>
  );
}
