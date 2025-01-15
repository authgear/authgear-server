import "graphiql/graphiql.css";
import "@graphiql/plugin-explorer/dist/style.css";

import React from "react";
import authgear from "@authgear/web";
import { render } from "react-dom";
import { GraphiQL } from "graphiql";
import { explorerPlugin } from "@graphiql/plugin-explorer";
import { createGraphiQLFetcher } from "@graphiql/toolkit";

(async function () {
  let fetch = window.fetch.bind(window);

  // There are 3 major use cases of GraphiQL.
  //
  // 1. GET <admin_api>/_api/admin/graphql, and expect it to work.
  //    This use case assumes ADMIN_API_AUTH=none, which is only true for development.
  //    In this use case, we use window.fetch.
  //
  // 2. GET <portal>/api/graphql. This use case requires inheritance of the authenticated state
  //    from the Authgear SDK, so we need to use authgear.fetch.
  //
  // 3. Access the Admin API GraphiQL proxied by the portal. This is similar to 2 that authenticated
  //    state is needed.
  //
  // We use the following flag to determine whether to use window.fetch or authgear.fetch.
  const metaElement = document.querySelector("meta[name=x-is-portal]");
  if (
    metaElement != null &&
    metaElement instanceof HTMLMetaElement &&
    metaElement.content === "true"
  ) {
    const resp = await fetch("/api/system-config.json");
    const systemConfig = await resp.json();
    const { authgearClientID, authgearEndpoint, authgearWebSDKSessionType } =
      systemConfig;
    await authgear.configure({
      clientID: authgearClientID,
      endpoint: authgearEndpoint,
      sessionType: authgearWebSDKSessionType,
    });

    // @ts-expect-error
    fetch = authgear.fetch.bind(authgear);
  }

  const fetcher = createGraphiQLFetcher({
    url: "",
    fetch,
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
