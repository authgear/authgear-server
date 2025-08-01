import { useCallback, useEffect } from "react";
import { useLocation, useNavigate, useSearchParams } from "react-router-dom";
import { IPivotItemProps } from "@fluentui/react";

function isKeyValid<K extends string>(
  validItemKeys: K[],
  key: string
): key is K {
  return validItemKeys.includes(key as K);
}

export function usePivotNavigation<K extends string = string>(
  validItemKeys: K[],
  onSwitchTab?: () => void,
  searchParamKey?: string
): {
  selectedKey: K;
  onLinkClick: (item?: { props: IPivotItemProps }) => void;
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
    (newKey: string) => {
      const newSearchParams = new URLSearchParams(searchParams);
      let newHash = location.hash;
      if (searchParamKey == null) {
        // Using hash
        newHash = newKey;
      } else {
        newSearchParams.set(searchParamKey, newKey);
      }
      // NOTE: avoid adding extra entry to history stack
      // NOTE: avoid changing other query string
      const queryStr = newSearchParams.toString();
      navigate(
        {
          search: queryStr,
          hash: newHash,
          pathname: location.pathname,
        },
        { replace: true }
      );
    },
    [location.hash, location.pathname, navigate, searchParamKey, searchParams]
  );

  useEffect(() => {
    if (!isKeyValid(validItemKeys, currentTabKey)) {
      changeTabKey(initialSelectedKey);
    }
  }, [validItemKeys, currentTabKey, initialSelectedKey, changeTabKey]);

  const onLinkClick = useCallback(
    (item?: { props: IPivotItemProps }) => {
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
