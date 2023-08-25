import React, { useCallback, useMemo } from "react";
import { Text, FontIcon, IButtonProps, Image, ImageFit } from "@fluentui/react";
import { useNavigate, useParams } from "react-router-dom";
import { FormattedMessage } from "@oursky/react-messageformat";
import { useQuery, useMutation } from "@apollo/client";
import Link from "../../Link";
import ScreenTitle from "../../ScreenTitle";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import { useSystemConfig } from "../../context/SystemConfigContext";
import { useAppAndSecretConfigQuery } from "./query/appAndSecretConfigQuery";
import {
  ScreenNavQueryQuery,
  ScreenNavQueryDocument,
} from "./query/screenNavQuery.generated";
import { client } from "./apollo";
import { TutorialStatusData } from "../../types";
import {
  SkipAppTutorialMutationMutation,
  SkipAppTutorialMutationMutationVariables,
  SkipAppTutorialMutationDocument,
} from "./mutations/skipAppTutorialMutation.generated";
import {
  SkipAppTutorialProgressMutationMutation,
  SkipAppTutorialProgressMutationMutationVariables,
  SkipAppTutorialProgressMutationDocument,
} from "./mutations/skipAppTutorialProgressMutation.generated";

// import iconKey from "../../images/getting-started-icon-key.png";
import iconCustomize from "../../images/getting-started-icon-customize.png";
import iconApplication from "../../images/getting-started-icon-application.png";
import iconSSO from "../../images/getting-started-icon-sso.png";
import iconTeam from "../../images/getting-started-icon-team.png";
import iconTick from "../../images/getting-started-icon-tick.png";
import styles from "./GetStartedScreen.module.css";
import {
  AuthgearGTMEventType,
  useMakeAuthgearGTMEventDataAttributes,
} from "../../GTMProvider";
import ScreenLayoutScrollView from "../../ScreenLayoutScrollView";
import ActionButton from "../../ActionButton";
import ExternalLink from "../../ExternalLink";
import LinkButton from "../../LinkButton";

type Progress = keyof TutorialStatusData["progress"];

interface CardSpec {
  key: Progress;
  iconSrc: string;
  internalHref: string | undefined;
  externalHref: string | undefined;
  canSkip: boolean;
  isDone: boolean;
}

interface MakeCardSpecsOptions {
  publicOrigin?: string;
  numberOfClients: number;
  tutorialStatusData: TutorialStatusData;
}

function makeCardSpecs(options: MakeCardSpecsOptions): CardSpec[] {
  const { numberOfClients, tutorialStatusData } = options;

  // This is disabled in https://github.com/authgear/authgear-server/issues/3319
  // const authui: CardSpec = {
  //   key: "authui",
  //   iconSrc: iconKey,
  //   internalHref: undefined,
  //   externalHref:
  //     publicOrigin != null ? `${publicOrigin}?x_tutorial=true` : undefined,
  //   canSkip: false,
  //   isDone: tutorialStatusData.progress["authui"] === true,
  // };

  const customize_ui: CardSpec = {
    key: "customize_ui",
    iconSrc: iconCustomize,
    internalHref: "~/configuration/ui-settings",
    externalHref: undefined,
    canSkip: false,
    isDone: tutorialStatusData.progress["customize_ui"] === true,
  };

  // Special handling for apps with applications.
  // https://github.com/authgear/authgear-server/issues/1976
  const create_application: CardSpec = {
    key: "create_application",
    iconSrc: iconApplication,
    internalHref: undefined,
    externalHref: undefined,
    canSkip: false,
    isDone:
      numberOfClients > 0 ||
      tutorialStatusData.progress["create_application"] === true,
  };
  create_application.internalHref = create_application.isDone
    ? "~/configuration/apps"
    : "~/configuration/apps/add";

  const sso: CardSpec = {
    key: "sso",
    iconSrc: iconSSO,
    internalHref: "~/configuration/authentication/external-oauth",
    externalHref: undefined,
    canSkip: true,
    isDone: tutorialStatusData.progress["sso"] === true,
  };

  const invite: CardSpec = {
    key: "invite",
    iconSrc: iconTeam,
    internalHref: "~/portal-admins/invite",
    externalHref: undefined,
    canSkip: false,
    isDone: tutorialStatusData.progress["invite"] === true,
  };

  return [/* authui, */ customize_ui, create_application, sso, invite];
}

