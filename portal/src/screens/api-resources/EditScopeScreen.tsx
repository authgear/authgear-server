import React, { useEffect, useMemo } from "react";
import {
  useParams,
  useNavigate,
  useLocation,
  createPath,
} from "react-router-dom";
import { FormattedMessage } from "../../intl";
import {
  ScopeForm,
  ScopeFormState,
  sanitizeScopeFormState,
} from "../../components/api-resources/ScopeForm";
import { useSimpleForm } from "../../hook/useSimpleForm";
import { FormContainerBase } from "../../FormContainerBase";
import { useUpdateScopeMutationMutation } from "../../graphql/adminapi/mutations/updateScopeMutation.generated";
import { Resource, Scope } from "../../graphql/adminapi/globalTypes.generated";
import APIResourceScreenLayout from "../../components/api-resources/APIResourceScreenLayout";
import { useScopeQueryQuery } from "../../graphql/adminapi/query/scopeQuery.generated";
import { useResourceQueryQuery } from "../../graphql/adminapi/query/resourceQuery.generated";
import { useLoadableView } from "../../hook/useLoadableView";

function EditScopeScreenContent({
  scope,
  resource,
}: {
  resource: Resource;
  scope: Scope;
}) {
  const { appID, resourceID } = useParams<{
    appID: string;
    resourceID: string;
    scopeID: string;
  }>();
  const navigate = useNavigate();
  const [updateScope] = useUpdateScopeMutationMutation();
  const location = useLocation();

  const initialState: ScopeFormState = useMemo(
    () => ({
      scope: scope.scope,
      description: scope.description ?? "",
    }),
    [scope]
  );

  const backURL = createPath({
    pathname: `/project/${appID}/api-resources/${encodeURIComponent(
      resourceID ?? ""
    )}`,
    hash: location.hash,
    search: location.search,
  });

  const form = useSimpleForm<ScopeFormState, Scope>({
    defaultState: initialState,
    submit: async (s) => {
      const state = sanitizeScopeFormState(s);
      const result = await updateScope({
        variables: {
          input: {
            resourceURI: resource.resourceURI,
            scope: state.scope,
            description: state.description,
          },
        },
      });
      if (result.data == null) {
        throw new Error("unexpected null data");
      }
      return result.data.updateScope.scope;
    },
    stateMode:
      "ConstantInitialStateAndResetCurrentStatetoInitialStateAfterSave",
  });

  useEffect(() => {
    if (form.isSubmitted) {
      navigate(backURL);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [form.isSubmitted]);

  return (
    <APIResourceScreenLayout
      breadcrumbItems={[
        {
          to: "~/api-resources",
          label: <FormattedMessage id="ScreenNav.api-resources" />,
        },
        {
          to: backURL,
          label: resource.name ?? resource.resourceURI,
        },
        {
          to: "",
          label: <FormattedMessage id="EditScopeScreen.title" />,
        },
      ]}
    >
      <FormContainerBase form={form}>
        <ScopeForm
          className="col-span-8 tablet:col-span-full py-6"
          mode="edit"
          state={form.state}
          setState={form.setState}
        />
      </FormContainerBase>
    </APIResourceScreenLayout>
  );
}

const EditScopeScreen: React.VFC = function EditScopeScreen() {
  const { resourceID, scopeID } = useParams<{
    appID: string;
    resourceID: string;
    scopeID: string;
  }>();
  const {
    data: scopeData,
    loading: scopeLoading,
    error: scopeError,
    refetch: scopeRefetch,
  } = useScopeQueryQuery({
    variables: { id: scopeID ?? "" },
  });
  const {
    data: resourceData,
    loading: resourceLoading,
    error: resourceError,
    refetch: resourceRefetch,
  } = useResourceQueryQuery({
    variables: { id: resourceID ?? "" },
  });

  return useLoadableView({
    loadables: [
      {
        isLoading: resourceLoading,
        loadError: resourceError,
        reload: resourceRefetch,
        data: resourceData,
      },
      {
        isLoading: scopeLoading,
        loadError: scopeError,
        reload: scopeRefetch,
        data: scopeData,
      },
    ] as const,
    render: ([resourceQuery, scopeQuery]) => {
      const { data: resourceData } = resourceQuery;
      const { data: scopeData } = scopeQuery;
      const resource =
        resourceData?.node?.__typename === "Resource"
          ? resourceData.node
          : null;
      const scope =
        scopeData?.node?.__typename === "Scope" ? scopeData.node : null;

      return <EditScopeScreenContent resource={resource!} scope={scope!} />;
    },
  });
};

export default EditScopeScreen;
