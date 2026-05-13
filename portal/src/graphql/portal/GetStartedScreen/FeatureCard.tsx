import React from "react";
import { FormattedMessage } from "../../../intl";
import { Text, useTheme } from "@fluentui/react";
import { LucideIcon } from "lucide-react";
import Link from "../../../Link";
import ExternalLink from "../../../ExternalLink";
import styles from "./GetStartedScreen.module.css";

interface FeatureCardProps {
  Icon: LucideIcon;
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
    Icon,
    titleMessageID,
    descriptionMessageID,
    actionMessageID,
    internalHref,
    externalHref,
    onClick,
  } = props;
  const theme = useTheme();

  return (
    <div className={styles.featureCard}>
      <Icon
        className={styles.featureIcon}
        size={28}
        strokeWidth={1.75}
        color={theme.palette.themePrimary}
        aria-hidden={true}
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
