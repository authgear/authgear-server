import React, { useEffect } from "react";
import {
  FormContainerBase,
  FormContainerBaseProps,
  FormModel,
  useFormContainerBaseContext,
} from "../../../FormContainerBase";
import { useFormTopErrors } from "../../../form";
import { useErrorMessageBarContext } from "../../../ErrorMessageBar";
import { useLoading } from "../../../hook/loading";

export interface RoleAndGroupsFormContainerProps<Form = FormModel>
  extends FormContainerBaseProps<Form> {}

function RoleAndGroupsFormContainerContent({
  children,
}: RoleAndGroupsFormContainerProps) {
  const { onSubmit, isUpdating } = useFormContainerBaseContext();

  useLoading(isUpdating);

  const errors = useFormTopErrors();
  const { setErrors } = useErrorMessageBarContext();
  useEffect(() => {
    setErrors(errors);
  }, [errors, setErrors]);

  return (
    <form onSubmit={onSubmit} noValidate={true}>
      {children}
    </form>
  );
}

export const RoleAndGroupsFormContainer: React.VFC<RoleAndGroupsFormContainerProps> =
  function RoleAndGroupsFormContainer(props) {
    const { form, canSave } = props;
    return (
      <FormContainerBase form={form} canSave={canSave}>
        <RoleAndGroupsFormContainerContent {...props} />
      </FormContainerBase>
    );
  };