function Title() {
  return (
    <ScreenTitle>
      <FormattedMessage id="GetStartedScreen.title" />
    </ScreenTitle>
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
  cardSpecs: CardSpec[];
}

function Counter(props: CounterProps) {
  const { cardSpecs } = props;
  const total = cardSpecs.length;
  const done = cardSpecs.reduce(
    (count, card) => count + (card.isDone ? 1 : 0),
    0
  );
  const remaining = total - done;
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
  cardKey: Progress;
  isDone: boolean;
  iconSrc: string;
  skipDisabled: boolean;
  skipProgress?: (progress: Progress) => Promise<void>;
  externalHref?: string;
  internalHref?: string;
}

function Card(props: CardProps) {
  const {
    cardKey,
    isDone,
    iconSrc,
    skipProgress,
    skipDisabled,
    externalHref,
    internalHref,
  } = props;
  const {
    themes: {
      main: {
        palette: { themePrimary },
      },
    },
  } = useSystemConfig();
  const { appID } = useParams() as { appID: string };
  const id = "GetStartedScreen.card." + cardKey;
  const onClickCard = useCallback(
    (e) => {
      const target = e.target;
      // Do not intercept clicks on links.
      if (target instanceof HTMLAnchorElement) {
        return;
      }
      // Do not intercept clicks on buttons.
      if (target instanceof HTMLButtonElement) {
        return;
      }

      // Otherwise clicking the card is the same as clicking the action button.
      const actionButton = document.getElementById(id);
      e.preventDefault();
      e.stopPropagation();
      actionButton?.click();
    },
    [id]
  );

  const onClickSkip = useCallback(
    (e) => {
      e.preventDefault();
      e.stopPropagation();

      skipProgress?.(cardKey).then(
        () => {},
        () => {}
      );
    },
    [skipProgress, cardKey]
  );

  const makeGTMEventDataAttributes = useMakeAuthgearGTMEventDataAttributes();
  const eventDataAttributes = useMemo(() => {
    return makeGTMEventDataAttributes({
      event: AuthgearGTMEventType.ClickedGetStarted,
      eventDataAttributes: {
        "get-started-type": cardKey,
      },
    });
  }, [makeGTMEventDataAttributes, cardKey]);

  return (
    <div
      className={styles.card}
      style={{
        // @ts-expect-error
        "--hover-color": themePrimary,
      }}
      role="button"
      onClick={onClickCard}
    >
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
      {internalHref != null ? (
        <Link
          id={id}
          to={internalHref.replace("~/", `/project/${appID}/`)}
          className={styles.cardActionButton}
          {...eventDataAttributes}
        >
          <FormattedMessage
            id={"GetStartedScreen.card.action-label." + cardKey}
          />
          {" >"}
        </Link>
      ) : null}
      {externalHref != null ? (
        <ExternalLink
          id={id}
          className={styles.cardActionButton}
          href={externalHref}
          target="_blank"
          {...eventDataAttributes}
        >
          <FormattedMessage
            id={"GetStartedScreen.card.action-label." + cardKey}
          />
          {" >"}
        </ExternalLink>
      ) : null}
      {skipProgress != null && !isDone ? (
        <LinkButton
          className={styles.cardSkipButton}
          onClick={onClickSkip}
          disabled={skipDisabled}
        >
          <FormattedMessage id="GetStartedScreen.card.skip-button.label" />
          {" >"}
        </LinkButton>
      ) : null}
    </div>
  );
}

interface CardsProps {
  cardSpecs: CardSpec[];
  skipProgress: (progress: Progress) => Promise<void>;
  loading: boolean;
}

function Cards(props: CardsProps) {
  const { cardSpecs, skipProgress, loading } = props;

  return (
    <div className={styles.cards}>
      {cardSpecs.map((card) => {
        return (
          <Card
            key={card.key}
            cardKey={card.key}
            isDone={card.isDone}
            skipProgress={card.canSkip ? skipProgress : undefined}
            skipDisabled={loading}
            iconSrc={card.iconSrc}
            externalHref={card.externalHref}
            internalHref={card.internalHref}
          />
        );
      })}
    </div>
  );
}

function HelpText() {
  return (
    <Text block={true}>
      <FontIcon className={styles.helpTextIcon} iconName="Lifesaver" />
      <FormattedMessage id="GetStartedScreen.help-text" />
    </Text>
  );
}

interface DismissButtonProps {
  onClick?: IButtonProps["onClick"];
  disabled?: IButtonProps["disabled"];
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
      text={<FormattedMessage id="GetStartedScreen.dismiss-button.label" />}
    />
  );
}

