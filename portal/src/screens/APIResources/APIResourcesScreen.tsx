import React from "react";
import ScreenContent from "../../ScreenContent";
import ScreenTitle from "../../ScreenTitle";
import ScreenDescription from "../../ScreenDescription";
import ScreenContentHeader from "../../ScreenContentHeader";
import { FormattedMessage } from "@oursky/react-messageformat";
import { ResourceList } from "../../components/api-resources/ResourceList";
import { useResourcesQueryQuery } from "../../graphql/adminapi/query/resourcesQuery.generated";
import ShowError from "../../ShowError";

const APIResourcesScreen: React.VFC = function APIResourcesScreen() {
  const { data, loading, error, refetch } = useResourcesQueryQuery();

  const resources =
    data?.resources?.edges
      ?.map((edge) => edge?.node)
      .filter((resource): resource is NonNullable<typeof resource> =>
        Boolean(resource)
      ) ?? [];

  if (error != null) {
    return <ShowError error={error} onRetry={refetch} />;
  }

  return (
    <ScreenContent>
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
      <div className="col-span-full p-8">
        <ResourceList resources={resources} loading={loading} />
      </div>
    </ScreenContent>
  );
};

export default APIResourcesScreen;
