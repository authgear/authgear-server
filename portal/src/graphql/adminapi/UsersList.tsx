import React, { useMemo, useContext, useState, useCallback } from "react";
import { graphql, QueryRenderer } from "react-relay";
import {
  DetailsList,
  DetailsListLayoutMode,
  SelectionMode,
  IColumn,
  DefaultButton,
} from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import AppContext from "../../AppContext";
import {
  UsersListQueryVariables,
  UsersListQueryResponse,
} from "./__generated__/UsersListQuery.graphql";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import styles from "./UsersList.module.scss";

interface PageInfo {
  hasNextPage: boolean;
  endCursor: string | null;
}

interface State {
  cursor: string | null;
}

interface Props extends UsersListQueryResponse {
  onClickNext: (pageInfo: PageInfo) => void;
}

const UsersList: React.FC<Props> = function UsersList(props: Props) {
  const edges = props.users?.edges;
  const pageInfo = props.users?.pageInfo;
  const { onClickNext } = props;

  const thisOnClickNext = useCallback(() => {
    if (pageInfo != null) {
      onClickNext(pageInfo);
    }
  }, [pageInfo, onClickNext]);

  const { renderToString } = useContext(Context);

  const columns: IColumn[] = [
    {
      key: "id",
      fieldName: "id",
      name: renderToString("UsersList.column.id"),
      minWidth: 400,
      maxWidth: 400,
    },
    {
      key: "createdAt",
      fieldName: "createdAt",
      name: renderToString("UsersList.column.created-at"),
      minWidth: 300,
    },
  ];

  const items: {
    id: string;
    createdAt: unknown;
  }[] = useMemo(() => {
    const items = [];
    if (edges != null) {
      for (const edge of edges) {
        const node = edge?.node;
        if (node != null) {
          items.push({
            id: node.id,
            createdAt: node.createdAt,
          });
        }
      }
    }
    return items;
  }, [edges]);

  return (
    <div className={styles.root}>
      <DetailsList
        selectionMode={SelectionMode.none}
        layoutMode={DetailsListLayoutMode.justified}
        columns={columns}
        items={items}
      />
      <div className={styles.pagination}>
        <DefaultButton
          onClick={thisOnClickNext}
          disabled={!(pageInfo?.hasNextPage ?? false)}
        >
          <FormattedMessage id="next" />
        </DefaultButton>
      </div>
    </div>
  );
};

const query = graphql`
  query UsersListQuery($pageSize: Int!, $cursor: String) {
    users(first: $pageSize, after: $cursor) {
      edges {
        node {
          id
          createdAt
        }
      }
      pageInfo {
        hasNextPage
        endCursor
      }
      totalCount
    }
  }
`;

// There is createRefetchContainer and createPaginationContainer.
// But the hasMore() createPaginationContainer for some reason always return false even hasNextPage is true.
// createRefetchContainer appends the refetch result. So it is not applicable too.
// So here we are using QueryRenderer to do pagination ourselves.
// FIXME(portal): However, the users query supports on infinite pagination.
// So we can only render a next button.
const RelayUsersList: React.FC = function RelayUsersList() {
  const environment = useContext(AppContext);

  const [{ cursor }, setState] = useState<State>({
    cursor: null,
  });

  const onClickNext = useCallback((pageInfo: PageInfo) => {
    if (pageInfo.endCursor != null) {
      setState({
        cursor: pageInfo.endCursor,
      });
    }
  }, []);

  return (
    <QueryRenderer<{
      variables: UsersListQueryVariables;
      response: UsersListQueryResponse;
    }>
      environment={environment}
      query={query}
      variables={{
        pageSize: 1,
        cursor,
      }}
      render={({ error, props, retry }) => {
        if (error != null) {
          return <ShowError error={error} onRetry={retry} />;
        }
        if (props == null) {
          // FIXME(portal): Use Skimmer
          return <ShowLoading />;
        }
        return <UsersList {...props} onClickNext={onClickNext} />;
      }}
    />
  );
};

export default RelayUsersList;
