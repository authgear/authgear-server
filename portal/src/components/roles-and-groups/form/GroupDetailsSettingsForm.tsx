import React, {
  useCallback,
  useContext,
  useMemo,
  useRef,
  useState,
} from "react";
import { useSystemConfig } from "../../../context/SystemConfigContext";
import {
  FormattedMessage,
  Context as MessageContext,
} from "@oursky/react-messageformat";
import { useFormContainerBaseContext } from "../../../FormContainerBase";
import { SimpleFormModel, useSimpleForm } from "../../../hook/useSimpleForm";
import {
  RoleAndGroupsFormFooter,
  RoleAndGroupsVeriticalFormLayout,
} from "../../../RoleAndGroupsLayout";
import FormTextField from "../../../FormTextField";
import WidgetDescription from "../../../WidgetDescription";
import PrimaryButton from "../../../PrimaryButton";
import DefaultButton from "../../../DefaultButton";
import { useNavigate, useParams } from "react-router-dom";
import { useUpdateGroupMutation } from "../../../graphql/adminapi/mutations/updateGroupMutation";
import { APIError } from "../../../error/error";
import { generateGroupKeyFromName, validateGroup } from "../../../model/group";
import { makeLocalValidationError } from "../../../error/validation";
import { GroupQueryNodeFragment } from "../../../graphql/adminapi/query/groupQuery.generated";
import DeleteGroupDialog, {
  DeleteGroupDialogData,
} from "../dialog/DeleteGroupDialog";
import { RoleAndGroupsFormContainer } from "./RoleAndGroupsFormContainer";

interface FormState {
  groupKey: string;
  groupName: string;
  groupDescription: string;
}

function GroupDetailsSettingsFormContent({
  onClickDeleteGroup,
}: {
  onClickDeleteGroup: () => void;
}) {
  const { themes } = useSystemConfig();
  const { renderToString } = useContext(MessageContext);

  const {
    form: { state: formState, setState: setFormState },
    isUpdating,
    canSave,
  } = useFormContainerBaseContext<SimpleFormModel<FormState, string | null>>();

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
      groupKey: createCallback("groupKey"),
      groupName: createCallback("groupName"),
      groupDescription: createCallback("groupDescription"),
    };
  }, [setFormState]);

  return (
    <div>
      <RoleAndGroupsVeriticalFormLayout>
        <div>
          <FormTextField
            required={true}
            fieldName="name"
            parentJSONPointer=""
            type="text"
            label={renderToString("GroupDetailsSettingsForm.groupName.title")}
            value={formState.groupName}
            onChange={onFormStateChangeCallbacks.groupName}
          />
          <WidgetDescription className="mt-2">
            <FormattedMessage id="GroupDetailsSettingsForm.groupName.description" />
          </WidgetDescription>
        </div>
        <div>
          <FormTextField
            required={true}
            fieldName="key"
            parentJSONPointer=""
            type="text"
            label={renderToString("GroupDetailsSettingsForm.groupKey.title")}
            placeholder={generateGroupKeyFromName(formState.groupName)}
            value={formState.groupKey}
            onChange={onFormStateChangeCallbacks.groupKey}
          />
          <WidgetDescription className="mt-2">
            <FormattedMessage id="GroupDetailsSettingsForm.groupKey.description" />
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
          label={renderToString(
            "GroupDetailsSettingsForm.groupDescription.title"
          )}
          value={formState.groupDescription}
          onChange={onFormStateChangeCallbacks.groupDescription}
        />
      </RoleAndGroupsVeriticalFormLayout>

      <RoleAndGroupsFormFooter className="mt-12">
        <PrimaryButton
          disabled={!canSave || isUpdating}
          type="submit"
          text={<FormattedMessage id="save" />}
        />
        <DefaultButton
          disabled={isUpdating}
          theme={themes.destructive}
          type="button"
          onClick={onClickDeleteGroup}
          text={
            <FormattedMessage id="GroupDetailsSettingsForm.button.deleteGroup" />
          }
        />
      </RoleAndGroupsFormFooter>
    </div>
  );
}

export const GroupDetailsSettingsForm: React.VFC<{
  group: GroupQueryNodeFragment;
}> = function GroupDetailsSettingsForm({ group }) {
  const { appID } = useParams() as { appID: string };
  const { updateGroup } = useUpdateGroupMutation();
  const navigate = useNavigate();

  const isDeletedRef = useRef(false);

  const validate = useCallback((rawState: FormState): APIError | null => {
    const [_, errors] = validateGroup({
      key: rawState.groupKey,
      name: rawState.groupName,
      description: rawState.groupDescription,
    });
    if (errors.length > 0) {
      return makeLocalValidationError(errors);
    }
    return null;
  }, []);

  const submit = useCallback(
    async (rawState: FormState) => {
      const [sanitizedGroup, errors] = validateGroup({
        key: rawState.groupKey,
        name: rawState.groupName,
        description: rawState.groupDescription,
      });
      if (errors.length > 0) {
        throw new Error("unexpected validation errors");
      }
      await updateGroup({
        id: group.id,
        key: sanitizedGroup.key,
        name: sanitizedGroup.name,
        description: sanitizedGroup.description,
      });
    },
    [group.id, updateGroup]
  );

  const defaultState = useMemo((): FormState => {
    return {
      groupKey: group.key,
      groupName: group.name ?? "",
      groupDescription: group.description ?? "",
    };
  }, [group]);

  const form = useSimpleForm({
    stateMode: "UpdateInitialStateWithUseEffect",
    defaultState,
    submit,
    validate,
  });

  const canSave = useMemo(
    () => form.state.groupName !== "",
    [form.state.groupName]
  );

  const [deleteGroupDialogData, setDeleteGroupDialogData] =
    useState<DeleteGroupDialogData | null>(null);
  const onClickDeleteGroup = useCallback(() => {
    setDeleteGroupDialogData({
      groupID: group.id,
      groupKey: group.key,
      groupName: group.name ?? null,
    });
  }, [group]);
  const dismissDeleteRoleDialog = useCallback((isDeleted: boolean) => {
    setDeleteGroupDialogData(null);
    isDeletedRef.current = isDeleted;
  }, []);

  const exitIfDeleted = useCallback(() => {
    if (isDeletedRef.current) {
      navigate(`/project/${appID}/user-management/groups`, { replace: true });
    }
  }, [navigate, appID]);

  return (
    <>
      <RoleAndGroupsFormContainer form={form} canSave={canSave}>
        <GroupDetailsSettingsFormContent
          onClickDeleteGroup={onClickDeleteGroup}
        />
      </RoleAndGroupsFormContainer>

      <DeleteGroupDialog
        onDismiss={dismissDeleteRoleDialog}
        onDismissed={exitIfDeleted}
        data={deleteGroupDialogData}
      />
    </>
  );
};
