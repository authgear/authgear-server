import React, { useContext, useMemo } from "react";
import { gql, useQuery } from "@apollo/client";
import { useNavigate } from "react-router-dom";
import {
  Context as LocaleContext,
  FormattedMessage,
} from "@oursky/react-messageformat";
import {
  CommandBar,
  ICommandBarItemProps,
  INavLink,
  INavLinkGroup,
  Nav,
  Text,
} from "@fluentui/react";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import { AppsScreenQuery } from "./__generated__/AppsScreenQuery";
import ScreenHeader from "../../ScreenHeader";
import styles from "./AppsScreen.module.scss";

const query = gql`
  query AppsScreenQuery {
    apps {
      edges {
        node {
          id
          effectiveAppConfig
        }
      }
    }
  }
`;

const AppList: React.FC<AppsScreenQuery> = function AppList(
  props: AppsScreenQuery
) {
  const navigate = useNavigate();
  const { renderToString } = useContext(LocaleContext);

  const commands: ICommandBarItemProps[] = useMemo(
    () => [
      {
        key: "create",
        text: renderToString("AppsScreen.create-app"),
        iconProps: { iconName: "NewFolder" },
      },
    ],
    [renderToString]
  );

  const groups: INavLinkGroup[] = useMemo(
    () => [
      {
        links:
          props.apps?.edges?.map(
            (edge): INavLink => {
              const appID = String(edge?.node?.id);
              const appOrigin =
                edge?.node?.effectiveAppConfig.http?.public_origin;
              const relPath = "/apps/" + encodeURIComponent(appID);
              return {
                name: appOrigin ?? appID,
                url: relPath,
                key: appID,
                onClick: (e) => {
                  e?.preventDefault();
                  e?.stopPropagation();
                  navigate(relPath);
                },
              };
            }
          ) ?? [],
      },
    ],
    [props.apps, navigate]
  );

  return (
    <main className={styles.root}>
      <ScreenHeader />
      <CommandBar
        className={styles.commandBar}
        items={[]}
        farItems={commands}
      />
      <section className={styles.body}>
        <Text as="h1" variant="xLarge" block={true}>
          <FormattedMessage id="AppsScreen.title" />
        </Text>
        <Nav groups={groups} />
      </section>
    </main>
  );
};

const AppsScreen: React.FC = function AppsScreen() {
  const { loading, error, data, refetch } = useQuery<AppsScreenQuery>(query);

  if (loading) {
    return <ShowLoading />;
  }

  if (error != null) {
    return <ShowError error={error} onRetry={refetch} />;
  }

  return <AppList apps={data?.apps ?? null} />;
};

export default AppsScreen;
