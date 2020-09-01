import React, { useMemo, useContext, useState, useCallback } from "react";
import { useQuery, gql } from "@apollo/client";
import {
  DetailsList,
  DetailsListLayoutMode,
  SelectionMode,
  IColumn,
  DefaultButton,
} from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import styles from "./UsersList.module.scss";

interface Data {
  users?: {
    edges?: ({
      node?: {
        id: string;
        createdAt: string;
      };
    } | null)[];
    pageInfo: PageInfo;
  };
}

interface Variables {
  pageSize: number;
  cursor: string | null;
}

interface PageInfo {
  hasNextPage: boolean;
  endCursor: string | null;
}

interface State {
  cursor: string | null;
}

interface Props extends Data {
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

const query = gql`
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

  const { loading, error, data, refetch } = useQuery<Data, Variables>(query, {
    variables: {
      pageSize: 1,
      cursor,
    },
  });

  if (loading) {
    // FIXME(portal): Use Skimmer
    return <ShowLoading />;
  }

  if (error != null) {
    return <ShowError error={error} onRetry={refetch} />;
  }

  return <UsersList {...data} onClickNext={onClickNext} />;
};

export default RelayUsersList;
