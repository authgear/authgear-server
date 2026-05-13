import React, { useEffect } from "react";
import { useParams, useNavigate, useLocation } from "react-router-dom";
import { useQuery } from "@apollo/client";
import {
  ScreenNavQueryQuery,
  ScreenNavQueryDocument,
} from "./query/screenNavQuery.generated";
import { usePortalClient } from "./apollo";
import ShowLoading from "../../ShowLoading";

const ProjectRootScreen: React.VFC = function ProjectRootScreen() {
  const { appID } = useParams() as { appID: string };
  const navigate = useNavigate();
  const location = useLocation();
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
  const path = `/project/${appID}/getting-started`;

  useEffect(() => {
    if (!loading && app != null) {
      navigate(
        { pathname: path, search: location.search.toString() },
        { replace: true }
      );
    }
  }, [loading, app, path, navigate, location.search]);

  return <ShowLoading />;
};

export default ProjectRootScreen;
