import { useCallback, useEffect } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import { IPivotItemProps } from "@fluentui/react";

function isHashValid(validItemKeys: string[], hash: string): boolean {
  return validItemKeys.includes(hash);
}

export function usePivotNavigation(
  validItemKeys: string[],
  switchTabEffect?: () => void
): {
  selectedKey: string;
  onLinkClick: (item?: { props: IPivotItemProps }) => void;
} {
  if (validItemKeys.length <= 0) {
    throw new Error("validItemKey must be non-empty");
  }
  const navigate = useNavigate();
  const location = useLocation();
  const hash = location.hash.slice(1);
  const initialSelectedKey = validItemKeys[0];

  useEffect(() => {
    if (!isHashValid(validItemKeys, hash)) {
      // NOTE: avoid adding extra entry to history stack
      navigate(`#${initialSelectedKey}`, { replace: true });
    }
  }, [validItemKeys, hash, initialSelectedKey, navigate]);

  const onLinkClick = useCallback(
    (item?: { props: IPivotItemProps }) => {
      const itemKey = item?.props.itemKey;
      if (typeof itemKey === "string") {
        if (itemKey !== hash) {
          switchTabEffect?.();
          navigate(`#${itemKey}`);
        }
      }
    },
    [navigate, hash, switchTabEffect]
  );

  const selectedKey = isHashValid(validItemKeys, hash)
    ? hash
    : initialSelectedKey;

  return { selectedKey, onLinkClick };
}
