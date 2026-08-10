import React, {
  useCallback,
  useContext,
  useMemo,
  useRef,
  useState,
} from "react";
import { Button } from "@radix-ui/themes";
import { FormattedMessage, Context as MessageContext } from "../../../intl";
import { useFormContainerBaseContext } from "../../../FormContainerBase";
import { SimpleFormModel } from "../../../hook/useSimpleForm";
import { useFormWithExternalInitialState } from "../../../hook/useFormWithExternalInitialState";
import {
  RoleAndGroupsFormFooter,
  RoleAndGroupsVeriticalFormLayout,
} from "../../../RoleAndGroupsLayout";
import { TextField } from "../../v2/TextField/TextField";
import { TextArea } from "../../v2/TextArea/TextArea";
import { PrimaryButton } from "../../v2/Button/PrimaryButton/PrimaryButton";
import { useNavigate, useParams } from "react-router-dom";
import { useUpdateRoleMutation } from "../../../graphql/adminapi/mutations/updateRoleMutation";
import { APIError } from "../../../error/error";
import { generateRoleKeyFromName, validateRole } from "../../../model/role";
import { makeLocalValidationError } from "../../../error/validation";
import { RoleQueryNodeFragment } from "../../../graphql/adminapi/query/roleQuery.generated";
import DeleteRoleDialog, {
  DeleteRoleDialogData,
} from "../dialog/DeleteRoleDialog";
import { RoleAndGroupsFormContainer } from "./RoleAndGroupsFormContainer";

interface FormState {
  roleKey: string;
  roleName: string;
  roleDescription: string;
}

function RoleDetailsSettingsFormContent({
  onClickDeleteRole,
}: {
  onClickDeleteRole: () => void;
}) {
  const { renderToString } = useContext(MessageContext);

  const {
    form: { state: formState, setState: setFormState },
    isUpdating,
    canSave,
  } = useFormContainerBaseContext<SimpleFormModel<FormState, string | null>>();

  const onFormStateChangeCallbacks = useMemo(() => {
    const createCallback = (key: keyof FormState) => {
      return (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
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
          required={true}
          fieldName="key"
          parentJSONPointer=""
          type="text"
          label={renderToString("AddRoleScreen.roleKey.title")}
          hint={<FormattedMessage id="AddRoleScreen.roleKey.description" />}
          placeholder={generateRoleKeyFromName(formState.roleName)}
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
          text={<FormattedMessage id="save" />}
        />
        <Button
          size="2"
          variant="outline"
          color="red"
          disabled={isUpdating}
          type="button"
          onClick={onClickDeleteRole}
        >
          <FormattedMessage id="RoleDetailsScreen.button.deleteRole" />
        </Button>
      </RoleAndGroupsFormFooter>
    </div>
  );
}

export const RoleDetailsSettingsForm: React.VFC<{
  role: RoleQueryNodeFragment;
}> = function RoleDetailsSettingsForm({ role }) {
  const { appID } = useParams() as { appID: string };
  const { updateRole } = useUpdateRoleMutation();
  const navigate = useNavigate();

  const isDeletedRef = useRef(false);

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
      await updateRole({
        id: role.id,
        key: sanitizedRole.key,
        name: sanitizedRole.name,
        description: sanitizedRole.description,
      });
      return { result: undefined };
    },
    [role.id, updateRole]
  );

  const defaultState = useMemo((): FormState => {
    return {
      roleKey: role.key,
      roleName: role.name ?? "",
      roleDescription: role.description ?? "",
    };
  }, [role]);

  const form = useFormWithExternalInitialState({
    defaultState,
    submit,
    validate,
  });

  const canSave = useMemo(
    () => form.state.roleName !== "",
    [form.state.roleName]
  );

  const [deleteRoleDialogData, setDeleteRoleDialogData] =
    useState<DeleteRoleDialogData | null>(null);
  const onClickDeleteRole = useCallback(() => {
    setDeleteRoleDialogData({
      roleID: role.id,
      roleKey: role.key,
      roleName: role.name ?? null,
    });
  }, [role]);
  const dismissDeleteRoleDialog = useCallback((isDeleted: boolean) => {
    setDeleteRoleDialogData(null);
    isDeletedRef.current = isDeleted;
  }, []);

  const exitIfDeleted = useCallback(() => {
    if (isDeletedRef.current) {
      navigate(`/project/${appID}/user-management/roles`, { replace: true });
    }
  }, [navigate, appID]);

  return (
    <>
      <RoleAndGroupsFormContainer form={form} canSave={canSave}>
        <RoleDetailsSettingsFormContent onClickDeleteRole={onClickDeleteRole} />
      </RoleAndGroupsFormContainer>

      <DeleteRoleDialog
        onDismiss={dismissDeleteRoleDialog}
        onDismissed={exitIfDeleted}
        data={deleteRoleDialogData}
      />
    </>
  );
};
