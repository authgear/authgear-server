import cn from "classnames";
import React, { useCallback, useContext, useMemo, useState } from "react";
import {
  useLocation,
  useNavigate,
  useParams,
  useSearchParams,
} from "react-router-dom";
import { Icon, Text, useTheme, Image, ImageFit } from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";

import ScreenContent from "../../ScreenContent";
import NavBreadcrumb, { BreadcrumbItem } from "../../NavBreadcrumb";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import EditOAuthClientForm from "./EditOAuthClientForm";
import { ApplicationType, OAuthClientConfig } from "../../types";
import { AppSecretConfigFormModel } from "../../hook/useAppSecretConfigForm";
import FormContainer from "../../FormContainer";
import ScreenLayoutScrollView from "../../ScreenLayoutScrollView";
import styles from "./EditOAuthClientScreen.module.css";
import Widget from "../../Widget";
import ExternalLink from "../../ExternalLink";
import flutterIconURL from "../../images/framework_flutter.svg";
import xamarinIconURL from "../../images/framework_xamarin.svg";
import PrimaryButton from "../../PrimaryButton";
import { useAppFeatureConfigQuery } from "./query/appFeatureConfigQuery";
import { AppSecretKey } from "./globalTypes.generated";
import { startReauthentication } from "./Authenticated";
import { useLocationEffect } from "../../hook/useLocationEffect";
import { useAppSecretVisitToken } from "./mutations/generateAppSecretVisitTokenMutation";
import { useOAuthClientForm } from "../../hook/useOAuthClientForm";

interface FormState {
  publicOrigin: string;
  clients: OAuthClientConfig[];
  editedClient: OAuthClientConfig | null;
  removeClientByID?: string;
  clientSecretMap: Partial<Record<string, string>>;
}

interface LocationState {
  isClientSecretRevealed: boolean;
}

function isLocationState(raw: unknown): raw is LocationState {
  return (
    raw != null &&
    typeof raw === "object" &&
    (raw as Partial<LocationState>).isClientSecretRevealed != null
  );
}

interface FrameworkItem {
  icon: React.ReactNode;
  name: string;
  docLink: string;
}

interface QuickStartFrameworkItemProps extends FrameworkItem {
  showOpenTutorialLabelWhenHover: boolean;
}

const QuickStartFrameworkItem: React.VFC<QuickStartFrameworkItemProps> =
  function QuickStartFrameworkItem(props) {
    const { icon, name, docLink, showOpenTutorialLabelWhenHover } = props;
    const [isHovering, setIsHovering] = useState(false);

    const onMouseOver = useCallback(() => {
      setIsHovering(true);
    }, [setIsHovering]);

    const onMouseOut = useCallback(() => {
      setIsHovering(false);
    }, [setIsHovering]);

    // when shouldShowArrowIcon is false, open tutorial label will be shown instead
    const shouldShowArrowIcon = useMemo(() => {
      // always show open tutorial label
      if (!showOpenTutorialLabelWhenHover) {
        return true;
      }

      // show open tutorial label when hover
      return !isHovering;
    }, [showOpenTutorialLabelWhenHover, isHovering]);

    return (
      <ExternalLink
        onMouseOver={onMouseOver}
        onMouseOut={onMouseOut}
        className={cn(styles.quickStartItem, {
          [styles.quickStartItemHovered]: isHovering,
        })}
        href={docLink}
        target="_blank"
      >
        <span
          className={styles.quickStartItemContainer}
          // assign css styles `pointer-events: none` here
          // to enforce the click event is triggered on the parent a element
          // for GTM click event tracking
          style={{ pointerEvents: "none" }}
        >
          <span className={styles.quickStartItemIcon}>{icon}</span>
          <Text variant="small" className={styles.quickStartItemText}>
            {name}
          </Text>
          {shouldShowArrowIcon ? (
            <Icon
              className={styles.quickStartItemArrowIcon}
              iconName="ChevronRightSmall"
            />
          ) : null}
          {!shouldShowArrowIcon ? (
            <Text className={styles.quickStartItemOpenTutorial}>
              <FormattedMessage id="EditOAuthClientScreen.quick-start.open-tutorial.label" />
            </Text>
          ) : null}
        </span>
      </ExternalLink>
    );
  };

