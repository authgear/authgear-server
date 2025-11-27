import cn from "classnames";
import React, { useCallback, useContext, useMemo, useState } from "react";
import {
  useLocation,
  useNavigate,
  useParams,
  useSearchParams,
} from "react-router-dom";
import {
  Icon,
  Text,
  useTheme,
  Image,
  ImageFit,
  PivotItem,
} from "@fluentui/react";
import { AGPivot } from "../../components/common/AGPivot";
import { Context, FormattedMessage } from "@oursky/react-messageformat";

import ScreenContent from "../../ScreenContent";
import NavBreadcrumb, { BreadcrumbItem } from "../../NavBreadcrumb";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import EditOAuthClientForm from "./EditOAuthClientForm";
import { EditOAuthClientFormResourcesContent } from "./EditOAuthClientFormResourcesContent";
import { EditOAuthClientFormQuickStartContent } from "./EditOAuthClientFormQuickStartContent";
import {
  ApplicationType,
  OAuthClientConfig,
  PortalAPIAppConfig,
  SAMLIdpSigningCertificate,
} from "../../types";
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
import { useLocationEffect } from "../../hook/useLocationEffect";
import { useAppSecretVisitToken } from "./mutations/generateAppSecretVisitTokenMutation";
import { useOAuthClientForm, FormState } from "../../hook/useOAuthClientForm";
import {
  OAuthClientSAMLForm,
  OAuthClientSAMLFormState,
  getDefaultOAuthClientSAMLFormState,
} from "../../components/applications/OAuthClientSAMLForm";
import { useAppAndSecretConfigQuery } from "./query/appAndSecretConfigQuery";
import iconSaml from "../../images/saml-logo.svg";
import { usePivotNavigation } from "../../hook/usePivot";
import {
  ClientSecretsHook,
  useClientSecret,
} from "../../hook/useClientSecrets";
import { useLoadableView } from "../../hook/useLoadableView";

