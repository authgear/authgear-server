import React, { useContext, useState } from "react";
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
import { useUpdateResourceMutationMutation } from "../../graphql/adminapi/mutations/updateResourceMutation.generated";
import { useSimpleForm } from "../../hook/useSimpleForm";
import { FormContainerBase } from "../../FormContainerBase";
import {
  ResourceForm,
  ResourceFormState,
  sanitizeFormState,
} from "../../components/api-resources/ResourceForm";

function APIResourceDetailsTab({ resource }: { resource: Resource }) {
  const [updateResource] = useUpdateResourceMutationMutation();

  const [initialState, setInitialState] = useState<ResourceFormState>({
    name: resource.name ?? "",
    resourceURI: resource.resourceURI,
  });

  const form = useSimpleForm<ResourceFormState, Resource>({
    defaultState: initialState,
    submit: async (s) => {
      const state = sanitizeFormState(s);
      const result = await updateResource({
        variables: {
          input: {
            name: state.name,
            resourceURI: state.resourceURI,
          },
        },
      });
      if (result.data == null) {
        throw new Error("unexpected null data");
      }
      setInitialState(state);
      return result.data.updateResource.resource;
    },
    stateMode: "UpdateInitialStateWithUseEffect",
  });
  return (
    <FormContainerBase form={form}>
      <ResourceForm
        className="justify-self-stretch"
        mode="edit"
        state={form.state}
        setState={form.setState}
      />
    </FormContainerBase>
  );
}

function APIResourceScopesTab() {
  return <div>TODO</div>;
}

function APIResourceApplicationsTab() {
  return <div>TODO</div>;
}

function APIResourceTestTab() {
  return <div>TODO</div>;
}

function APIResourceDetailsContent({ resource }: { resource: Resource }) {
  const { selectedKey, onLinkClick } = usePivotNavigation([
    "details",
    "scopes",
    "applications",
    "test",
  ]);
  const { renderToString } = useContext(MessageContext);
  return (
    <div className="py-6 grid content-start grid-flow-row col-span-8 tablet:col-span-full">
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
        <APIResourceDetailsTab resource={resource} />
      ) : null}
      {selectedKey === "scopes" ? <APIResourceScopesTab /> : null}
      {selectedKey === "applications" ? <APIResourceApplicationsTab /> : null}
      {selectedKey === "test" ? <APIResourceTestTab /> : null}
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
