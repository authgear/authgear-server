import React, { useContext } from "react";
import { useParams } from "react-router-dom";
import { useResourceQueryQuery } from "../../graphql/adminapi/query/resourceQuery.generated";
import { useLoadableView } from "../../hook/useLoadableView";
import {
  FormattedMessage,
  Context as MessageContext,
} from "@oursky/react-messageformat";
import APIResourceScreenLayout from "../../components/api-resources/APIResourceScreenLayout";
import { Resource } from "../../graphql/adminapi/globalTypes.generated";
import { Pivot, PivotItem } from "@fluentui/react";
import { usePivotNavigation } from "../../hook/usePivot";
import { APIResourceDetailsScreenDetailsTab } from "./APIResourceDetailsScreenDetailsTab";
import { APIResourceDetailsScreenScopesTab } from "./APIResourceDetailsScreenScopesTab";
import { APIResourceDetailsScreenApplicationsTab } from "./APIResourceDetailsScreenApplicationsTab";
import { APIResourceDetailsScreenTestTab } from "./APIResourceDetailsScreenTestTab";

function APIResourceDetailsContent({ resource }: { resource: Resource }) {
  const { selectedKey, onLinkClick } = usePivotNavigation([
    "details",
    "scopes",
    "applications",
    "test",
  ]);
  const { renderToString } = useContext(MessageContext);
  return (
    <div className="pt-6 flex flex-col col-span-full">
      <Pivot selectedKey={selectedKey} onLinkClick={onLinkClick}>
        <PivotItem
          headerText={renderToString("APIResourceDetailsScreen.tab.details")}
          itemKey="details"
        />
        <PivotItem
          headerText={renderToString("APIResourceDetailsScreen.tab.scopes")}
          itemKey="scopes"
        />
        <PivotItem
          headerText={renderToString(
            "APIResourceDetailsScreen.tab.applications"
          )}
          itemKey="applications"
        />
        <PivotItem
          headerText={renderToString("APIResourceDetailsScreen.tab.test")}
          itemKey="test"
        />
      </Pivot>
      {selectedKey === "details" ? (
        <APIResourceDetailsScreenDetailsTab resource={resource} />
      ) : null}
      {selectedKey === "scopes" ? (
        <APIResourceDetailsScreenScopesTab resource={resource} />
      ) : null}
      {selectedKey === "applications" ? (
        <APIResourceDetailsScreenApplicationsTab resource={resource} />
      ) : null}
      {selectedKey === "test" ? (
        <APIResourceDetailsScreenTestTab resource={resource} />
      ) : null}
    </div>
  );
}

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
          data?.node && data.node.__typename === "Resource" ? data.node : null;
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
            <APIResourceDetailsContent resource={resource} />
          </APIResourceScreenLayout>
        );
      },
    });
  };

export default APIResourceDetailsScreen;
