import React from "react";
import { gql, useQuery } from "@apollo/client";
import { Link } from "react-router-dom";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import styles from "./AppsScreen.module.scss";

const query = gql`
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

interface Data {
  apps?: {
    edges?: ({
      node?: {
        id: string;
      };
    } | null)[];
  };
}

interface Variables {}

const AppList: React.FC<Data> = function AppList(props: Data) {
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
  const { loading, error, data, refetch } = useQuery<Data, Variables>(query);

  if (loading) {
    return <ShowLoading />;
  }

  if (error != null) {
    return <ShowError error={error} onRetry={refetch} />;
  }

  return <AppList {...data} />;
};

export default AppsScreen;
