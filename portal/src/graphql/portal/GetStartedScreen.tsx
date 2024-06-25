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
  CheckFirstUserQuery,
  CheckFirstUserDocument,
} from "../adminapi/query/checkFirstUserQuery.generated";
import {
  ScreenNavQueryQuery,
  ScreenNavQueryDocument,
} from "./query/screenNavQuery.generated";
import { usePortalClient } from "./apollo";
import { TutorialStatusData, UIImplementation } from "../../types";
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

import iconKey from "../../images/getting-started-icon-key.png";
import iconCustomize from "../../images/getting-started-icon-customize.png";
import iconApplication from "../../images/getting-started-icon-application.png";
import iconSSO from "../../images/getting-started-icon-sso.png";
import iconTeam from "../../images/getting-started-icon-team.png";
import iconTick from "../../images/getting-started-icon-tick.png";
import styles from "./GetStartedScreen.module.css";
import ScreenLayoutScrollView from "../../ScreenLayoutScrollView";
import ActionButton from "../../ActionButton";
import LinkButton from "../../LinkButton";
import { useGenerateTesterTokenMutation } from "./mutations/generateTesterTokenMutation";
import { useCapture } from "../../gtm_v2";

type Progress = keyof TutorialStatusData["progress"];

interface CardSpec {
  key: Progress;
  iconSrc: string;
  canSkip: boolean;
  isDone: boolean;

  internalHref?: string;
  onClick?: (e: React.MouseEvent<HTMLElement>) => void;
}

interface MakeCardSpecsOptions {
  appID: string;
  publicOrigin: string;
  numberOfClients: number;
  tutorialStatusData: TutorialStatusData;
  userTotalCount: number;
  authUIImplementation: UIImplementation;
}

function useCardSpecs(options: MakeCardSpecsOptions): CardSpec[] {
  const {
    appID,
    publicOrigin,
    numberOfClients,
    tutorialStatusData,
    userTotalCount,
    authUIImplementation,
  } = options;

  const { generateTesterToken } = useGenerateTesterTokenMutation(appID);
  const capture = useCapture();
  const onTryAuth = useCallback(async () => {
    const token = await generateTesterToken(window.location.href);
    const destination = new URL(publicOrigin);
    destination.pathname = "/tester";
    destination.search = new URLSearchParams({ token }).toString();
    window.location.assign(destination);
  }, [generateTesterToken, publicOrigin]);

  const authui: CardSpec = useMemo(
    () => ({
      key: "authui",
      iconSrc: iconKey,
      canSkip: false,
      isDone: userTotalCount > 0,
      onClick: (e) => {
        e.preventDefault();
        e.stopPropagation();

        capture("getStarted.clicked-signup");

        onTryAuth().catch((e) => console.error(e));
      },
    }),
    [userTotalCount, onTryAuth, capture]
  );

  const customize_ui: CardSpec = useMemo(
    () => ({
      key: "customize_ui",
      iconSrc: iconCustomize,
      internalHref:
        authUIImplementation === "authflowv2"
          ? "~/branding/design"
          : "~/branding/ui-settings",
      onClick: (_e) => {
        capture("getStarted.clicked-customize");
      },
      canSkip: false,
      isDone: tutorialStatusData.progress["customize_ui"] === true,
    }),
    [authUIImplementation, tutorialStatusData.progress, capture]
  );

  // Special handling for apps with applications.
  // https://github.com/authgear/authgear-server/issues/1976
  const create_application: CardSpec = useMemo(() => {
    const spec: CardSpec = {
      key: "create_application",
      iconSrc: iconApplication,
      internalHref: undefined,
      canSkip: false,
      isDone:
        numberOfClients > 0 ||
        tutorialStatusData.progress["create_application"] === true,
      onClick: (_e) => {
        capture("getStarted.clicked-create_app");
      },
    };
    spec.internalHref = spec.isDone
      ? "~/configuration/apps"
      : "~/configuration/apps/add";
    return spec;
  }, [numberOfClients, tutorialStatusData.progress, capture]);

  const sso: CardSpec = useMemo(
    () => ({
      key: "sso",
      iconSrc: iconSSO,
      internalHref: "~/configuration/authentication/external-oauth",
      canSkip: true,
      isDone: tutorialStatusData.progress["sso"] === true,
      onClick: (_e) => {
        capture("getStarted.clicked-social_login");
      },
    }),
    [tutorialStatusData.progress, capture]
  );

  const invite: CardSpec = useMemo(
    () => ({
      key: "invite",
      iconSrc: iconTeam,
      internalHref: "~/portal-admins/invite",
      canSkip: false,
      isDone: tutorialStatusData.progress["invite"] === true,
      onClick: (_e) => {
        capture("getStarted.clicked-add_members");
      },
    }),
    [tutorialStatusData.progress, capture]
  );

  return useMemo(
    () => [authui, customize_ui, create_application, sso, invite],
    [authui, create_application, customize_ui, invite, sso]
  );
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

  internalHref?: string;
  onClick?: (e: React.MouseEvent<HTMLElement>) => void;
}

