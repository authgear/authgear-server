import React, { useState } from "react";
import { useUpdateResourceMutationMutation } from "../../graphql/adminapi/mutations/updateResourceMutation.generated";
import { useSimpleForm } from "../../hook/useSimpleForm";
import { FormContainerBase } from "../../FormContainerBase";
import WidgetTitle from "../../WidgetTitle";
import { FormattedMessage } from "../../intl";
import { Resource } from "../../graphql/adminapi/globalTypes.generated";
import {
  ResourceForm,
  ResourceFormState,
  sanitizeFormState,
} from "../../components/api-resources/ResourceForm";

export function APIResourceDetailsScreenDetailsTab({
  resource,
}: {
  resource: Resource;
}): JSX.Element {
  const [updateResource] = useUpdateResourceMutationMutation();

  const [initialState, setInitialState] = useState<ResourceFormState>({
    name: resource.name ?? "",
    resourceURI: resource.resourceURI,
  });

  const form = useSimpleForm<ResourceFormState, null>({
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
      return null;
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
