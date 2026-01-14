import React, {
  useCallback,
  useContext,
  useMemo,
  useState,
  useRef,
} from "react";
import {
  RoleAndGroupsFormFooter,
  RoleAndGroupsLayout,
  RoleAndGroupsVeriticalFormLayout,
} from "../../RoleAndGroupsLayout";
import { useFormContainerBaseContext } from "../../FormContainerBase";
import { BreadcrumbItem } from "../../NavBreadcrumb";
import {
  Context as MessageContext,
  FormattedMessage,
} from "../../intl";
import { useNavigate, useParams } from "react-router-dom";
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
import { RoleAndGroupsFormContainer } from "../../components/roles-and-groups/form/RoleAndGroupsFormContainer";
import PrimaryButton from "../../PrimaryButton";
import DefaultButton from "../../DefaultButton";
import { useSystemConfig } from "../../context/SystemConfigContext";
import { useUpdateRoleMutation } from "./mutations/updateRoleMutation";
import { usePivotNavigation } from "../../hook/usePivot";
import { PivotItem, SearchBox } from "@fluentui/react";
import { AGPivot } from "../../components/common/AGPivot";
import DeleteRoleDialog, {
  DeleteRoleDialogData,
} from "../../components/roles-and-groups/dialog/DeleteRoleDialog";
import { GroupsEmptyView } from "../../components/roles-and-groups/empty-view/GroupsEmptyView";
import { useQuery } from "@apollo/client";
import {
  GroupsListQueryDocument,
  GroupsListQueryQuery,
  GroupsListQueryQueryVariables,
} from "./query/groupsListQuery.generated";
import {
  RoleGroupsList,
  RoleGroupsListItem,
} from "../../components/roles-and-groups/list/RoleGroupsList";
import { AddRoleGroupsDialog } from "../../components/roles-and-groups/dialog/AddRoleGroupsDialog";
import { searchGroups } from "../../model/group";

interface FormState {
  roleKey: string;
  roleName: string;
  roleDescription: string;
}

const SETTINGS_KEY = "settings";
const GROUPS_KEY = "groups";

function RoleDetailsScreenSettingsForm({
  onClickDeleteRole,
}: {
  onClickDeleteRole: () => void;
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
      roleKey: createCallback("roleKey"),
      roleName: createCallback("roleName"),
      roleDescription: createCallback("roleDescription"),
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
            label={renderToString("AddRoleScreen.roleName.title")}
            value={formState.roleName}
            onChange={onFormStateChangeCallbacks.roleName}
          />
          <WidgetDescription className="mt-2">
            <FormattedMessage id="AddRoleScreen.roleName.description" />
          </WidgetDescription>
        </div>
        <div>
          <FormTextField
            required={true}
            fieldName="key"
            parentJSONPointer=""
            type="text"
            label={renderToString("AddRoleScreen.roleKey.title")}
            value={formState.roleKey}
            onChange={onFormStateChangeCallbacks.roleKey}
          />
          <WidgetDescription className="mt-2">
            <FormattedMessage id="AddRoleScreen.roleKey.description" />
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
          label={renderToString("AddRoleScreen.roleDescription.title")}
          value={formState.roleDescription}
          onChange={onFormStateChangeCallbacks.roleDescription}
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
          onClick={onClickDeleteRole}
          text={<FormattedMessage id="RoleDetailsScreen.button.deleteRole" />}
        />
      </RoleAndGroupsFormFooter>
    </div>
  );
}

function RoleDetailsScreenSettingsFormContainer({
  role,
}: {
  role: RoleQueryNodeFragment;
}) {
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

  const form = useSimpleForm({
    stateMode: "UpdateInitialStateWithUseEffect",
    defaultState,
    submit,
    validate,
  });

  const canSave = useMemo(
    () => form.state.roleName !== "" && form.state.roleKey !== "",
    [form.state.roleKey, form.state.roleName]
  );

  const [deleteRoleDialogData, setDeleteRoleDialogData] =
    useState<DeleteRoleDialogData | null>(null);
  const onClickDeleteRole = useCallback(() => {
    setDeleteRoleDialogData({
      roleID: role.id,
      roleKey: role.key,
      roleName: role.name ?? null,
    });
  }, [role.id, role.key, role.name]);
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
        <RoleDetailsScreenSettingsForm onClickDeleteRole={onClickDeleteRole} />
      </RoleAndGroupsFormContainer>

      <DeleteRoleDialog
        onDismiss={dismissDeleteRoleDialog}
        onDismissed={exitIfDeleted}
        data={deleteRoleDialogData}
      />
    </>
  );
}

