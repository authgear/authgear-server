import { useCallback, useEffect } from "react";
import { useLocation, useNavigate, useSearchParams } from "react-router-dom";
// Structural stand-in for the old FluentUI IPivotItemProps; callers only
// ever read itemKey.
interface PivotItemPropsLike {
  itemKey?: string;
}

function isKeyValid<K extends string>(
  validItemKeys: K[],
  key: string
): key is K {
  return validItemKeys.includes(key as K);
}

export function usePivotNavigation<K extends string = string>(
  validItemKeys: K[],
  onSwitchTab?: () => void,
  searchParamKey?: string,
  // When true, a user tab switch pushes a new history entry so the browser
  // back button returns to the previously selected tab. Defaults to false
  // (replace) to preserve the historical behavior of existing callers.
  pushHistory: boolean = false
): {
  selectedKey: K;
  onLinkClick: (item?: { props: PivotItemPropsLike }) => void;
  onChangeKey: (key: K) => void;
} {
  if (validItemKeys.length <= 0) {
    throw new Error("validItemKey must be non-empty");
  }
  const navigate = useNavigate();
  const location = useLocation();
  const [searchParams] = useSearchParams();
  const initialSelectedKey = validItemKeys[0];

  const currentTabKey =
    (searchParamKey
      ? searchParams.get(searchParamKey)
      : location.hash.slice(1)) ?? initialSelectedKey;

  const changeTabKey = useCallback(
    (newKey: string, replace: boolean = !pushHistory) => {
      const newSearchParams = new URLSearchParams(searchParams);
      let newHash = location.hash;
      if (searchParamKey == null) {
        // Using hash
        newHash = newKey;
      } else {
        newSearchParams.set(searchParamKey, newKey);
      }
      // NOTE: avoid changing other query string
      const queryStr = newSearchParams.toString();
      navigate(
        {
          search: queryStr,
          hash: newHash,
          pathname: location.pathname,
        },
        { replace }
      );
    },
    [
      location.hash,
      location.pathname,
      navigate,
      searchParamKey,
      searchParams,
      pushHistory,
    ]
  );

  useEffect(() => {
    if (!isKeyValid(validItemKeys, currentTabKey)) {
      // Correcting an invalid key must never add a history entry.
      changeTabKey(initialSelectedKey, true);
    }
  }, [validItemKeys, currentTabKey, initialSelectedKey, changeTabKey]);

  const onLinkClick = useCallback(
    (item?: { props: PivotItemPropsLike }) => {
      const itemKey = item?.props.itemKey;
      if (typeof itemKey === "string") {
        if (itemKey !== currentTabKey) {
          onSwitchTab?.();
          // NOTE: avoid changing query string
          changeTabKey(itemKey);
        }
      }
    },
    [currentTabKey, onSwitchTab, changeTabKey]
  );

  const selectedKey = isKeyValid(validItemKeys, currentTabKey)
    ? currentTabKey
    : initialSelectedKey;

  return { selectedKey, onLinkClick, onChangeKey: changeTabKey };
}
