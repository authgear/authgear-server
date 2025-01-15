import "graphiql/graphiql.css";
import "@graphiql/plugin-explorer/dist/style.css";

import React from "react";
import authgear from "@authgear/web";
import { render } from "react-dom";
import { GraphiQL } from "graphiql";
import { explorerPlugin } from "@graphiql/plugin-explorer";
import { createGraphiQLFetcher } from "@graphiql/toolkit";

(async function () {
  const resp = await fetch("/api/system-config.json");
  const systemConfig = await resp.json();
  const { authgearClientID, authgearEndpoint, authgearWebSDKSessionType } =
    systemConfig;
  await authgear.configure({
    clientID: authgearClientID,
    endpoint: authgearEndpoint,
    sessionType: authgearWebSDKSessionType,
  });

  const fetcher = createGraphiQLFetcher({
    url: "",
    // @ts-expect-error
    fetch: authgear.fetch.bind(authgear),
  });

  const explorer = explorerPlugin();
  const plugins = [explorer];

  const query = new URLSearchParams(window.location.search).get("query") || "";

  render(
    <GraphiQL
      fetcher={fetcher}
      defaultEditorToolsVisibility={true}
      plugins={plugins}
      query={query}
    />,
    document.getElementById("react-app-root")
  );
})();
