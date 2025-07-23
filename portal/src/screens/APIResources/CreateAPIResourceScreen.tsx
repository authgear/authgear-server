import React from "react";
import ScreenContent from "../../ScreenContent";
import ScreenContentHeader from "../../ScreenContentHeader";
import { FormattedMessage } from "@oursky/react-messageformat";
import NavBreadcrumb from "../../NavBreadcrumb";
import {
  ResourceForm,
  ResourceFormState,
} from "../../components/api-resources/ResourceForm";
import { useSimpleForm } from "../../hook/useSimpleForm";
import { FormContainerBase } from "../../FormContainerBase";
import { ErrorMessageBarContextProvider } from "../../ErrorMessageBar";

const defaultState: ResourceFormState = {
  name: "",
  resourceURI: "",
};

const CreateAPIResourceScreen: React.VFC = function CreateAPIResourceScreen() {
  const form = useSimpleForm<ResourceFormState>({
    defaultState,
    submit: async () => {
      // TODO
    },
    stateMode:
      "ConstantInitialStateAndResetCurrentStatetoInitialStateAfterSave",
  });

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
