import React, { useCallback, useContext, useMemo, useState } from "react";
import cn from "classnames";
import {
  Dialog,
  DialogFooter,
  IDialogContentProps,
  Text,
} from "@fluentui/react";
import { produce } from "immer";
import { Context, FormattedMessage } from "../../intl";
import ExternalLink from "../../ExternalLink";
import PrimaryButton from "../../PrimaryButton";
import DefaultButton from "../../DefaultButton";
import { AppSecretConfigFormModel } from "../../hook/useAppSecretConfigForm";
import {
  findFramework,
  frameworksForType,
  type FrameworkEntry,
} from "./CreateOAuthClientScreen/frameworks";
import { FrameworkCard } from "./CreateOAuthClientScreen/FrameworkCard";
import type { Framework, OAuthClientConfig } from "../../types";
import styles from "./EditOAuthClientFormSPAQuickStart.module.css";

interface FormStateShape {
  clients: OAuthClientConfig[];
  editedClient: OAuthClientConfig | null;
}

export interface EditOAuthClientFormSPAQuickStartProps<S extends FormStateShape> {
  className?: string;
  client: OAuthClientConfig;
  form: AppSecretConfigFormModel<S>;
}

export function EditOAuthClientFormSPAQuickStart<S extends FormStateShape>({
  className,
  client,
  form,
}: EditOAuthClientFormSPAQuickStartProps<S>): React.ReactElement {
  const [dialogVisible, setDialogVisible] = useState(false);
  const [applying, setApplying] = useState(false);

  const framework = findFramework(client.x_framework);

  const openDialog = useCallback(() => setDialogVisible(true), []);
  const closeDialog = useCallback(() => {
    if (!applying) setDialogVisible(false);
  }, [applying]);

  const applyFramework = useCallback(
    async (newFrameworkId: Framework) => {
      setApplying(true);
      try {
        const newState = produce(form.state, (draft) => {
          draft.clients = draft.clients.map((c) =>
            c.client_id === client.client_id
              ? { ...c, x_framework: newFrameworkId }
              : c
          );
          if (draft.editedClient?.client_id === client.client_id) {
            draft.editedClient.x_framework = newFrameworkId;
          }
        });
        form.setState(() => newState);
        await form.saveWithState(newState);
        setDialogVisible(false);
      } finally {
        setApplying(false);
      }
    },
    [client.client_id, form]
  );

  if (framework == null) {
    return (
      <div className={cn(styles.root, className)}>
        <div className={styles.frameworkCard}>
          <div className={styles.iconWrap}>
            <i className={cn("ti", "ti-app-window", styles.frameworkIcon)} />
          </div>
          <div className={styles.frameworkText}>
            <Text variant="large" block={true} styles={titleStyles}>
              <FormattedMessage id="EditOAuthClientFormSPAQuickStart.no-framework.title" />
            </Text>
            <Text block={true} className={styles.helperText}>
              <FormattedMessage id="EditOAuthClientFormSPAQuickStart.no-framework.body" />
            </Text>
            <PrimaryButton
              className={styles.changeButton}
              onClick={openDialog}
              text={
                <FormattedMessage id="EditOAuthClientFormSPAQuickStart.choose-framework" />
              }
            />
          </div>
        </div>
        <ChangeFrameworkDialog
          visible={dialogVisible}
          applicationType="spa"
          currentFrameworkId={null}
          applying={applying}
          onApply={applyFramework}
          onDismiss={closeDialog}
        />
      </div>
    );
  }

  return (
    <div className={cn(styles.root, className)}>
      <div className={styles.frameworkRow}>
        <div className={styles.iconWrap}>
          <i
            className={cn("ti", `ti-${framework.iconName}`, styles.frameworkIcon)}
            aria-hidden={true}
          />
        </div>
        <div className={styles.frameworkText}>
          <Text variant="large" block={true} styles={titleStyles}>
            {framework.displayName}
          </Text>
          <Text block={true} className={styles.helperText}>
            {framework.helperText}
          </Text>
        </div>
        <DefaultButton
          className={styles.changeButtonInline}
          onClick={openDialog}
          text={
            <FormattedMessage id="EditOAuthClientFormSPAQuickStart.change-button" />
          }
        />
      </div>

      <Text variant="xLarge" block={true} className={styles.sectionHeading}>
        <FormattedMessage id="EditOAuthClientFormSPAQuickStart.step-by-step.title" />
      </Text>
      <div className={styles.tutorialCard}>
        <div className={styles.tutorialHeader}>
          <i
            className={cn("ti", "ti-clock", styles.tutorialIcon)}
            aria-hidden={true}
          />
          <Text styles={tutorialDurationStyles}>
            <FormattedMessage id="EditOAuthClientFormSPAQuickStart.tutorial.duration" />
          </Text>
        </div>
        <Text block={true} className={styles.tutorialBody}>
          <FormattedMessage
            id="EditOAuthClientFormSPAQuickStart.tutorial.body"
            values={{
              displayName: framework.displayName,
              // eslint-disable-next-line react/no-unstable-nested-components
              docLink: (chunks: React.ReactNode) => (
                <ExternalLink href={framework.docLink}>{chunks}</ExternalLink>
              ),
            }}
          />
        </Text>
      </div>

      <ChangeFrameworkDialog
        visible={dialogVisible}
        applicationType="spa"
        currentFrameworkId={framework.id}
        applying={applying}
        onApply={applyFramework}
        onDismiss={closeDialog}
      />
    </div>
  );
}

