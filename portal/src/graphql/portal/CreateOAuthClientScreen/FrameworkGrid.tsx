import React, { useCallback, useMemo } from "react";
import cn from "classnames";
import { Text } from "@radix-ui/themes";
import { FormattedMessage } from "react-intl";
import {
  IconRadioCards,
  type IconRadioCardOption,
} from "../../../components/v2/IconRadioCards/IconRadioCards";
import {
  frameworks,
  type FrameworkEntry,
  type FrameworkSection,
} from "./frameworks";
import type { Framework } from "../../../types";
import styles from "./FrameworkGrid.module.css";

export interface FrameworkGridProps {
  selectedId: Framework | null;
  m2mSelected: boolean;
  onSelect: (id: Framework) => void;
  onSelectM2M: () => void;
}

const M2M_VALUE = "__m2m__" as const;

type GridValue = Framework | typeof M2M_VALUE;

const sectionsOrder: FrameworkSection[] = ["website", "mobile", "integration"];
const sectionLabelKey: Record<FrameworkSection, string> = {
  website: "CreateOAuthClientScreen.framework.section.website",
  mobile: "CreateOAuthClientScreen.framework.section.mobile",
  integration: "CreateOAuthClientScreen.framework.section.integration",
};

function frameworkOption(f: FrameworkEntry): IconRadioCardOption<GridValue> {
  return {
    value: f.id,
    icon: (
      <i
        className={cn("ti", `ti-${f.iconName}`, styles.icon)}
        aria-hidden={true}
      />
    ),
    title: <FormattedMessage id={f.displayNameMessageId} />,
    subtitle: <FormattedMessage id={f.helperTextMessageId} />,
  };
}

const m2mOption: IconRadioCardOption<GridValue> = {
  value: M2M_VALUE,
  icon: <i className={cn("ti", "ti-server", styles.icon)} aria-hidden={true} />,
  title: <FormattedMessage id="CreateOAuthClientScreen.framework.m2m.title" />,
  subtitle: (
    <FormattedMessage id="CreateOAuthClientScreen.framework.m2m.description" />
  ),
};

export const FrameworkGrid: React.FC<FrameworkGridProps> = ({
  selectedId,
  m2mSelected,
  onSelect,
  onSelectM2M,
}) => {
  const grouped = useMemo(() => {
    const acc: Record<FrameworkSection, FrameworkEntry[]> = {
      website: [],
      mobile: [],
      integration: [],
    };
    frameworks.forEach((f) => {
      acc[f.section].push(f);
    });
    return acc;
  }, []);

  const onValueChange = useCallback(
    (value: GridValue) => {
      if (value === M2M_VALUE) {
        onSelectM2M();
      } else {
        onSelect(value);
      }
    },
    [onSelect, onSelectM2M]
  );

  return (
    <div className={styles.root}>
      {sectionsOrder.map((section) => {
        const options: IconRadioCardOption<GridValue>[] =
          grouped[section].map(frameworkOption);
        if (section === "integration") {
          options.push(m2mOption);
        }
        // Each section is its own radio group; the shared controlled value
        // makes at most one card checked across all sections.
        const sectionValue: GridValue | null =
          m2mSelected && section === "integration"
            ? M2M_VALUE
            : selectedId != null &&
              grouped[section].some((f) => f.id === selectedId)
            ? selectedId
            : null;
        return (
          <div key={section} className={styles.section}>
            <Text size="2" weight="medium" className={styles.sectionLabel}>
              <FormattedMessage id={sectionLabelKey[section]} />
            </Text>
            <IconRadioCards
              size="2"
              options={options}
              value={sectionValue}
              onValueChange={onValueChange}
              itemMinWidth={200}
              itemFillSpaces={true}
            />
          </div>
        );
      })}
    </div>
  );
};
