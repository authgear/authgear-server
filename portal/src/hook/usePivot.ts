import { useCallback } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import { IPivotItemProps } from "@fluentui/react";

export function usePivot(): {
  hash: string;
  onLinkClick: (item?: { props: IPivotItemProps }) => void;
} {
  const navigate = useNavigate();
  const location = useLocation();
  const hash = location.hash.slice(1);

  const onLinkClick = useCallback(
    (item) => {
      const itemKey = item?.props.itemKey;
      if (typeof itemKey === "string") {
        navigate(`./#${itemKey}`);
      }
    },
    [navigate]
  );

  return { hash, onLinkClick };
}
