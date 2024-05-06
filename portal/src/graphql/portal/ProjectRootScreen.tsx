import React, { useEffect } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { useQuery } from "@apollo/client";
import { useSystemConfig } from "../../context/SystemConfigContext";
import {
  ScreenNavQueryQuery,
  ScreenNavQueryDocument,
} from "./query/screenNavQuery.generated";
import { usePortalClient } from "./apollo";
import ShowLoading from "../../ShowLoading";

const ProjectRootScreen: React.VFC = function ProjectRootScreen() {
  const { appID } = useParams() as { appID: string };
  const { analyticEnabled } = useSystemConfig();
  const navigate = useNavigate();
  const client = usePortalClient();
  const queryResult = useQuery<ScreenNavQueryQuery>(ScreenNavQueryDocument, {
    client,
    variables: {
      id: appID,
    },
    fetchPolicy: "network-only",
  });
  const app =
    queryResult.data?.node?.__typename === "App" ? queryResult.data.node : null;
  const { loading } = queryResult;
  const skippedTutorial = app?.tutorialStatus.data.skipped === true;
  const path = !skippedTutorial
    ? `/project/${appID}/getting-started`
    : analyticEnabled
    ? `/project/${appID}/analytics`
    : `/project/${appID}/users/`;

  useEffect(() => {
    if (!loading && app != null) {
      navigate(path, { replace: true });
    }
  }, [loading, app, path, navigate]);

  return <ShowLoading />;
};

export default ProjectRootScreen;
