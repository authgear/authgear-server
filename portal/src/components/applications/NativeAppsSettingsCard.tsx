import React, { useCallback, useContext, useMemo, useState } from "react";
import { produce } from "immer";
import { IconButton as RadixIconButton, Text } from "@radix-ui/themes";
import { Pencil1Icon, TrashIcon } from "@radix-ui/react-icons";
import { Context, FormattedMessage } from "../../intl";
import ErrorRenderer from "../../ErrorRenderer";
import { useFormField } from "../../form";
import { makeValidationErrorCustomMessageIDRule } from "../../error/parse";
import { SettingsSectionCard } from "../v2/SettingsSectionCard/SettingsSectionCard";
import { TextField } from "../v2/TextField/TextField";
import { CopyIconButton } from "../v2/CopyIconButton/CopyIconButton";
import { Badge } from "../v2/Badge/Badge";
import { SecondaryButton } from "../v2/Button/SecondaryButton/SecondaryButton";
import {
  OAuthClientNativeAppConfig,
  OAuthClientNativeAppPlatform,
} from "../../types";
import { nativeAppDisplayID } from "../../util/nativeApps";
import { NativeAppDialog } from "./NativeAppDialog";

const platformMessageIDs: Record<OAuthClientNativeAppPlatform, string> = {
  ios: "NativeAppsSettings.platform.ios",
  android: "NativeAppsSettings.platform.android",
};

const NATIVE_APP_FIELD_NAMES = [
  "team_id",
  "bundle_id",
  "package_name",
  "sha256_cert_fingerprints",
] as const;

interface FieldErrorsProps {
  parentPointer: string;
  fieldName: string;
}

// Binds server validation errors at <parentPointer>/<fieldName> to the row,
// so they do not fall through to the generic error bar.
function FieldErrors({ parentPointer, fieldName }: FieldErrorsProps) {
  const field = useMemo(() => {
    return {
      parentJSONPointer: new RegExp(`^${parentPointer}$`),
      fieldName,
      // A rule's result is attached to the field that owns it, so the location
      // must be this field's own. validateOAuthClientNativeApps in
      // pkg/lib/config/config.go emits the duplicate as a general cause.
      rules:
        fieldName === "bundle_id" || fieldName === "package_name"
          ? [
              makeValidationErrorCustomMessageIDRule(
                "general",
                new RegExp(`^${parentPointer}/${fieldName}$`),
                "NativeAppDialog.error.duplicated"
              ),
            ]
          : undefined,
    };
  }, [parentPointer, fieldName]);
  const { errors } = useFormField(field);
  if (errors.length === 0) {
    return null;
  }
  return (
    <Text as="p" size="1" color="red">
      <ErrorRenderer errors={errors} />
    </Text>
  );
}

interface NativeAppRowProps {
  app: OAuthClientNativeAppConfig;
  index: number;
  clientIndex: number;
  onEdit: (index: number) => void;
  onDelete: (index: number) => void;
}

function NativeAppRow({
  app,
  index,
  clientIndex,
  onEdit,
  onDelete,
}: NativeAppRowProps) {
  const { renderToString } = useContext(Context);

  const entryPointer = `/oauth/clients/${clientIndex}/x_native_apps/${index}`;

  const fingerprintCount =
    app.platform === "android" ? app.sha256_cert_fingerprints.length : 0;

  return (
    <div className="flex flex-col gap-1">
      <div className="flex flex-row items-center gap-2">
        <Badge
          size="1"
          variant="neutral"
          text={<FormattedMessage id={platformMessageIDs[app.platform]} />}
        />
        <Text size="2" className="min-w-0 flex-1 break-all">
          {nativeAppDisplayID(app)}
        </Text>
        {app.platform === "android" ? (
          <Text size="1" color="gray">
            <FormattedMessage
              id="NativeAppsSettings.certificate-count"
              values={{ count: fingerprintCount }}
            />
          </Text>
        ) : null}
        <RadixIconButton
          type="button"
          variant="ghost"
          size="1"
          aria-label={renderToString("edit")}
          onClick={() => onEdit(index)}
        >
          <Pencil1Icon width="1rem" height="1rem" />
        </RadixIconButton>
        <RadixIconButton
          type="button"
          variant="ghost"
          color="red"
          size="1"
          aria-label={renderToString("delete")}
          onClick={() => onDelete(index)}
        >
          <TrashIcon width="1rem" height="1rem" />
        </RadixIconButton>
      </div>
      {NATIVE_APP_FIELD_NAMES.map((name) => (
        <FieldErrors key={name} parentPointer={entryPointer} fieldName={name} />
      ))}
      {Array.from({ length: fingerprintCount }, (_, k) => (
        <FieldErrors
          key={k}
          parentPointer={`${entryPointer}/sha256_cert_fingerprints`}
          fieldName={String(k)}
        />
      ))}
    </div>
  );
}

