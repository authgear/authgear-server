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
import { SettingsSectionCard } from "../../v2/SettingsSectionCard/SettingsSectionCard";
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
      groupKey: createCallback("groupKey"),
      groupName: createCallback("groupName"),
      groupDescription: createCallback("groupDescription"),
    };
  }, [setFormState]);

  return (
    <div>
      <SettingsSectionCard
        title={<FormattedMessage id="GroupDetailsScreen.tabs.settings" />}
      >
        <RoleAndGroupsVeriticalFormLayout>
          <TextField
            size="2"
            required={true}
            fieldName="name"
            parentJSONPointer=""
            type="text"
            label={renderToString("GroupDetailsSettingsForm.groupName.title")}
            hint={
              <FormattedMessage id="GroupDetailsSettingsForm.groupName.description" />
            }
            value={formState.groupName}
            onChange={onFormStateChangeCallbacks.groupName}
          />
          <TextField
            size="2"
            required={true}
            fieldName="key"
            parentJSONPointer=""
            type="text"
            label={renderToString("GroupDetailsSettingsForm.groupKey.title")}
            hint={
              <FormattedMessage id="GroupDetailsSettingsForm.groupKey.description" />
            }
            placeholder={generateGroupKeyFromName(formState.groupName)}
            value={formState.groupKey}
            onChange={onFormStateChangeCallbacks.groupKey}
          />
          <TextArea
            size="2"
            fieldName="description"
            parentJSONPointer=""
            label={renderToString(
              "GroupDetailsSettingsForm.groupDescription.title"
            )}
            value={formState.groupDescription}
            onChange={onFormStateChangeCallbacks.groupDescription}
          />
        </RoleAndGroupsVeriticalFormLayout>
      </SettingsSectionCard>

      <RoleAndGroupsFormFooter className="mt-8">
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
          onClick={onClickDeleteGroup}
        >
          <FormattedMessage id="GroupDetailsSettingsForm.button.deleteGroup" />
        </Button>
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
      return { result: undefined };
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

  const form = useFormWithExternalInitialState({
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
