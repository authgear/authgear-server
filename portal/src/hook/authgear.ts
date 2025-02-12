import authgear, { Page } from "@authgear/web";
import { useSystemConfig } from "../context/SystemConfigContext";
import { useCallback } from "react";

export function useSettingsAnchor(): {
  href: string;
  onClick: (e?: React.SyntheticEvent<HTMLElement>) => void;
} {
  const { authgearEndpoint, authgearWebSDKSessionType } = useSystemConfig();
  const settingURL = authgearEndpoint + "/settings";
  const onClickSettings = useCallback(
    (e?: React.SyntheticEvent<HTMLElement>) => {
      if (e == null) {
        return;
      }
      switch (authgearWebSDKSessionType) {
        case "cookie":
          // We already have idp session,
          // just let the anchor handle the navigation.
          return;
        case "refresh_token":
          // We don't have idp session if using refresh token.
          // Use the sdk to open settings.
          e.preventDefault();
          e.stopPropagation();
          authgear.open(Page.Settings, { openInSameTab: true });
      }
    },
    [authgearWebSDKSessionType]
  );

  return {
    href: settingURL,
    onClick: onClickSettings,
  };
}
