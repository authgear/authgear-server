import React, { useCallback } from "react";
import { FormattedMessage } from "../../../intl";
import { Text } from "@fluentui/react";
import { PlayIcon } from "@radix-ui/react-icons";
import Link from "../../../Link";
import { useTester } from "../../../hook/tester";
import { useCapture } from "../../../gtm_v2";
import styles from "./GetStartedScreen.module.css";

interface HeroLoginCardProps {
  appID: string;
  publicOrigin: string;
  hasApp: boolean;
}

export default function HeroLoginCard(
  props: HeroLoginCardProps
): React.ReactElement {
  const { appID, publicOrigin, hasApp } = props;
  const capture = useCapture();
  const { triggerTester, isLoading } = useTester(appID, publicOrigin);

  const onClickPreview = useCallback(
    (e: React.MouseEvent<HTMLElement>) => {
      e.preventDefault();
      e.stopPropagation();
      capture("getStarted.clicked-preview_login_page", { has_app: hasApp });
      triggerTester().catch((err) => {
        console.error(err);
      });
    },
    [capture, hasApp, triggerTester]
  );

  const onClickCustomize = useCallback(() => {
    capture("getStarted.clicked-customize_branding", { has_app: hasApp });
  }, [capture, hasApp]);

  return (
    <div className={`${styles.heroCard} ${styles.heroCardLogin}`}>
      <div className={styles.heroCardBody}>
        <Text block={true} className={styles.heroBadge}>
          <FormattedMessage id="GetStartedScreen.hero.login.badge" />
        </Text>
        <Text as="h2" block={true} className={styles.heroTitle}>
          <FormattedMessage id="GetStartedScreen.hero.login.title" />
        </Text>
        <Text block={true} className={styles.heroSubtitle}>
          <FormattedMessage id="GetStartedScreen.hero.login.subtitle" />
        </Text>
        <div className={styles.heroButtonRow}>
          <button
            type="button"
            className={styles.heroButtonWhite}
            onClick={onClickPreview}
            disabled={isLoading}
          >
            <PlayIcon width={14} height={14} aria-hidden={true} />
            <FormattedMessage id="GetStartedScreen.hero.login.preview-button" />
          </button>
          <Link
            to={`/project/${appID}/branding/design`}
            onClick={onClickCustomize}
            className={styles.heroButtonOutline}
          >
            <FormattedMessage id="GetStartedScreen.hero.login.customize-button" />
          </Link>
        </div>
      </div>
    </div>
  );
}
