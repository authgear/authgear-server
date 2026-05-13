import React, { useMemo } from "react";
import { useParams } from "react-router-dom";
import { FormattedMessage } from "../../../intl";
import { Text } from "@fluentui/react";
import ScreenTitle from "../../../ScreenTitle";
import ShowLoading from "../../../ShowLoading";
import ShowError from "../../../ShowError";
import ScreenLayoutScrollView from "../../../ScreenLayoutScrollView";
import { useCapture } from "../../../gtm_v2";
import { useAppAndSecretConfigQuery } from "../query/appAndSecretConfigQuery";

import HeroLoginCard from "./HeroLoginCard";
import HeroIntegrateCard from "./HeroIntegrateCard";
import FeatureCard from "./FeatureCard";
import ResourceColumn, { ResourceRowProps } from "./ResourceColumn";

import feature2faIcon from "../../../images/getting-started/feature-2fa.svg";
import featureBotProtectionIcon from "../../../images/getting-started/feature-bot-protection.svg";
import featureUserManagementIcon from "../../../images/getting-started/feature-user-management.svg";
import featureAdminApiIcon from "../../../images/getting-started/feature-admin-api.svg";
import featureCustomUiIcon from "../../../images/getting-started/feature-custom-ui.svg";
import featureHooksIcon from "../../../images/getting-started/feature-hooks.svg";
import resourceDiscordIcon from "../../../images/getting-started/resource-discord.svg";
import resourceEmailIcon from "../../../images/getting-started/resource-email.svg";
import resourceSalesIcon from "../../../images/getting-started/resource-sales.svg";
import resourceDocsIcon from "../../../images/getting-started/resource-docs.svg";
import resourceApiRefIcon from "../../../images/getting-started/resource-api-ref.svg";
import resourceQuickstartIcon from "../../../images/getting-started/resource-quickstart.svg";

import styles from "./GetStartedScreen.module.css";

const DOCS_HOME = "https://docs.authgear.com/";
const DOCS_QUICKSTART = "https://docs.authgear.com/get-started";
const DOCS_API_REFERENCE = "https://docs.authgear.com/reference/apis/admin-api";
const DOCS_CUSTOM_UI = "https://docs.authgear.com/get-started/native-mobile-app";
const MAILTO_HELLO = "mailto:hello@authgear.com";
const URL_SALES =
  "https://www.authgear.com/talk-with-us?utm_source=portal&utm_medium=link&utm_campaign=getting_started";
const URL_DISCORD = "https://discord.gg/authgear";

interface GetStartedScreenContentProps {
  publicOrigin: string;
  numberOfClients: number;
}

