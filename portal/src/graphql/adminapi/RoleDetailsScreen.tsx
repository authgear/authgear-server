import React, { useCallback, useContext, useMemo } from "react";
import {
  RoleAndGroupsLayout,
  RoleAndGroupsVeriticalFormLayout,
} from "../../RoleAndGroupsLayout";
import {
  FormContainerBase,
  useFormContainerBaseContext,
} from "../../FormContainerBase";
import { BreadcrumbItem } from "../../NavBreadcrumb";
import {
  Context as MessageContext,
  FormattedMessage,
} from "@oursky/react-messageformat";
import { useParams } from "react-router-dom";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import { useRoleQuery } from "./query/roleQuery";
import { RoleQueryNodeFragment } from "./query/roleQuery.generated";
import { validateRole } from "../../model/role";
import { APIError } from "../../error/error";
import { makeLocalValidationError } from "../../error/validation";
import { SimpleFormModel, useSimpleForm } from "../../hook/useSimpleForm";
import WidgetDescription from "../../WidgetDescription";
import FormTextField from "../../FormTextField";

interface FormState {
  roleKey: string;
  roleName: string;
  roleDescription: string;
}

function RoleDetailsScreenSettingsForm() {
  const { renderToString } = useContext(MessageContext);

  const {
    onSubmit,
    form: { state: formState, setState: setFormState },
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
      roleKey: createCallback("roleKey"),
      roleName: createCallback("roleName"),
      roleDescription: createCallback("roleDescription"),
    };
  }, [setFormState]);

  return (
    <form onSubmit={onSubmit} noValidate={true}>
      <RoleAndGroupsVeriticalFormLayout>
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
      </RoleAndGroupsVeriticalFormLayout>
    </form>
  );
}

function RoleDetailsScreenSettingsFormContainer({
  role,
}: {
  role: RoleQueryNodeFragment;
}) {
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

  const submit = useCallback(async (rawState: FormState) => {
    const [_, errors] = validateRole({
      key: rawState.roleKey,
      name: rawState.roleName,
      description: rawState.roleDescription,
    });
    if (errors.length > 0) {
      throw new Error("unexpected validation errors");
    }
    // TODO: Call api
  }, []);

  const defaultState = useMemo((): FormState => {
    return {
      roleKey: role.key,
      roleName: role.name ?? "",
      roleDescription: role.description ?? "",
    };
  }, [role]);

  const form = useSimpleForm({
    stateMode:
      "ConstantInitialStateAndResetCurrentStatetoInitialStateAfterSave",
    defaultState,
    submit,
    validate,
  });

  return (
    <FormContainerBase form={form}>
      <RoleDetailsScreenSettingsForm />
    </FormContainerBase>
  );
}

const RoleDetailsScreenLoaded: React.VFC<{ role: RoleQueryNodeFragment }> =
  function RoleDetailsScreenLoaded({ role }) {
    const breadcrumbs = useMemo<BreadcrumbItem[]>(() => {
      return [
        {
          to: "~/user-management/roles",
          label: <FormattedMessage id="RolesScreen.title" />,
        },
        { to: ".", label: role.name ?? role.key },
      ];
    }, [role]);

    return (
      <RoleAndGroupsLayout breadcrumbs={breadcrumbs}>
        <RoleDetailsScreenSettingsFormContainer role={role} />
      </RoleAndGroupsLayout>
    );
  };

const RoleDetailsScreen: React.VFC = function RoleDetailsScreen() {
  const { roleID } = useParams() as { roleID: string };
  const { role, loading, error, refetch } = useRoleQuery(roleID);

  if (error != null) {
    return <ShowError error={error} onRetry={refetch} />;
  }

  if (loading) {
    return <ShowLoading />;
  }

  if (role == null) {
    return <ShowLoading />;
  }

  return <RoleDetailsScreenLoaded role={role} />;
};

export default RoleDetailsScreen;