function RoleDetailsScreenGroupListContainer({
  role,
}: {
  role: RoleQueryNodeFragment;
}) {
  const { renderToString } = useContext(MessageContext);
  const {
    data: groupsQueryData,
    loading,
    error,
    refetch,
  } = useQuery<GroupsListQueryQuery, GroupsListQueryQueryVariables>(
    GroupsListQueryDocument,
    {
      variables: {
        pageSize: 0,
        searchKeyword: "",
      },
      fetchPolicy: "network-only",
    }
  );

  const [searchKeyword, setSearchKeyword] = useState<string>("");
  const onChangeSearchKeyword = useCallback(
    (e?: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
      if (e === undefined) {
        return;
      }
      const value = e.currentTarget.value;
      setSearchKeyword(value);
    },
    []
  );
  const onClearSearchKeyword = useCallback(() => {
    setSearchKeyword("");
  }, []);

  const [isAddGroupDialogHidden, setIsAddGroupDialogHidden] = useState(true);
  const showAddGroupDialog = useCallback(
    () => setIsAddGroupDialogHidden(false),
    []
  );
  const hideAddGroupDialog = useCallback(
    () => setIsAddGroupDialogHidden(true),
    []
  );

  const filteredRoleGroups = useMemo(() => {
    const roleGroups =
      role.groups?.edges?.flatMap<RoleGroupsListItem>((edge) => {
        if (edge?.node != null) {
          return [edge.node];
        }
        return [];
      }) ?? [];
    return searchGroups(roleGroups, searchKeyword);
  }, [role.groups?.edges, searchKeyword]);

  const roleGroups = useMemo(() => {
    return (
      role.groups?.edges?.flatMap((e) => {
        if (e?.node) {
          return [e.node];
        }
        return [];
      }) ?? []
    );
  }, [role.groups?.edges]);

  if (error != null) {
    return <ShowError error={error} onRetry={refetch} />;
  }

  if (loading) {
    return <ShowLoading />;
  }

  const totalCount = groupsQueryData?.groups?.totalCount ?? 0;

  if (totalCount === 0) {
    return <GroupsEmptyView />;
  }

  return (
    <>
      <section className="flex-1 flex flex-col">
        <header className="flex flex-row items-center justify-between mb-8">
          <SearchBox
            className="max-w-[300px] min-w-0 flex-1 mr-2"
            placeholder={renderToString("search")}
            value={searchKeyword}
            onChange={onChangeSearchKeyword}
            onClear={onClearSearchKeyword}
          />
          <PrimaryButton
            text={<FormattedMessage id="RoleDetailsScreen.groups.add" />}
            onClick={showAddGroupDialog}
          />
        </header>
        <RoleGroupsList
          className="flex-1 min-h-0"
          role={role}
          groups={filteredRoleGroups}
        />
      </section>
      <AddRoleGroupsDialog
        roleID={role.id}
        roleKey={role.key}
        roleName={role.name ?? null}
        roleGroups={roleGroups}
        isHidden={isAddGroupDialogHidden}
        onDismiss={hideAddGroupDialog}
      />
    </>
  );
}

const RoleDetailsScreenLoaded: React.VFC<{
  role: RoleQueryNodeFragment;
  reload: ReturnType<typeof useRoleQuery>["refetch"];
}> = function RoleDetailsScreenLoaded({ role }) {
  const { renderToString } = useContext(MessageContext);

  const { selectedKey, onLinkClick } = usePivotNavigation([
    SETTINGS_KEY,
    GROUPS_KEY,
  ]);

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
    <RoleAndGroupsLayout headerBreadcrumbs={breadcrumbs}>
      <AGPivot
        overflowBehavior="menu"
        selectedKey={selectedKey}
        onLinkClick={onLinkClick}
        className="mb-8"
      >
        <PivotItem
          itemKey={SETTINGS_KEY}
          headerText={renderToString("RoleDetailsScreen.tabs.settings")}
        />
        <PivotItem
          itemKey={GROUPS_KEY}
          headerText={renderToString("RoleDetailsScreen.tabs.groups")}
        />
      </AGPivot>
      {selectedKey === GROUPS_KEY ? (
        <RoleDetailsScreenGroupListContainer role={role} />
      ) : (
        <RoleDetailsScreenSettingsFormContainer role={role} />
      )}
    </RoleAndGroupsLayout>
  );
};

const RoleDetailsScreen: React.VFC = function RoleDetailsScreen() {
  const { roleID } = useParams() as { roleID: string };
  const { role, loading, error, refetch } = useRoleQuery(roleID, {
    fetchPolicy: "network-only",
  });

  if (error != null) {
    return <ShowError error={error} onRetry={refetch} />;
  }

  if (loading) {
    return <ShowLoading />;
  }

  if (role == null) {
    return <ShowLoading />;
  }

  return <RoleDetailsScreenLoaded role={role} reload={refetch} />;
};

export default RoleDetailsScreen;
