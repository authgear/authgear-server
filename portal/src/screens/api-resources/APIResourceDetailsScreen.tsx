import React, { useContext, useState, useMemo } from "react";
import { useParams } from "react-router-dom";
import { useResourceQueryQuery } from "../../graphql/adminapi/query/resourceQuery.generated";
import { useLoadableView } from "../../hook/useLoadableView";
import {
  FormattedMessage,
  Context as MessageContext,
} from "@oursky/react-messageformat";
import APIResourceScreenLayout from "../../components/api-resources/APIResourceScreenLayout";
import { Resource, Scope } from "../../graphql/adminapi/globalTypes.generated";
import { Pivot, PivotItem, Text } from "@fluentui/react";
import { usePivotNavigation } from "../../hook/usePivot";
import { useUpdateResourceMutationMutation } from "../../graphql/adminapi/mutations/updateResourceMutation.generated";
import { useSimpleForm } from "../../hook/useSimpleForm";
import { FormContainerBase } from "../../FormContainerBase";
import {
  ResourceForm,
  ResourceFormState,
  sanitizeFormState,
} from "../../components/api-resources/ResourceForm";
import WidgetTitle from "../../WidgetTitle";
import {
  CreateScopeForm,
  CreateScopeFormState,
  sanitizeCreateScopeFormState,
} from "../../components/api-resources/CreateScopeForm";
import { useCreateScopeMutationMutation } from "../../graphql/adminapi/mutations/createScopeMutation.generated";
import {
  useResourceScopesQueryQuery,
  ResourceScopesQueryDocument,
} from "../../graphql/adminapi/query/resourceScopesQuery.generated";
import { ScopeList } from "../../components/api-resources/ScopeList";
import { encodeOffsetToCursor } from "../../util/pagination";
import ShowError from "../../ShowError";

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
      <div className="justify-self-stretch py-5 max-w-180">
        <WidgetTitle className="mb-4">
          <FormattedMessage id="APIResourceDetailsScreen.tab.details" />
        </WidgetTitle>
        <ResourceForm mode="edit" state={form.state} setState={form.setState} />
      </div>
    </FormContainerBase>
  );
}

function APIResourceScopesTab({ resource }: { resource: Resource }) {
  const [createScope] = useCreateScopeMutationMutation();
  const [initialState] = useState<CreateScopeFormState>({
    scope: "",
    description: "",
  });
  const form = useSimpleForm<CreateScopeFormState, any>({
    defaultState: initialState,
    submit: async (state) => {
      const sanitized = sanitizeCreateScopeFormState(state);
      await createScope({
        variables: {
          input: {
            resourceURI: resource.resourceURI,
            scope: sanitized.scope,
            description: sanitized.description,
          },
        },
        awaitRefetchQueries: true,
        refetchQueries: [ResourceScopesQueryDocument],
      });
    },
    stateMode:
      "ConstantInitialStateAndResetCurrentStatetoInitialStateAfterSave",
  });

  const [offset, setOffset] = useState(0);
  const pageSize = 10;

  const { data, loading, error, refetch } = useResourceScopesQueryQuery({
    variables: {
      resourceID: resource.id,
      first: pageSize,
      after: offset === 0 ? undefined : encodeOffsetToCursor(offset),
    },
  });

  const scopes = useMemo(() => {
    return data?.node && data.node.__typename === "Resource"
      ? data.node.scopes?.edges
          ?.map((edge) => edge?.node)
          .filter((n): n is Scope => !!n) ?? []
      : [];
  }, [data]);

  const totalCount = useMemo(() => {
    return data?.node && data.node.__typename === "Resource"
      ? data.node.scopes?.totalCount ?? 0
      : 0;
  }, [data]);

  const pagination = useMemo(() => {
    return {
      offset,
      pageSize,
      totalCount,
      onChangeOffset: setOffset,
    };
  }, [offset, pageSize, totalCount, setOffset]);

  if (error != null) {
    return <ShowError error={error} onRetry={refetch} />;
  }

  return (
    <FormContainerBase form={form}>
      <div className="pt-5 flex-1 flex flex-col space-y-2">
        <header>
          <WidgetTitle className="mb-2">
            <FormattedMessage id="APIResourceDetailsScreen.tab.scopes" />
          </WidgetTitle>
          <Text>
            <FormattedMessage id="APIResourceDetailsScreen.scopes.description" />
          </Text>
        </header>
        <div className="flex-1 flex flex-col space-y-8">
          <CreateScopeForm state={form.state} setState={form.setState} />
          {scopes.length > 0 ? (
            <ScopeList
              className="flex-1 min-h-0"
              scopes={scopes}
              loading={loading}
              pagination={pagination}
              onEdit={() => {}}
              onDelete={() => {}}
            />
          ) : null}
        </div>
      </div>
    </FormContainerBase>
  );
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
        <APIResourceDetailsTab resource={resource} />
      ) : null}
      {selectedKey === "scopes" ? (
        <APIResourceScopesTab resource={resource} />
      ) : null}
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
