import React, { useCallback } from "react";
import ActionButton from "./ActionButton";
import { useSystemConfig } from "./context/SystemConfigContext";

export interface FoldableDivProps {
  folded: boolean;
  label: React.ReactElement;
  setFolded: (folded: boolean) => void;
  className?: string;
  children?: React.ReactNode;
}

export default function FoldableDiv(
  props: FoldableDivProps
): React.ReactElement {
  const { folded, setFolded, label, className, children } = props;
  const { themes } = useSystemConfig();
  const onClick = useCallback(() => {
    setFolded(!folded);
  }, [folded, setFolded]);

  return (
    <div className={className}>
      <ActionButton
        text={label}
        iconProps={{
          iconName: folded ? "ChevronDown" : "ChevronUp",
        }}
        styles={{
          root: {
            padding: "0",
          },
          flexContainer: {
            flexDirection: "row-reverse",
            gap: "8px",
          },
          label: {
            color: themes.main.palette.themePrimary,
            margin: "0",
          },
        }}
        onClick={onClick}
      />
      {folded ? null : children}
    </div>
  );
}
