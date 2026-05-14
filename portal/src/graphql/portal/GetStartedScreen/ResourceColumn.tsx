import React from "react";
import { FormattedMessage } from "../../../intl";
import { Text, useTheme } from "@fluentui/react";
import Link from "../../../Link";
import ExternalLink from "../../../ExternalLink";
import styles from "./GetStartedScreen.module.css";

// IconProps is not exported from @radix-ui/react-icons, so we duplicate it.
interface IconProps extends React.SVGAttributes<SVGElement> {
  children?: never;
  color?: string;
}

export interface ResourceRowProps {
  Icon: React.ComponentType<IconProps>;
  titleMessageID: string;
  descriptionMessageID?: string;
  internalHref?: string;
  externalHref?: string;
  onClick?: (e: React.MouseEvent<HTMLElement>) => void;
}

interface ResourceColumnProps {
  headingMessageID: string;
  rows: ResourceRowProps[];
}

function ResourceRow(props: ResourceRowProps): React.ReactElement {
  const {
    Icon,
    titleMessageID,
    descriptionMessageID,
    internalHref,
    externalHref,
    onClick,
  } = props;
  const theme = useTheme();

  const body = (
    <>
      <div className={styles.resourceIconWrapper}>
        <Icon
          className={styles.resourceIcon}
          width={20}
          height={20}
          color={theme.palette.themePrimary}
          aria-hidden={true}
        />
      </div>
      <div className={styles.resourceTextBlock}>
        <Text block={true} className={styles.resourceTitle}>
          <FormattedMessage id={titleMessageID} />
        </Text>
        {descriptionMessageID != null ? (
          <Text block={true} className={styles.resourceDescription}>
            <FormattedMessage id={descriptionMessageID} />
          </Text>
        ) : null}
      </div>
    </>
  );

  if (internalHref != null) {
    return (
      <Link to={internalHref} onClick={onClick} className={styles.resourceRow}>
        {body}
      </Link>
    );
  }
  if (externalHref != null) {
    return (
      <ExternalLink
        href={externalHref}
        onClick={onClick}
        className={styles.resourceRow}
      >
        {body}
      </ExternalLink>
    );
  }
  return <div className={styles.resourceRow}>{body}</div>;
}

export default function ResourceColumn(
  props: ResourceColumnProps
): React.ReactElement {
  const { headingMessageID, rows } = props;
  return (
    <div className={styles.resourceColumn}>
      <Text as="h3" block={true} className={styles.resourceHeading}>
        <FormattedMessage id={headingMessageID} />
      </Text>
      <div className={styles.resourceRows}>
        {rows.map((row) => (
          <ResourceRow key={row.titleMessageID} {...row} />
        ))}
      </div>
    </div>
  );
}
