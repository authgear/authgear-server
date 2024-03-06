import React, { useCallback, useMemo } from "react";
import { useSimpleForm } from "../../hook/useSimpleForm";
import { FormContainerBase } from "../../FormContainerBase";
import { RoleAndGroupsFormLayout } from "../../RoleAndGroupsFormLayout";
import { BreadcrumbItem } from "../../NavBreadcrumb";
import { FormattedMessage } from "@oursky/react-messageformat";

interface FormState {}

const defaultState: FormState = {};

const AddRolesScreen: React.VFC = function AddRolesScreen() {
  const submit = useCallback(async () => {}, []);

  const form = useSimpleForm({
    stateMode:
      "ConstantInitialStateAndResetCurrentStatetoInitialStateAfterSave",
    defaultState,
    submit,
  });

  const breadcrumbs = useMemo<BreadcrumbItem[]>(() => {
    return [
      {
        to: "~/user-management/roles",
        label: <FormattedMessage id="RolesScreen.title" />,
      },
      { to: ".", label: <FormattedMessage id="AddRolesScreen.title" /> },
    ];
  }, []);

  return (
    <FormContainerBase form={form}>
      <RoleAndGroupsFormLayout
        breadcrumbs={breadcrumbs}
      ></RoleAndGroupsFormLayout>
    </FormContainerBase>
  );
};

export default AddRolesScreen;
