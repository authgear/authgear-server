import React, {
  useMemo,
  useContext,
  useState,
  useCallback,
  useRef,
  useEffect,
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
  ProgressIndicator,
} from "@fluentui/react";
import { useParams, useNavigate } from "react-router-dom";
import axios from "axios";
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
import { useAppAndSecretConfigQuery } from "../portal/query/appAndSecretConfigQuery";
import { jsonPointerToString } from "../../util/jsonpointer";
import { AccessControlLevelString } from "../../types";

import styles from "./EditPictureScreen.module.scss";

interface FormState {
  picture?: string;
  selected?: string;
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
  appID: string;
}

interface UploadState {
  error: unknown;
  loading: boolean;
  percentComplete?: number;
}

const DEFAULT_UPLOAD_STATE: UploadState = {
  error: undefined,
  loading: false,
};

function EditPictureScreenContent(props: EditPictureScreenContentProps) {
  const { user, appID } = props;
  const { renderToString } = useContext(Context);
  const { themes } = useSystemConfig();
  const navigate = useNavigate();
  const [uploadState, setUploadState] = useState(DEFAULT_UPLOAD_STATE);
  const fileInputRef = useRef<HTMLInputElement | null>(null);
  const cropperjsRef = useRef<ReactCropperjs | null>(null);
  const uploadedURLRef = useRef<string | null>(null);
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
    async (_state: FormState) => {
      if (uploadedURLRef.current != null) {
        const standardAttributes = {
          ...user.standardAttributes,
          picture: uploadedURLRef.current,
        };
        await updateUser(user.id, standardAttributes, user.customAttributes);
      } else {
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

  const { updateError, save, state, setState, isUpdating } = useSimpleForm({
    stateMode: "UpdateInitialStateWithUseEffect",
    defaultState,
    submit,
  });

  const isDirty = useMemo(() => {
    return state.selected != null;
  }, [state.selected]);

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

  const onProgress = useCallback(
    (e: ProgressEvent) => {
      if (e.lengthComputable) {
        setState((prev) => {
          return {
            ...prev,
            percentComplete: e.loaded / e.total,
          };
        });
      }
    },
    [setState]
  );

  const upload = useCallback(async () => {
    if (uploadState.loading) {
      return;
    }

    try {
      const blob = await cropperjsRef.current?.getBlob();
      if (blob == null) {
        return;
      }

      setUploadState({
        error: undefined,
        loading: true,
        percentComplete: undefined,
      });

      const resp = await axios(`/api/apps/${appID}/_api/admin/images/upload`, {
        method: "GET",
        onUploadProgress: onProgress,
        onDownloadProgress: onProgress,
      });

      const { upload_url } = resp.data.result;
      const formData = new FormData();
      formData.append("file", blob);
      const uploadResp = await axios(upload_url, {
        method: "POST",
        data: formData,
        onUploadProgress: onProgress,
        onDownloadProgress: onProgress,
      });

      const {
        result: { url },
      } = uploadResp.data;
      uploadedURLRef.current = url;
      save().then(
        () => {
          navigate("..", { replace: true });
        },
        () => {}
      );
      // eslint-disable-next-line @typescript-eslint/no-implicit-any-catch
    } catch (e: any) {
      if (e?.response?.data?.error != null) {
        setUploadState((prev) => {
          return {
            ...prev,
            error: e.response.data.error,
          };
        });
      } else {
        setUploadState((prev) => {
          return {
            ...prev,
            error: e,
          };
        });
      }
    } finally {
      setUploadState((prev) => {
        return {
          ...prev,
          loading: false,
        };
      });
    }
  }, [appID, uploadState.loading, save, navigate, onProgress]);

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
        disabled: uploadState.loading || isUpdating,
        onClick: () => {
          upload().catch(() => {});
        },
      },
    ];
  }, [
    renderToString,
    pictureIsSet,
    themes.destructive,
    themes.main,
    state.selected,
    uploadState.loading,
    upload,
    isUpdating,
  ]);
  return (
    <FormProvider error={updateError || uploadState.error}>
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
            <ProgressIndicator
              className={styles.widget}
              percentComplete={uploadState.percentComplete}
              progressHidden={!uploadState.loading}
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
  const navigate = useNavigate();
  const { appID, userID } = useParams();
  const {
    user,
    loading: loadingUser,
    error: userError,
    refetch: refetchUser,
  } = useUserQuery(userID);

  const {
    effectiveAppConfig,
    loading: loadingAppConfig,
    error: appConfigError,
    refetch: refetchAppConfig,
  } = useAppAndSecretConfigQuery(appID);

  const standardAttributeAccessControl = useMemo(() => {
    const record: Record<string, AccessControlLevelString> = {};
    for (const item of effectiveAppConfig?.user_profile?.standard_attributes
      ?.access_control ?? []) {
      record[item.pointer] = item.access_control.portal_ui;
    }
    return record;
  }, [effectiveAppConfig]);

  const profileImageEditable = useMemo(() => {
    const ptr = jsonPointerToString(["picture"]);
    const level = standardAttributeAccessControl[ptr];
    return level === "readwrite";
  }, [standardAttributeAccessControl]);

  useEffect(() => {
    if (!profileImageEditable) {
      navigate("..");
    }
  }, [navigate, profileImageEditable]);

  if (loadingUser || loadingAppConfig) {
    return <ShowLoading />;
  }

  if (user == null || effectiveAppConfig == null) {
    return <ShowLoading />;
  }

  if (userError != null) {
    return <ShowError error={userError} onRetry={refetchUser} />;
  }

  if (appConfigError != null) {
    return <ShowError error={appConfigError} onRetry={refetchAppConfig} />;
  }

  return <EditPictureScreenContent user={user} appID={appID} />;
};

export default EditPictureScreen;
