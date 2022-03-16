import React, {
  useMemo,
  useContext,
  useState,
  useCallback,
  useRef,
  ChangeEvent,
} from "react";
import cn from "classnames";
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
import ReactCropperjs from "../../ReactCropperjs";
import { UserQuery_node_User } from "./query/__generated__/UserQuery";
import { useSystemConfig } from "../../context/SystemConfigContext";
import { useUserQuery } from "./query/userQuery";
import { useSimpleForm } from "../../hook/useSimpleForm";
import { useUpdateUserMutation } from "./mutations/updateUserMutation";

import styles from "./EditPictureScreen.module.scss";

interface FormState {
  picture?: string;
  selected?: string;
  uploaded?: string;
}

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
  const fileInputRef = useRef<HTMLInputElement | null>(null);
  const cropperjsRef = useRef<ReactCropperjs | null>(null);
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
      if (state.uploaded == null) {
        const standardAttributes = {
          ...user.standardAttributes,
        };
        delete standardAttributes.picture;
        await updateUser(user.id, standardAttributes, user.customAttributes);
      }
    },
    [user.id, user.standardAttributes, user.customAttributes, updateUser]
  );

  const picture = user.standardAttributes.picture;
  const pictureIsSet = picture != null && picture !== "";

  const defaultState = useMemo(() => {
    return {
      picture,
    };
  }, [picture]);

  const { updateError, save, state, setState } = useSimpleForm({
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

  const onChangeFile = useCallback(
    (e: ChangeEvent<HTMLInputElement>) => {
      const target = e.currentTarget;
      if (!(target instanceof HTMLInputElement)) {
        return;
      }

      const file = target.files?.[0];
      if (file == null) {
        return;
      }

      const reader = new FileReader();
      reader.addEventListener("load", () => {
        if (typeof reader.result === "string") {
          const selected: string = reader.result;
          setState((prev) => {
            return {
              ...prev,
              selected,
            };
          });
        }
      });
      reader.readAsDataURL(file);
    },
    [setState]
  );

  const items: ICommandBarItemProps[] = useMemo(() => {
    if (state.selected == null) {
      return [
        {
          key: "upload",
          text: renderToString("EditPictureScreen.upload-new-picture.label"),
          iconProps: { iconName: "Upload" },
          onClick: () => {
            fileInputRef.current?.click();
          },
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
    }
    return [
      {
        key: "save",
        text: renderToString("save"),
        iconProps: { iconName: "Save" },
        onClick: () => {
          // FIXME(images): get signed URL and upload image.
          // cropperjsRef.current?.getBlob();
        },
      },
    ];
  }, [
    renderToString,
    pictureIsSet,
    themes.destructive,
    themes.main,
    state.selected,
  ]);
  return (
    <FormProvider error={updateError}>
      <CommandBarContainer
        primaryItems={items}
        messageBar={<FormErrorMessageBar />}
      >
        <form>
          <ScreenContent>
            <NavBreadcrumb
              className={styles.widget}
              items={navBreadcrumbItems}
            />
            <ReactCropperjs
              ref={cropperjsRef}
              className={cn(styles.widget, styles.cropperjs)}
              editSrc={state.selected}
              displaySrc={state.picture}
            />
          </ScreenContent>
          <input
            ref={fileInputRef}
            className={styles.fileInput}
            type="file"
            accept="image/png, image/jpeg"
            onChange={onChangeFile}
          />
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
