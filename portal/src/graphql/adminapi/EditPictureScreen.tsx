import React, { useMemo, useContext, useState, useCallback } from "react";
import { FormattedMessage, Context } from "@oursky/react-messageformat";
import {
  Dialog,
  DialogFooter,
  PrimaryButton,
  DefaultButton,
  ICommandBarItemProps,
} from "@fluentui/react";
import { useParams, useNavigate } from "react-router-dom";
import CommandBarContainer from "../../CommandBarContainer";
import { FormProvider } from "../../form";
import { FormErrorMessageBar } from "../../FormErrorMessageBar";
import NavBreadcrumb from "../../NavBreadcrumb";
import NavigationBlockerDialog from "../../NavigationBlockerDialog";
import ScreenContent from "../../ScreenContent";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import { UserQuery_node_User } from "./query/__generated__/UserQuery";
import { useSystemConfig } from "../../context/SystemConfigContext";
import { useUserQuery } from "./query/userQuery";
import { useSimpleForm } from "../../hook/useSimpleForm";
import { useUpdateUserMutation } from "./mutations/updateUserMutation";

import styles from "./EditPictureScreen.module.scss";

interface FormState {
  picture: string;
}

const defaultState: FormState = {
  picture: "",
};

interface RemoveDialogProps {
  hidden: boolean;
  onDismiss: () => void;
  onConfirm: () => void;
}

function RemoveDialog(props: RemoveDialogProps) {
  const { hidden, onDismiss, onConfirm } = props;
  const { renderToString } = useContext(Context);
  const dialogContentProps = useMemo(() => {
    return {
      title: <FormattedMessage id="EditPictureScreen.remove-picture.label" />,
      subText: renderToString(
        "EditPictureScreen.remove-picture.dialog.description"
      ),
    };
  }, [renderToString]);
  const { themes } = useSystemConfig();
  return (
    <Dialog
      hidden={hidden}
      dialogContentProps={dialogContentProps}
      onDismiss={onDismiss}
    >
      <DialogFooter>
        <PrimaryButton onClick={onConfirm} theme={themes.destructive}>
          <FormattedMessage id="remove" />
        </PrimaryButton>
        <DefaultButton onClick={onDismiss}>
          <FormattedMessage id="cancel" />
        </DefaultButton>
      </DialogFooter>
    </Dialog>
  );
}

interface EditPictureScreenContentProps {
  user: UserQuery_node_User;
}

function EditPictureScreenContent(props: EditPictureScreenContentProps) {
  const { user } = props;
  const { renderToString } = useContext(Context);
  const { themes } = useSystemConfig();
  const navigate = useNavigate();
  const isDirty = false;
  const navBreadcrumbItems = useMemo(() => {
    return [
      { to: "../../..", label: <FormattedMessage id="UsersScreen.title" /> },
      { to: "..", label: <FormattedMessage id="UserDetailsScreen.title" /> },
      { to: ".", label: <FormattedMessage id="EditPictureScreen.title" /> },
    ];
  }, []);

  const [isRemoveDialogVisible, setIsRemoveDialogVisible] = useState(false);
  const onDismissRemoveDialog = useCallback(() => {
    setIsRemoveDialogVisible(false);
  }, []);

  const { updateUser } = useUpdateUserMutation();

  const submit = useCallback(
    async (state: FormState) => {
      if (state.picture === "") {
        const standardAttributes = {
          ...user.standardAttributes,
        };
        delete standardAttributes.picture;
        await updateUser(user.id, standardAttributes, user.customAttributes);
      }
    },
    [user.id, user.standardAttributes, user.customAttributes, updateUser]
  );

  const { updateError, save } = useSimpleForm({
    stateMode: "UpdateInitialStateWithUseEffect",
    defaultState,
    submit,
  });

  const onConfirmRemove = useCallback(() => {
    save().then(
      () => {
        setIsRemoveDialogVisible(false);
        navigate("..", { replace: true });
      },
      () => {
        setIsRemoveDialogVisible(false);
      }
    );
  }, [save, navigate]);

  const picture = user.standardAttributes.picture;
  const pictureIsSet = picture != null && picture !== "";

  const initialItems: ICommandBarItemProps[] = useMemo(() => {
    return [
      {
        key: "upload",
        text: renderToString("EditPictureScreen.upload-new-picture.label"),
        iconProps: { iconName: "Upload" },
      },
      {
        key: "remove",
        text: renderToString("EditPictureScreen.remove-picture.label"),
        iconProps: { iconName: "Delete" },
        disabled: !pictureIsSet,
        theme: pictureIsSet ? themes.destructive : themes.main,
        onClick: () => {
          setIsRemoveDialogVisible(true);
        },
      },
    ];
  }, [renderToString, pictureIsSet, themes.destructive, themes.main]);
  return (
    <FormProvider error={updateError}>
      <CommandBarContainer
        primaryItems={initialItems}
        messageBar={<FormErrorMessageBar />}
      >
        <form>
          <ScreenContent>
            <NavBreadcrumb
              className={styles.widget}
              items={navBreadcrumbItems}
            />
          </ScreenContent>
        </form>
      </CommandBarContainer>
      <NavigationBlockerDialog blockNavigation={isDirty} />
      <RemoveDialog
        hidden={!isRemoveDialogVisible}
        onDismiss={onDismissRemoveDialog}
        onConfirm={onConfirmRemove}
      />
    </FormProvider>
  );
}

const EditPictureScreen: React.FC = function EditPictureScreen() {
  const { appID: _appID, userID } = useParams();
  const {
    user,
    loading: loadingUser,
    error: userError,
    refetch: refetchUser,
  } = useUserQuery(userID);

  if (loadingUser) {
    return <ShowLoading />;
  }

  if (user == null) {
    return <ShowLoading />;
  }

  if (userError != null) {
    return <ShowError error={userError} onRetry={refetchUser} />;
  }

  return <EditPictureScreenContent user={user} />;
};

export default EditPictureScreen;