export interface LocationState {
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
              docLink:
                "https://docs.authgear.com/get-started/single-page-app/react",
            },
            {
              icon: <i className={cn("fab", "fa-vuejs")} />,
              name: renderToString(
                "EditOAuthClientScreen.quick-start.framework.vue"
              ),
              docLink:
                "https://docs.authgear.com/get-started/single-page-app/vue",
            },
            {
              icon: <i className={cn("fab", "fa-angular")} />,
              name: renderToString(
                "EditOAuthClientScreen.quick-start.framework.angular"
              ),
              docLink:
                "https://docs.authgear.com/get-started/single-page-app/angular",
            },
            {
              icon: <i className={cn("fab", "fa-js")} />,
              name: renderToString(
                "EditOAuthClientScreen.quick-start.framework.other-js"
              ),
              docLink:
                "https://docs.authgear.com/get-started/single-page-app/website",
            },
          ];
        case "traditional_webapp":
          return [
            {
              icon: <Icon iconName="Globe" />,
              name: renderToString(
                "EditOAuthClientScreen.quick-start.framework.traditional-webapp"
              ),
              docLink:
                "https://docs.authgear.com/get-started/single-page-app/website",
            },
          ];
        case "native":
          return [
            {
              icon: <i className={cn("fab", "fa-react")} />,
              name: renderToString(
                "EditOAuthClientScreen.quick-start.framework.react-native"
              ),
              docLink:
                "https://docs.authgear.com/get-started/native-mobile-app/react-native",
            },
            {
              icon: <i className={cn("fab", "fa-apple")} />,
              name: renderToString(
                "EditOAuthClientScreen.quick-start.framework.ios"
              ),
              docLink:
                "https://docs.authgear.com/get-started/native-mobile-app/ios",
            },
            {
              icon: <i className={cn("fab", "fa-android")} />,
              name: renderToString(
                "EditOAuthClientScreen.quick-start.framework.android"
              ),
              docLink:
                "https://docs.authgear.com/get-started/native-mobile-app/android",
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
              docLink:
                "https://docs.authgear.com/get-started/native-mobile-app/flutter",
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
              docLink:
                "https://docs.authgear.com/get-started/native-mobile-app/xamarin",
            },
          ];
        case "confidential":
          return [
            {
              icon: <i className={cn("fab", "fa-openid")} />,
              name: renderToString(
                "EditOAuthClientScreen.quick-start.framework.oidc"
              ),
              docLink: "https://docs.authgear.com/get-started/oidc-provider",
            },
            {
              icon: (
                <img className="w-6.5 h-6.5 object-contain" src={iconSaml} />
              ),
              name: renderToString(
                "EditOAuthClientScreen.quick-start.framework.saml"
              ),
              docLink:
                "https://docs.authgear.com/get-started/single-sign-on-with-saml",
            },
          ];
        case "third_party_app":
          return [
            {
              icon: <i className={cn("fab", "fa-openid")} />,
              name: renderToString(
                "EditOAuthClientScreen.quick-start.framework.oidc"
              ),
              docLink: "https://docs.authgear.com/get-started/oidc-provider",
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
  rawAppConfig: PortalAPIAppConfig;
  clientID: string;
  samlIdpEntityID: string;
  samlIdpSigningCertificates: SAMLIdpSigningCertificate[];
  customUIEnabled: boolean;
  showAppLogoSection: boolean;
  app2appEnabled: boolean;
  onGeneratedNewIdpSigningCertificate: () => void;
  clientSecretHook: ClientSecretsHook;
}

enum FormTab {
  SETTINGS = "settings",
  SAML2 = "saml2",
  API_RESOURCES = "api-resources",
  QUICK_START = "quick-start",
}

interface FormTabContextType {
  formTab: FormTab;
  setFormTab: (tab: FormTab) => void;
}

const FormTabContext = React.createContext<FormTabContextType>(undefined!);

const EditOAuthClientContent: React.VFC<EditOAuthClientContentProps> =
  function EditOAuthClientContent(props) {
    const {
      clientID,
      samlIdpEntityID,
      rawAppConfig,
      samlIdpSigningCertificates,
      form: { state, setState },
      customUIEnabled,
      showAppLogoSection,
      app2appEnabled,
      onGeneratedNewIdpSigningCertificate,
      clientSecretHook,
    } = props;
    const { renderToString } = useContext(Context);

    const { formTab, setFormTab } = useContext(FormTabContext);

    const client =
      state.editedClient ?? state.clients.find((c) => c.client_id === clientID);

    const clientSecret = useMemo(() => {
      return client?.client_id
        ? state.clientSecretMap[client.client_id]
        : undefined;
    }, [client, state.clientSecretMap]);

    const onFormTabChange = useCallback(
      (item?: PivotItem) => {
        if (item == null) {
          return;
        }
        const { itemKey } = item.props;
        setFormTab(itemKey as FormTab);
      },
      [setFormTab]
    );

    const onClientConfigChange = useCallback(
      (editedClient: OAuthClientConfig) => {
        setState((state) => ({ ...state, editedClient }));
      },
      [setState]
    );

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
      <ScreenContent
        className="flex-1-0-auto"
        layout={formTab === FormTab.API_RESOURCES ? "list" : "auto-rows"}
      >
        <header
          className={cn(styles.widget, styles["widget--wide"], "space-y-5")}
        >
          <EditOAuthClientNavBreadcrumb clientName={client.name ?? ""} />
          <AGPivot
            className={styles.widget}
            selectedKey={formTab}
            onLinkClick={onFormTabChange}
          >
            {client.x_application_type === "m2m" ? (
              <PivotItem
                itemKey={FormTab.QUICK_START}
                headerText={renderToString(
                  "EditOAuthClientScreen.tabs.quick-start"
                )}
              />
            ) : null}
            <PivotItem
              itemKey={FormTab.SETTINGS}
              headerText={renderToString("EditOAuthClientScreen.tabs.settings")}
            />
            {client.x_application_type === "confidential" ? (
              <PivotItem
                itemKey={FormTab.SAML2}
                headerText={renderToString("EditOAuthClientScreen.tabs.saml2")}
              />
            ) : null}
            {client.x_application_type === "m2m" ? (
              <PivotItem
                itemKey={FormTab.API_RESOURCES}
                headerText={renderToString(
                  "EditOAuthClientScreen.tabs.api-resources"
                )}
              />
            ) : null}
          </AGPivot>
        </header>
        {formTab === FormTab.QUICK_START ? (
          <EditOAuthClientFormQuickStartContent
            className={cn(styles.widget, styles["widget--wide"])}
            client={client}
            appConfig={rawAppConfig}
            clientSecrets={clientSecret}
          />
        ) : null}
        {formTab === FormTab.SETTINGS ? (
          <OAuthClientSettingsForm
            client={client}
            state={state}
            app2appEnabled={app2appEnabled}
            customUIEnabled={customUIEnabled}
            showAppLogoSection={showAppLogoSection}
            onClientConfigChange={onClientConfigChange}
            clientSecretHook={clientSecretHook}
          />
        ) : null}
        {formTab === FormTab.SAML2 ? (
          <OAuthClientSAML2Content
            clientID={client.client_id}
            samlIdpEntityID={samlIdpEntityID}
            rawAppConfig={rawAppConfig}
            samlIdpSigningCertificates={samlIdpSigningCertificates}
            state={state}
            setState={setState}
            onGeneratedNewIdpSigningCertificate={
              onGeneratedNewIdpSigningCertificate
            }
          />
        ) : null}
        {formTab === FormTab.API_RESOURCES ? (
          <EditOAuthClientFormResourcesContent
            className={cn(styles["widget--wide"])}
            client={client}
          />
        ) : null}
      </ScreenContent>
    );
  };

