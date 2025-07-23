import React, { useState, useCallback, useMemo } from "react";
import ScreenContent from "../../ScreenContent";
import { encodeOffsetToCursor } from "../../util/pagination";
import ScreenTitle from "../../ScreenTitle";
import ScreenDescription from "../../ScreenDescription";
import ScreenContentHeader from "../../ScreenContentHeader";
import { FormattedMessage } from "@oursky/react-messageformat";
import { ResourceList } from "../../components/api-resources/ResourceList";
import { useResourcesQueryQuery } from "../../graphql/adminapi/query/resourcesQuery.generated";
import ShowError from "../../ShowError";
import { PaginationProps } from "../../PaginationWidget";

const PAGE_SIZE = 20;

const APIResourcesScreen: React.VFC = function APIResourcesScreen() {
  const [offset, setOffset] = useState(0);

  const { data, loading, error, refetch } = useResourcesQueryQuery({
    variables: {
      first: PAGE_SIZE,
      after: useMemo(() => {
        if (offset === 0) {
          return undefined;
        }
        return encodeOffsetToCursor(offset - 1);
      }, [offset]),
    },
  });

  const resources =
    data?.resources?.edges
      ?.map((edge) => edge?.node)
      .filter(
        (resource): resource is NonNullable<typeof resource> => !!resource
      ) ?? [];

  const onChangeOffset = useCallback((offset: number) => {
    setOffset(offset);
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
    <ScreenContent className="flex-1" layout="list">
      <ScreenContentHeader
        title={
          <ScreenTitle>
            <FormattedMessage id="APIResourcesScreen.title" />
          </ScreenTitle>
        }
        description={
          <ScreenDescription>
            <FormattedMessage id="APIResourcesScreen.description" />
          </ScreenDescription>
        }
      />
      <div className="col-span-full flex flex-col">
        <ResourceList
          className="flex-1"
          resources={resources}
          loading={loading}
          pagination={pagination}
        />
      </div>
    </ScreenContent>
  );
};

export default APIResourcesScreen;
