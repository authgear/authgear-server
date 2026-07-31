import React, { useCallback, useEffect, useRef, useState } from "react";
import { Dialog } from "@radix-ui/themes";
import axios, { RawAxiosRequestHeaders } from "axios";
import authgear from "@authgear/web";
import { FormattedMessage } from "../../intl";
import ReactCropperjs from "../../ReactCropperjs";
import { Callout } from "../../components/v2/Callout/Callout";
import { PrimaryButton } from "../../components/v2/Button/PrimaryButton/PrimaryButton";
import { SecondaryButton } from "../../components/v2/Button/SecondaryButton/SecondaryButton";
import { useUpdateUserMutation } from "./mutations/updateUserMutation";
import { UserQueryNodeFragment } from "./query/userQuery.generated";
import styles from "./ProfilePictureDialog.module.css";

const MAX_FILE_SIZE = 500 * 1024;
const ACCEPTED_IMAGE_TYPES = new Set([
  "image/svg+xml",
  "image/png",
  "image/jpeg",
  "image/x-icon",
  "image/vnd.microsoft.icon",
]);

function isAcceptedFile(file: File): boolean {
  return (
    file.size <= MAX_FILE_SIZE &&
    (ACCEPTED_IMAGE_TYPES.has(file.type) ||
      /\.(svg|png|jpe?g|ico)$/i.test(file.name))
  );
}

export const PROFILE_PICTURE_ACCEPT =
  ".svg,.png,.jpeg,.jpg,.ico,image/svg+xml,image/png,image/jpeg,image/x-icon,image/vnd.microsoft.icon";

export interface ProfilePictureDialogProps {
  appID: string;
  user: UserQueryNodeFragment;
  file: File | null;
  onDismiss: () => void;
  onSaved?: () => unknown;
}

export function ProfilePictureDialog(
  props: ProfilePictureDialogProps
): React.ReactElement {
  const { appID, user, file, onDismiss, onSaved } = props;
  const { updateUser } = useUpdateUserMutation();
  const cropperRef = useRef<ReactCropperjs | null>(null);
  const [source, setSource] = useState<string>();
  const [imageError, setImageError] = useState(false);
  const [uploadError, setUploadError] = useState(false);
  const [isSaving, setIsSaving] = useState(false);

  useEffect(() => {
    setSource(undefined);
    setImageError(false);
    setUploadError(false);

    if (file == null) {
      return;
    }
    if (!isAcceptedFile(file)) {
      setImageError(true);
      return;
    }

    const reader = new FileReader();
    reader.addEventListener("load", () => {
      if (typeof reader.result === "string") {
        setSource(reader.result);
      }
    });
    reader.addEventListener("error", () => {
      setImageError(true);
    });
    reader.readAsDataURL(file);
  }, [file]);

  const onOpenChange = useCallback(
    (open: boolean) => {
      if (!open && !isSaving) {
        onDismiss();
      }
    },
    [isSaving, onDismiss]
  );

  const onSave = useCallback(async () => {
    if (isSaving || source == null || imageError) {
      return;
    }

    setIsSaving(true);
    setUploadError(false);
    try {
      const blob = await cropperRef.current?.getBlob();
      if (blob == null) {
        setImageError(true);
        return;
      }

      await authgear.refreshAccessTokenIfNeeded();
      const headers: RawAxiosRequestHeaders = {};
      if (authgear.accessToken != null) {
        headers.Authorization = `Bearer ${authgear.accessToken}`;
      }

      const response = await axios(
        `/api/apps/${appID}/_api/admin/images/upload`,
        {
          method: "GET",
          headers,
        }
      );
      const formData = new FormData();
      formData.append("file", blob);
      const uploadResponse = await axios(response.data.result.upload_url, {
        method: "POST",
        headers,
        data: formData,
      });
      const picture = uploadResponse.data.result.url;

      await updateUser(
        user.id,
        {
          ...user.standardAttributes,
          picture,
        },
        user.customAttributes
      );
      await onSaved?.();
      onDismiss();
    } catch {
      setUploadError(true);
    } finally {
      setIsSaving(false);
    }
  }, [
    appID,
    imageError,
    isSaving,
    onDismiss,
    onSaved,
    source,
    updateUser,
    user.customAttributes,
    user.id,
    user.standardAttributes,
  ]);

  return (
    <Dialog.Root open={file != null} onOpenChange={onOpenChange}>
      <Dialog.Content maxWidth="519px" size="3">
        <Dialog.Title>
          <FormattedMessage id="ProfilePictureDialog.title" />
        </Dialog.Title>
        <Dialog.Description size="2" className={styles.description}>
          <FormattedMessage id="ProfilePictureDialog.description" />
        </Dialog.Description>

        {source != null ? (
          <ReactCropperjs
            ref={cropperRef}
            className={styles.cropArea}
            height={174}
            editSrc={source}
            onLoad={() => {
              setImageError(false);
            }}
            onError={() => {
              setImageError(true);
            }}
          />
        ) : null}

        {imageError ? (
          <Callout
            className={styles.error}
            type="error"
            showCloseButton={false}
            text={<FormattedMessage id="errors.invalid-selected-image" />}
          />
        ) : null}
        {uploadError ? (
          <Callout
            className={styles.error}
            type="error"
            showCloseButton={false}
            text={<FormattedMessage id="errors.network" />}
          />
        ) : null}

        <div className={styles.actions}>
          <SecondaryButton
            size="2"
            disabled={isSaving}
            onClick={onDismiss}
            text={<FormattedMessage id="cancel" />}
          />
          <PrimaryButton
            size="2"
            loading={isSaving}
            disabled={source == null || imageError || isSaving}
            onClick={() => {
              onSave().catch(() => {});
            }}
            text={<FormattedMessage id="save" />}
          />
        </div>
      </Dialog.Content>
    </Dialog.Root>
  );
}