interface OAuthClientSettingsFormProps {
  client: OAuthClientConfig;
  state: FormState;
  app2appEnabled: boolean;
  customUIEnabled: boolean;
  showAppLogoSection: boolean;
  onClientConfigChange: (newClientConfig: OAuthClientConfig) => void;
  clientSecretHook: ClientSecretsHook;
}

function OAuthClientSettingsForm({
  client,
  state,
  app2appEnabled,
  customUIEnabled,
  showAppLogoSection,
  onClientConfigChange,
  clientSecretHook,
}: OAuthClientSettingsFormProps): React.ReactElement {
  const theme = useTheme();
  const hideQuickStart = useMemo(
    () =>
      (["m2m"] as OAuthClientConfig["x_application_type"][]).includes(
        client.x_application_type
      ),
    [client.x_application_type]
  );
  return (
    <>
      <div
        className={cn(
          styles.widget,
          styles.widgetColumn,
          hideQuickStart ? styles["widget--wide"] : null
        )}
      >
        <EditOAuthClientForm
          publicOrigin={state.publicOrigin}
          clientConfig={client}
          customUIEnabled={customUIEnabled}
          app2appEnabled={app2appEnabled}
          showAppLogoSection={showAppLogoSection}
          onClientConfigChange={onClientConfigChange}
          clientSecretHook={clientSecretHook}
        />
      </div>
      {!hideQuickStart ? (
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
      ) : null}
    </>
  );
}

interface OAuthClientSAML2ContentProps {
  clientID: string;
  rawAppConfig: PortalAPIAppConfig;
  samlIdpEntityID: string;
  samlIdpSigningCertificates: SAMLIdpSigningCertificate[];
  state: FormState;
  setState: (fn: (state: FormState) => FormState) => void;
  onGeneratedNewIdpSigningCertificate: () => void;
}

