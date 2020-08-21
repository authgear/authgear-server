import React, { useCallback } from "react";
import authgear from "@authgear/web";

const AppsScreen: React.FC = function AppsScreen() {
  const redirectURI = window.location.origin + "/";

  const onClickLogout = useCallback(() => {
    authgear
      .logout({
        redirectURI,
      })
      .catch((err) => {
        console.error(err);
      });
  }, [redirectURI]);

  return (
    <div>
      <p>This is /apps</p>
      <button type="button" onClick={onClickLogout}>
        Click here to logout
      </button>
    </div>
  );
};

export default AppsScreen;
