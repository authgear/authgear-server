import React, { useCallback, useContext, useMemo, useRef } from "react";
import { useNavigate, useParams } from "react-router-dom";
import cn from "classnames";
import { produce } from "immer";
import {
  DropdownMenu,
  Flex,
  IconButton as RadixIconButton,
  RadioGroup,
  Separator,
  Text,
} from "@radix-ui/themes";
import {
  DotsVerticalIcon,
  InfoCircledIcon,
  Pencil1Icon,
} from "@radix-ui/react-icons";
import { Context, FormattedMessage } from "../../intl";
import {
  isPromotionConflictBehaviour,
  PortalAPIAppConfig,
  PromotionConflictBehaviour,
  promotionConflictBehaviours,
  OAuthClientConfig,
} from "../../types";
import { clearEmptyObject } from "../../util/misc";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import ScreenContent from "../../ScreenContent";
import {
  AppConfigFormModel,
  useAppConfigForm,
} from "../../hook/useAppConfigForm";
import { Tooltip } from "../../components/v2/Tooltip/Tooltip";
import { Toggle } from "../../components/v2/Toggle/Toggle";
import { FormField } from "../../components/v2/FormField/FormField";
import { SettingsSectionCard } from "../../components/v2/SettingsSectionCard/SettingsSectionCard";
import { SaveFunctionBar } from "../../components/v2/SaveFunctionBar/SaveFunctionBar";
import ShowOnlyIfSIWEIsDisabled from "./ShowOnlyIfSIWEIsDisabled";
import styles from "./AnonymousUsersConfigurationScreen.module.css";
import FormContainer from "../../FormContainer";
import { useFormContainerBaseContext } from "../../FormContainerBase";
import Link from "../../Link";
import { formatSeconds } from "../../util/formatDuration";

interface FormState {
  enabled: boolean;
  promotionConflictBehaviour: PromotionConflictBehaviour;
  oauthClients: OAuthClientConfig[];
  sessionLifetimeSeconds: number | undefined;
  sessionIdleTimeoutEnabled: boolean;
  sessionIdleTimeoutSeconds: number | undefined;
}

function constructFormState(config: PortalAPIAppConfig): FormState {
  const enabled =
    config.authentication?.identities?.includes("anonymous") ?? false;
  const promotionConflictBehaviour =
    config.identity?.on_conflict?.promotion ?? "error";
  const oauthClients = config.oauth?.clients ?? [];
  return {
    enabled,
    promotionConflictBehaviour,
    oauthClients,
    sessionLifetimeSeconds: config.session?.lifetime_seconds,
    sessionIdleTimeoutEnabled: config.session?.idle_timeout_enabled ?? false,
    sessionIdleTimeoutSeconds: config.session?.idle_timeout_seconds,
  };
}

function constructConfig(
  config: PortalAPIAppConfig,
  initialState: FormState,
  currentState: FormState,
  effectiveConfig: PortalAPIAppConfig
): PortalAPIAppConfig {
  return produce(config, (config) => {
    if (initialState.enabled !== currentState.enabled) {
      const identities = (
        effectiveConfig.authentication?.identities ?? []
      ).slice();
      const index = identities.indexOf("anonymous");
      if (currentState.enabled && index === -1) {
        identities.push("anonymous");
      } else if (!currentState.enabled && index >= 0) {
        identities.splice(index, 1);
      }
      config.authentication ??= {};
      config.authentication.identities = identities;
    }
    if (
      currentState.enabled &&
      initialState.promotionConflictBehaviour !==
        currentState.promotionConflictBehaviour
    ) {
      config.identity ??= {};
      config.identity.on_conflict ??= {};
      config.identity.on_conflict.promotion =
        currentState.promotionConflictBehaviour;
    }
    clearEmptyObject(config);
  });
}

const conflictBehaviourMessageId: Record<PromotionConflictBehaviour, string> = {
  login: "AnonymousIdentityConflictBehaviour.login",
  error: "AnonymousIdentityConflictBehaviour.error",
};

interface OAuthClientListItem {
  clientID: string;
  name: string;
  refreshTokenIdleTimeout: string;
  refreshTokenLifetime: string;
}

