import React, { useCallback } from "react";
import {
  Text,
  FontIcon,
  ActionButton,
  IButtonProps,
  Image,
  ImageFit,
  Link,
} from "@fluentui/react";
import { useParams } from "react-router-dom";
import { FormattedMessage } from "@oursky/react-messageformat";
import ScreenContent from "../../ScreenContent";
import ReactRouterLink from "../../ReactRouterLink";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import { useSystemConfig } from "../../context/SystemConfigContext";
import { useAppAndSecretConfigQuery } from "./query/appAndSecretConfigQuery";

import iconKey from "../../images/getting-started-icon-key.png";
import iconCustomize from "../../images/getting-started-icon-customize.png";
import iconApplication from "../../images/getting-started-icon-application.png";
import iconSSO from "../../images/getting-started-icon-sso.png";
import iconTeam from "../../images/getting-started-icon-team.png";
import iconTick from "../../images/getting-started-icon-tick.png";
import styles from "./GetStartedScreen.module.scss";

const cards = [
  {
    key: "authui",
    iconSrc: iconKey,
  },
  {
    key: "customize_ui",
    iconSrc: iconCustomize,
    internalHref: "../configuration/ui-settings",
  },
  {
    key: "create_application",
    iconSrc: iconApplication,
    internalHref: "../configuration/apps/add",
  },
  {
    key: "sso",
    iconSrc: iconSSO,
    internalHref: "../configuration/single-sign-on",
  },
  {
    key: "invite",
    iconSrc: iconTeam,
    internalHref: "../portal-admins/invite",
  },
];

function Title() {
  return (
    <Text as="h1" variant="large" block={true} className={styles.title}>
      <FormattedMessage id="GetStartedScreen.title" />
    </Text>
  );
}

function Description() {
  return (
    <Text block={true} className={styles.description}>
      <FormattedMessage id="GetStartedScreen.description" />
    </Text>
  );
}

interface CounterProps {
  remaining: number;
}

function Counter(props: CounterProps) {
  const { remaining } = props;
  return (
    <Text block={true} className={styles.counter}>
      <FormattedMessage
        id="GetStartedScreen.counter"
        values={{
          remaining,
        }}
      />
    </Text>
  );
}

interface CardProps {
  cardKey: string;
  isDone: boolean;
  iconSrc: string;
  skipEnabled: boolean;
  externalHref?: string;
  internalHref?: string;
}

function Card(props: CardProps) {
  const { cardKey, isDone, iconSrc, skipEnabled, externalHref, internalHref } =
    props;
  const id = "GetStartedScreen.card." + cardKey;
  const onClickCard = useCallback(
    (e) => {
      const target = e.target;
      const actionButton = document.getElementById(id);
      if (target === actionButton) {
        // The element being clicked is the action button.
        // Let the event does its default and propagate.
        return;
      }

      // Clicking the card is the same as clicking the action button.
      e.preventDefault();
      e.stopPropagation();
      actionButton?.click();
    },
    [id]
  );
  const onClickSkip = useCallback((e) => {
    e.preventDefault();
    e.stopPropagation();
  }, []);
  return (
    <div className={styles.card} role="button" onClick={onClickCard}>
      <Image
        className={styles.cardIcon}
        src={isDone ? iconTick : iconSrc}
        imageFit={ImageFit.cover}
      />
      <Text className={styles.cardTitle} variant="mediumPlus">
        <FormattedMessage id={"GetStartedScreen.card.title." + cardKey} />
      </Text>
      <Text className={styles.cardDescription}>
        <FormattedMessage id={"GetStartedScreen.card.description." + cardKey} />
      </Text>
      {internalHref != null && (
        <ReactRouterLink
          id={id}
          to={internalHref}
          component={Link}
          className={styles.cardActionButton}
        >
          <FormattedMessage
            id={"GetStartedScreen.card.action-label." + cardKey}
          />
          {" >"}
        </ReactRouterLink>
      )}
      {externalHref != null && (
        <Link
          id={id}
          className={styles.cardActionButton}
          href={externalHref}
          target="_blank"
        >
          <FormattedMessage
            id={"GetStartedScreen.card.action-label." + cardKey}
          />
          {" >"}
        </Link>
      )}
      {skipEnabled && (
        <Link
          className={styles.cardSkipButton}
          as="button"
          onClick={onClickSkip}
        >
          <FormattedMessage id="GetStartedScreen.card.skip-button.label" />
          {" >"}
        </Link>
      )}
    </div>
  );
}

interface CardsProps {
  publicOrigin?: string;
}

function Cards(props: CardsProps) {
  const { publicOrigin } = props;
  return (
    <div className={styles.cards}>
      {cards.map((card) => {
        return (
          <Card
            key={card.key}
            cardKey={card.key}
            isDone={false}
            skipEnabled={card.key === "sso"}
            iconSrc={card.iconSrc}
            externalHref={
              card.internalHref == null && publicOrigin != null
                ? `${publicOrigin}?x_tutorial=true`
                : undefined
            }
            internalHref={card.internalHref}
          />
        );
      })}
    </div>
  );
}

function HelpText() {
  return (
    <Text block={true} className={styles.helpText}>
      <FontIcon iconName="Lifesaver" />
      <FormattedMessage id="GetStartedScreen.help-text" />
    </Text>
  );
}

interface DismissButtonProps {
  onClick?: IButtonProps["onClick"];
}

function DismissButton(props: DismissButtonProps) {
  const { onClick } = props;
  const {
    themes: { main },
  } = useSystemConfig();
  return (
    <ActionButton
      className={styles.dismissButton}
      styles={{
        root: {
          color: main.semanticColors.bodySubtext,
        },
      }}
      onClick={onClick}
    >
      <FormattedMessage id="GetStartedScreen.dismiss-button.label" />
    </ActionButton>
  );
}

export default function GetStartedScreen(): React.ReactElement {
  const { appID } = useParams();

  const { effectiveAppConfig, loading, error, refetch } =
    useAppAndSecretConfigQuery(appID);

  if (loading || !effectiveAppConfig) {
    return <ShowLoading />;
  }

  if (error) {
    return <ShowError error={error} onRetry={refetch} />;
  }

  return (
    <ScreenContent className={styles.root}>
      <Title />
      <Description />
      <Counter remaining={5} />
      <Cards publicOrigin={effectiveAppConfig.http?.public_origin} />
      <HelpText />
      <DismissButton />
    </ScreenContent>
  );
}
