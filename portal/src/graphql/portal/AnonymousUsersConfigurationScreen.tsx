/* global JSX */
import React, { useCallback, useContext, useMemo } from "react";
import { useParams } from "react-router-dom";
import { produce } from "immer";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import {
  Dropdown,
  IDropdownOption,
  Text,
  DetailsList,
  SelectionMode,
  IColumn,
  IDetailsHeaderProps,
  DetailsHeader,
  IRenderFunction,
  IDetailsColumnRenderTooltipProps,
} from "@fluentui/react";
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
import ScreenTitle from "../../ScreenTitle";
import ScreenDescription from "../../ScreenDescription";
import WidgetTitle from "../../WidgetTitle";
import Widget from "../../Widget";
import {
  AppConfigFormModel,
  useAppConfigForm,
} from "../../hook/useAppConfigForm";
import FormContainer from "../../FormContainer";
import Tooltip from "../../Tooltip";
import Toggle from "../../Toggle";
import ShowOnlyIfSIWEIsDisabled from "./ShowOnlyIfSIWEIsDisabled";
import styles from "./AnonymousUsersConfigurationScreen.module.css";
import {
  FormContainerBase,
  useFormContainerBaseContext,
} from "../../FormContainerBase";
import PrimaryButton from "../../PrimaryButton";
import HorizontalDivider from "../../HorizontalDivider";

const dropDownStyles = {
  dropdown: {
    maxWidth: "300px",
  },
};

interface FormState {
  enabled: boolean;
  promotionConflictBehaviour: PromotionConflictBehaviour;
  oauthClients: OAuthClientConfig[];
  sessionPersistentCookie: boolean;
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
    sessionPersistentCookie: !(config.session?.cookie_non_persistent ?? false),
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
  // eslint-disable-next-line complexity
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
  name: string;
  refreshTokenIdleTimeout: string;
  refreshTokenLifetime: string;
}

interface AnonymousUserLifeTimeDescriptionProps {
  form: AppConfigFormModel<FormState>;
}

