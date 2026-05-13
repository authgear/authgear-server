import React, { useEffect, useMemo } from "react";
import { useParams } from "react-router-dom";
import { FormattedMessage } from "../../../intl";
import { Text } from "@fluentui/react";
import ShowLoading from "../../../ShowLoading";
import ShowError from "../../../ShowError";
import ScreenLayoutScrollView from "../../../ScreenLayoutScrollView";
import { useCapture } from "../../../gtm_v2";
import { useAppAndSecretConfigQuery } from "../query/appAndSecretConfigQuery";

import HeroLoginCard from "./HeroLoginCard";
import HeroIntegrateCard from "./HeroIntegrateCard";
import FeatureCard from "./FeatureCard";
import ResourceColumn, { ResourceRowProps } from "./ResourceColumn";

import {
  BotMessageSquare,
  Code,
  FileJson2,
  LockOpen,
  Mail,
  MessageCircle,
  BookOpen,
  Rocket,
  Headphones,
  Settings,
  UserCog,
  Webhook,
} from "lucide-react";

import styles from "./GetStartedScreen.module.css";

const DOCS_HOME = "https://docs.authgear.com/";
const DOCS_QUICKSTART = "https://docs.authgear.com/get-started/start-building";
const DOCS_API_REFERENCE = "https://docs.authgear.com/reference/apis";
const DOCS_CUSTOM_UI = "https://docs.authgear.com/customization/custom-ui";
const MAILTO_SUPPORT = "mailto:support@authgear.com";
const URL_SALES = "https://www.authgear.com/schedule-demo";
const URL_DISCORD = "https://discord.gg/Kdn5vcYwAS";

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
  const hasApp = numberOfClients > 0;
  const captureData = useMemo(() => ({ has_app: hasApp }), [hasApp]);

  useEffect(() => {
    capture("getStarted.viewed", captureData);
  }, [capture, captureData]);

  const featureCards = useMemo(
    () => [
      {
        Icon: LockOpen,
        titleMessageID: "GetStartedScreen.feature.2fa.title",
        descriptionMessageID: "GetStartedScreen.feature.2fa.description",
        actionMessageID: "GetStartedScreen.feature.2fa.action",
        internalHref: `/project/${appID}/configuration/authentication/2fa`,
        onClick: () => capture("getStarted.clicked-feature_2fa", captureData),
      },
      {
        Icon: BotMessageSquare,
        titleMessageID: "GetStartedScreen.feature.bot-protection.title",
        descriptionMessageID:
          "GetStartedScreen.feature.bot-protection.description",
        actionMessageID: "GetStartedScreen.feature.bot-protection.action",
        internalHref: `/project/${appID}/attack-protection/bot-protection`,
        onClick: () =>
          capture("getStarted.clicked-feature_bot_protection", captureData),
      },
      {
        Icon: UserCog,
        titleMessageID: "GetStartedScreen.feature.user-management.title",
        descriptionMessageID:
          "GetStartedScreen.feature.user-management.description",
        actionMessageID: "GetStartedScreen.feature.user-management.action",
        internalHref: `/project/${appID}/users`,
        onClick: () =>
          capture("getStarted.clicked-feature_user_management", captureData),
      },
      {
        Icon: Settings,
        titleMessageID: "GetStartedScreen.feature.admin-api.title",
        descriptionMessageID: "GetStartedScreen.feature.admin-api.description",
        actionMessageID: "GetStartedScreen.feature.admin-api.action",
        internalHref: `/project/${appID}/advanced/admin-api`,
        onClick: () =>
          capture("getStarted.clicked-feature_admin_api", captureData),
      },
      {
        Icon: FileJson2,
        titleMessageID: "GetStartedScreen.feature.custom-ui.title",
        descriptionMessageID: "GetStartedScreen.feature.custom-ui.description",
        actionMessageID: "GetStartedScreen.feature.custom-ui.action",
        externalHref: DOCS_CUSTOM_UI,
        onClick: () =>
          capture("getStarted.clicked-feature_custom_ui", captureData),
      },
      {
        Icon: Webhook,
        titleMessageID: "GetStartedScreen.feature.hooks.title",
        descriptionMessageID: "GetStartedScreen.feature.hooks.description",
        actionMessageID: "GetStartedScreen.feature.hooks.action",
        internalHref: `/project/${appID}/advanced/hooks`,
        onClick: () => capture("getStarted.clicked-feature_hooks", captureData),
      },
    ],
    [appID, capture, captureData]
  );

  const contactRows: ResourceRowProps[] = useMemo(
    () => [
      {
        Icon: MessageCircle,
        titleMessageID: "GetStartedScreen.get-in-touch.discord.title",
        descriptionMessageID:
          "GetStartedScreen.get-in-touch.discord.description",
        externalHref: URL_DISCORD,
        onClick: () => capture("getStarted.clicked-discord", captureData),
      },
      {
        Icon: Mail,
        titleMessageID: "GetStartedScreen.get-in-touch.email.title",
        descriptionMessageID: "GetStartedScreen.get-in-touch.email.description",
        externalHref: MAILTO_SUPPORT,
        onClick: () => capture("getStarted.clicked-email", captureData),
      },
      {
        Icon: Headphones,
        titleMessageID: "GetStartedScreen.get-in-touch.sales.title",
        descriptionMessageID: "GetStartedScreen.get-in-touch.sales.description",
        externalHref: URL_SALES,
        onClick: () => capture("getStarted.clicked-sales", captureData),
      },
    ],
    [capture, captureData]
  );

  const resourceRows: ResourceRowProps[] = useMemo(
    () => [
      {
        Icon: BookOpen,
        titleMessageID: "GetStartedScreen.resource.documentation.title",
        descriptionMessageID:
          "GetStartedScreen.resource.documentation.description",
        externalHref: DOCS_HOME,
        onClick: () => capture("getStarted.clicked-docs", captureData),
      },
      {
        Icon: Code,
        titleMessageID: "GetStartedScreen.resource.api-reference.title",
        descriptionMessageID:
          "GetStartedScreen.resource.api-reference.description",
        externalHref: DOCS_API_REFERENCE,
        onClick: () => capture("getStarted.clicked-api_reference", captureData),
      },
      {
        Icon: Rocket,
        titleMessageID: "GetStartedScreen.resource.quickstart.title",
        descriptionMessageID:
          "GetStartedScreen.resource.quickstart.description",
        externalHref: DOCS_QUICKSTART,
        onClick: () => capture("getStarted.clicked-quickstart", captureData),
      },
    ],
    [capture, captureData]
  );

  return (
    <ScreenLayoutScrollView>
      <div className={styles.root}>
        <div className={styles.heroRow}>
          <HeroLoginCard
            appID={appID}
            publicOrigin={publicOrigin}
            hasApp={hasApp}
          />
          <HeroIntegrateCard appID={appID} hasApp={hasApp} />
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

  const { effectiveAppConfig, isLoading, loadError, refetch } =
    useAppAndSecretConfigQuery(appID);

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
