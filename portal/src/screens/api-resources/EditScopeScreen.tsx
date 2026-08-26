import React from "react";
import { Navigate, useParams } from "react-router-dom";

// Kept for backwards-compatible deep links; edit now opens as a modal on the details Scopes tab.
const EditScopeScreen: React.VFC = function EditScopeScreen() {
  const { appID, resourceID } = useParams<{
    appID: string;
    resourceID: string;
  }>();
  return (
    <Navigate
      to={`/project/${encodeURIComponent(
        appID ?? ""
      )}/api-resources/${encodeURIComponent(resourceID ?? "")}`}
      replace={true}
    />
  );
};

export default EditScopeScreen;
