import React, { useCallback, useMemo } from "react";
import { FormattedMessage } from "../../../intl";
import { Text } from "@fluentui/react";
import Link from "../../../Link";
import { useCapture } from "../../../gtm_v2";
import styles from "./GetStartedScreen.module.css";

interface HeroIntegrateCardProps {
  appID: string;
  numberOfClients: number;
}

export default function HeroIntegrateCard(
  props: HeroIntegrateCardProps
): React.ReactElement {
  const { appID, numberOfClients } = props;
  const capture = useCapture();

  const hasApp = numberOfClients > 0;

  const { href, labelMessageID } = useMemo(() => {
    if (hasApp) {
      return {
        href: `/project/${appID}/configuration/apps`,
        labelMessageID: "GetStartedScreen.hero.integrate.view-application",
      };
    }
    return {
      href: `/project/${appID}/configuration/apps/add`,
      labelMessageID: "GetStartedScreen.hero.integrate.start-integration",
    };
  }, [appID, hasApp]);

  const onClick = useCallback(() => {
    capture("getStarted.clicked-create_app");
  }, [capture]);

  return (
    <div className={styles.heroCard}>
      <div className={styles.heroCardBody}>
        <Text block={true} className={styles.heroBadge}>
          <FormattedMessage id="GetStartedScreen.hero.integrate.badge" />
        </Text>
        <Text as="h2" block={true} className={styles.heroTitle}>
          <FormattedMessage id="GetStartedScreen.hero.integrate.title" />
        </Text>
        <Text block={true} className={styles.heroSubtitle}>
          <FormattedMessage id="GetStartedScreen.hero.integrate.subtitle" />
        </Text>
        <div className={styles.heroButtonRow}>
          <Link to={href} onClick={onClick} className={styles.heroPrimaryButton}>
            <FormattedMessage id={labelMessageID} />
          </Link>
        </div>
      </div>
    </div>
  );
}
