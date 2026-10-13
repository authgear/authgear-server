import React, { useCallback, useContext, useState } from "react";
import { Dialog, Flex } from "@radix-ui/themes";
import { Context, FormattedMessage } from "../../intl";
import { TextField } from "../v2/TextField/TextField";
import { TextArea } from "../v2/TextArea/TextArea";
import { PrimaryButton } from "../v2/Button/PrimaryButton/PrimaryButton";
import { SecondaryButton } from "../v2/Button/SecondaryButton/SecondaryButton";
import {
  OAuthClientNativeAppConfig,
  OAuthClientNativeAppPlatform,
} from "../../types";
import {
  NativeAppDraft,
  NativeAppFieldError,
  nativeAppFromDraft,
  nativeAppToDraft,
  validateNativeApp,
} from "../../util/nativeApps";

const titleMessageIDs: Record<
  OAuthClientNativeAppPlatform,
  { add: string; edit: string }
> = {
  ios: {
    add: "NativeAppDialog.title.add-ios",
    edit: "NativeAppDialog.title.edit-ios",
  },
  android: {
    add: "NativeAppDialog.title.add-android",
    edit: "NativeAppDialog.title.edit-android",
  },
};

export interface NativeAppDialogProps {
  open: boolean;
  platform: OAuthClientNativeAppPlatform;
  // null opens the dialog in add mode.
  initialValue: OAuthClientNativeAppConfig | null;
  // Every entry other than the one being edited, across all clients.
  otherApps: OAuthClientNativeAppConfig[];
  onSubmit: (value: OAuthClientNativeAppConfig) => void;
  onDismiss: () => void;
}

function renderFieldErrors(
  errors: NativeAppFieldError[],
  field: NativeAppFieldError["field"]
): React.ReactNode | undefined {
  const matched = errors.filter((e) => e.field === field);
  if (matched.length === 0) {
    return undefined;
  }
  return matched.map((e, i) => (
    <React.Fragment key={i}>
      {i > 0 ? <br /> : null}
      <FormattedMessage
        id={e.messageID}
        values={"fingerprint" in e ? { fingerprint: e.fingerprint } : {}}
      />
    </React.Fragment>
  ));
}

export function NativeAppDialog(
  props: NativeAppDialogProps
): React.ReactElement {
  const { open, platform, initialValue, onDismiss } = props;

  const onOpenChange = useCallback(
    (nextOpen: boolean) => {
      if (!nextOpen) {
        onDismiss();
      }
    },
    [onDismiss]
  );

  const titles = titleMessageIDs[platform];

  return (
    <Dialog.Root open={open} onOpenChange={onOpenChange}>
      <Dialog.Content maxWidth="480px" size="3">
        <Dialog.Title>
          <FormattedMessage
            id={initialValue == null ? titles.add : titles.edit}
          />
        </Dialog.Title>
        {/* Dialog.Content unmounts on close, so each open starts a fresh draft. */}
        <NativeAppDialogForm {...props} />
      </Dialog.Content>
    </Dialog.Root>
  );
}

function NativeAppDialogForm(props: NativeAppDialogProps): React.ReactElement {
  const { platform, initialValue, otherApps, onSubmit, onDismiss } = props;
  const { renderToString } = useContext(Context);

  const [draft, setDraft] = useState<NativeAppDraft>(() =>
    nativeAppToDraft(initialValue)
  );
  const [errors, setErrors] = useState<NativeAppFieldError[]>([]);

  const makeOnChange = useCallback(
    (key: keyof NativeAppDraft) =>
      (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
        const value = e.currentTarget.value;
        setDraft((prev) => ({ ...prev, [key]: value }));
      },
    []
  );

  const onFormSubmit = useCallback(
    (e: React.FormEvent<HTMLFormElement>) => {
      e.preventDefault();
      e.stopPropagation();
      const app = nativeAppFromDraft(platform, draft);
      const validationErrors = validateNativeApp(app, otherApps);
      if (validationErrors.length > 0) {
        setErrors(validationErrors);
        return;
      }
      onSubmit(app);
    },
    [platform, draft, otherApps, onSubmit]
  );

  return (
    <form onSubmit={onFormSubmit}>
      <Flex direction="column" gap="4">
        {platform === "ios" ? (
          <>
            <TextField
              size="2"
              label={renderToString("NativeAppDialog.team-id.label")}
              hint={renderToString("NativeAppDialog.team-id.hint")}
              placeholder="ABCDE12345"
              value={draft.teamID}
              onChange={makeOnChange("teamID")}
              error={renderFieldErrors(errors, "team_id")}
              required={true}
            />
            <TextField
              size="2"
              label={renderToString("NativeAppDialog.bundle-id.label")}
              placeholder="com.example.app"
              value={draft.bundleID}
              onChange={makeOnChange("bundleID")}
              error={renderFieldErrors(errors, "bundle_id")}
              required={true}
            />
          </>
        ) : (
          <>
            <TextField
              size="2"
              label={renderToString("NativeAppDialog.package-name.label")}
              placeholder="com.example.app"
              value={draft.packageName}
              onChange={makeOnChange("packageName")}
              error={renderFieldErrors(errors, "package_name")}
              required={true}
            />
            <TextArea
              size="2"
              label={renderToString("NativeAppDialog.fingerprints.label")}
              hint={renderToString("NativeAppDialog.fingerprints.hint")}
              value={draft.fingerprints}
              onChange={makeOnChange("fingerprints")}
              error={renderFieldErrors(errors, "sha256_cert_fingerprints")}
              required={true}
            />
          </>
        )}
        <Flex gap="3" justify="end">
          <SecondaryButton
            size="2"
            text={<FormattedMessage id="cancel" />}
            onClick={onDismiss}
          />
          <PrimaryButton
            type="submit"
            size="2"
            text={
              <FormattedMessage id={initialValue == null ? "add" : "done"} />
            }
          />
        </Flex>
      </Flex>
    </form>
  );
}
