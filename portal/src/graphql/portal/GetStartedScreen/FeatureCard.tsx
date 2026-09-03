import { Heading } from "@radix-ui/themes";
import React from "react";
import { FormattedMessage } from "../../../intl";
import Link from "../../../Link";
import ExternalLink from "../../../ExternalLink";
import styles from "./GetStartedScreen.module.css";

// IconProps is not exported from @radix-ui/react-icons, so we duplicate it.
interface IconProps extends React.SVGAttributes<SVGElement> {
  children?: never;
  color?: string;
}

interface FeatureCardProps {
  Icon: React.ComponentType<IconProps>;
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
  return (
    <div className={styles.featureCard}>
      <div className={styles.featureHeader}>
        <Icon
          className={styles.featureIcon}
          width={18}
          height={18}
          aria-hidden={true}
        />
        <Heading as="h3" className={styles.featureTitle}>
          <FormattedMessage id={titleMessageID} />
        </Heading>
      </div>
      <p className={styles.featureDescription}>
        <FormattedMessage id={descriptionMessageID} />
      </p>
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
