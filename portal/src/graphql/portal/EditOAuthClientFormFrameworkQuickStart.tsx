import React, {
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import cn from "classnames";
import { Dialog, IconButton as RadixIconButton, Text } from "@radix-ui/themes";
import { EyeOpenIcon } from "@radix-ui/react-icons";
import { produce } from "immer";
import { Context, FormattedMessage } from "../../intl";
import ExternalLink from "../../ExternalLink";
import { PrimaryButton } from "../../components/v2/Button/PrimaryButton/PrimaryButton";
import { SecondaryButton } from "../../components/v2/Button/SecondaryButton/SecondaryButton";
import { CopyIconButton } from "../../components/v2/CopyIconButton/CopyIconButton";
import { TextField } from "../../components/v2/TextField/TextField";
import {
  IconRadioCards,
  type IconRadioCardOption,
} from "../../components/v2/IconRadioCards/IconRadioCards";
import { AppSecretConfigFormModel } from "../../hook/useAppSecretConfigForm";
import {
  findFramework,
  frameworksForType,
  getQuickStartGuide,
  type FrameworkEntry,
} from "./CreateOAuthClientScreen/frameworks";
import { StarterKitSection } from "./StarterKitSection";
import { appendRedirectURI } from "./CreateOAuthClientScreen/starterKit";
import type {
  ApplicationType,
  Framework,
  OAuthClientConfig,
  OAuthClientSecretKey,
} from "../../types";
import { useCapture } from "../../gtm_v2";
import { useEndpoints } from "../../hook/useEndpoints";
import { useStartReauthentication } from "../../graphql/portal/Authenticated";
import { useNavigate, useParams } from "react-router-dom";
import type { LocationState } from "./EditOAuthClientScreen";
import PortalLink from "../../Link";
import styles from "./EditOAuthClientFormFrameworkQuickStart.module.css";

const MASKED_SECRET = "***************";
const OIDC_RECOMMENDED_SCOPE =
  "openid offline_access https://authgear.com/scopes/full-userinfo";
const OIDC_DOCS_URL = "https://docs.authgear.com/get-started/oidc-provider";

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
  onGoToSettings?: () => void;
}

export function EditOAuthClientFormFrameworkQuickStart<
  S extends FormStateShape
