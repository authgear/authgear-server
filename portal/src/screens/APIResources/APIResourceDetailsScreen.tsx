import React from "react";
import { useParams } from "react-router-dom";
import { useResourceQueryQuery } from "../../graphql/adminapi/query/resourceQuery.generated";
import { useLoadableView } from "../../hook/useLoadableView";
import { FormattedMessage } from "@oursky/react-messageformat";
import APIResourceScreenLayout from "../../components/api-resources/APIResourceScreenLayout";
import { Resource } from "../../graphql/adminapi/globalTypes.generated";

const APIResourceDetailsScreen: React.VFC =
  function APIResourceDetailsScreen() {
    const { resourceID } = useParams<{ resourceID: string }>();
    const { data, loading, error, refetch } = useResourceQueryQuery({
      variables: { id: resourceID! },
    });

    return useLoadableView({
      loadables: [
        {
          isLoading: loading,
          loadError: error,
          reload: refetch,
          data: data,
        },
      ],
      render: ([query]) => {
        const { data } = query;
        const resource =
          data?.node && data.node.__typename === "Resource"
            ? (data.node as Resource)
            : null;
        if (!resource) {
          return null;
        }
        return (
          <APIResourceScreenLayout
            breadcrumbItems={[
              {
                to: "~/api-resources",
                label: <FormattedMessage id="ScreenNav.api-resources" />,
              },
              {
                to: "",
                label: resource.name ?? resource.resourceURI,
              },
            ]}
          >
            <div>{resource.name}</div>
          </APIResourceScreenLayout>
        );
      },
    });
  };

export default APIResourceDetailsScreen;