interface GetStartedScreenContentProps {
  loading: boolean;
  publicOrigin?: string;
  numberOfClients: number;
  tutorialStatusData: TutorialStatusData;
}

function GetStartedScreenContent(props: GetStartedScreenContentProps) {
  const { appID } = useParams() as { appID: string };
  const navigate = useNavigate();

  const {
    loading: propLoading,
    publicOrigin,
    numberOfClients,
    tutorialStatusData,
  } = props;

  const [
    skipAppTutorialMutationFunction,
    { loading: skipAppTutorialMutationLoading },
  ] = useMutation<
    SkipAppTutorialMutationMutation,
    SkipAppTutorialMutationMutationVariables
  >(SkipAppTutorialMutationDocument, {
    client,
    refetchQueries: [ScreenNavQueryDocument],
  });

  const loading = propLoading || skipAppTutorialMutationLoading;

  const [
    skipAppTutorialProgressMutationFuction,
    { loading: skipAppTutorialProgressMutationLoading },
  ] = useMutation<
    SkipAppTutorialProgressMutationMutation,
    SkipAppTutorialProgressMutationMutationVariables
  >(SkipAppTutorialProgressMutationDocument, {
    client,
    refetchQueries: [ScreenNavQueryDocument],
  });

  const skipProgress = useCallback(
    async (progress: Progress) => {
      await skipAppTutorialProgressMutationFuction({
        variables: {
          appID,
          progress,
        },
      });
    },
    [appID, skipAppTutorialProgressMutationFuction]
  );

  const onClickDismissButton = useCallback(
    (e) => {
      e.preventDefault();
      e.stopPropagation();

      skipAppTutorialMutationFunction({
        variables: {
          appID,
        },
      }).then(
        () => {
          navigate("./..");
        },
        () => {}
      );
    },
    [appID, skipAppTutorialMutationFunction, navigate]
  );

  const cardSpecs = makeCardSpecs({
    publicOrigin,
    numberOfClients,
    tutorialStatusData,
  });

  return (
    <ScreenLayoutScrollView>
      <div className={styles.root}>
        <Title />
        <div className={styles.descriptionRow}>
          <Description />
          <Counter cardSpecs={cardSpecs} />
        </div>
        <Cards
          cardSpecs={cardSpecs}
          skipProgress={skipProgress}
          loading={skipAppTutorialProgressMutationLoading}
        />
        <HelpText />
        <DismissButton onClick={onClickDismissButton} disabled={loading} />
      </div>
    </ScreenLayoutScrollView>
  );
}

export default function GetStartedScreen(): React.ReactElement {
  const { appID } = useParams() as { appID: string };

  const {
    effectiveAppConfig,
    loading: appConfigLoading,
    error,
    refetch,
  } = useAppAndSecretConfigQuery(appID);

  const queryResult = useQuery<ScreenNavQueryQuery>(ScreenNavQueryDocument, {
    client,
    variables: {
      id: appID,
    },
    // Refresh each time this screen is visited.
    fetchPolicy: "network-only",
  });

  const app =
    queryResult.data?.node?.__typename === "App" ? queryResult.data.node : null;

  const tutorialStatusData = app?.tutorialStatus.data;

  const loading = queryResult.loading || appConfigLoading;

  if (loading || !tutorialStatusData || !effectiveAppConfig) {
    return <ShowLoading />;
  }

  if (error) {
    return <ShowError error={error} onRetry={refetch} />;
  }

  return (
    <GetStartedScreenContent
      loading={loading}
      publicOrigin={effectiveAppConfig.http?.public_origin}
      numberOfClients={effectiveAppConfig.oauth?.clients?.length ?? 0}
      tutorialStatusData={tutorialStatusData}
    />
  );
}
