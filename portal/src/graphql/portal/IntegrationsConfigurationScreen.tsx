import React, { useContext, useMemo, useCallback, useState } from "react";
import { useParams } from "react-router-dom";
import { Button, Dialog, Flex, Text } from "@radix-ui/themes";
import { FormattedMessage, Context } from "../../intl";
import { produce } from "immer";
import {
  AppConfigFormModel,
  useAppConfigForm,
} from "../../hook/useAppConfigForm";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import ScreenContent from "../../ScreenContent";
import ScreenLayoutScrollView from "../../ScreenLayoutScrollView";
import { Badge } from "../../components/v2/Badge/Badge";
import { PortalAPIAppConfig } from "../../types";
import { TextField } from "../../components/v2/TextField/TextField";
import { PrimaryButton } from "../../components/v2/Button/PrimaryButton/PrimaryButton";
import { SecondaryButton } from "../../components/v2/Button/SecondaryButton/SecondaryButton";
import { parseAPIErrors, parseRawError } from "../../error/parse";
import ErrorRenderer from "../../ErrorRenderer";
import { clearEmptyObject } from "../../util/misc";
import styles from "./IntegrationsConfigurationScreen.module.css";

import gtmLogoURL from "../../images/gtm_logo.png";

function isValidGTMContainerID(containerID: string): boolean {
  return /^GTM-.+/.test(containerID);
}

function gtmContainerIDFormatError(): React.ReactNode {
  return (
    <FormattedMessage
      id="errors.validation.format"
      values={{ format: "google_tag_manager_container_id" }}
    />
  );
}

function gtmContainerIDRequiredError(): React.ReactNode {
  return <FormattedMessage id="errors.validation.required" />;
}

interface FormState {
  googleTagManagerContainerID: string;
}

function constructFormState(config: PortalAPIAppConfig): FormState {
  return {
    googleTagManagerContainerID: config.google_tag_manager?.container_id ?? "",
  };
}

function constructConfig(
  config: PortalAPIAppConfig,
  _initialState: FormState,
  currentState: FormState,
  _effectiveConfig: PortalAPIAppConfig
): PortalAPIAppConfig {
  return produce(config, (config) => {
    config.google_tag_manager ??= {};
    if (currentState.googleTagManagerContainerID !== "") {
      config.google_tag_manager.container_id =
        currentState.googleTagManagerContainerID;
    } else {
      delete config.google_tag_manager.container_id;
    }
    clearEmptyObject(config);
  });
}

interface Item {
  iconURL: string;
  name: string;
  description: string;
  connected: boolean;
  hasSavedConnection: boolean;
}

interface AddonProps {
  item: Item;
}

function Addon(props: AddonProps) {
  const { item } = props;
  return (
    <div className={styles.addon}>
      <div className={styles.addonLogo}>
        <img className={styles.addonLogoImage} src={item.iconURL} alt="" />
      </div>
      <Text as="div" size="2" weight="medium" className={styles.addonName}>
        {item.name}
      </Text>
      <Text as="div" size="2" className={styles.addonDescription}>
        {item.description}
      </Text>
    </div>
  );
}

export interface IntegrationsConfigurationContentProps {
  form: AppConfigFormModel<FormState>;
}

