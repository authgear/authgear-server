import React from "react";
import { FormattedMessage } from "../../../intl";
import { Text, Image, ImageFit } from "@fluentui/react";
import Link from "../../../Link";
import ExternalLink from "../../../ExternalLink";
import styles from "./GetStartedScreen.module.css";

interface FeatureCardProps {
  iconSrc: string;
  titleMessageID: string;
  descriptionMessageID: string;
  actionMessageID: string;
  internalHref?: string;
  externalHref?: string;
  onClick?: (e: React.MouseEvent<HTMLElement>) => void;
}

export default function FeatureCard(
  props: FeatureCardProps
): React.ReactElement {
  const {
    iconSrc,
    titleMessageID,
    descriptionMessageID,
    actionMessageID,
    internalHref,
    externalHref,
    onClick,
  } = props;

  return (
    <div className={styles.featureCard}>
      <Image
        className={styles.featureIcon}
        src={iconSrc}
        imageFit={ImageFit.contain}
        alt=""
      />
      <Text as="h3" block={true} className={styles.featureTitle}>
        <FormattedMessage id={titleMessageID} />
      </Text>
      <Text block={true} className={styles.featureDescription}>
        <FormattedMessage id={descriptionMessageID} />
      </Text>
      {internalHref != null ? (
        <Link
          to={internalHref}
          onClick={onClick}
          className={styles.featureAction}
        >
          <FormattedMessage id={actionMessageID} />
        </Link>
      ) : externalHref != null ? (
        <ExternalLink
          href={externalHref}
          onClick={onClick}
          className={styles.featureAction}
        >
          <FormattedMessage id={actionMessageID} />
        </ExternalLink>
      ) : null}
    </div>
  );
}
