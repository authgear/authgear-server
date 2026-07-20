import React from "react";
import { Navigate, useParams } from "react-router-dom";

// Kept for backwards-compatible deep links; create now opens as a modal on the list screen.
const CreateAPIResourceScreen: React.VFC = function CreateAPIResourceScreen() {
  const { appID } = useParams<{ appID: string }>();
  return (
    <Navigate
      to={`/project/${encodeURIComponent(appID ?? "")}/api-resources?create=1`}
      replace={true}
    />
  );
};

export default CreateAPIResourceScreen;
