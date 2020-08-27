import React from "react";
import { graphql, QueryRenderer } from "react-relay";
import { Link } from "react-router-dom";
import { AppsScreenQueryResponse } from "./__generated__/AppsScreenQuery.graphql";
import { environment } from "./relay";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import styles from "./AppsScreen.module.scss";

const query = graphql`
  query AppsScreenQuery {
    apps {
      edges {
        node {
          id
        }
      }
    }
  }
`;

interface Empty {}

const AppList: React.FC<AppsScreenQueryResponse> = function AppList(
  props: AppsScreenQueryResponse
) {
  return (
    <div className={styles.appList}>
      {props.apps?.edges?.map((edge) => {
        const appID = String(edge?.node?.id);
        return (
          <Link
            to={"/apps/" + encodeURIComponent(appID)}
            key={appID}
            className={styles.appItem}
          >
            {appID}
          </Link>
        );
      })}
    </div>
  );
};

const AppsScreen: React.FC = function AppsScreen() {
  return (
    <QueryRenderer<{ variables: Empty; response: AppsScreenQueryResponse }>
      environment={environment}
      query={query}
      variables={{}}
      render={({ error, props }) => {
        if (error != null) {
          return <ShowError error={error} />;
        }
        if (props == null) {
          return <ShowLoading />;
        }
        return <AppList {...props} />;
      }}
    />
  );
};

export default AppsScreen;
