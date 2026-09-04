import React, { useRef, useState } from "react";
import { useUpdateResourceMutationMutation } from "../../graphql/adminapi/mutations/updateResourceMutation.generated";
import { useFormWithExternalInitialState } from "../../hook/useFormWithExternalInitialState";
import { FormContainerBase } from "../../FormContainerBase";
import { FormattedMessage } from "../../intl";
import { Resource } from "../../graphql/adminapi/globalTypes.generated";
import {
  ResourceForm,
  ResourceFormState,
  sanitizeFormState,
} from "../../components/api-resources/ResourceForm";
import { SaveFunctionBar } from "../../components/v2/SaveFunctionBar/SaveFunctionBar";
import { SettingsSectionCard } from "../../components/v2/SettingsSectionCard/SettingsSectionCard";
import styles from "./APIResourceDetailsDetailsSection.module.css";

export function APIResourceDetailsScreenDetailsSection({
  resource,
}: {
  resource: Resource;
}): JSX.Element {
  const contentWidthAnchorRef = useRef<HTMLDivElement>(null);
  const [updateResource] = useUpdateResourceMutationMutation();

  const [initialState] = useState<ResourceFormState>({
    name: resource.name ?? "",
    resourceURI: resource.resourceURI,
  });

  const form = useFormWithExternalInitialState<ResourceFormState, null>({
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
      return { result: null, nextInitialState: state };
    },
  });

  return (
    <FormContainerBase form={form} canSave={form.state.name.trim() !== ""}>
      <div ref={contentWidthAnchorRef} className={styles.root}>
        <SettingsSectionCard
          title={
            <FormattedMessage id="APIResourceDetailsScreen.section.details" />
          }
        >
          <ResourceForm
            mode="edit"
            state={form.state}
            setState={form.setState}
          />
        </SettingsSectionCard>
        <SaveFunctionBar anchorRef={contentWidthAnchorRef} />
      </div>
    </FormContainerBase>
  );
}
