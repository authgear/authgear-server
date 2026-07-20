import React, { PropsWithChildren } from "react";
import * as Collapsible from "@radix-ui/react-collapsible";
import { ChevronDownIcon } from "@radix-ui/react-icons";
import { Heading, Text } from "@radix-ui/themes";
import { FormattedMessage } from "../../intl";
import styles from "./ConfigurationGroup.module.css";

interface ConfigurationGroupProps {
  labelKey: string;
  /** When true, renders as a Radix Collapsible accordion section. */
  collapsible?: boolean;
  defaultOpen?: boolean;
}
const ConfigurationGroup: React.VFC<
  PropsWithChildren<ConfigurationGroupProps>
> = function ConfigurationGroup(props) {
  const {
    labelKey,
    collapsible = false,
    defaultOpen = false,
    children,
  } = props;

  const label = <FormattedMessage id={labelKey} />;

  if (!collapsible) {
    return (
      <div className={styles.staticRoot}>
        <Heading as="h2" size="3" weight="medium" className={styles.label}>
          {label}
        </Heading>
        <div className={styles.content}>{children}</div>
      </div>
    );
  }

  return (
    <Collapsible.Root defaultOpen={defaultOpen} className={styles.root}>
      <Collapsible.Trigger className={styles.trigger}>
        <Text as="span" size="3" weight="medium" className={styles.label}>
          {label}
        </Text>
        <ChevronDownIcon className={styles.chevron} aria-hidden />
      </Collapsible.Trigger>
      <Collapsible.Content className={styles.content}>
        {children}
      </Collapsible.Content>
    </Collapsible.Root>
  );
};

export default ConfigurationGroup;