const AnonymousUserLifeTimeDescription: React.VFC<AnonymousUserLifeTimeDescriptionProps> =
  function AnonymousUserLifeTimeDescription(props) {
    const { renderToString } = useContext(Context);
    const { appID } = useParams() as { appID: string };
    const {
      sessionIdleTimeoutEnabled,
      sessionIdleTimeoutSeconds,
      sessionLifetimeSeconds,
      sessionPersistentCookie,
      oauthClients,
    } = props.form.state;

    const columns: IColumn[] = useMemo(
      () => [
        {
          key: "name",
          name: renderToString(
            "AnonymousUsersConfigurationScreen.user-lifetime.applications-list.label.name"
          ),
          minWidth: 200,
          maxWidth: 200,
          isMultiline: true,
        },
        {
          key: "refresh-token-idle-timeout",
          name: "",
          minWidth: 170,
          maxWidth: 170,
        },
        {
          key: "refresh-token-lifetime",
          name: "",
          minWidth: 170,
          maxWidth: 170,
        },
      ],
      [renderToString]
    );

    const items: OAuthClientListItem[] = useMemo(() => {
      return oauthClients.map((client) => {
        return {
          name: client.name ?? "",
          refreshTokenIdleTimeout: client.refresh_token_idle_timeout_enabled
            ? client.refresh_token_idle_timeout_seconds?.toFixed(0) ?? ""
            : "-",
          refreshTokenLifetime:
            client.refresh_token_lifetime_seconds?.toFixed(0) ?? "",
        };
      });
    }, [oauthClients]);

    const onRenderItemColumn = useCallback(
      (item?: OAuthClientListItem, _index?: number, column?: IColumn) => {
        if (item == null) {
          return null;
        }
        switch (column?.key) {
          case "name":
            return item.name;
          case "refresh-token-idle-timeout":
            return item.refreshTokenIdleTimeout;
          case "refresh-token-lifetime":
            return item.refreshTokenLifetime;
          default:
            return null;
        }
      },
      []
    );

    const onRenderColumnHeaderTooltip: IRenderFunction<IDetailsColumnRenderTooltipProps> =
      useCallback(
        (
          props?: IDetailsColumnRenderTooltipProps,
          defaultRender?: (
            props: IDetailsColumnRenderTooltipProps
          ) => JSX.Element | null
        ) => {
          if (props == null || defaultRender == null || props.column == null) {
            return null;
          }
          if (
            props.column.key === "refresh-token-idle-timeout" ||
            props.column.key === "refresh-token-lifetime"
          ) {
            return (
              <span className={styles.tooltipHeader}>
                <Text variant="medium" className={styles.bold}>
                  <FormattedMessage
                    id={
                      "AnonymousUsersConfigurationScreen.user-lifetime.applications-list.label." +
                      props.column.key
                    }
                  />
                </Text>
                <Tooltip
                  tooltipMessageId={
                    "AnonymousUsersConfigurationScreen.user-lifetime.applications-list.tooltip." +
                    props.column.key
                  }
                />
              </span>
            );
          }
          return defaultRender(props);
        },
        []
      );

    const onRenderDetailsHeader = useCallback(
      (props?: IDetailsHeaderProps) => {
        if (props == null) {
          return null;
        }
        return (
          <DetailsHeader
            {...props}
            className={styles.detailsHeader}
            onRenderColumnHeaderTooltip={onRenderColumnHeaderTooltip}
          />
        );
      },
      [onRenderColumnHeaderTooltip]
    );

    return (
      <Widget className={styles.widget}>
        <WidgetTitle>
          <FormattedMessage id="AnonymousUsersConfigurationScreen.user-lifetime.title" />
        </WidgetTitle>
        <Text
          variant="medium"
          block={true}
          className={styles.widgetDescription}
        >
          <FormattedMessage id="AnonymousUsersConfigurationScreen.user-lifetime.description" />
        </Text>
        <div>
          <Text className={styles.title} variant="medium" block={true}>
            <FormattedMessage id="AnonymousUsersConfigurationScreen.user-lifetime.cookie.title" />
          </Text>
          <div className={styles.sessionInfo}>
            {sessionIdleTimeoutEnabled ? (
              <>
                <div className={styles.tooltipLabel}>
                  <Text variant="medium">
                    <FormattedMessage id="AnonymousUsersConfigurationScreen.user-lifetime.cookie.label.idle-timeout" />
                  </Text>
                  <Tooltip tooltipMessageId="AnonymousUsersConfigurationScreen.user-lifetime.cookie.tooltip.idle-timeout" />
                </div>
                <Text variant="medium">
                  <FormattedMessage
                    id="AnonymousUsersConfigurationScreen.user-lifetime.cookie.value.seconds"
                    values={{
                      seconds: sessionIdleTimeoutSeconds?.toFixed(0) ?? "",
                    }}
                  />
                </Text>
              </>
            ) : null}
            <div className={styles.tooltipLabel}>
              <Text variant="medium">
                <FormattedMessage id="AnonymousUsersConfigurationScreen.user-lifetime.cookie.label.session-lifetime" />
              </Text>
              <Tooltip tooltipMessageId="AnonymousUsersConfigurationScreen.user-lifetime.cookie.tooltip.session-lifetime" />
            </div>
            <Text variant="medium">
              <FormattedMessage
                id="AnonymousUsersConfigurationScreen.user-lifetime.cookie.value.seconds"
                values={{
                  seconds: sessionLifetimeSeconds?.toFixed(0) ?? "",
                }}
              />
            </Text>
            <div className={styles.tooltipLabel}>
              <Text variant="medium">
                <FormattedMessage id="AnonymousUsersConfigurationScreen.user-lifetime.cookie.label.persistent-cookie" />
              </Text>
              <Tooltip tooltipMessageId="AnonymousUsersConfigurationScreen.user-lifetime.cookie.tooltip.persistent-cookie" />
            </div>
            <Text variant="medium">
              <FormattedMessage
                id={sessionPersistentCookie ? "enabled" : "disabled"}
              />
            </Text>
          </div>
        </div>
        <div>
          <Text className={styles.title} variant="medium" block={true}>
            <FormattedMessage id="AnonymousUsersConfigurationScreen.user-lifetime.token.title" />
          </Text>
          <DetailsList
            columns={columns}
            items={items}
            selectionMode={SelectionMode.none}
            onRenderItemColumn={onRenderItemColumn}
            onRenderDetailsHeader={onRenderDetailsHeader}
          />
        </div>
        <Text variant="medium" block={true}>
          <FormattedMessage
            id="AnonymousUsersConfigurationScreen.user-lifetime.go-to-applications.description"
            values={{
              applicationsPath: `/project/${appID}/configuration/apps`,
            }}
          />
        </Text>
      </Widget>
    );
  };