>({
  className,
  client,
  applicationType,
  form,
  clientSecrets,
  onGoToSettings,
}: EditOAuthClientFormFrameworkQuickStartProps<S>): React.ReactElement {
  const { renderToString } = useContext(Context);
  const [dialogVisible, setDialogVisible] = useState(false);
  const [applying, setApplying] = useState(false);

  const framework = findFramework(client.x_framework);

  const capture = useCapture();
  const viewedRef = useRef(false);
  useEffect(() => {
    if (viewedRef.current) {
      return;
    }
    viewedRef.current = true;
    capture("quickstart.viewed", {
      client_id: client.client_id,
      framework_id: client.x_framework ?? "",
      // This tab also renders for clients created by the legacy flow
      // (no x_framework); keep the cohort split accurate.
      wizard_version: client.x_framework != null ? "framework_first" : "legacy",
    });
  }, [capture, client.client_id, client.x_framework]);

  const openDialog = useCallback(() => setDialogVisible(true), []);
  const closeDialog = useCallback(() => {
    if (!applying) setDialogVisible(false);
  }, [applying]);

  const applyFramework = useCallback(
    async (newFrameworkId: Framework) => {
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
      setApplying(true);
      try {
        // Do not touch the form state before the save: saveWithState reloads
        // the config on success, so the content behind the dialog updates at
        // the same moment the dialog dismisses. Updating the state upfront
        // made the dialog look hung while the request was still in flight.
        await form.saveWithState(newState);
        setDialogVisible(false);
      } finally {
        setApplying(false);
      }
    },
    [client.client_id, form]
  );

  const { appID } = useParams() as { appID: string };
  const [settingRedirect, setSettingRedirect] = useState(false);

  const onAuthorizeRedirectURIs = useCallback(() => {
    const uris = findFramework(client.x_framework)?.starterKit?.redirectURIs;
    if (uris == null || uris.length === 0) {
      return;
    }
    setSettingRedirect(true);
    const appendAll = (existing: string[] | undefined): string[] => {
      let next = existing ?? [];
      for (const uri of uris) {
        next = appendRedirectURI(next, uri);
      }
      return next;
    };
    const newState = produce(form.state, (draft) => {
      const target = draft.clients.find(
        (c) => c.client_id === client.client_id
      );
      if (target != null) {
        target.redirect_uris = appendAll(target.redirect_uris);
      }
      if (draft.editedClient?.client_id === client.client_id) {
        draft.editedClient.redirect_uris = appendAll(
          draft.editedClient.redirect_uris
        );
      }
    });
    form
      .saveWithState(newState)
      .catch(() => {})
      .finally(() => setSettingRedirect(false));
  }, [client.x_framework, client.client_id, form]);

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
          <Text as="p" size="4" weight="bold" className={styles.emptyTitle}>
            <FormattedMessage id="EditOAuthClientFormFrameworkQuickStart.no-framework.title" />
          </Text>
          <Text as="p" size="2" className={styles.emptyBody}>
            <FormattedMessage id="EditOAuthClientFormFrameworkQuickStart.no-framework.body" />
          </Text>
          <span className={styles.emptyButton}>
            <PrimaryButton
              size="2"
              onClick={openDialog}
              text={
                <FormattedMessage id="EditOAuthClientFormFrameworkQuickStart.choose-framework" />
              }
            />
          </span>
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
          <Text as="p" size="3" weight="bold" className={styles.frameworkName}>
            <FormattedMessage id={framework.displayNameMessageId} />
          </Text>
          <Text as="p" size="2" className={styles.helperText}>
            <FormattedMessage id={framework.helperTextMessageId} />
          </Text>
        </div>
        <span className={styles.changeButtonInline}>
          <SecondaryButton
            size="2"
            onClick={openDialog}
            text={
              <FormattedMessage id="EditOAuthClientFormFrameworkQuickStart.change-button" />
            }
          />
        </span>
      </div>

      {framework.id === "other-oidc" ? (
        <OIDCProviderSection
          client={client}
          publicOrigin={form.state.publicOrigin}
          clientSecrets={clientSecrets}
        />
      ) : (
        <>
          <Text as="p" size="4" weight="bold" className={styles.sectionHeading}>
            <FormattedMessage id="EditOAuthClientFormFrameworkQuickStart.step-by-step.title" />
          </Text>
          <div className={styles.tutorialCard}>
            <div className={styles.tutorialHeader}>
              <i
                className={cn("ti", "ti-clock", styles.tutorialIcon)}
                aria-hidden={true}
              />
              <Text size="2" weight="bold">
                <FormattedMessage id="EditOAuthClientFormFrameworkQuickStart.tutorial.duration" />
              </Text>
            </div>
            <Text as="p" size="2" className={styles.tutorialBody}>
              <FormattedMessage
                id={
                  getQuickStartGuide({
                    x_application_type: applicationType,
                    x_framework: framework.id,
                  }).bodyMessageId
                }
                values={{
                  displayName: renderToString(framework.displayNameMessageId),
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

      {framework.starterKit != null ? (
        <StarterKitSection
          starterKit={framework.starterKit}
          frameworkDisplayName={renderToString(framework.displayNameMessageId)}
          clientID={client.client_id}
          publicOrigin={form.state.publicOrigin}
          usersPath={`/project/${appID}/users`}
          redirectURIIsSet={framework.starterKit.redirectURIs.every((uri) =>
            (client.redirect_uris ?? []).includes(uri)
          )}
          saving={settingRedirect}
          onAuthorize={onAuthorizeRedirectURIs}
          onGoToSettings={onGoToSettings ?? (() => {})}
        />
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
  const { appID } = useParams() as { appID: string };
  const [selected, setSelected] = useState<Framework | null>(
    currentFrameworkId
  );

  // Reset selection when dialog opens with a different current framework.
  React.useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    if (visible) setSelected(currentFrameworkId);
  }, [visible, currentFrameworkId]);

  // Only frameworks compatible with this client's application type can be
  // switched to; the application type is fixed at creation. Other platforms
  // require a new application (see the note below the grid).
  const options = useMemo<IconRadioCardOption<Framework>[]>(
    () =>
      frameworksForType(applicationType).map((f: FrameworkEntry) => ({
        value: f.id,
        icon: (
          <i
            className={cn("ti", `ti-${f.iconName}`, styles.dialogCardIcon)}
            aria-hidden={true}
          />
        ),
        title: <FormattedMessage id={f.displayNameMessageId} />,
        subtitle: <FormattedMessage id={f.helperTextMessageId} />,
      })),
    [applicationType]
  );

  const onApplyClick = useCallback(() => {
    if (selected == null) return;
    void onApply(selected);
  }, [onApply, selected]);

  const canApply =
    selected != null && selected !== currentFrameworkId && !applying;

  const onOpenChange = useCallback(
    (open: boolean) => {
      if (!open && !applying) {
        onDismiss();
      }
    },
    [applying, onDismiss]
  );

  return (
    <Dialog.Root open={visible} onOpenChange={onOpenChange}>
      <Dialog.Content maxWidth="620px" size="3">
        <Dialog.Title>
          <FormattedMessage id="EditOAuthClientFormFrameworkQuickStart.change-dialog.title" />
        </Dialog.Title>
        <div className={styles.dialogGrid}>
          <IconRadioCards
            size="2"
            options={options}
            value={selected}
            onValueChange={setSelected}
            itemMinWidth={200}
            itemFillSpaces={true}
          />
        </div>
        <div className={styles.changeFrameworkNote}>
          <FormattedMessage
            id="EditOAuthClientFormFrameworkQuickStart.change-dialog.other-platform-note"
            values={{
              // eslint-disable-next-line react/no-unstable-nested-components
              createAppLink: (chunks: React.ReactNode) => (
                <PortalLink to={`/project/${appID}/configuration/apps/add`}>
                  {chunks}
                </PortalLink>
              ),
            }}
          />
        </div>
        <div className={styles.dialogActions}>
          <SecondaryButton
            size="2"
            onClick={onDismiss}
            disabled={applying}
            text={<FormattedMessage id="cancel" />}
          />
          <PrimaryButton
            size="2"
            onClick={onApplyClick}
            disabled={!canApply}
            loading={applying}
            text={
              <FormattedMessage id="EditOAuthClientFormFrameworkQuickStart.change-dialog.apply" />
            }
          />
        </div>
      </Dialog.Content>
    </Dialog.Root>
  );
}

interface CookieSnippetSectionProps {
  snippet: { language: string; code: string };
}

function CookieSnippetSection({ snippet }: CookieSnippetSectionProps) {
  return (
    <>
      <Text as="p" size="4" weight="bold" className={styles.sectionHeading}>
        <FormattedMessage id="EditOAuthClientFormFrameworkQuickStart.snippet.title" />
      </Text>
      <Text as="p" size="2" className={styles.snippetDescription}>
        <FormattedMessage id="EditOAuthClientFormFrameworkQuickStart.snippet.description" />
      </Text>
      <div className={styles.snippetCard}>
        <div className={styles.snippetHeader}>
          <span className={styles.snippetLanguage}>{snippet.language}</span>
          <div className={styles.snippetCopyWrap}>
            <CopyIconButton textToCopy={snippet.code} />
          </div>
        </div>
        <pre className={styles.snippetCode}>
          <code>{snippet.code}</code>
        </pre>
      </div>
    </>
  );
}

interface CopyFieldProps {
  label: string;
  value: string;
  suffix?: React.ReactNode;
}

// A read-only v2 TextField with a copy button (or a custom suffix),
// replacing the FluentUI TextFieldWithCopyButton.
function CopyField({ label, value, suffix }: CopyFieldProps) {
  return (
    <TextField
      size="2"
      label={label}
      value={value}
      readOnly={true}
      suffixPlain={true}
      suffix={suffix ?? <CopyIconButton textToCopy={value} />}
    />
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

  return (
    <div className={styles.oidcSection}>
      <Text as="p" size="4" weight="bold" className={styles.sectionHeading}>
        <FormattedMessage id="EditOAuthClientFormFrameworkQuickStart.oidc.title" />
      </Text>
      <Text as="p" size="2" className={styles.oidcDescription}>
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
      <CopyField
        label={renderToString(
          "EditOAuthClientFormFrameworkQuickStart.oidc.client-id"
        )}
        value={client.client_id}
      />
      {showSecret ? (
        <CopyField
          label={renderToString(
            "EditOAuthClientFormFrameworkQuickStart.oidc.client-secret"
          )}
          value={secretValue}
          suffix={
            isRevealed ? undefined : (
              <RadixIconButton
                variant="ghost"
                color="gray"
                size="2"
                aria-label={renderToString("reveal")}
                disabled={isRevealing}
                onClick={onRevealClick}
              >
                <EyeOpenIcon width="1rem" height="1rem" />
              </RadixIconButton>
            )
          }
        />
      ) : null}
      <CopyField
        label={renderToString(
          "EditOAuthClientFormFrameworkQuickStart.oidc.scope"
        )}
        value={OIDC_RECOMMENDED_SCOPE}
      />
      {endpoints.authorize != null ? (
        <CopyField
          label={renderToString(
            "EditOAuthClientFormFrameworkQuickStart.oidc.login-endpoint"
          )}
          value={endpoints.authorize}
        />
      ) : null}
      {endpoints.userinfo != null ? (
        <CopyField
          label={renderToString(
            "EditOAuthClientFormFrameworkQuickStart.oidc.userinfo-endpoint"
          )}
          value={endpoints.userinfo}
        />
      ) : null}
      <CopyField
        label={renderToString(
          "EditOAuthClientFormFrameworkQuickStart.oidc.token-endpoint"
        )}
        value={endpoints.token}
      />
      {endpoints.endSession != null ? (
        <CopyField
          label={renderToString(
            "EditOAuthClientFormFrameworkQuickStart.oidc.end-session-endpoint"
          )}
          value={endpoints.endSession}
        />
      ) : null}
      <CopyField
        label={renderToString(
          "EditOAuthClientFormFrameworkQuickStart.oidc.jwks-uri"
        )}
        value={endpoints.jwksUri}
      />
    </div>
  );
}