interface QuickStartFrameworkListProps {
  applicationType?: ApplicationType;
  showOpenTutorialLabelWhenHover: boolean;
}

const QuickStartFrameworkList: React.VFC<QuickStartFrameworkListProps> =
  function QuickStartFrameworkList(props) {
    const { applicationType, showOpenTutorialLabelWhenHover } = props;
    const { renderToString } = useContext(Context);

    const items: FrameworkItem[] = useMemo(() => {
      switch (applicationType) {
        case "spa":
          return [
            {
              icon: <i className={cn("fab", "fa-react")} />,
              name: renderToString(
                "EditOAuthClientScreen.quick-start.framework.react"
              ),
              docLink: "https://docs.authgear.com/tutorials/spa/react",
            },
            {
              icon: <i className={cn("fab", "fa-vuejs")} />,
              name: renderToString(
                "EditOAuthClientScreen.quick-start.framework.vue"
              ),
              docLink: "https://docs.authgear.com/tutorials/spa/vue",
            },
            {
              icon: <i className={cn("fab", "fa-angular")} />,
              name: renderToString(
                "EditOAuthClientScreen.quick-start.framework.angular"
              ),
              docLink: "https://docs.authgear.com/tutorials/spa/angular",
            },
            {
              icon: <i className={cn("fab", "fa-js")} />,
              name: renderToString(
                "EditOAuthClientScreen.quick-start.framework.other-js"
              ),
              docLink: "https://docs.authgear.com/get-started/website",
            },
          ];
        case "traditional_webapp":
          return [
            {
              icon: <Icon iconName="Globe" />,
              name: renderToString(
                "EditOAuthClientScreen.quick-start.framework.traditional-webapp"
              ),
              docLink: "https://docs.authgear.com/get-started/website",
            },
          ];
        case "native":
          return [
            {
              icon: <i className={cn("fab", "fa-react")} />,
              name: renderToString(
                "EditOAuthClientScreen.quick-start.framework.react-native"
              ),
              docLink: "https://docs.authgear.com/get-started/react-native",
            },
            {
              icon: <i className={cn("fab", "fa-apple")} />,
              name: renderToString(
                "EditOAuthClientScreen.quick-start.framework.ios"
              ),
              docLink: "https://docs.authgear.com/get-started/ios",
            },
            {
              icon: <i className={cn("fab", "fa-android")} />,
              name: renderToString(
                "EditOAuthClientScreen.quick-start.framework.android"
              ),
              docLink: "https://docs.authgear.com/get-started/android",
            },
            {
              icon: (
                <Image
                  src={flutterIconURL}
                  imageFit={ImageFit.contain}
                  className={styles.frameworkImage}
                />
              ),
              name: renderToString(
                "EditOAuthClientScreen.quick-start.framework.flutter"
              ),
              docLink: "https://docs.authgear.com/get-started/flutter",
            },
            {
              icon: (
                <Image
                  src={xamarinIconURL}
                  imageFit={ImageFit.contain}
                  className={styles.frameworkImage}
                />
              ),
              name: renderToString(
                "EditOAuthClientScreen.quick-start.framework.xamarin"
              ),
              docLink: "https://docs.authgear.com/get-started/xamarin",
            },
          ];
        case "confidential":
          return [
            {
              icon: <i className={cn("fab", "fa-openid")} />,
              name: renderToString(
                "EditOAuthClientScreen.quick-start.framework.oidc"
              ),
              docLink: "https://docs.authgear.com/integrate/oidc-provider",
            },
          ];
        case "third_party_app":
          return [
            {
              icon: <i className={cn("fab", "fa-openid")} />,
              name: renderToString(
                "EditOAuthClientScreen.quick-start.framework.oidc"
              ),
              docLink: "https://docs.authgear.com/integrate/oidc-provider",
            },
          ];
        default:
          return [];
      }
    }, [applicationType, renderToString]);

    if (applicationType == null) {
      return null;
    }

    return (
      <>
        {items.map((item, index) => (
          <QuickStartFrameworkItem
            key={`quick-start-${index}`}
            showOpenTutorialLabelWhenHover={showOpenTutorialLabelWhenHover}
            {...item}
          />
        ))}
      </>
    );
  };

