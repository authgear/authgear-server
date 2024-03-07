import React, { useCallback, useContext, useMemo } from "react";
import { useSimpleForm } from "../../hook/useSimpleForm";
import {
  FormContainerBase,
  useFormContainerBaseContext,
} from "../../FormContainerBase";
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
import { useCreateRoleMutation } from "./mutations/createRoleMutation";
import { APIError } from "../../error/error";
import { makeLocalValidationError } from "../../error/validation";
import { sanitizeRole, validateRole } from "../../model/role";
import { useNavigate } from "react-router-dom";

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
  const navigate = useNavigate();

  const { createRole } = useCreateRoleMutation();

  const afterSave = useCallback(() => {
    // TODO(tung): Should navigate to the edit screen of the created role
    navigate("..", { replace: true });
  }, [navigate]);

  const cancel = useCallback(() => {
    navigate("..", { replace: true });
  }, [navigate]);

  const validate = useCallback((rawState: FormState): APIError | null => {
    const errors = validateRole({
      key: rawState.roleKey,
      name: rawState.roleName,
      description: rawState.roleDescription,
    });
    if (errors.length > 0) {
      return makeLocalValidationError(errors);
    }
    return null;
  }, []);

  const submit = useCallback(
    async (rawState: FormState) => {
      const sanitizedRole = sanitizeRole({
        key: rawState.roleKey,
        name: rawState.roleName,
        description: rawState.roleDescription,
      });
      return createRole(
        sanitizedRole.key,
        sanitizedRole.name,
        sanitizedRole.description
      );
    },
    [createRole]
  );

  const form = useSimpleForm({
    stateMode:
      "ConstantInitialStateAndResetCurrentStatetoInitialStateAfterSave",
    defaultState,
    submit,
    validate,
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
    <FormContainerBase form={form} afterSave={afterSave}>
      <RoleAndGroupsFormLayout
        breadcrumbs={breadcrumbs}
        Footer={<AddRolesScreenFooter onCancel={cancel} />}
      >
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

function AddRolesScreenFooter({ onCancel }: { onCancel: () => void }) {
  const { canSave, isUpdating } = useFormContainerBaseContext();

  return (
    <>
      <PrimaryButton
        disabled={!canSave || isUpdating}
        type="submit"
        text={<FormattedMessage id="create" />}
      />
      <DefaultButton
        disabled={isUpdating}
        type="button"
        onClick={onCancel}
        text={<FormattedMessage id="cancel" />}
      />
    </>
  );
}