function OAuthClientSAML2Content({
  clientID,
  rawAppConfig,
  samlIdpEntityID,
  samlIdpSigningCertificates,
  state,
  setState,
  onGeneratedNewIdpSigningCertificate,
}: OAuthClientSAML2ContentProps): React.ReactElement {
  const formState =
    useMemo<OAuthClientSAMLFormState>((): OAuthClientSAMLFormState => {
      const samlConfig = state.samlServiceProviders.find(
        (sp) => sp.clientID === clientID
      );
      const defaults = getDefaultOAuthClientSAMLFormState();
      if (samlConfig == null) {
        return defaults;
      }
      return {
        isSAMLEnabled: samlConfig.isEnabled,
        nameIDFormat: samlConfig.nameIDFormat,
        nameIDAttributePointer:
          samlConfig.nameIDAttributePointer ?? defaults.nameIDAttributePointer,
        acsURLs: samlConfig.acsURLs,
        destination: samlConfig.desitination ?? defaults.destination,
        recipient: samlConfig.recipient ?? defaults.recipient,
        audience: samlConfig.audience ?? defaults.audience,
        assertionValidDurationSeconds:
          samlConfig.assertionValidDurationSeconds ??
          defaults.assertionValidDurationSeconds,
        isSLOEnabled: samlConfig.isSLOEnabled ?? defaults.isSLOEnabled,
        sloCallbackURL: samlConfig.sloCallbackURL ?? defaults.sloCallbackURL,
        sloCallbackBinding:
          samlConfig.sloCallbackBinding ?? defaults.sloCallbackBinding,
        signatureVerificationEnabled:
          samlConfig.signatureVerificationEnabled ??
          defaults.signatureVerificationEnabled,
        signingCertificates:
          samlConfig.certificates ?? defaults.signingCertificates,
        isMetadataUploaded: samlConfig.isMetadataUploaded,
      };
    }, [clientID, state.samlServiceProviders]);

  const onFormStateChange = useCallback(
    (newState: OAuthClientSAMLFormState) => {
      setState((prevState): FormState => {
        const newSAMLConfig: (typeof state.samlServiceProviders)[number] = {
          clientID: clientID,
          isEnabled: newState.isSAMLEnabled,
          nameIDFormat: newState.nameIDFormat,
          nameIDAttributePointer: newState.nameIDAttributePointer,
          acsURLs: newState.acsURLs,
          desitination: newState.destination,
          recipient: newState.recipient,
          audience: newState.audience,
          assertionValidDurationSeconds: newState.assertionValidDurationSeconds,
          isSLOEnabled: newState.isSLOEnabled,
          sloCallbackURL: newState.sloCallbackURL,
          sloCallbackBinding: newState.sloCallbackBinding,
          signatureVerificationEnabled: newState.signatureVerificationEnabled,
          certificates: newState.signingCertificates,
          isMetadataUploaded: newState.isMetadataUploaded,
        };
        const newServiceProviders = [...prevState.samlServiceProviders];
        const existingConfigIndex = state.samlServiceProviders.findIndex(
          (sp) => sp.clientID === clientID
        );
        if (existingConfigIndex === -1) {
          newServiceProviders.push(newSAMLConfig);
        } else {
          newServiceProviders[existingConfigIndex] = newSAMLConfig;
        }
        return {
          ...prevState,
          samlServiceProviders: newServiceProviders,
        };
      });
    },
    [clientID, setState, state]
  );

  const jsonPointer = useMemo(() => {
    const idx = state.samlServiceProviders.findIndex(
      (sp) => sp.clientID === clientID
    );

    return `/saml/service_providers/${idx}`;
  }, [clientID, state.samlServiceProviders]);

  return (
    <div className={cn(styles.widget)}>
      <OAuthClientSAMLForm
        clientID={clientID}
        rawAppConfig={rawAppConfig}
        samlIdpEntityID={samlIdpEntityID}
        samlIdpSigningCertificates={samlIdpSigningCertificates}
        publicOrigin={state.publicOrigin}
        parentJSONPointer={jsonPointer}
        formState={formState}
        onFormStateChange={onFormStateChange}
        onGeneratedNewIdpSigningCertificate={
          onGeneratedNewIdpSigningCertificate
        }
      />
    </div>
  );
}

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
  const {
    isLoading: appQueryLoading,
    loadError: appQueryLoadError,
    samlIdpEntityID,
    secretConfig,
    rawAppConfig,
    refetch: refetchAppAndSecretConfig,
  } = useAppAndSecretConfigQuery(appID, secretToken);
  const generateClientSecretHook = useClientSecret(appID, secretToken);

  const featureConfig = useAppFeatureConfigQuery(appID);

  const [searchParams] = useSearchParams();
  const isQuickScreenVisible = useMemo(() => {
    const quickstart = searchParams.get("quickstart");
    return quickstart === "true";
  }, [searchParams]);

  const customUIEnabled = useMemo(() => {
    if (featureConfig.isLoading) {
      return false;
    }
    return featureConfig.effectiveFeatureConfig?.oauth?.client
      ?.custom_ui_enabled;
  }, [
    featureConfig.isLoading,
    featureConfig.effectiveFeatureConfig?.oauth?.client?.custom_ui_enabled,
  ]);

  const showAppLogoSection = useMemo(() => {
    if (featureConfig.isLoading) {
      return false;
    }
    return featureConfig.effectiveFeatureConfig?.ui
      ?.allow_showing_logo_uri_as_project_logo;
  }, [
    featureConfig.isLoading,
    featureConfig.effectiveFeatureConfig?.ui
      ?.allow_showing_logo_uri_as_project_logo,
  ]);

  const app2appEnabled = useMemo(() => {
    if (featureConfig.isLoading) {
      return false;
    }
    return featureConfig.effectiveFeatureConfig?.oauth?.client?.app2app_enabled;
  }, [featureConfig]);

  const samlIdPSigningCertificates = useMemo(() => {
    return secretConfig?.samlIdpSigningSecrets?.certificates ?? [];
  }, [secretConfig]);

  return useLoadableView({
    loadables: [
      form,
      {
        isLoading: appQueryLoading,
        loadError: appQueryLoadError,
        reload: refetchAppAndSecretConfig,
      },
      generateClientSecretHook,
    ],
    render: () => {
      if (isQuickScreenVisible) {
        return <OAuthQuickStartScreenContent form={form} clientID={clientID} />;
      }

      return (
        <FormContainerContent
          form={form}
          rawAppConfig={rawAppConfig!}
          clientID={clientID}
          samlIdpEntityID={samlIdpEntityID}
          samlIdpSigningCertificates={samlIdPSigningCertificates}
          customUIEnabled={customUIEnabled}
          showAppLogoSection={showAppLogoSection}
          app2appEnabled={app2appEnabled}
          refetchAppAndSecretConfig={refetchAppAndSecretConfig}
          clientSecretHook={generateClientSecretHook}
        />
      );
    },
  });
};