interface AnonymousUserConfigurationContentProps {
  form: AppConfigFormModel<FormState>;
}

const AnonymousUserConfigurationContent: React.VFC<AnonymousUserConfigurationContentProps> =
  function AnonymousUserConfigurationContent(props) {
    const { state, setState } = props.form;

    const { renderToString } = useContext(Context);
    const { canSave, onSave } = useFormContainerBaseContext();

    const conflictBehaviourOptions = useMemo(
      () =>
        promotionConflictBehaviours.map((behaviour) => {
          const selectedBehaviour = state.promotionConflictBehaviour;
          return {
            key: behaviour,
            text: renderToString(conflictBehaviourMessageId[behaviour]),
            isSelected: selectedBehaviour === behaviour,
          };
        }),
      [state, renderToString]
    );

    const onEnableChange = useCallback(
      (_event, checked?: boolean) =>
        setState((state) => ({
          ...state,
          enabled: checked ?? false,
        })),
      [setState]
    );

    const onConflictOptionChange = useCallback(
      (_event, option?: IDropdownOption) => {
        const key = option?.key;
        if (key && isPromotionConflictBehaviour(key)) {
          setState((state) => ({
            ...state,
            promotionConflictBehaviour: key,
          }));
        }
      },
      [setState]
    );

    return (
      <ScreenContent>
        <ScreenTitle className={styles.widget}>
          <FormattedMessage id="AnonymousUsersConfigurationScreen.title" />
        </ScreenTitle>
        <ShowOnlyIfSIWEIsDisabled className={styles.widget}>
          <ScreenDescription className={styles.widget}>
            <FormattedMessage id="AnonymousUsersConfigurationScreen.description" />
          </ScreenDescription>
          <Widget className={styles.widget}>
            <WidgetTitle>
              <FormattedMessage id="AnonymousUsersConfigurationScreen.title" />
            </WidgetTitle>
            <Toggle
              checked={state.enabled}
              onChange={onEnableChange}
              label={renderToString(
                "AnonymousUsersConfigurationScreen.enable.label"
              )}
              inlineLabel={false}
            />
            <Dropdown
              styles={dropDownStyles}
              label={renderToString(
                "AnonymousUsersConfigurationScreen.conflict-droplist.label"
              )}
              disabled={!state.enabled}
              options={conflictBehaviourOptions}
              selectedKey={state.promotionConflictBehaviour}
              onChange={onConflictOptionChange}
            />
          </Widget>
          <Widget className={styles.widget}>
            <div>
              <PrimaryButton
                text={renderToString("save")}
                disabled={!canSave}
                onClick={onSave}
              />
            </div>
          </Widget>
          <Widget className={styles.widget}>
            <HorizontalDivider />
          </Widget>
          <AnonymousUserLifeTimeDescription form={props.form} />
        </ShowOnlyIfSIWEIsDisabled>
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
      <FormContainerBase form={form}>
        <AnonymousUserConfigurationContent form={form} />
      </FormContainerBase>
    );
  };

export default AnonymousUserConfigurationScreen;