const IntegrationsConfigurationContent: React.VFC<IntegrationsConfigurationContentProps> =
  function IntegrationsConfigurationContent({ form }) {
    const { renderToString } = useContext(Context);
    const { initialState, isUpdating, updateError, reset } = form;

    const [gtmDialogOpen, setGtmDialogOpen] = useState(false);
    const [draftContainerID, setDraftContainerID] = useState("");
    const [localContainerIDError, setLocalContainerIDError] =
      useState<React.ReactNode>(null);
    const [pendingGTMAction, setPendingGTMAction] = useState<
      "save" | "delete" | null
    >(null);

    const gtmContainerIDField = useMemo(
      () => ({
        parentJSONPointer: "/google_tag_manager",
        fieldName: "container_id",
      }),
      []
    );

    const containerIDError = useMemo(() => {
      if (updateError == null) {
        return null;
      }
      const apiErrors = parseRawError(updateError);
      const { fieldErrors } = parseAPIErrors(
        apiErrors,
        [gtmContainerIDField],
        []
      );
      for (const [field, errors] of fieldErrors.entries()) {
        if (
          field.fieldName === gtmContainerIDField.fieldName &&
          field.parentJSONPointer === gtmContainerIDField.parentJSONPointer
        ) {
          return errors.length > 0 ? <ErrorRenderer errors={errors} /> : null;
        }
      }
      return null;
    }, [updateError, gtmContainerIDField]);

    const displayContainerIDError = localContainerIDError ?? containerIDError;

    const items: Item[] = useMemo(() => {
      const savedConnected = initialState.googleTagManagerContainerID !== "";
      return [
        {
          iconURL: gtmLogoURL,
          name: renderToString(
            "IntegrationsConfigurationScreen.add-on.gtm.name"
          ),
          description: renderToString(
            "IntegrationsConfigurationScreen.add-on.gtm.description"
          ),
          connected: savedConnected,
          hasSavedConnection: savedConnected,
        },
      ];
    }, [renderToString, initialState.googleTagManagerContainerID]);

    const showStatusColumn = useMemo(
      () => items.some((item) => item.connected),
      [items]
    );

    const onOpenGtmDialog = useCallback(() => {
      reset();
      setLocalContainerIDError(null);
      setPendingGTMAction(null);
      setDraftContainerID(initialState.googleTagManagerContainerID);
      setGtmDialogOpen(true);
    }, [initialState.googleTagManagerContainerID, reset]);

    const onCloseGtmDialog = useCallback(() => {
      if (!isUpdating) {
        setGtmDialogOpen(false);
        setLocalContainerIDError(null);
        setPendingGTMAction(null);
        reset();
      }
    }, [isUpdating, reset]);

    const onGtmDialogOpenChange = useCallback(
      (open: boolean) => {
        if (!open) {
          onCloseGtmDialog();
        }
      },
      [onCloseGtmDialog]
    );

    const onDraftContainerIDChange = useCallback(
      (e: React.ChangeEvent<HTMLInputElement>) => {
        setDraftContainerID(e.target.value);
        setLocalContainerIDError(null);
      },
      []
    );

    const hasSavedGTMConnection =
      initialState.googleTagManagerContainerID !== "";

    const onSaveGTM = useCallback(() => {
      const newID = draftContainerID.trim();
      if (newID === "") {
        setLocalContainerIDError(gtmContainerIDRequiredError());
        return;
      }
      if (!isValidGTMContainerID(newID)) {
        setLocalContainerIDError(gtmContainerIDFormatError());
        return;
      }

      setLocalContainerIDError(null);
      setPendingGTMAction("save");
      form
        .saveWith((prev) => ({
          ...prev,
          googleTagManagerContainerID: newID,
        }))
        .then(() => {
          setGtmDialogOpen(false);
          setPendingGTMAction(null);
        })
        .catch(() => {
          setPendingGTMAction(null);
        });
    }, [draftContainerID, form]);

    const onGtmFormSubmit = useCallback(
      (e: React.FormEvent) => {
        e.preventDefault();
        onSaveGTM();
      },
      [onSaveGTM]
    );

    const onDeleteGTM = useCallback(() => {
      setLocalContainerIDError(null);
      setPendingGTMAction("delete");
      form
        .saveWith((prev) => ({
          ...prev,
          googleTagManagerContainerID: "",
        }))
        .then(() => {
          setGtmDialogOpen(false);
          setPendingGTMAction(null);
        })
        .catch(() => {
          setPendingGTMAction(null);
        });
    }, [form]);

    return (
      <ScreenLayoutScrollView>
        <ScreenContent layout="list">
          <div className={styles.widget}>
            <Text as="p" size="5" weight="bold" className={styles.pageTitle}>
              <FormattedMessage id="IntegrationsConfigurationScreen.title" />
            </Text>
          </div>
          <div className={styles.widget}>
            <div className={styles.tableWrapper}>
              <div className={styles.table}>
                <div className={styles.tableHeader}>
                  <div className={styles.headerCellAddon}>
                    <FormattedMessage id="IntegrationsConfigurationScreen.add-on" />
                  </div>
                  {showStatusColumn ? (
                    <div className={styles.headerCellStatus}>
                      <FormattedMessage id="IntegrationsConfigurationScreen.status" />
                    </div>
                  ) : null}
                  <div className={styles.headerCellAction}>
                    <FormattedMessage id="IntegrationsConfigurationScreen.action" />
                  </div>
                </div>
                {items.map((item) => (
                  <div key={item.name} className={styles.tableRow}>
                    <div className={styles.cellAddon}>
                      <Addon item={item} />
                    </div>
                    {showStatusColumn ? (
                      <div className={styles.cellStatus}>
                        {item.connected ? (
                          <Badge
                            size="1"
                            variant="success"
                            text={
                              <FormattedMessage id="IntegrationsConfigurationScreen.status.connected" />
                            }
                          />
                        ) : null}
                      </div>
                    ) : null}
                    <div className={styles.cellAction}>
                      <button
                        type="button"
                        className={styles.action}
                        onClick={onOpenGtmDialog}
                      >
                        {item.hasSavedConnection ? (
                          <FormattedMessage id="edit" />
                        ) : (
                          <FormattedMessage id="connect" />
                        )}
                      </button>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </ScreenContent>

        <Dialog.Root open={gtmDialogOpen} onOpenChange={onGtmDialogOpenChange}>
          <Dialog.Content maxWidth="400px" size="3">
            <Dialog.Title>
              <FormattedMessage id="IntegrationsConfigurationScreen.add-on.gtm.dialog.title" />
            </Dialog.Title>
            <form onSubmit={onGtmFormSubmit}>
              <TextField
                size="2"
                label={
                  <FormattedMessage id="IntegrationsConfigurationScreen.add-on.gtm.dialog.container-id.label" />
                }
                placeholder={renderToString(
                  "GoogleTagManagerConfigurationScreen.container-id.placeholder"
                )}
                value={draftContainerID}
                onChange={onDraftContainerIDChange}
                error={displayContainerIDError}
              />
              <Flex
                gap="3"
                mt="4"
                justify={hasSavedGTMConnection ? "between" : "end"}
              >
                {hasSavedGTMConnection ? (
                  <Button
                    type="button"
                    size="2"
                    variant="soft"
                    color="red"
                    onClick={onDeleteGTM}
                    loading={pendingGTMAction === "delete"}
                    disabled={isUpdating}
                  >
                    <FormattedMessage id="delete" />
                  </Button>
                ) : null}
                <Flex gap="3">
                  <SecondaryButton
                    size="2"
                    text={<FormattedMessage id="cancel" />}
                    onClick={onCloseGtmDialog}
                    disabled={isUpdating}
                  />
                  <PrimaryButton
                    type="submit"
                    size="2"
                    text={<FormattedMessage id="save" />}
                    loading={pendingGTMAction === "save"}
                    disabled={isUpdating}
                  />
                </Flex>
              </Flex>
            </form>
          </Dialog.Content>
        </Dialog.Root>
      </ScreenLayoutScrollView>
    );
  };

const IntegrationsConfigurationScreen: React.VFC =
  function IntegrationsConfigurationScreen() {
    const { appID } = useParams() as { appID: string };
    const form = useAppConfigForm({
      appID,
      constructFormState,
      constructConfig,
    });

    if (form.isLoading) {
      return <ShowLoading />;
    }

    if (form.loadError) {
      return <ShowError error={form.loadError} onRetry={form.reload} />;
    }

    return <IntegrationsConfigurationContent form={form} />;
  };

export default IntegrationsConfigurationScreen;