function GetStartedScreenContent(
  props: GetStartedScreenContentProps
): React.ReactElement {
  const { publicOrigin, numberOfClients } = props;
  const { appID } = useParams() as { appID: string };
  const capture = useCapture();

  const featureCards = useMemo(
    () => [
      {
        iconSrc: feature2faIcon,
        titleMessageID: "GetStartedScreen.feature.2fa.title",
        descriptionMessageID: "GetStartedScreen.feature.2fa.description",
        actionMessageID: "GetStartedScreen.feature.2fa.action",
        internalHref: `/project/${appID}/configuration/authentication/2fa`,
        onClick: () => capture("getStarted.clicked-feature_2fa"),
      },
      {
        iconSrc: featureBotProtectionIcon,
        titleMessageID: "GetStartedScreen.feature.bot-protection.title",
        descriptionMessageID:
          "GetStartedScreen.feature.bot-protection.description",
        actionMessageID: "GetStartedScreen.feature.bot-protection.action",
        internalHref: `/project/${appID}/attack-protection/bot-protection`,
        onClick: () => capture("getStarted.clicked-feature_bot_protection"),
      },
      {
        iconSrc: featureUserManagementIcon,
        titleMessageID: "GetStartedScreen.feature.user-management.title",
        descriptionMessageID:
          "GetStartedScreen.feature.user-management.description",
        actionMessageID: "GetStartedScreen.feature.user-management.action",
        internalHref: `/project/${appID}/users`,
        onClick: () => capture("getStarted.clicked-feature_user_management"),
      },
      {
        iconSrc: featureAdminApiIcon,
        titleMessageID: "GetStartedScreen.feature.admin-api.title",
        descriptionMessageID: "GetStartedScreen.feature.admin-api.description",
        actionMessageID: "GetStartedScreen.feature.admin-api.action",
        internalHref: `/project/${appID}/advanced/admin-api`,
        onClick: () => capture("getStarted.clicked-feature_admin_api"),
      },
      {
        iconSrc: featureCustomUiIcon,
        titleMessageID: "GetStartedScreen.feature.custom-ui.title",
        descriptionMessageID: "GetStartedScreen.feature.custom-ui.description",
        actionMessageID: "GetStartedScreen.feature.custom-ui.action",
        externalHref: DOCS_CUSTOM_UI,
        onClick: () => capture("getStarted.clicked-feature_custom_ui"),
      },
      {
        iconSrc: featureHooksIcon,
        titleMessageID: "GetStartedScreen.feature.hooks.title",
        descriptionMessageID: "GetStartedScreen.feature.hooks.description",
        actionMessageID: "GetStartedScreen.feature.hooks.action",
        internalHref: `/project/${appID}/advanced/hooks`,
        onClick: () => capture("getStarted.clicked-feature_hooks"),
      },
    ],
    [appID, capture]
  );

  const contactRows: ResourceRowProps[] = useMemo(
    () => [
      {
        iconSrc: resourceDiscordIcon,
        titleMessageID: "GetStartedScreen.get-in-touch.discord.title",
        descriptionMessageID:
          "GetStartedScreen.get-in-touch.discord.description",
        externalHref: URL_DISCORD,
        onClick: () => capture("getStarted.clicked-discord"),
      },
      {
        iconSrc: resourceEmailIcon,
        titleMessageID: "GetStartedScreen.get-in-touch.email.title",
        descriptionMessageID:
          "GetStartedScreen.get-in-touch.email.description",
        externalHref: MAILTO_HELLO,
        onClick: () => capture("getStarted.clicked-email"),
      },
      {
        iconSrc: resourceSalesIcon,
        titleMessageID: "GetStartedScreen.get-in-touch.sales.title",
        descriptionMessageID:
          "GetStartedScreen.get-in-touch.sales.description",
        externalHref: URL_SALES,
        onClick: () => capture("getStarted.clicked-sales"),
      },
    ],
    [capture]
  );

  const resourceRows: ResourceRowProps[] = useMemo(
    () => [
      {
        iconSrc: resourceDocsIcon,
        titleMessageID: "GetStartedScreen.resource.documentation.title",
        descriptionMessageID:
          "GetStartedScreen.resource.documentation.description",
        externalHref: DOCS_HOME,
        onClick: () => capture("getStarted.clicked-docs"),
      },
      {
        iconSrc: resourceApiRefIcon,
        titleMessageID: "GetStartedScreen.resource.api-reference.title",
        descriptionMessageID:
          "GetStartedScreen.resource.api-reference.description",
        externalHref: DOCS_API_REFERENCE,
        onClick: () => capture("getStarted.clicked-api_reference"),
      },
      {
        iconSrc: resourceQuickstartIcon,
        titleMessageID: "GetStartedScreen.resource.quickstart.title",
        descriptionMessageID:
          "GetStartedScreen.resource.quickstart.description",
        externalHref: DOCS_QUICKSTART,
        onClick: () => capture("getStarted.clicked-quickstart"),
      },
    ],
    [capture]
  );

  return (
    <ScreenLayoutScrollView>
      <div className={styles.root}>
        <ScreenTitle>
          <FormattedMessage id="GetStartedScreen.title" />
        </ScreenTitle>

        <div className={styles.heroRow}>
          <HeroLoginCard publicOrigin={publicOrigin} />
          <HeroIntegrateCard numberOfClients={numberOfClients} />
        </div>

        <section className={styles.featureSection}>
          <Text as="h2" block={true} className={styles.sectionHeading}>
            <FormattedMessage id="GetStartedScreen.features.heading" />
          </Text>
          <div className={styles.featureGrid}>
            {featureCards.map((card) => (
              <FeatureCard key={card.titleMessageID} {...card} />
            ))}
          </div>
        </section>

        <section className={styles.bottomRow}>
          <ResourceColumn
            headingMessageID="GetStartedScreen.get-in-touch.heading"
            rows={contactRows}
          />
          <ResourceColumn
            headingMessageID="GetStartedScreen.resource.heading"
            rows={resourceRows}
          />
        </section>
      </div>
    </ScreenLayoutScrollView>
  );
}

export default function GetStartedScreen(): React.ReactElement {
  const { appID } = useParams() as { appID: string };

  const {
    effectiveAppConfig,
    isLoading,
    loadError,
    refetch,
  } = useAppAndSecretConfigQuery(appID);

  if (isLoading || !effectiveAppConfig) {
    return <ShowLoading />;
  }

  if (loadError) {
    return <ShowError error={loadError} onRetry={refetch} />;
  }

  return (
    <GetStartedScreenContent
      publicOrigin={effectiveAppConfig.http?.public_origin ?? ""}
      numberOfClients={effectiveAppConfig.oauth?.clients?.length ?? 0}
    />
  );
}