const titleStyles = { root: { fontWeight: 600 as const } };
const tutorialDurationStyles = { root: { fontWeight: 600 as const } };

interface ChangeFrameworkDialogProps {
  visible: boolean;
  applicationType: "spa" | "native" | "confidential" | "traditional_webapp";
  currentFrameworkId: Framework | null;
  applying: boolean;
  onApply: (frameworkId: Framework) => Promise<void>;
  onDismiss: () => void;
}

function ChangeFrameworkDialog(props: ChangeFrameworkDialogProps) {
  const {
    visible,
    applicationType,
    currentFrameworkId,
    applying,
    onApply,
    onDismiss,
  } = props;
  const { renderToString } = useContext(Context);
  const [selected, setSelected] = useState<Framework | null>(currentFrameworkId);

  // Reset selection when dialog opens with a different current framework.
  React.useEffect(() => {
    if (visible) setSelected(currentFrameworkId);
  }, [visible, currentFrameworkId]);

  const options = useMemo<FrameworkEntry[]>(
    () => frameworksForType(applicationType),
    [applicationType]
  );

  const dialogContent: IDialogContentProps = useMemo(
    () => ({
      title: renderToString("EditOAuthClientFormSPAQuickStart.change-dialog.title"),
    }),
    [renderToString]
  );

  const onApplyClick = useCallback(() => {
    if (selected == null) return;
    void onApply(selected);
  }, [onApply, selected]);

  const canApply = selected != null && selected !== currentFrameworkId && !applying;

  return (
    <Dialog
      hidden={!visible}
      dialogContentProps={dialogContent}
      modalProps={{ isBlocking: applying }}
      onDismiss={onDismiss}
      maxWidth={620}
    >
      <div className={styles.dialogGrid}>
        {options.map((f) => (
          <FrameworkCard
            key={f.id}
            framework={f}
            selected={selected === f.id}
            onSelect={() => setSelected(f.id)}
          />
        ))}
      </div>
      <DialogFooter>
        <PrimaryButton
          onClick={onApplyClick}
          disabled={!canApply}
          text={
            <FormattedMessage id="EditOAuthClientFormSPAQuickStart.change-dialog.apply" />
          }
        />
        <DefaultButton
          onClick={onDismiss}
          disabled={applying}
          text={<FormattedMessage id="cancel" />}
        />
      </DialogFooter>
    </Dialog>
  );
}