function CopyField({ label, value }: { label: string; value: string }) {
  return (
    <TextField
      size="2"
      label={label}
      value={value}
      readOnly={true}
      suffixPlain={true}
      suffix={<CopyIconButton textToCopy={value} />}
    />
  );
}

export interface NativeAppsSettingsCardProps {
  className?: string;
  contentClassName?: string;
  publicOrigin: string;
  clientIndex: number;
  // The platforms the admin may add; existing entries of others stay editable.
  platforms: OAuthClientNativeAppPlatform[];
  apps: OAuthClientNativeAppConfig[];
  otherClientsApps: OAuthClientNativeAppConfig[];
  // undefined removes x_native_apps, matching the server's omitempty.
  onAppsChange: (apps: OAuthClientNativeAppConfig[] | undefined) => void;
}

export function NativeAppsSettingsCard(
  props: NativeAppsSettingsCardProps
): React.ReactElement {
  const {
    className,
    contentClassName,
    publicOrigin,
    clientIndex,
    platforms,
    apps,
    otherClientsApps,
    onAppsChange,
  } = props;
  const { renderToString } = useContext(Context);

  const [dialogOpen, setDialogOpen] = useState(false);
  // Kept after close so the dialog keeps its content while animating out.
  // index null = add.
  const [dialogTarget, setDialogTarget] = useState<{
    platform: OAuthClientNativeAppPlatform;
    index: number | null;
  }>({ platform: "ios", index: null });

  const openDialog = useCallback(
    (platform: OAuthClientNativeAppPlatform, index: number | null) => {
      setDialogTarget({ platform, index });
      setDialogOpen(true);
    },
    []
  );

  const onAddIOS = useCallback(() => openDialog("ios", null), [openDialog]);
  const onAddAndroid = useCallback(
    () => openDialog("android", null),
    [openDialog]
  );
  const onEdit = useCallback(
    (index: number) => openDialog(apps[index].platform, index),
    [apps, openDialog]
  );
  const onDelete = useCallback(
    (index: number) => {
      const next = apps.filter((_, i) => i !== index);
      onAppsChange(next.length > 0 ? next : undefined);
    },
    [apps, onAppsChange]
  );
  const onDismiss = useCallback(() => setDialogOpen(false), []);

  const onSubmit = useCallback(
    (app: OAuthClientNativeAppConfig) => {
      const next = produce(apps, (draft) => {
        if (dialogTarget.index == null) {
          draft.push(app);
        } else {
          draft[dialogTarget.index] = app;
        }
      });
      onAppsChange(next);
      setDialogOpen(false);
    },
    [apps, dialogTarget.index, onAppsChange]
  );

  const otherApps = useMemo(
    () => [
      ...otherClientsApps,
      ...apps.filter((_, i) => i !== dialogTarget.index),
    ],
    [otherClientsApps, apps, dialogTarget.index]
  );

  const hasIOS = apps.some((app) => app.platform === "ios");
  const publicHost = useMemo(() => {
    try {
      return new URL(publicOrigin).host;
    } catch {
      return publicOrigin;
    }
  }, [publicOrigin]);

  return (
    <SettingsSectionCard
      className={className}
      contentClassName={contentClassName}
      title={<FormattedMessage id="NativeAppsSettings.title" />}
      description={<FormattedMessage id="NativeAppsSettings.description" />}
    >
      {apps.length > 0 ? (
        <div className="flex flex-col gap-3">
          {apps.map((app, index) => (
            <NativeAppRow
              key={index}
              app={app}
              index={index}
              clientIndex={clientIndex}
              onEdit={onEdit}
              onDelete={onDelete}
            />
          ))}
        </div>
      ) : (
        <Text as="p" size="2" color="gray">
          <FormattedMessage id="NativeAppsSettings.empty" />
        </Text>
      )}

      <div className="flex flex-row flex-wrap gap-2">
        {platforms.includes("ios") ? (
          <SecondaryButton
            size="2"
            text={<FormattedMessage id="NativeAppsSettings.add-ios" />}
            onClick={onAddIOS}
          />
        ) : null}
        {platforms.includes("android") ? (
          <SecondaryButton
            size="2"
            text={<FormattedMessage id="NativeAppsSettings.add-android" />}
            onClick={onAddAndroid}
          />
        ) : null}
      </div>

      {hasIOS ? (
        <CopyField
          label={renderToString("NativeAppsSettings.ios-entitlement.label")}
          value={`webcredentials:${publicHost}`}
        />
      ) : null}

      <NativeAppDialog
        open={dialogOpen}
        platform={dialogTarget.platform}
        initialValue={
          dialogTarget.index == null ? null : apps[dialogTarget.index] ?? null
        }
        otherApps={otherApps}
        onSubmit={onSubmit}
        onDismiss={onDismiss}
      />
    </SettingsSectionCard>
  );
}
