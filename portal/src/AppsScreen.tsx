import React from "react";
import { graphql, QueryRenderer } from "react-relay";
import { AppsScreenQueryResponse } from "./__generated__/AppsScreenQuery.graphql";
import { environment } from "./relay";
import ScreenHeader from "./ScreenHeader";
import ShowError from "./ShowError";
import ShowLoading from "./ShowLoading";
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
        const appID = edge?.node?.id;
        return (
          <div key={appID} className={styles.appItem}>
            {appID}
          </div>
        );
      })}
    </div>
  );
};

const AppsScreen: React.FC = function AppsScreen() {
  return (
    <div>
      <ScreenHeader />
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
    </div>
  );
};

export default AppsScreen;