function FormContainerContent({
  form,
  rawAppConfig,
  clientID,
  samlIdpEntityID,
  samlIdpSigningCertificates,
  customUIEnabled,
  showAppLogoSection,
  app2appEnabled,
  refetchAppAndSecretConfig,
  clientSecretHook,
}: {
  form: AppSecretConfigFormModel<FormState>;
  rawAppConfig: PortalAPIAppConfig;
  clientID: string;
  samlIdpEntityID: string | null | undefined;
  samlIdpSigningCertificates: SAMLIdpSigningCertificate[];
  customUIEnabled: boolean;
  showAppLogoSection: boolean;
  app2appEnabled: boolean;
  refetchAppAndSecretConfig: () => void;
  clientSecretHook: ClientSecretsHook;
}) {
  const { state } = form;
  const { selectedKey: formTab, onChangeKey: setFormTab } = usePivotNavigation(
    useMemo(() => {
      const client =
        state.editedClient ??
        state.clients.find((c) => c.client_id === clientID);
      switch (client?.x_application_type) {
        case "m2m":
          return [FormTab.QUICK_START, FormTab.SETTINGS, FormTab.API_RESOURCES];
        case "confidential":
          return [FormTab.SETTINGS, FormTab.SAML2];
        default:
          return [FormTab.SETTINGS];
      }
    }, [clientID, state.clients, state.editedClient]),
    undefined,
    "tab"
  );

  const contextValue = useMemo<FormTabContextType>(
    () => ({
      formTab,
      setFormTab: setFormTab,
    }),
    [formTab, setFormTab]
  );

  const hideFooter = (() => {
    switch (formTab) {
      case FormTab.API_RESOURCES:
        return true;
      case FormTab.QUICK_START:
        return true;
      default:
        return false;
    }
  })();

  return (
    <FormTabContext.Provider value={contextValue}>
      <FormContainer
        className="flex-1-0-auto flex flex-col"
        form={form}
        stickyFooterComponent={true}
        showDiscardButton={true}
        hideFooterComponent={hideFooter}
        localError={clientSecretHook.saveError}
      >
        <EditOAuthClientContent
          form={form}
          rawAppConfig={rawAppConfig}
          clientID={clientID}
          samlIdpEntityID={samlIdpEntityID ?? ""}
          samlIdpSigningCertificates={samlIdpSigningCertificates}
          customUIEnabled={customUIEnabled}
          showAppLogoSection={showAppLogoSection}
          app2appEnabled={app2appEnabled}
          onGeneratedNewIdpSigningCertificate={refetchAppAndSecretConfig}
          clientSecretHook={clientSecretHook}
        />
      </FormContainer>
    </FormTabContext.Provider>
  );
}

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