interface EditOAuthClientNavBreadcrumbProps {
  clientName: string;
}

const EditOAuthClientNavBreadcrumb: React.VFC<EditOAuthClientNavBreadcrumbProps> =
  function EditOAuthClientNavBreadcrumb(props) {
    const navBreadcrumbItems: BreadcrumbItem[] = useMemo(() => {
      return [
        {
          to: "~/configuration/apps",
          label: (
            <FormattedMessage id="ApplicationsConfigurationScreen.title" />
          ),
        },
        {
          to: ".",
          label: props.clientName,
        },
      ];
    }, [props.clientName]);

    return (
      <NavBreadcrumb className={styles.widget} items={navBreadcrumbItems} />
    );
  };

interface EditOAuthClientContentProps {
  form: AppSecretConfigFormModel<FormState>;
  clientID: string;
  customUIEnabled: boolean;
  app2appEnabled: boolean;
}

const EditOAuthClientContent: React.VFC<EditOAuthClientContentProps> =
  function EditOAuthClientContent(props) {
    const {
      clientID,
      form: { state, setState },
      customUIEnabled,
      app2appEnabled,
    } = props;
    const theme = useTheme();

    const navigate = useNavigate();

    const client =
      state.editedClient ?? state.clients.find((c) => c.client_id === clientID);

    const clientSecret = useMemo(() => {
      return client?.client_id
        ? state.clientSecretMap[client.client_id]
        : undefined;
    }, [client, state.clientSecretMap]);

    const onClientConfigChange = useCallback(
      (editedClient: OAuthClientConfig) => {
        setState((state) => ({ ...state, editedClient }));
      },
      [setState]
    );

    const onRevealSecret = useCallback(() => {
      const state: LocationState = {
        isClientSecretRevealed: true,
      };
      startReauthentication(navigate, state).catch((e) => {
        // Normally there should not be any error.
        console.error(e);
      });
    }, [navigate]);

    if (client == null) {
      return (
        <Text>
          <FormattedMessage
            id="EditOAuthClientScreen.client-not-found"
            values={{ clientID }}
          />
        </Text>
      );
    }

    return (
      <ScreenContent>
        <EditOAuthClientNavBreadcrumb clientName={client.name ?? ""} />
        <div className={cn(styles.widget, styles.widgetColumn)}>
          <EditOAuthClientForm
            publicOrigin={state.publicOrigin}
            clientConfig={client}
            clientSecret={clientSecret}
            customUIEnabled={customUIEnabled}
            app2appEnabled={app2appEnabled}
            onClientConfigChange={onClientConfigChange}
            onRevealSecret={onRevealSecret}
          />
        </div>
        <div className={styles.quickStartColumn}>
          <Widget>
            <div className={styles.quickStartWidget}>
              <Text className={styles.quickStartWidgetTitle}>
                <Icon
                  className={styles.quickStartWidgetTitleIcon}
                  styles={{ root: { color: theme.palette.themePrimary } }}
                  iconName="Lightbulb"
                />
                <FormattedMessage id="EditOAuthClientScreen.quick-start-widget.title" />
              </Text>
              <Text>
                <FormattedMessage
                  id="EditOAuthClientScreen.quick-start-widget.question"
                  values={{
                    applicationType: client.x_application_type ?? "",
                  }}
                />
              </Text>
              <QuickStartFrameworkList
                applicationType={client.x_application_type}
                showOpenTutorialLabelWhenHover={false}
              />
            </div>
          </Widget>
        </div>
      </ScreenContent>
    );
  };

interface OAuthQuickStartScreenContentProps {
  form: AppSecretConfigFormModel<FormState>;
  clientID: string;
}

