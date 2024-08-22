import React, {
  useMemo,
  useContext,
  useState,
  useCallback,
  useRef,
  useEffect,
  ChangeEvent,
} from "react";
import { FormattedMessage, Context } from "@oursky/react-messageformat";
import { Dialog, DialogFooter, Spinner, SpinnerSize } from "@fluentui/react";
import { useParams, useNavigate } from "react-router-dom";
import axios, { AxiosProgressEvent } from "axios";
import PrimaryButton from "../../PrimaryButton";
import DefaultButton from "../../DefaultButton";
import { FormProvider } from "../../form";
import { FormErrorMessageBar } from "../../FormErrorMessageBar";
import NavBreadcrumb from "../../NavBreadcrumb";
import NavigationBlockerDialog from "../../NavigationBlockerDialog";
import ScreenContent from "../../ScreenContent";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import ReactCropperjs from "../../ReactCropperjs";
import { UserQueryNodeFragment } from "./query/userQuery.generated";
import { useSystemConfig } from "../../context/SystemConfigContext";
import { useUserQuery } from "./query/userQuery";
import { useSimpleForm } from "../../hook/useSimpleForm";
import { useUpdateUserMutation } from "./mutations/updateUserMutation";
import { useAppAndSecretConfigQuery } from "../portal/query/appAndSecretConfigQuery";
import { jsonPointerToString } from "../../util/jsonpointer";
import { AccessControlLevelString } from "../../types";
import { APIError } from "../../error/error";
import { ErrorParseRule, makeLocalErrorParseRule } from "../../error/parse";

import styles from "./EditPictureScreen.module.css";
import DefaultLayout from "../../DefaultLayout";

interface FormState {
  picture?: string;
  selected?: string;
}

interface RemoveDialogProps {
  hidden: boolean;
  onDismiss: () => void;
  onConfirm: () => void;
}

const SENTINEL: APIError = {
  errorName: "__local",
  reason: "__local",
  info: {
    error: {
      messageID: "errors.invalid-selected-image",
    },
  },
};

const RULES: ErrorParseRule[] = [
  makeLocalErrorParseRule(SENTINEL, SENTINEL.info.error),
];

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
        <PrimaryButton
          onClick={onConfirm}
          theme={themes.destructive}
          text={<FormattedMessage id="remove" />}
        />
        <DefaultButton
          onClick={onDismiss}
          text={<FormattedMessage id="cancel" />}
        />
      </DialogFooter>
    </Dialog>
  );
}

interface EditPictureScreenContentProps {
  user: UserQueryNodeFragment;
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
  const [reactCropperjsError, setReactCropperjsError] = useState<
    typeof SENTINEL | null
  >(null);
  const [uploadState, setUploadState] = useState(DEFAULT_UPLOAD_STATE);
  const fileInputRef = useRef<HTMLInputElement | null>(null);
  const cropperjsRef = useRef<ReactCropperjs | null>(null);
  const uploadedURLRef = useRef<string | null>(null);
  const navBreadcrumbItems = useMemo(() => {
    return [
      { to: "~/users", label: <FormattedMessage id="UsersScreen.title" /> },
      {
        to: `~/users/${user.id}/details`,
        label: <FormattedMessage id="UserDetailsScreen.title" />,
      },
      { to: ".", label: <FormattedMessage id="EditPictureScreen.title" /> },
    ];
  }, [user.id]);

  const [isRemoveDialogVisible, setIsRemoveDialogVisible] = useState(false);
  const onDismissRemoveDialog = useCallback(() => {
    setIsRemoveDialogVisible(false);
  }, []);

  const { updateUser } = useUpdateUserMutation();

  const onReactCropperjsError = useCallback(() => {
    setReactCropperjsError(SENTINEL);
  }, []);

  const onReactCropperjsLoad = useCallback(() => {
    setReactCropperjsError(null);
  }, []);

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
        navigate("./..", { replace: true });
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
    (e: AxiosProgressEvent) => {
      const { loaded, total } = e;
      if (total != null) {
        setState((prev) => {
          return {
            ...prev,
            percentComplete: loaded / total,
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
          navigate("./..", { replace: true });
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

  const onClickSelectImage = useCallback(() => {
    fileInputRef.current?.click();
  }, []);
  const loading = uploadState.loading || isUpdating;

  const showUpload = useMemo(
    () => state.selected == null || reactCropperjsError != null,
    [state, reactCropperjsError]
  );
  const showRemove = useMemo(
    () => state.selected == null && pictureIsSet,
    [state, pictureIsSet]
  );
  const showSave = useMemo(
    () => state.selected != null && reactCropperjsError == null,
    [state, reactCropperjsError]
  );

  return (
    <FormProvider
      loading={loading}
      error={updateError || uploadState.error || reactCropperjsError}
      rules={RULES}
    >
      <DefaultLayout
        position="end"
        messageBar={<FormErrorMessageBar />}
        footer={
          <>
            {showUpload ? (
              <PrimaryButton
                text={renderToString(
                  "EditPictureScreen.upload-new-picture.label"
                )}
                iconProps={{ iconName: "Upload" }}
                onClick={() => {
                  fileInputRef.current?.click();
                }}
              />
            ) : null}
            {showRemove ? (
              <DefaultButton
                text={renderToString("EditPictureScreen.remove-picture.label")}
                iconProps={{ iconName: "Delete" }}
                disabled={!pictureIsSet}
                theme={themes.destructive}
                useThemePrimaryForBorderColor={true}
                onClick={() => {
                  setIsRemoveDialogVisible(true);
                }}
              />
            ) : null}
            {showSave ? (
              <PrimaryButton
                text={
                  <div className={styles.saveButton}>
                    {loading ? (
                      <Spinner size={SpinnerSize.xSmall} ariaLive="assertive" />
                    ) : null}
                    <span>
                      <FormattedMessage id="save" />
                    </span>
                  </div>
                }
                disabled={loading}
                onClick={() => {
                  upload().catch(() => {});
                }}
              />
            ) : null}
          </>
        }
      >
        <form>
          <ScreenContent>
            <NavBreadcrumb
              className={styles.widget}
              items={navBreadcrumbItems}
            />
            <ReactCropperjs
              ref={cropperjsRef}
              className={styles.widget}
              editSrc={state.selected}
              displaySrc={state.picture}
              onError={onReactCropperjsError}
              onLoad={onReactCropperjsLoad}
              onClickSelectImage={onClickSelectImage}
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
      </DefaultLayout>
      <NavigationBlockerDialog blockNavigation={isDirty} />
      <RemoveDialog
        hidden={!isRemoveDialogVisible}
        onDismiss={onDismissRemoveDialog}
        onConfirm={onConfirmRemove}
      />
    </FormProvider>
  );
}

const EditPictureScreen: React.VFC = function EditPictureScreen() {
  const navigate = useNavigate();
  const { appID, userID } = useParams() as { appID: string; userID: string };
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
      navigate("./..");
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
