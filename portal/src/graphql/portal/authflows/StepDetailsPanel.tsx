import React, { useContext, useMemo } from "react";
import yaml from "js-yaml";
import { Flex, Heading, IconButton, Text } from "@radix-ui/themes";
import { Cross2Icon } from "@radix-ui/react-icons";
import { Context, FormattedMessage } from "../../../intl";
import { FlowStepItem } from "./tree";
import { optionLabel, stepTitle } from "./labels";
import styles from "./StepDetailsPanel.module.css";

export interface StepDetailsPanelProps {
  item: FlowStepItem;
  onClose: () => void;
}

export function StepDetailsPanel({
  item,
  onClose,
}: StepDetailsPanelProps): React.ReactElement {
  const { renderToString } = useContext(Context);
  const { step } = item;
  const options = step.one_of ?? [];
  const stepYAML = useMemo(() => yaml.dump(step), [step]);
  return (
    <aside className={styles.panel}>
      <Flex justify="between" align="center">
        <Heading size="3" as="h2">
          {stepTitle(step.type, renderToString)}
        </Heading>
        <IconButton
          variant="ghost"
          color="gray"
          aria-label={renderToString("AuthFlowsScreen.panel.close")}
          onClick={onClose}
        >
          <Cross2Icon />
        </IconButton>
      </Flex>
      {step.name != null ? (
        <Text as="p" size="1" color="gray">
          <FormattedMessage
            id="AuthFlowsScreen.panel.step-name"
            values={{ name: step.name }}
          />
        </Text>
      ) : null}
      {item.optional ? (
        <Text as="p" size="2">
          <FormattedMessage id="AuthFlowsScreen.panel.optional" />
        </Text>
      ) : null}
      {step.target_step != null ? (
        <Text as="p" size="2">
          <FormattedMessage
            id="AuthFlowsScreen.panel.target-step"
            values={{ name: step.target_step }}
          />
        </Text>
      ) : null}
      {options.length > 0 ? (
        <>
          <Heading size="2" as="h3">
            <FormattedMessage id="AuthFlowsScreen.panel.options" />
          </Heading>
          <ul className={styles.options}>
            {options.map((o, i) => (
              <li key={i}>{optionLabel(o, renderToString)}</li>
            ))}
          </ul>
        </>
      ) : null}
      <Heading size="2" as="h3">
        <FormattedMessage id="AuthFlowsScreen.panel.yaml" />
      </Heading>
      <pre className={styles.yaml}>{stepYAML}</pre>
    </aside>
  );
}
