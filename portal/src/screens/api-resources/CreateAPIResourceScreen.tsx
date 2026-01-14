import React, { useEffect } from "react";
import { FormattedMessage } from "../../intl";
import {
  ResourceForm,
  ResourceFormState,
  sanitizeFormState,
} from "../../components/api-resources/ResourceForm";
import { useSimpleForm } from "../../hook/useSimpleForm";
import { FormContainerBase } from "../../FormContainerBase";
import { useCreateResourceMutationMutation } from "../../graphql/adminapi/mutations/createResourceMutation.generated";
import { makeReasonErrorParseRule } from "../../error/parse";
import { useNavigate, useParams } from "react-router-dom";
import APIResourceScreenLayout from "../../components/api-resources/APIResourceScreenLayout";

const defaultState: ResourceFormState = {
  name: "",
  resourceURI: "",
};

const errorRules = [
  makeReasonErrorParseRule(
    "ResourceDuplicateURI",
    "errors.resources.duplicateURI"
  ),
];

const CreateAPIResourceScreen: React.VFC = function CreateAPIResourceScreen() {
  const [createResource] = useCreateResourceMutationMutation();
  const navigate = useNavigate();
  const { appID } = useParams<{ appID: string }>();

  const form = useSimpleForm<ResourceFormState, string>({
    defaultState,
    submit: async (s) => {
      const state = sanitizeFormState(s);
      const result = await createResource({
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
      return result.data.createResource.resource.id;
    },
    stateMode:
      "ConstantInitialStateAndResetCurrentStatetoInitialStateAfterSave",
  });

  useEffect(() => {
    if (form.submissionResult && appID) {
      navigate(
        `/project/${appID}/api-resources/${encodeURIComponent(
          form.submissionResult
        )}`
      );
    }
  }, [form.isSubmitted, navigate, form.submissionResult, appID]);

  return (
    <APIResourceScreenLayout
      breadcrumbItems={[
        {
          to: "~/api-resources",
          label: <FormattedMessage id="ScreenNav.api-resources" />,
        },
        {
          to: "",
          label: <FormattedMessage id="CreateAPIResourceScreen.title" />,
        },
      ]}
    >
      <FormContainerBase form={form} errorRules={errorRules}>
        <ResourceForm
          className="col-span-8 tablet:col-span-full py-6"
          mode="create"
          state={form.state}
          setState={form.setState}
        />
      </FormContainerBase>
    </APIResourceScreenLayout>
  );
};

export default CreateAPIResourceScreen;
