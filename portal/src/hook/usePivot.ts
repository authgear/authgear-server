import { useCallback, useEffect } from "react";
import { useLocation, useNavigate, useSearchParams } from "react-router-dom";
import { IPivotItemProps } from "@fluentui/react";

function isHashValid<K extends string>(
  validItemKeys: K[],
  hash: string
): hash is K {
  return validItemKeys.includes(hash as K);
}

export function usePivotNavigation<K extends string = string>(
  validItemKeys: K[],
  onSwitchTab?: () => void
): {
  selectedKey: K;
  onLinkClick: (item?: { props: IPivotItemProps }) => void;
} {
  if (validItemKeys.length <= 0) {
    throw new Error("validItemKey must be non-empty");
  }
  const navigate = useNavigate();
  const location = useLocation();
  const [searchParams] = useSearchParams();
  const hash = location.hash.slice(1);
  const initialSelectedKey = validItemKeys[0];

  const changeHashKeepSearchParam = useCallback(
    (hash: string, options?: { replace?: boolean }) => {
      const queryStr = searchParams.toString();
      navigate(
        {
          search: queryStr,
          hash: hash,
          pathname: location.pathname,
        },
        options
      );
    },
    [location.pathname, navigate, searchParams]
  );

  useEffect(() => {
    if (!isHashValid(validItemKeys, hash)) {
      // NOTE: avoid adding extra entry to history stack
      // NOTE: avoid changing query string
      changeHashKeepSearchParam(initialSelectedKey, { replace: true });
    }
  }, [validItemKeys, hash, initialSelectedKey, changeHashKeepSearchParam]);

  const onLinkClick = useCallback(
    (item?: { props: IPivotItemProps }) => {
      const itemKey = item?.props.itemKey;
      if (typeof itemKey === "string") {
        if (itemKey !== hash) {
          onSwitchTab?.();
          // NOTE: avoid changing query string
          changeHashKeepSearchParam(itemKey);
        }
      }
    },
    [hash, onSwitchTab, changeHashKeepSearchParam]
  );

  const selectedKey = isHashValid(validItemKeys, hash)
    ? hash
    : initialSelectedKey;

  return { selectedKey, onLinkClick };
}
