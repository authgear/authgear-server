import React, { useMemo } from "react";
import { FormattedMessage } from "react-intl";
import { FrameworkCard } from "./FrameworkCard";
import { M2MCard } from "./M2MCard";
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
  m2mSelected: boolean;
  onSelect: (id: Framework) => void;
  onSelectM2M: () => void;
}

const sectionsOrder: FrameworkSection[] = ["website", "mobile", "integration"];
const sectionLabelKey: Record<FrameworkSection, string> = {
  website: "CreateOAuthClientScreen.framework.section.website",
  mobile: "CreateOAuthClientScreen.framework.section.mobile",
  integration: "CreateOAuthClientScreen.framework.section.integration",
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
            {section === "integration" ? (
              <M2MCard selected={m2mSelected} onSelect={onSelectM2M} />
            ) : null}
          </div>
        </div>
      ))}
    </div>
  );
};
