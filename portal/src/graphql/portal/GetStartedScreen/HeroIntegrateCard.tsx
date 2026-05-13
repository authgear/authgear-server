import React, { useCallback, useMemo } from "react";
import { useParams } from "react-router-dom";
import { FormattedMessage } from "../../../intl";
import { Text, Image, ImageFit } from "@fluentui/react";
import Link from "../../../Link";
import PrimaryButton from "../../../PrimaryButton";
import { useCapture } from "../../../gtm_v2";
import heroIntegrateImage from "../../../images/getting-started/hero-integrate.svg";
import styles from "./GetStartedScreen.module.css";

interface HeroIntegrateCardProps {
  numberOfClients: number;
}

export default function HeroIntegrateCard(
  props: HeroIntegrateCardProps
): React.ReactElement {
  const { numberOfClients } = props;
  const { appID } = useParams() as { appID: string };
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
          <Link to={href} onClick={onClick} className={styles.heroPrimaryLink}>
            <PrimaryButton
              text={<FormattedMessage id={labelMessageID} />}
            />
          </Link>
        </div>
      </div>
      <Image
        className={styles.heroIllustration}
        src={heroIntegrateImage}
        imageFit={ImageFit.contain}
        alt=""
      />
    </div>
  );
}