function Card(props: CardProps) {
  const {
    cardKey,
    isDone,
    iconSrc,
    skipProgress,
    skipDisabled,
    onClick,
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
          onClick={onClick}
          className={styles.cardActionButton}
        >
          <FormattedMessage
            id={"GetStartedScreen.card.action-label." + cardKey}
          />
          {" >"}
        </Link>
      ) : (
        <LinkButton
          id={id}
          className={styles.cardActionButton}
          onClick={onClick}
          target="_blank"
        >
          <FormattedMessage
            id={"GetStartedScreen.card.action-label." + cardKey}
          />
          {" >"}
        </LinkButton>
      )}
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
            onClick={card.onClick}
            internalHref={card.internalHref}
          />
        );
      })}
    </div>
  );
}

function HelpText() {
  const capture = useCapture();
  const onClickForum = useCallback(() => {
    capture("getStarted.clicked-forum");
  }, [capture]);
  const onClickContactUs = useCallback(() => {
    capture("getStarted.clicked-contact_us");
  }, [capture]);

  return (
    <Text block={true}>
      <FontIcon className={styles.helpTextIcon} iconName="Lifesaver" />
      <FormattedMessage
        id="GetStartedScreen.help-text"
        values={{
          onClickForum,
          onClickContactUs,
        }}
      />
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
  publicOrigin: string;
  numberOfClients: number;
  tutorialStatusData: TutorialStatusData;
  userTotalCount: number;
  authUIImplementation: UIImplementation;
}

function GetStartedScreenContent(props: GetStartedScreenContentProps) {
  const { appID } = useParams() as { appID: string };
  const navigate = useNavigate();
  const capture = useCapture();

  const {
    loading: propLoading,
    publicOrigin,
    numberOfClients,
    tutorialStatusData,
    userTotalCount,
    authUIImplementation,
  } = props;

  const client = usePortalClient();

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

      capture("getStarted.clicked-done");

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
    [appID, skipAppTutorialMutationFunction, navigate, capture]
  );

  const cardSpecs = useCardSpecs({
    appID,
    publicOrigin,
    numberOfClients,
    tutorialStatusData,
    userTotalCount,
    authUIImplementation,
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

// eslint-disable-next-line complexity
export default function GetStartedScreen(): React.ReactElement {
  const { appID } = useParams() as { appID: string };
  const client = usePortalClient();

  const {
    effectiveAppConfig,
    loading: appConfigLoading,
    error,
    refetch,
  } = useAppAndSecretConfigQuery(appID);

  const queryResult0 = useQuery<CheckFirstUserQuery>(CheckFirstUserDocument, {
    // Refresh each time this screen is visited.
    fetchPolicy: "network-only",
  });

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

  const loading =
    queryResult0.loading || queryResult.loading || appConfigLoading;

  if (loading || !tutorialStatusData || !effectiveAppConfig) {
    return <ShowLoading />;
  }

  if (error) {
    return <ShowError error={error} onRetry={refetch} />;
  }

  return (
    <GetStartedScreenContent
      loading={loading}
      publicOrigin={effectiveAppConfig.http?.public_origin ?? ""}
      numberOfClients={effectiveAppConfig.oauth?.clients?.length ?? 0}
      authUIImplementation={
        effectiveAppConfig.ui?.implementation ?? "interaction"
      }
      tutorialStatusData={tutorialStatusData}
      userTotalCount={queryResult0.data?.users?.totalCount ?? 0}
    />
  );
}
