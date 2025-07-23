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
import { ErrorMessageBarContextProvider } from "../../ErrorMessageBar";
import { useCreateResourceMutationMutation } from "../../graphql/adminapi/mutations/createResourceMutation.generated";

const defaultState: ResourceFormState = {
  name: "",
  resourceURI: "",
};

const CreateAPIResourceScreen: React.VFC = function CreateAPIResourceScreen() {
  const [createResource] = useCreateResourceMutationMutation();

  const form = useSimpleForm<ResourceFormState>({
    defaultState,
    submit: async (s) => {
      const state = sanitizeFormState(s);
      await createResource({
        variables: {
          input: {
            name: state.name,
            resourceURI: state.resourceURI,
          },
        },
      });
    },
    stateMode:
      "ConstantInitialStateAndResetCurrentStatetoInitialStateAfterSave",
  });

  useEffect(() => {
    // TODO: Go to edit screen
  }, [form.isSubmitted]);

  return (
    <ErrorMessageBarContextProvider>
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
        <FormContainerBase form={form}>
          <ResourceForm
            className="col-span-8 tablet:col-span-full"
            state={form.state}
            setState={form.setState}
          />
        </FormContainerBase>
      </ScreenContent>
    </ErrorMessageBarContextProvider>
  );
};

export default CreateAPIResourceScreen;
