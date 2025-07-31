import React, { useContext, useState, useCallback, useMemo } from "react";
import cn from "classnames";
import styles from "./EditOAuthClientFormResourcesSection.module.css";
import WidgetTitle from "../../WidgetTitle";
import {
  Context as MessageContext,
  FormattedMessage,
} from "@oursky/react-messageformat";
import { SearchBox } from "@fluentui/react/lib/SearchBox";
import { useResourcesQueryQuery } from "../adminapi/query/resourcesQuery.generated";
import {
  ApplicationResourcesList,
  ApplicationResourceListItem,
} from "../../components/api-resources/ApplicationResourcesList";
import { OAuthClientConfig } from "../../types";

import { encodeOffsetToCursor } from "../../util/pagination";
import { PaginationProps } from "../../PaginationWidget";
import { useDebounced } from "../../hook/useDebounced";
import ShowError from "../../ShowError";

const PAGE_SIZE = 10;

export const EditOAuthClientFormResourcesSection: React.FC<{
  className?: string;
  client: OAuthClientConfig;
}> = ({ className, client }) => {
  const { renderToString } = useContext(MessageContext);
  const [searchKeyword, setSearchKeyword] = useState("");
  const [offset, setOffset] = useState(0);
  const [debouncedSearchKeyword] = useDebounced(searchKeyword, 300);
  const handleSearchChange = useCallback(
    (_event?: React.ChangeEvent<HTMLInputElement>, newValue?: string): void => {
      setSearchKeyword(newValue ?? "");
      setOffset(0); // Reset offset on search change
    },
    [setSearchKeyword, setOffset]
  );

  const { data, loading, error, refetch } = useResourcesQueryQuery({
    variables: {
      first: PAGE_SIZE,
      after: encodeOffsetToCursor(offset),
      searchKeyword:
        debouncedSearchKeyword === "" ? undefined : debouncedSearchKeyword,
    },
    fetchPolicy: "network-only",
  });

  const resourceListData: ApplicationResourceListItem[] = useMemo(() => {
    const resources =
      data?.resources?.edges
        ?.map((edge) => edge?.node)
        .filter((node) => !!node) ?? [];
    return resources.map((resource) => {
      const isAuthorized = resource.clientIDs.includes(client.client_id);
      return {
        id: resource.id,
        name: resource.name,
        resourceURI: resource.resourceURI,
        isAuthorized: isAuthorized,
      };
    });
  }, [client.client_id, data?.resources?.edges]);

  const handleToggleAuthorization = useCallback(
    (_item: ApplicationResourceListItem, _isAuthorized: boolean) => {
      // TODO: Implement mutation to update authorization status
    },
    []
  );

  const onChangeOffset = useCallback((newOffset: number) => {
    setOffset(newOffset);
  }, []);

  const pagination: PaginationProps = {
    offset,
    pageSize: PAGE_SIZE,
    totalCount: data?.resources?.totalCount ?? undefined,
    onChangeOffset,
  };
  if (error != null) {
    return <ShowError error={error} onRetry={refetch} />;
  }
  return (
    <section className={cn(styles.resourcesSection, className)}>
      <WidgetTitle id="resoucres">
        <FormattedMessage id="EditOAuthClientForm.resources.title" />
      </WidgetTitle>
      <SearchBox
        placeholder={renderToString("search")}
        styles={{
          root: {
            width: 300,
          },
        }}
        onChange={handleSearchChange}
      />
      <ApplicationResourcesList
        resources={resourceListData}
        loading={loading}
        pagination={pagination}
        onToggleAuthorization={handleToggleAuthorization}
      />
    </section>
  );
};
