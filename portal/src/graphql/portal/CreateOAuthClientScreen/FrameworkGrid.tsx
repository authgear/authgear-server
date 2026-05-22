import React, { useMemo } from "react";
import { FormattedMessage } from "react-intl";
import { FrameworkCard } from "./FrameworkCard";
import WidgetSubtitle from "../../../WidgetSubtitle";
import {
  frameworks,
  type FrameworkEntry,
  type FrameworkSection,
} from "./frameworks";
import type { Framework } from "../../../types";
import styles from "./FrameworkGrid.module.css";

export interface FrameworkGridProps {
  selectedId: Framework | null;
  onSelect: (id: Framework) => void;
}

const sectionsOrder: FrameworkSection[] = ["website", "mobile"];
const sectionLabelKey: Record<FrameworkSection, string> = {
  website: "CreateOAuthClientScreen.framework.section.website",
  mobile: "CreateOAuthClientScreen.framework.section.mobile",
};

export const FrameworkGrid: React.FC<FrameworkGridProps> = ({
  selectedId,
  onSelect,
}) => {
  const grouped = useMemo(() => {
    const acc: Record<FrameworkSection, FrameworkEntry[]> = {
      website: [],
      mobile: [],
    };
    frameworks.forEach((f) => {
      acc[f.section].push(f);
    });
    return acc;
  }, []);
  return (
    <div className={styles.root} role="radiogroup">
      {sectionsOrder.map((section) => (
        <div key={section} className={styles.section}>
          <WidgetSubtitle>
            <FormattedMessage id={sectionLabelKey[section]} />
          </WidgetSubtitle>
          <div className={styles.grid}>
            {grouped[section].map((f) => (
              <FrameworkCard
                key={f.id}
                framework={f}
                selected={selectedId === f.id}
                onSelect={() => onSelect(f.id)}
              />
            ))}
          </div>
        </div>
      ))}
    </div>
  );
};
