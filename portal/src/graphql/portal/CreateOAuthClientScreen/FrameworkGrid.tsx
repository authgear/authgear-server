import React, { useMemo } from "react";
import { useIntl } from "react-intl";
import { FrameworkCard } from "./FrameworkCard";
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
  const { formatMessage } = useIntl();
  const grouped = useMemo(() => {
    const acc: Record<FrameworkSection, FrameworkEntry[]> = {
      website: [],
      mobile: [],
    };
    frameworks.forEach((f) => acc[f.section].push(f));
    return acc;
  }, []);
  return (
    <div className={styles.root} role="radiogroup">
      {sectionsOrder.map((section) => (
        <div key={section} className={styles.section}>
          <div className={styles.sectionLabel}>
            {formatMessage({ id: sectionLabelKey[section] })}
          </div>
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
