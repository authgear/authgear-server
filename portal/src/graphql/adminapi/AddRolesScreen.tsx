import React, { useCallback, useContext, useMemo } from "react";
import { useSimpleForm } from "../../hook/useSimpleForm";
import { FormContainerBase } from "../../FormContainerBase";
import { RoleAndGroupsFormLayout } from "../../RoleAndGroupsFormLayout";
import { BreadcrumbItem } from "../../NavBreadcrumb";
import {
  Context as MessageContext,
  FormattedMessage,
} from "@oursky/react-messageformat";
import FormTextField from "../../FormTextField";
import WidgetDescription from "../../WidgetDescription";
import PrimaryButton from "../../PrimaryButton";
import DefaultButton from "../../DefaultButton";

interface FormState {
  roleKey: string;
  roleName: string;
  roleDescription: string;
}

const defaultState: FormState = {
  roleKey: "",
  roleName: "",
  roleDescription: "",
};

const AddRolesScreen: React.VFC = function AddRolesScreen() {
  const { renderToString } = useContext(MessageContext);

  const submit = useCallback(async () => {}, []);

  const form = useSimpleForm({
    stateMode:
      "ConstantInitialStateAndResetCurrentStatetoInitialStateAfterSave",
    defaultState,
    submit,
  });

  const { state: formState, setState: setFormState } = form;

  const breadcrumbs = useMemo<BreadcrumbItem[]>(() => {
    return [
      {
        to: "~/user-management/roles",
        label: <FormattedMessage id="RolesScreen.title" />,
      },
      { to: ".", label: <FormattedMessage id="AddRolesScreen.title" /> },
    ];
  }, []);

  const onFormStateChangeCallbacks = useMemo(() => {
    const createCallback = (key: keyof FormState) => {
      return (e: React.FormEvent<HTMLInputElement | HTMLTextAreaElement>) => {
        const newValue = e.currentTarget.value;
        setFormState((prev) => {
          return { ...prev, [key]: newValue };
        });
      };
    };
    return {
      roleKey: createCallback("roleKey"),
      roleName: createCallback("roleName"),
      roleDescription: createCallback("roleDescription"),
    };
  }, [setFormState]);

  return (
    <FormContainerBase form={form}>
      <RoleAndGroupsFormLayout
        breadcrumbs={breadcrumbs}
        Footer={
          <>
            <PrimaryButton
              type="submit"
              text={<FormattedMessage id="create" />}
            />
            <DefaultButton
              type="button"
              text={<FormattedMessage id="cancel" />}
            />
          </>
        }
      >
        <div>
          <FormTextField
            required={true}
            fieldName="key"
            parentJSONPointer=""
            type="text"
            label={renderToString("AddRolesScreen.roleKey.title")}
            value={formState.roleKey}
            onChange={onFormStateChangeCallbacks.roleKey}
          />
          <WidgetDescription className="mt-2">
            <FormattedMessage id="AddRolesScreen.roleKey.description" />
          </WidgetDescription>
        </div>
        <div>
          <FormTextField
            required={true}
            fieldName="name"
            parentJSONPointer=""
            type="text"
            label={renderToString("AddRolesScreen.roleName.title")}
            value={formState.roleName}
            onChange={onFormStateChangeCallbacks.roleName}
          />
          <WidgetDescription className="mt-2">
            <FormattedMessage id="AddRolesScreen.roleName.description" />
          </WidgetDescription>
        </div>
        <FormTextField
          multiline={true}
          resizable={false}
          autoAdjustHeight={true}
          rows={3}
          fieldName="description"
          parentJSONPointer=""
          type="text"
          label={renderToString("AddRolesScreen.roleDescription.title")}
          value={formState.roleDescription}
          onChange={onFormStateChangeCallbacks.roleDescription}
        />
      </RoleAndGroupsFormLayout>
    </FormContainerBase>
  );
};

export default AddRolesScreen;
