import React, { useEffect, useContext } from "react";
import { useLoading } from "../../hook/loading";
import { useFormContainerBaseContext } from "../../FormContainerBase";
import { useErrorMessageBarContext } from "../../ErrorMessageBar";
import { useFormTopErrors } from "../../form";
import FormTextField from "../../FormTextField";
import PrimaryButton from "../../PrimaryButton";
import { Context as MessageContext } from "@oursky/react-messageformat";

export const CreateScopeForm: React.VFC = function CreateScopeForm() {
  const { renderToString } = useContext(MessageContext);
  const { onSubmit, isUpdating } = useFormContainerBaseContext();
  useLoading(isUpdating);
  const errors = useFormTopErrors();
  const { setErrors } = useErrorMessageBarContext();
  useEffect(() => {
    setErrors(errors);
  }, [errors, setErrors]);

  return (
    <form onSubmit={onSubmit} className="flex items-end max-w-200 gap-x-4">
      <FormTextField
        className="flex-1"
        required={true}
        label={renderToString("CreateScopeForm.scope.label")}
        fieldName="scope"
        parentJSONPointer=""
        type="text"
        placeholder={renderToString("CreateScopeForm.scope.placeholder")}
      />
      <FormTextField
        className="flex-1"
        label={renderToString("CreateScopeForm.description.label")}
        fieldName="description"
        parentJSONPointer=""
        type="text"
        placeholder={renderToString("CreateScopeForm.description.placeholder")}
      />
      <PrimaryButton
        className="flex-none"
        type="submit"
        text={renderToString("CreateScopeForm.add.button")}
        iconProps={{ iconName: "Add" }}
      />
    </form>
  );
};
