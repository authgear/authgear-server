import React, { useCallback, useContext, useEffect, useMemo } from "react";
import { SimpleFormModel, useSimpleForm } from "../../hook/useSimpleForm";
import { useFormContainerBaseContext } from "../../FormContainerBase";
import {
  RoleAndGroupsFormFooter,
  RoleAndGroupsLayout,
  RoleAndGroupsVeriticalFormLayout,
} from "../../RoleAndGroupsLayout";
import { BreadcrumbItem } from "../../NavBreadcrumb";
import { Context as MessageContext, FormattedMessage } from "../../intl";
import { TextField } from "../../components/v2/TextField/TextField";
import { TextArea } from "../../components/v2/TextArea/TextArea";
import { PrimaryButton } from "../../components/v2/Button/PrimaryButton/PrimaryButton";
import { SecondaryButton } from "../../components/v2/Button/SecondaryButton/SecondaryButton";
import { useCreateRoleMutation } from "./mutations/createRoleMutation";
import { APIError } from "../../error/error";
import { makeLocalValidationError } from "../../error/validation";
import { generateRoleKeyFromName, validateRole } from "../../model/role";
import { useNavigate } from "react-router-dom";
import { useErrorMessageBarContext } from "../../ErrorMessageBar";
import { useFormTopErrors } from "../../form";
import { RoleAndGroupsFormContainer } from "../../components/roles-and-groups/form/RoleAndGroupsFormContainer";
import { makeReasonErrorParseRule } from "../../error/parse";

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

function AddRolesScreenForm() {
  const {
    form: { state: formState, setState: setFormState },
    canSave,
    isUpdating,
  } = useFormContainerBaseContext<SimpleFormModel<FormState, string | null>>();
  const navigate = useNavigate();

  const errors = useFormTopErrors();
  const { setErrors } = useErrorMessageBarContext();
  useEffect(() => {
    setErrors(errors);
  }, [errors, setErrors]);

  const cancel = useCallback(() => {
    navigate("..", { replace: true });
  }, [navigate]);

  const onFormStateChangeCallbacks = useMemo(() => {
    const createCallback = (key: keyof FormState) => {
      return (
        e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>
      ) => {
        const newValue = e.target.value;
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

  const { renderToString } = useContext(MessageContext);

  const roleKeyFieldErrorRules = useMemo(
    () => [
      makeReasonErrorParseRule(
        "RoleDuplicateKey",
        "errors.roles.key.duplicated"
      ),
    ],
    []
  );

  return (
    <div>
      <RoleAndGroupsVeriticalFormLayout>
        <TextField
          size="2"
          required={true}
          fieldName="name"
          parentJSONPointer=""
          type="text"
          label={renderToString("AddRoleScreen.roleName.title")}
          hint={<FormattedMessage id="AddRoleScreen.roleName.description" />}
          value={formState.roleName}
          onChange={onFormStateChangeCallbacks.roleName}
        />
        <TextField
          size="2"
          errorRules={roleKeyFieldErrorRules}
          required={true}
          fieldName="key"
          parentJSONPointer=""
          placeholder={generateRoleKeyFromName(formState.roleName)}
          type="text"
          label={renderToString("AddRoleScreen.roleKey.title")}
          hint={<FormattedMessage id="AddRoleScreen.roleKey.description" />}
          value={formState.roleKey}
          onChange={onFormStateChangeCallbacks.roleKey}
        />
        <TextArea
          size="2"
          fieldName="description"
          parentJSONPointer=""
          label={renderToString("AddRoleScreen.roleDescription.title")}
          value={formState.roleDescription}
          onChange={onFormStateChangeCallbacks.roleDescription}
        />
      </RoleAndGroupsVeriticalFormLayout>
      <RoleAndGroupsFormFooter className="mt-12">
        <PrimaryButton
          size="2"
          disabled={!canSave || isUpdating}
          type="submit"
          text={<FormattedMessage id="create" />}
        />
        <SecondaryButton
          size="2"
          disabled={isUpdating}
          type="button"
          onClick={cancel}
          text={<FormattedMessage id="cancel" />}
        />
      </RoleAndGroupsFormFooter>
    </div>
  );
}

const AddRoleScreen: React.VFC = function AddRoleScreen() {
  const navigate = useNavigate();

  const { createRole } = useCreateRoleMutation();

  const validate = useCallback((rawState: FormState): APIError | null => {
    const [_, errors] = validateRole({
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
      const [sanitizedRole, errors] = validateRole({
        key: rawState.roleKey,
        name: rawState.roleName,
        description: rawState.roleDescription,
      });
      if (errors.length > 0) {
        throw new Error("unexpected validation errors");
      }
      return createRole({
        key: sanitizedRole.key,
        name: sanitizedRole.name,
        description: sanitizedRole.description,
      });
    },
    [createRole]
  );

  const form = useSimpleForm({
    defaultState,
    submit,
    validate,
  });

  const canSave = useMemo(
    () => form.state.roleName !== "",
    [form.state.roleName]
  );

  const breadcrumbs = useMemo<BreadcrumbItem[]>(() => {
    return [
      {
        to: "~/user-management/roles",
        label: <FormattedMessage id="RolesScreen.title" />,
      },
      { to: ".", label: <FormattedMessage id="AddRoleScreen.title" /> },
    ];
  }, []);

  useEffect(() => {
    if (form.submissionResult != null) {
      const roleID = form.submissionResult;
      navigate(`../${encodeURIComponent(roleID)}/details`, { replace: true });
    }
  }, [form.submissionResult, navigate]);

  return (
    <RoleAndGroupsLayout headerBreadcrumbs={breadcrumbs}>
      <RoleAndGroupsFormContainer form={form} canSave={canSave}>
        <AddRolesScreenForm />
      </RoleAndGroupsFormContainer>
    </RoleAndGroupsLayout>
  );
};

export default AddRoleScreen;