function LabelWithTooltip(props: {
  labelId: string;
  tooltipId: string;
}): React.ReactElement {
  const { labelId, tooltipId } = props;
  return (
    <div className={styles.tooltipLabel}>
      <Text as="span" size="2">
        <FormattedMessage id={labelId} />
      </Text>
      <Tooltip content={<FormattedMessage id={tooltipId} />}>
        <InfoCircledIcon className={styles.infoIcon} />
      </Tooltip>
    </div>
  );
}

interface AnonymousUserLifeTimeDescriptionProps {
  form: AppConfigFormModel<FormState>;
  className?: string;
}

const AnonymousUserLifeTimeDescription: React.VFC<AnonymousUserLifeTimeDescriptionProps> =
  function AnonymousUserLifeTimeDescription(props) {
    const { form, className } = props;
    const { locale, renderToString } = useContext(Context);
    const { appID } = useParams() as { appID: string };
    const navigate = useNavigate();
    const {
      sessionIdleTimeoutEnabled,
      sessionIdleTimeoutSeconds,
      sessionLifetimeSeconds,
      oauthClients,
    } = form.state;

    const items: OAuthClientListItem[] = useMemo(() => {
      return oauthClients.map((client) => {
        return {
          clientID: client.client_id,
          name: client.name ?? "",
          refreshTokenIdleTimeout: client.refresh_token_idle_timeout_enabled
            ? client.refresh_token_idle_timeout_seconds != null
              ? formatSeconds(
                  locale,
                  client.refresh_token_idle_timeout_seconds
                ) ?? "-"
              : "-"
            : "-",
          refreshTokenLifetime:
            client.refresh_token_lifetime_seconds != null
              ? formatSeconds(locale, client.refresh_token_lifetime_seconds) ??
                ""
              : "",
        };
      });
    }, [locale, oauthClients]);

    const onEditApplication = useCallback(
      (clientID: string) => {
        navigate(
          `/project/${appID}/configuration/apps/${encodeURIComponent(
            clientID
          )}/edit`
        );
      },
      [appID, navigate]
    );

    return (
      <SettingsSectionCard
        className={cn(styles.widget, className)}
        contentClassName="gap-4"
        title={
          <FormattedMessage id="AnonymousUsersConfigurationScreen.user-lifetime.title" />
        }
      >
        <Text
          as="p"
          size="2"
          color="gray"
          className={styles.sectionDescription}
        >
          <FormattedMessage
            id="AnonymousUsersConfigurationScreen.user-lifetime.description"
            values={{
              // eslint-disable-next-line react/no-unstable-nested-components
              b: (chunks: React.ReactNode) => <b>{chunks}</b>,
            }}
          />
        </Text>
        <div className={styles.lifetimeSection}>
          <Text as="p" size="2" weight="medium" className={styles.sectionTitle}>
            <FormattedMessage id="AnonymousUsersConfigurationScreen.user-lifetime.cookie.title" />
          </Text>
          <div className={styles.sessionInfo}>
            {sessionIdleTimeoutEnabled ? (
              <>
                <LabelWithTooltip
                  labelId="AnonymousUsersConfigurationScreen.user-lifetime.cookie.label.idle-timeout"
                  tooltipId="AnonymousUsersConfigurationScreen.user-lifetime.cookie.tooltip.idle-timeout"
                />
                <Text as="span" size="2">
                  <FormattedMessage
                    id="AnonymousUsersConfigurationScreen.user-lifetime.cookie.value.seconds"
                    values={{
                      formattedDuration:
                        sessionIdleTimeoutSeconds != null
                          ? formatSeconds(locale, sessionIdleTimeoutSeconds) ??
                            ""
                          : "",
                    }}
                  />
                </Text>
              </>
            ) : null}
            <LabelWithTooltip
              labelId="AnonymousUsersConfigurationScreen.user-lifetime.cookie.label.session-lifetime"
              tooltipId="AnonymousUsersConfigurationScreen.user-lifetime.cookie.tooltip.session-lifetime"
            />
            <Text as="span" size="2">
              <FormattedMessage
                id="AnonymousUsersConfigurationScreen.user-lifetime.cookie.value.seconds"
                values={{
                  formattedDuration:
                    sessionLifetimeSeconds != null
                      ? formatSeconds(locale, sessionLifetimeSeconds) ?? ""
                      : "",
                }}
              />
            </Text>
          </div>
        </div>

        <Separator size="4" className={styles.sectionSeparator} />

        <div className={styles.lifetimeSection}>
          <Text as="p" size="2" weight="medium" className={styles.sectionTitle}>
            <FormattedMessage id="AnonymousUsersConfigurationScreen.user-lifetime.token.title" />
          </Text>
          <div className={styles.tableWrapper}>
            <div className={styles.table}>
              <div className={styles.tableHeader}>
                <div className={styles.headerCellName}>
                  <FormattedMessage id="AnonymousUsersConfigurationScreen.user-lifetime.applications-list.label.name" />
                </div>
                <div className={styles.headerCellIdleTimeout}>
                  <div className={styles.tooltipLabel}>
                    <FormattedMessage id="AnonymousUsersConfigurationScreen.user-lifetime.applications-list.label.refresh-token-idle-timeout" />
                    <Tooltip
                      content={
                        <FormattedMessage id="AnonymousUsersConfigurationScreen.user-lifetime.applications-list.tooltip.refresh-token-idle-timeout" />
                      }
                    >
                      <InfoCircledIcon className={styles.infoIcon} />
                    </Tooltip>
                  </div>
                </div>
                <div className={styles.headerCellLifetime}>
                  <div className={styles.tooltipLabel}>
                    <FormattedMessage id="AnonymousUsersConfigurationScreen.user-lifetime.applications-list.label.refresh-token-lifetime" />
                    <Tooltip
                      content={
                        <FormattedMessage id="AnonymousUsersConfigurationScreen.user-lifetime.applications-list.tooltip.refresh-token-lifetime" />
                      }
                    >
                      <InfoCircledIcon className={styles.infoIcon} />
                    </Tooltip>
                  </div>
                </div>
                <div className={styles.headerCellActions} aria-hidden={true} />
              </div>
              {items.map((item) => (
                <div key={item.clientID} className={styles.tableRow}>
                  <div className={styles.cellName}>
                    <Text size="2" className={styles.cellNameText}>
                      {item.name}
                    </Text>
                  </div>
                  <div className={styles.cellIdleTimeout}>
                    <Text size="2">{item.refreshTokenIdleTimeout}</Text>
                  </div>
                  <div className={styles.cellLifetime}>
                    <Text size="2">{item.refreshTokenLifetime}</Text>
                  </div>
                  <div className={styles.cellActions}>
                    <DropdownMenu.Root>
                      <DropdownMenu.Trigger>
                        <RadixIconButton
                          className={styles.rowActionsButton}
                          variant="soft"
                          color="gray"
                          size="2"
                          aria-label={renderToString(
                            "AnonymousUsersConfigurationScreen.user-lifetime.applications-list.row-actions"
                          )}
                        >
                          <DotsVerticalIcon width="1rem" height="1rem" />
                        </RadixIconButton>
                      </DropdownMenu.Trigger>
                      <DropdownMenu.Content align="end">
                        <DropdownMenu.Item
                          onSelect={() => {
                            onEditApplication(item.clientID);
                          }}
                        >
                          <Pencil1Icon />
                          <FormattedMessage id="edit" />
                        </DropdownMenu.Item>
                      </DropdownMenu.Content>
                    </DropdownMenu.Root>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>

        <Text as="p" size="2" color="gray" className={styles.applicationsLink}>
          <FormattedMessage
            id="AnonymousUsersConfigurationScreen.user-lifetime.go-to-applications.description"
            values={{
              // eslint-disable-next-line react/no-unstable-nested-components
              reactRouterLink: (chunks: React.ReactNode) => (
                <Link to={`/project/${appID}/configuration/apps`}>
                  {chunks}
                </Link>
              ),
            }}
          />
        </Text>
      </SettingsSectionCard>
    );
  };

interface AnonymousUserConfigurationContentProps {
  form: AppConfigFormModel<FormState>;
}

const AnonymousUserConfigurationContent: React.VFC<AnonymousUserConfigurationContentProps> =
  function AnonymousUserConfigurationContent(props) {
    const { state, setState } = props.form;

    const { getIsDirty } = useFormContainerBaseContext();
    const isDirty = useMemo(() => getIsDirty(), [getIsDirty]);
    const contentWidthAnchorRef = useRef<HTMLDivElement>(null);

    const conflictBehaviourOptions = useMemo(
      () =>
        promotionConflictBehaviours.map((behaviour) => ({
          value: behaviour,
          labelId: conflictBehaviourMessageId[behaviour],
        })),
      []
    );

    const onEnableChange = useCallback(
      (checked: boolean) =>
        setState((state) => ({
          ...state,
          enabled: checked,
        })),
      [setState]
    );

    const onConflictOptionChange = useCallback(
      (value: string) => {
        if (isPromotionConflictBehaviour(value)) {
          setState((state) => ({
            ...state,
            promotionConflictBehaviour: value,
          }));
        }
      },
      [setState]
    );

    return (
      <ScreenContent className={cn(isDirty ? styles.contentWithSaveBar : null)}>
        <div
          ref={contentWidthAnchorRef}
          className={cn(styles.widget, styles.pageHeader)}
        >
          <Text as="p" size="5" weight="bold" className={styles.pageTitle}>
            <FormattedMessage id="AnonymousUsersConfigurationScreen.title" />
          </Text>
          <Text as="p" size="2" color="gray" className={styles.pageDescription}>
            <FormattedMessage id="AnonymousUsersConfigurationScreen.description" />
          </Text>
        </div>
        <ShowOnlyIfSIWEIsDisabled className={styles.widget}>
          <SettingsSectionCard
            className={styles.widget}
            contentClassName="gap-4"
            title={
              <FormattedMessage id="AnonymousUsersConfigurationScreen.settings.label" />
            }
          >
            <Toggle
              checked={state.enabled}
              onCheckedChange={onEnableChange}
              textWeight="medium"
              text={
                <FormattedMessage id="AnonymousUsersConfigurationScreen.enable.label" />
              }
            />
            {state.enabled ? (
              <FormField
                size="2"
                labelSize="2"
                labelSpace="1"
                label={
                  <FormattedMessage id="AnonymousUsersConfigurationScreen.conflict-droplist.label" />
                }
              >
                <RadioGroup.Root
                  value={state.promotionConflictBehaviour}
                  onValueChange={onConflictOptionChange}
                >
                  <Flex direction="column" gap="2">
                    {conflictBehaviourOptions.map((opt) => (
                      <Text as="label" size="2" key={opt.value}>
                        <Flex gap="2" align="center">
                          <RadioGroup.Item value={opt.value} />
                          <FormattedMessage id={opt.labelId} />
                        </Flex>
                      </Text>
                    ))}
                  </Flex>
                </RadioGroup.Root>
              </FormField>
            ) : null}
          </SettingsSectionCard>
          <AnonymousUserLifeTimeDescription
            form={props.form}
            className={
              isDirty ? styles.settingsCardSaveBarClearance : undefined
            }
          />
        </ShowOnlyIfSIWEIsDisabled>
        <SaveFunctionBar anchorRef={contentWidthAnchorRef} />
      </ScreenContent>
    );
  };

const AnonymousUserConfigurationScreen: React.VFC =
  function AnonymousUserConfigurationScreen() {
    const { appID } = useParams() as { appID: string };
    const form = useAppConfigForm({
      appID,
      constructFormState,
      constructConfig,
    });

    if (form.isLoading) {
      return <ShowLoading />;
    }

    if (form.loadError) {
      return <ShowError error={form.loadError} onRetry={form.reload} />;
    }

    return (
      <FormContainer form={form} hideFooterComponent={true}>
        <AnonymousUserConfigurationContent form={form} />
      </FormContainer>
    );
  };

export default AnonymousUserConfigurationScreen;
