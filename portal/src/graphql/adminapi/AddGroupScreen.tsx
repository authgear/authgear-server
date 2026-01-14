import React, { useCallback, useContext, useEffect, useMemo } from "react";
import {
  RoleAndGroupsFormFooter,
  RoleAndGroupsLayout,
  RoleAndGroupsVeriticalFormLayout,
} from "../../RoleAndGroupsLayout";
import { RoleAndGroupsFormContainer } from "../../components/roles-and-groups/form/RoleAndGroupsFormContainer";
import { BreadcrumbItem } from "../../NavBreadcrumb";
import {
  FormattedMessage,
  Context as MessageContext,
} from "../../intl";
import { SimpleFormModel, useSimpleForm } from "../../hook/useSimpleForm";
import { useNavigate } from "react-router-dom";
import { useCreateGroupMutation } from "./mutations/createGroupMutation";
import { APIError } from "../../error/error";
import { makeLocalValidationError } from "../../error/validation";
import { generateGroupKeyFromName, validateGroup } from "../../model/group";
import { useFormContainerBaseContext } from "../../FormContainerBase";
import { useFormTopErrors } from "../../form";
import { useErrorMessageBarContext } from "../../ErrorMessageBar";
import FormTextField from "../../FormTextField";
import PrimaryButton from "../../PrimaryButton";
import DefaultButton from "../../DefaultButton";
import WidgetDescription from "../../WidgetDescription";
import { makeReasonErrorParseRule } from "../../error/parse";

interface FormState {
  groupKey: string;
  groupName: string;
  groupDescription: string;
}

const defaultState: FormState = {
  groupKey: "",
  groupName: "",
  groupDescription: "",
};

function AddGroupScreenForm() {
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

  const { renderToString } = useContext(MessageContext);

  const groupKeyFieldErrorRules = useMemo(
    () => [
      makeReasonErrorParseRule(
        "GroupDuplicateKey",
        "errors.groups.key.duplicated"
      ),
    ],
    []
  );

  return (
    <div>
      <RoleAndGroupsVeriticalFormLayout>
        <div>
          <FormTextField
            required={true}
            fieldName="name"
            parentJSONPointer=""
            type="text"
            label={renderToString("AddGroupScreen.groupName.title")}
            value={formState.groupName}
            onChange={onFormStateChangeCallbacks.groupName}
          />
          <WidgetDescription className="mt-2">
            <FormattedMessage id="AddGroupScreen.groupName.description" />
          </WidgetDescription>
        </div>
        <div>
          <FormTextField
            errorRules={groupKeyFieldErrorRules}
            required={true}
            fieldName="key"
            parentJSONPointer=""
            placeholder={generateGroupKeyFromName(formState.groupName)}
            type="text"
            label={renderToString("AddGroupScreen.groupKey.title")}
            value={formState.groupKey}
            onChange={onFormStateChangeCallbacks.groupKey}
          />
          <WidgetDescription className="mt-2">
            <FormattedMessage id="AddGroupScreen.groupKey.description" />
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
          label={renderToString("AddGroupScreen.groupDescription.title")}
          value={formState.groupDescription}
          onChange={onFormStateChangeCallbacks.groupDescription}
        />
      </RoleAndGroupsVeriticalFormLayout>
      <RoleAndGroupsFormFooter className="mt-12">
        <PrimaryButton
          disabled={!canSave || isUpdating}
          type="submit"
          text={<FormattedMessage id="create" />}
        />
        <DefaultButton
          disabled={isUpdating}
          type="button"
          onClick={cancel}
          text={<FormattedMessage id="cancel" />}
        />
      </RoleAndGroupsFormFooter>
    </div>
  );
}

const AddGroupScreen: React.VFC = function AddGroupScreen() {
  const breadcrumbs = useMemo<BreadcrumbItem[]>(() => {
    return [
      {
        to: "~/user-management/groups",
        label: <FormattedMessage id="GroupsScreen.title" />,
      },
      { to: ".", label: <FormattedMessage id="AddGroupScreen.title" /> },
    ];
  }, []);

  const navigate = useNavigate();

  const { createGroup } = useCreateGroupMutation();

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
      return createGroup({
        key: sanitizedGroup.key,
        name: sanitizedGroup.name,
        description: sanitizedGroup.description,
      });
    },
    [createGroup]
  );

  const form = useSimpleForm({
    stateMode:
      "ConstantInitialStateAndResetCurrentStatetoInitialStateAfterSave",
    defaultState,
    submit,
    validate,
  });

  const canSave = useMemo(
    () => form.state.groupName !== "",
    [form.state.groupName]
  );

  useEffect(() => {
    if (form.submissionResult != null) {
      const groupID = form.submissionResult;
      navigate(`../${encodeURIComponent(groupID)}/details`, { replace: true });
    }
  }, [form.submissionResult, navigate]);

  return (
    <RoleAndGroupsLayout headerBreadcrumbs={breadcrumbs}>
      <RoleAndGroupsFormContainer form={form} canSave={canSave}>
        <AddGroupScreenForm />
      </RoleAndGroupsFormContainer>
    </RoleAndGroupsLayout>
  );
};

export default AddGroupScreen;
