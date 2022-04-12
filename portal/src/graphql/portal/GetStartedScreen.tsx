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
import { useNavigate, useParams } from "react-router-dom";
import { FormattedMessage } from "@oursky/react-messageformat";
import { gql, useQuery, useMutation } from "@apollo/client";
import ScreenContent from "../../ScreenContent";
import ReactRouterLink from "../../ReactRouterLink";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import { useSystemConfig } from "../../context/SystemConfigContext";
import { useAppAndSecretConfigQuery } from "./query/appAndSecretConfigQuery";
import query from "./query/ScreenNavQuery";
import { client } from "./apollo";
import { ScreenNavQuery } from "./query/__generated__/ScreenNavQuery";
import { TutorialStatusData } from "../../types";
import {
  SkipAppTutorialMutation,
  SkipAppTutorialMutationVariables,
} from "./__generated__/SkipAppTutorialMutation";
import {
  SkipAppTutorialProgressMutation,
  SkipAppTutorialProgressMutationVariables,
} from "./__generated__/SkipAppTutorialProgressMutation";

import iconKey from "../../images/getting-started-icon-key.png";
import iconCustomize from "../../images/getting-started-icon-customize.png";
import iconApplication from "../../images/getting-started-icon-application.png";
import iconSSO from "../../images/getting-started-icon-sso.png";
import iconTeam from "../../images/getting-started-icon-team.png";
import iconTick from "../../images/getting-started-icon-tick.png";
import styles from "./GetStartedScreen.module.scss";

const skipAppTutorialMutation = gql`
  mutation SkipAppTutorialMutation($appID: String!) {
    skipAppTutorial(input: { id: $appID }) {
      app {
        id
      }
    }
  }
`;

const skipAppTutorialProgressMutation = gql`
  mutation SkipAppTutorialProgressMutation(
    $appID: String!
    $progress: String!
  ) {
    skipAppTutorialProgress(input: { id: $appID, progress: $progress }) {
      app {
        id
        tutorialStatus {
          data
        }
      }
    }
  }
`;

type Progress = keyof TutorialStatusData["progress"];

interface CardSpec {
  key: Progress;
  iconSrc: string;
  internalHref?: string;
}

const cards: CardSpec[] = [
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
  tutorialStatusData: TutorialStatusData;
}

function Counter(props: CounterProps) {
  const total = 5;
  const { tutorialStatusData } = props;
  let done = 0;
  for (const [, val] of Object.entries(tutorialStatusData.progress)) {
    if (val === true) {
      ++done;
    }
  }
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
      {skipProgress != null && (
        <Link
          className={styles.cardSkipButton}
          as="button"
          onClick={onClickSkip}
          disabled={skipDisabled}
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
  tutorialStatusData: TutorialStatusData;
  skipProgress: (progress: Progress) => Promise<void>;
  skipDisabled: boolean;
}

function Cards(props: CardsProps) {
  const { publicOrigin, tutorialStatusData, skipProgress, skipDisabled } =
    props;
  return (
    <div className={styles.cards}>
      {cards.map((card) => {
        return (
          <Card
            key={card.key}
            cardKey={card.key}
            isDone={tutorialStatusData.progress[card.key] === true}
            skipProgress={card.key === "sso" ? skipProgress : undefined}
            skipDisabled={skipDisabled}
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
    >
      <FormattedMessage id="GetStartedScreen.dismiss-button.label" />
    </ActionButton>
  );
}

export default function GetStartedScreen(): React.ReactElement {
  const { appID } = useParams();
  const navigate = useNavigate();

  const {
    effectiveAppConfig,
    loading: appConfigLoading,
    error,
    refetch,
  } = useAppAndSecretConfigQuery(appID);

  const queryResult = useQuery<ScreenNavQuery>(query, {
    client,
    variables: {
      id: appID,
    },
    // Refresh each time this screen is visited.
    fetchPolicy: "network-only",
  });

  const [
    skipAppTutorialMutationFunction,
    { loading: skipAppTutorialMutationLoading },
  ] = useMutation<SkipAppTutorialMutation, SkipAppTutorialMutationVariables>(
    skipAppTutorialMutation,
    {
      client,
      refetchQueries: [{ query }],
    }
  );

  const [
    skipAppTutorialProgressMutationFuction,
    { loading: skipAppTutorialProgressMutationLoading },
  ] = useMutation<
    SkipAppTutorialProgressMutation,
    SkipAppTutorialProgressMutationVariables
  >(skipAppTutorialProgressMutation, {
    client,
    refetchQueries: [{ query }],
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

  const app =
    queryResult.data?.node?.__typename === "App" ? queryResult.data.node : null;

  const tutorialStatusData = app?.tutorialStatus.data;

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
          navigate("../");
        },
        () => {}
      );
    },
    [appID, skipAppTutorialMutationFunction, navigate]
  );

  const loading =
    queryResult.loading || appConfigLoading || skipAppTutorialMutationLoading;

  if (loading || !tutorialStatusData || !effectiveAppConfig) {
    return <ShowLoading />;
  }

  if (error) {
    return <ShowError error={error} onRetry={refetch} />;
  }

  return (
    <ScreenContent className={styles.root}>
      <Title />
      <Description />
      <Counter tutorialStatusData={tutorialStatusData} />
      <Cards
        publicOrigin={effectiveAppConfig.http?.public_origin}
        tutorialStatusData={tutorialStatusData}
        skipProgress={skipProgress}
        skipDisabled={skipAppTutorialProgressMutationLoading}
      />
      <HelpText />
      <DismissButton onClick={onClickDismissButton} disabled={loading} />
    </ScreenContent>
  );
}
