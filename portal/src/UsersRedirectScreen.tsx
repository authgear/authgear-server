import React from "react";
import { Navigate, useLocation } from "react-router-dom";

// NOTE: This screen redirect users/* to user-management/users/* for backward compatibility purpose
const UsersRedirectScreen: React.VFC = function UsersRedirectScreen() {
  const location = useLocation();

  return (
    <Navigate
      to={{
        pathname: location.pathname.replace("users", "user-management/users"),
        search: location.search,
        hash: location.hash,
      }}
      replace={true}
    />
  );
};

export default UsersRedirectScreen;
