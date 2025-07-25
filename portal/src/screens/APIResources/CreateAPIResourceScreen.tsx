import React, { useEffect } from "react";
import ScreenContent from "../../ScreenContent";
import ScreenContentHeader from "../../ScreenContentHeader";
import { FormattedMessage } from "@oursky/react-messageformat";
import NavBreadcrumb from "../../NavBreadcrumb";
import {
  ResourceForm,
  ResourceFormState,
  sanitizeFormState,
} from "../../components/api-resources/ResourceForm";
import { useSimpleForm } from "../../hook/useSimpleForm";
import { FormContainerBase } from "../../FormContainerBase";
import {
  ErrorMessageBar,
  ErrorMessageBarContextProvider,
} from "../../ErrorMessageBar";
import { useCreateResourceMutationMutation } from "../../graphql/adminapi/mutations/createResourceMutation.generated";
import { makeReasonErrorParseRule } from "../../error/parse";
import { useNavigate, useParams } from "react-router-dom";
import { Resource } from "../../graphql/adminapi/globalTypes.generated";

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

  const form = useSimpleForm<ResourceFormState, Resource>({
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
      return result.data.createResource.resource;
    },
    stateMode:
      "ConstantInitialStateAndResetCurrentStatetoInitialStateAfterSave",
  });

  useEffect(() => {
    if (form.submissionResult?.id && appID) {
      navigate(
        `/project/${appID}/api-resources/${encodeURIComponent(
          form.submissionResult.id
        )}`
      );
    }
  }, [form.isSubmitted, navigate, form.submissionResult, appID]);

  return (
    <ErrorMessageBarContextProvider>
      <div className="flex-1 flex flex-col">
        <ErrorMessageBar />
        <ScreenContent className="flex-1" layout="list">
          <ScreenContentHeader
            title={
              <NavBreadcrumb
                items={[
                  {
                    to: "~/api-resources",
                    label: <FormattedMessage id="APIResourcesScreen.title" />,
                  },
                  {
                    to: "",
                    label: (
                      <FormattedMessage id="CreateAPIResourceScreen.title" />
                    ),
                  },
                ]}
              />
            }
          />
          <FormContainerBase form={form} errorRules={errorRules}>
            <ResourceForm
              className="col-span-8 tablet:col-span-full"
              state={form.state}
              setState={form.setState}
            />
          </FormContainerBase>
        </ScreenContent>
      </div>
    </ErrorMessageBarContextProvider>
  );
};

export default CreateAPIResourceScreen;