const OAuthQuickStartScreenContent: React.VFC<OAuthQuickStartScreenContentProps> =
  function OAuthQuickStartScreenContent(props) {
    const {
      clientID,
      form: { state },
    } = props;
    const navigate = useNavigate();
    const theme = useTheme();
    const client =
      state.editedClient ?? state.clients.find((c) => c.client_id === clientID);

    const onNextButtonClick = useCallback(
      (e) => {
        e.preventDefault();
        e.stopPropagation();
        navigate(".");
      },
      [navigate]
    );

    return (
      <ScreenLayoutScrollView>
        <ScreenContent>
          <EditOAuthClientNavBreadcrumb clientName={client?.name ?? ""} />
          <Widget className={styles.widget}>
            <Text variant="xLarge" block={true}>
              <Icon
                className={styles.quickStartScreenTitleIcon}
                styles={{ root: { color: theme.palette.themePrimary } }}
                iconName="Lightbulb"
              />
              <FormattedMessage id="EditOAuthClientScreen.quick-start-screen.title" />
            </Text>
            <Text className={styles.quickStartScreenDescription} block={true}>
              <FormattedMessage
                id="EditOAuthClientScreen.quick-start-screen.question"
                values={{
                  applicationType: client?.x_application_type ?? "",
                }}
              />
            </Text>
            <QuickStartFrameworkList
              applicationType={client?.x_application_type}
              showOpenTutorialLabelWhenHover={true}
            />
            <div className={styles.quickStartScreenButtons}>
              <PrimaryButton
                onClick={onNextButtonClick}
                text={<FormattedMessage id="next" />}
              />
            </div>
          </Widget>
        </ScreenContent>
      </ScreenLayoutScrollView>
    );
  };

const EditOAuthClientScreen1: React.VFC<{
  appID: string;
  clientID: string;
  secretToken: string | null;
}> = function EditOAuthClientScreen1({ appID, clientID, secretToken }) {
  const form = useOAuthClientForm(appID, secretToken);

  const featureConfig = useAppFeatureConfigQuery(appID);

  const [searchParams] = useSearchParams();
  const isQuickScreenVisible = useMemo(() => {
    const quickstart = searchParams.get("quickstart");
    return quickstart === "true";
  }, [searchParams]);

  const customUIEnabled = useMemo(() => {
    if (featureConfig.loading) {
      return false;
    }
    return featureConfig.effectiveFeatureConfig?.oauth?.client
      ?.custom_ui_enabled;
  }, [
    featureConfig.loading,
    featureConfig.effectiveFeatureConfig?.oauth?.client?.custom_ui_enabled,
  ]);

  const app2appEnabled = useMemo(() => {
    if (featureConfig.loading) {
      return false;
    }
    return featureConfig.effectiveFeatureConfig?.oauth?.client?.app2app_enabled;
  }, [featureConfig]);

  if (form.isLoading) {
    return <ShowLoading />;
  }

  if (form.loadError) {
    return <ShowError error={form.loadError} onRetry={form.reload} />;
  }

  if (isQuickScreenVisible) {
    return <OAuthQuickStartScreenContent form={form} clientID={clientID} />;
  }

  return (
    <FormContainer form={form} stickyFooterComponent={true}>
      <EditOAuthClientContent
        form={form}
        clientID={clientID}
        customUIEnabled={customUIEnabled}
        app2appEnabled={app2appEnabled}
      />
    </FormContainer>
  );
};

const SECRETS = [AppSecretKey.OauthClientSecrets];

const EditOAuthClientScreen: React.VFC = function EditOAuthClientScreen() {
  const { appID, clientID } = useParams() as {
    appID: string;
    clientID: string;
  };
  const location = useLocation();
  const [shouldRefreshToken] = useState<boolean>(() => {
    const { state } = location;
    if (isLocationState(state) && state.isClientSecretRevealed) {
      return true;
    }
    return false;
  });
  useLocationEffect<LocationState>(() => {
    // Pop the location state if exist
  });
  const { token, loading, error, retry } = useAppSecretVisitToken(
    appID,
    SECRETS,
    shouldRefreshToken
  );

  if (error) {
    return <ShowError error={error} onRetry={retry} />;
  }

  if (loading || token === undefined) {
    return <ShowLoading />;
  }

  return (
    <EditOAuthClientScreen1
      appID={appID}
      clientID={clientID}
      secretToken={token}
    />
  );
};

export default EditOAuthClientScreen;
