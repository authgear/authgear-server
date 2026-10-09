import React, {
  useContext,
  useMemo,
  useCallback,
  useId,
  useState,
} from "react";
import { useLocation, useNavigate, useParams } from "react-router-dom";
import {
  Button,
  Dialog,
  Flex,
  Heading,
  Select,
  Text,
  VisuallyHidden,
} from "@radix-ui/themes";
import { FormattedMessage, Context } from "../../intl";
import { produce } from "immer";
import {
  AppSecretConfigFormModel,
  useAppSecretConfigForm,
} from "../../hook/useAppSecretConfigForm";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import ScreenContent from "../../ScreenContent";
import ScreenLayoutScrollView from "../../ScreenLayoutScrollView";
import { Badge } from "../../components/v2/Badge/Badge";
import {
  PortalAPIAppConfig,
  PortalAPISecretConfig,
  PortalAPISecretConfigUpdateInstruction,
} from "../../types";
import { TextField } from "../../components/v2/TextField/TextField";
import { FormField } from "../../components/v2/FormField/FormField";
import { Callout } from "../../components/v2/Callout/Callout";
import { PrimaryButton } from "../../components/v2/Button/PrimaryButton/PrimaryButton";
import { SecondaryButton } from "../../components/v2/Button/SecondaryButton/SecondaryButton";
import {
  FormField as ErrorFormField,
  ParsedAPIError,
  parseAPIErrors,
  parseRawError,
} from "../../error/parse";
import ErrorRenderer from "../../ErrorRenderer";
import ExternalLink from "../../ExternalLink";
import { clearEmptyObject } from "../../util/misc";
import {
  isAuditLogStreamingAvailable,
  isGoogleTagManagerAvailable,
} from "../../util/integrations";
import { useAppFeatureConfigQuery } from "./query/appFeatureConfigQuery";
import { useAppSecretVisitToken } from "./mutations/generateAppSecretVisitTokenMutation";
import { AppSecretKey } from "./globalTypes.generated";
import { useLocationEffect } from "../../hook/useLocationEffect";
import { useSessionStorage } from "../../hook/useSessionStorage";
import { startReauthentication } from "./Authenticated";
import {
  DATADOG_SITE_OPTIONS,
  DATADOG_SITE_OTHER,
  applyDatadogConnect,
  applyDatadogDelete,
  applyDatadogEdit,
  datadogAppOrigin,
  findManagedDatadogStream,
  formValueToSite,
  remainingStreamNames,
  siteToFormValue,
} from "./integrations/datadog";
import styles from "./IntegrationsConfigurationScreen.module.css";

import gtmLogoURL from "../../images/gtm_logo.png";
import datadogLogoURL from "../../images/datadog_logo.svg";

const DATADOG_API_KEYS_PATH = "/organization-settings/api-keys";

const MASKED_SECRET = "***************";

const DATADOG_DRAFT_STORAGE_KEY = "integrations-config-screen-datadog-draft";

interface LocationState {
  isOAuthRedirect: boolean;
}
function isLocationState(raw: unknown): raw is LocationState {
  return (
    raw != null &&
    typeof raw === "object" &&
    (raw as Partial<LocationState>).isOAuthRedirect != null
  );
}

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

function requiredError(): React.ReactNode {
  return <FormattedMessage id="errors.validation.required" />;
}

type DatadogAction = "none" | "connect" | "edit" | "delete";

interface DatadogFormState {
  connected: boolean;
  usesEndpoint: boolean;
  endpoint: string;
  siteOption: string;
  siteOther: string;
  // null until the key is revealed with a secret visit token.
  originalApiKey: string | null;
  apiKey: string;
  action: DatadogAction;
}

interface FormState {
  googleTagManagerContainerID: string;
  datadog: DatadogFormState;
}

function constructFormState(
  config: PortalAPIAppConfig,
  secrets: PortalAPISecretConfig
): FormState {
  const stream = findManagedDatadogStream(config);
  const storedSecret =
    stream != null
      ? secrets.telemetryAuditLogStreamSecrets?.datadog?.find(
          (s) => s.streamName === stream.name
        )
      : undefined;
  const originalApiKey = storedSecret?.apiKey ?? null;
  const endpoint = stream?.http?.endpoint;
  const site = siteToFormValue(stream?.datadog?.site);
  return {
    googleTagManagerContainerID: config.google_tag_manager?.container_id ?? "",
    datadog: {
      connected: storedSecret != null,
      usesEndpoint: endpoint != null,
      endpoint: endpoint ?? "",
      siteOption: site.option,
      siteOther: site.other,
      originalApiKey,
      apiKey: originalApiKey ?? "",
      action: "none",
    },
  };
}

function constructConfig(
  config: PortalAPIAppConfig,
  secrets: PortalAPISecretConfig,
  _initialState: FormState,
  currentState: FormState,
  _effectiveConfig: PortalAPIAppConfig
): [PortalAPIAppConfig, PortalAPISecretConfig] {
  let newConfig = produce(config, (config) => {
    config.google_tag_manager ??= {};
    if (currentState.googleTagManagerContainerID !== "") {
      config.google_tag_manager.container_id =
        currentState.googleTagManagerContainerID;
    } else {
      delete config.google_tag_manager.container_id;
    }
    clearEmptyObject(config);
  });

  const { datadog } = currentState;
  const site = formValueToSite(datadog.siteOption, datadog.siteOther);
  switch (datadog.action) {
    case "connect":
      newConfig = applyDatadogConnect(newConfig, site);
      break;
    case "edit":
      newConfig = applyDatadogEdit(newConfig, site);
      break;
    case "delete":
      newConfig = applyDatadogDelete(newConfig);
      break;
    case "none":
      break;
  }

  return [newConfig, secrets];
}

function constructSecretUpdateInstruction(
  config: PortalAPIAppConfig,
  _secrets: PortalAPISecretConfig,
  currentState: FormState
): PortalAPISecretConfigUpdateInstruction | undefined {
  const { datadog } = currentState;
  const apiKey = datadog.apiKey.trim();
  switch (datadog.action) {
    case "connect":
    case "edit": {
      // A blank key on edit means the masked key was never unlocked.
      if (apiKey === "" || apiKey === datadog.originalApiKey) {
        return undefined;
      }
      const stream = findManagedDatadogStream(config);
      if (stream == null) {
        return undefined;
      }
      return {
        telemetryAuditLogStreamSecrets: {
          action: "set",
          setData: { datadog: [{ streamName: stream.name, apiKey }] },
        },
      };
    }
    case "delete":
      return {
        telemetryAuditLogStreamSecrets: {
          action: "cleanup",
          cleanupData: { keepStreamNames: remainingStreamNames(config) },
        },
      };
    case "none":
      return undefined;
  }
}

const datadogSiteField: ErrorFormField = {
  parentJSONPointer: /^\/telemetry\/audit_logs\/streams\/\d+\/datadog$/,
  fieldName: "site",
};

interface Item {
  id: "gtm" | "datadog";
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
  form: AppSecretConfigFormModel<FormState>;
  gtmAvailable: boolean;
  datadogAvailable: boolean;
}

const IntegrationsConfigurationContent: React.VFC<IntegrationsConfigurationContentProps> =
  function IntegrationsConfigurationContent({
    form,
    gtmAvailable,
    datadogAvailable,
  }) {
    const { renderToString } = useContext(Context);
    // Nothing on this screen calls setState, so state is the saved state.
    const { state: initialState, isUpdating, updateError, reset } = form;

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
      const result: Item[] = [];
      if (gtmAvailable) {
        const savedConnected = initialState.googleTagManagerContainerID !== "";
        result.push({
          id: "gtm",
          iconURL: gtmLogoURL,
          name: renderToString(
            "IntegrationsConfigurationScreen.add-on.gtm.name"
          ),
          description: renderToString(
            "IntegrationsConfigurationScreen.add-on.gtm.description"
          ),
          connected: savedConnected,
          hasSavedConnection: savedConnected,
        });
      }
      if (datadogAvailable) {
        result.push({
          id: "datadog",
          iconURL: datadogLogoURL,
          name: renderToString(
            "IntegrationsConfigurationScreen.add-on.datadog.name"
          ),
          description: renderToString(
            "IntegrationsConfigurationScreen.add-on.datadog.description"
          ),
          connected: initialState.datadog.connected,
          hasSavedConnection: initialState.datadog.connected,
        });
      }
      return result;
    }, [
      renderToString,
      gtmAvailable,
      datadogAvailable,
      initialState.googleTagManagerContainerID,
      initialState.datadog.connected,
    ]);

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
        setLocalContainerIDError(requiredError());
        return;
      }
      if (!isValidGTMContainerID(newID)) {
        setLocalContainerIDError(gtmContainerIDFormatError());
        return;
      }

      setLocalContainerIDError(null);
      setPendingGTMAction("save");
      form
        .saveWithState({
          ...form.state,
          googleTagManagerContainerID: newID,
        })
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
        .saveWithState({
          ...form.state,
          googleTagManagerContainerID: "",
        })
        .then(() => {
          setGtmDialogOpen(false);
          setPendingGTMAction(null);
        })
        .catch(() => {
          setPendingGTMAction(null);
        });
    }, [form]);

    const [datadogDialogOpen, setDatadogDialogOpen] = useState(false);
    const [datadogDraft, setDatadogDraft] = useState<DatadogFormState>(
      initialState.datadog
    );
    const [localApiKeyError, setLocalApiKeyError] =
      useState<React.ReactNode>(null);
    const [localSiteOtherError, setLocalSiteOtherError] =
      useState<React.ReactNode>(null);
    const [pendingDatadogAction, setPendingDatadogAction] = useState<
      "save" | "delete" | null
    >(null);
    const [datadogApiKeyEditing, setDatadogApiKeyEditing] = useState(false);
    const datadogSiteSelectID = useId();
    const datadogApiKeyInputID = useId();
    const datadogSiteOtherInputID = useId();

    // Revealing the key needs a reauth redirect, which unmounts this dialog.
    const [
      storedDatadogDraft,
      setStoredDatadogDraft,
      removeStoredDatadogDraft,
    ] = useSessionStorage<DatadogFormState | null>(
      DATADOG_DRAFT_STORAGE_KEY,
      null
    );

    useLocationEffect((state: LocationState) => {
      if (state.isOAuthRedirect && storedDatadogDraft != null) {
        setDatadogDraft({
          ...storedDatadogDraft,
          originalApiKey: initialState.datadog.originalApiKey,
          apiKey: initialState.datadog.originalApiKey ?? "",
        });
        setDatadogApiKeyEditing(true);
        setDatadogDialogOpen(true);
        removeStoredDatadogDraft();
      }
    });

    const { datadogSiteError, datadogTopErrors } = useMemo((): {
      datadogSiteError: React.ReactNode;
      datadogTopErrors: ParsedAPIError[];
    } => {
      if (updateError == null) {
        return { datadogSiteError: null, datadogTopErrors: [] };
      }
      const apiErrors = parseRawError(updateError);
      const { fieldErrors, topErrors, conflictErrors } = parseAPIErrors(
        apiErrors,
        [datadogSiteField],
        []
      );
      const siteErrors = fieldErrors.get(datadogSiteField) ?? [];
      return {
        datadogSiteError:
          siteErrors.length > 0 ? <ErrorRenderer errors={siteErrors} /> : null,
        datadogTopErrors: [
          ...topErrors,
          // The project changed since this page loaded; reloading fixes it.
          ...(conflictErrors.length > 0
            ? [{ messageID: "FormConfirmOverridingDialog.title" }]
            : []),
        ],
      };
    }, [updateError]);

    const resetDatadogDialog = useCallback(() => {
      setLocalApiKeyError(null);
      setLocalSiteOtherError(null);
      setPendingDatadogAction(null);
    }, []);

    const onOpenDatadogDialog = useCallback(() => {
      reset();
      resetDatadogDialog();
      setDatadogDraft(initialState.datadog);
      setDatadogApiKeyEditing(!initialState.datadog.connected);
      setDatadogDialogOpen(true);
    }, [initialState.datadog, reset, resetDatadogDialog]);

    const navigate = useNavigate();
    const onClickEditDatadogAPIKey = useCallback(() => {
      if (datadogDraft.originalApiKey != null) {
        setDatadogApiKeyEditing(true);
        return;
      }

      const locationState: LocationState = {
        isOAuthRedirect: true,
      };

      setStoredDatadogDraft(datadogDraft);

      startReauthentication(navigate, locationState).catch((e) => {
        console.error(e);
        removeStoredDatadogDraft();
      });
    }, [
      datadogDraft,
      navigate,
      removeStoredDatadogDraft,
      setStoredDatadogDraft,
    ]);

    const onCloseDatadogDialog = useCallback(() => {
      if (!isUpdating) {
        setDatadogDialogOpen(false);
        resetDatadogDialog();
        reset();
      }
    }, [isUpdating, reset, resetDatadogDialog]);

    const onDatadogDialogOpenChange = useCallback(
      (open: boolean) => {
        if (!open) {
          onCloseDatadogDialog();
        }
      },
      [onCloseDatadogDialog]
    );

    const onDatadogSiteOptionChange = useCallback((value: string) => {
      setDatadogDraft((prev) => ({ ...prev, siteOption: value }));
      setLocalSiteOtherError(null);
    }, []);

    const onDatadogSiteOtherChange = useCallback(
      (e: React.ChangeEvent<HTMLInputElement>) => {
        const value = e.target.value;
        setDatadogDraft((prev) => ({ ...prev, siteOther: value }));
        setLocalSiteOtherError(null);
      },
      []
    );

    const onDatadogAPIKeyChange = useCallback(
      (e: React.ChangeEvent<HTMLInputElement>) => {
        const value = e.target.value;
        setDatadogDraft((prev) => ({ ...prev, apiKey: value }));
        setLocalApiKeyError(null);
      },
      []
    );

    const hasSavedDatadogConnection = initialState.datadog.connected;

    const onSaveDatadog = useCallback(() => {
      const apiKeyError =
        datadogApiKeyEditing && datadogDraft.apiKey.trim() === ""
          ? requiredError()
          : null;
      const siteOtherError =
        !datadogDraft.usesEndpoint &&
        datadogDraft.siteOption === DATADOG_SITE_OTHER &&
        datadogDraft.siteOther.trim() === ""
          ? requiredError()
          : null;
      setLocalApiKeyError(apiKeyError);
      setLocalSiteOtherError(siteOtherError);
      if (apiKeyError != null || siteOtherError != null) {
        return;
      }

      setPendingDatadogAction("save");
      form
        .saveWithState({
          ...form.state,
          datadog: {
            ...datadogDraft,
            action: hasSavedDatadogConnection ? "edit" : "connect",
          },
        })
        .then(() => {
          setDatadogDialogOpen(false);
          setPendingDatadogAction(null);
        })
        .catch(() => {
          setPendingDatadogAction(null);
        });
    }, [datadogApiKeyEditing, datadogDraft, form, hasSavedDatadogConnection]);

    const onDatadogFormSubmit = useCallback(
      (e: React.FormEvent) => {
        e.preventDefault();
        onSaveDatadog();
      },
      [onSaveDatadog]
    );

    const onDeleteDatadog = useCallback(() => {
      setLocalApiKeyError(null);
      setLocalSiteOtherError(null);
      setPendingDatadogAction("delete");
      form
        .saveWithState({
          ...form.state,
          datadog: { ...form.state.datadog, action: "delete" },
        })
        .then(() => {
          setDatadogDialogOpen(false);
          setPendingDatadogAction(null);
        })
        .catch(() => {
          setPendingDatadogAction(null);
        });
    }, [form]);

    const onClickItemAction = useCallback(
      (id: Item["id"]) => {
        switch (id) {
          case "gtm":
            onOpenGtmDialog();
            break;
          case "datadog":
            onOpenDatadogDialog();
            break;
        }
      },
      [onOpenGtmDialog, onOpenDatadogDialog]
    );

    const isDatadogSiteOther = datadogDraft.siteOption === DATADOG_SITE_OTHER;
    // Endpoint streams have no site; an empty Other value also falls back.
    const datadogAPIKeysURL =
      datadogAppOrigin(
        datadogDraft.usesEndpoint
          ? ""
          : formValueToSite(datadogDraft.siteOption, datadogDraft.siteOther)
      ) + DATADOG_API_KEYS_PATH;

    return (
      <ScreenLayoutScrollView>
        <ScreenContent layout="list">
          <div className={styles.widget}>
            <Heading
              as="h1"
              size="5"
              weight="bold"
              className={styles.pageTitle}
            >
              <FormattedMessage id="IntegrationsConfigurationScreen.title" />
            </Heading>
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
                  <div key={item.id} className={styles.tableRow}>
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
                        onClick={() => onClickItemAction(item.id)}
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

        <Dialog.Root
          open={datadogDialogOpen}
          onOpenChange={onDatadogDialogOpenChange}
        >
          <Dialog.Content maxWidth="400px" size="3">
            <Dialog.Title>
              <FormattedMessage id="IntegrationsConfigurationScreen.add-on.datadog.dialog.title" />
            </Dialog.Title>
            <form onSubmit={onDatadogFormSubmit}>
              <Flex direction="column" gap="4">
                {datadogTopErrors.length > 0 ? (
                  <Callout
                    type="error"
                    showCloseButton={false}
                    text={<ErrorRenderer errors={datadogTopErrors} />}
                  />
                ) : null}
                {datadogDraft.usesEndpoint ? (
                  <TextField
                    size="2"
                    readOnly={true}
                    label={
                      <FormattedMessage id="IntegrationsConfigurationScreen.add-on.datadog.dialog.endpoint.label" />
                    }
                    value={datadogDraft.endpoint}
                  />
                ) : (
                  <>
                    <FormField
                      size="2"
                      labelSpace="1"
                      htmlFor={datadogSiteSelectID}
                      label={
                        <FormattedMessage id="IntegrationsConfigurationScreen.add-on.datadog.dialog.site.label" />
                      }
                      error={isDatadogSiteOther ? null : datadogSiteError}
                    >
                      <Select.Root
                        value={datadogDraft.siteOption}
                        onValueChange={onDatadogSiteOptionChange}
                      >
                        <Select.Trigger
                          id={datadogSiteSelectID}
                          variant="surface"
                        />
                        <Select.Content>
                          {DATADOG_SITE_OPTIONS.map((option) => (
                            <Select.Item key={option.site} value={option.site}>
                              <FormattedMessage
                                id="IntegrationsConfigurationScreen.add-on.datadog.dialog.site.option"
                                values={{
                                  label: option.label,
                                  site: option.site,
                                }}
                              />
                            </Select.Item>
                          ))}
                          <Select.Item value={DATADOG_SITE_OTHER}>
                            <FormattedMessage id="IntegrationsConfigurationScreen.add-on.datadog.dialog.site.other" />
                          </Select.Item>
                        </Select.Content>
                      </Select.Root>
                    </FormField>
                    {isDatadogSiteOther ? (
                      <>
                        <VisuallyHidden>
                          <label htmlFor={datadogSiteOtherInputID}>
                            <FormattedMessage id="IntegrationsConfigurationScreen.add-on.datadog.dialog.site.label" />
                          </label>
                        </VisuallyHidden>
                        <TextField
                          id={datadogSiteOtherInputID}
                          size="2"
                          placeholder={renderToString(
                            "IntegrationsConfigurationScreen.add-on.datadog.dialog.site.other.placeholder"
                          )}
                          value={datadogDraft.siteOther}
                          onChange={onDatadogSiteOtherChange}
                          error={localSiteOtherError ?? datadogSiteError}
                        />
                      </>
                    ) : null}
                  </>
                )}
                <FormField
                  size="2"
                  labelSpace="1"
                  htmlFor={datadogApiKeyInputID}
                  label={
                    <FormattedMessage id="IntegrationsConfigurationScreen.add-on.datadog.dialog.api-key.label" />
                  }
                  error={localApiKeyError}
                  hint={
                    <FormattedMessage
                      id="IntegrationsConfigurationScreen.add-on.datadog.dialog.api-key.hint"
                      values={{
                        // eslint-disable-next-line react/no-unstable-nested-components
                        ExternalLink: (chunks: React.ReactNode) => (
                          <ExternalLink href={datadogAPIKeysURL}>
                            {chunks}
                          </ExternalLink>
                        ),
                      }}
                    />
                  }
                >
                  <div className={styles.apiKeyInputRow}>
                    <TextField.Input
                      id={datadogApiKeyInputID}
                      size="2"
                      inputClassName={styles.apiKeyInput}
                      type="text"
                      value={
                        datadogApiKeyEditing ||
                        datadogDraft.originalApiKey != null
                          ? datadogDraft.apiKey
                          : MASKED_SECRET
                      }
                      onChange={onDatadogAPIKeyChange}
                      readOnly={!datadogApiKeyEditing}
                      error={localApiKeyError}
                    >
                      {null}
                    </TextField.Input>
                    {!datadogApiKeyEditing ? (
                      <SecondaryButton
                        size="2"
                        onClick={onClickEditDatadogAPIKey}
                        text={<FormattedMessage id="edit" />}
                        disabled={isUpdating}
                      />
                    ) : null}
                  </div>
                </FormField>
                <Text as="p" size="1" color="gray">
                  <FormattedMessage id="IntegrationsConfigurationScreen.add-on.datadog.dialog.delivery-note" />
                </Text>
              </Flex>
              <Flex
                gap="3"
                mt="4"
                justify={hasSavedDatadogConnection ? "between" : "end"}
              >
                {hasSavedDatadogConnection ? (
                  <Button
                    type="button"
                    size="2"
                    variant="soft"
                    color="red"
                    onClick={onDeleteDatadog}
                    loading={pendingDatadogAction === "delete"}
                    disabled={isUpdating}
                  >
                    <FormattedMessage id="delete" />
                  </Button>
                ) : null}
                <Flex gap="3">
                  <SecondaryButton
                    size="2"
                    text={<FormattedMessage id="cancel" />}
                    onClick={onCloseDatadogDialog}
                    disabled={isUpdating}
                  />
                  <PrimaryButton
                    type="submit"
                    size="2"
                    text={<FormattedMessage id="save" />}
                    loading={pendingDatadogAction === "save"}
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

const IntegrationsConfigurationScreen1: React.VFC<{
  appID: string;
  secretToken: string | null;
}> = function IntegrationsConfigurationScreen1({ appID, secretToken }) {
  const form = useAppSecretConfigForm({
    appID,
    secretVisitToken: secretToken,
    constructFormState,
    constructConfig,
    constructSecretUpdateInstruction,
  });
  const featureConfig = useAppFeatureConfigQuery(appID);

  if (form.isLoading || featureConfig.isLoading) {
    return <ShowLoading />;
  }

  if (form.loadError) {
    return <ShowError error={form.loadError} onRetry={form.reload} />;
  }

  if (featureConfig.loadError) {
    return (
      <ShowError
        error={featureConfig.loadError}
        onRetry={featureConfig.reload}
      />
    );
  }

  return (
    <IntegrationsConfigurationContent
      form={form}
      gtmAvailable={isGoogleTagManagerAvailable(
        featureConfig.effectiveFeatureConfig
      )}
      datadogAvailable={isAuditLogStreamingAvailable(
        featureConfig.effectiveFeatureConfig
      )}
    />
  );
};

const SECRETS = [AppSecretKey.TelemetryAuditLogStreamSecrets];

const IntegrationsConfigurationScreen: React.VFC =
  function IntegrationsConfigurationScreen() {
    const { appID } = useParams() as { appID: string };
    const location = useLocation();
    const [shouldRefreshToken] = useState<boolean>(() => {
      const { state } = location;
      if (isLocationState(state) && state.isOAuthRedirect) {
        return true;
      }
      return false;
    });
    const { token, error, retry } = useAppSecretVisitToken(
      appID,
      SECRETS,
      shouldRefreshToken
    );
    if (error) {
      return <ShowError error={error} onRetry={retry} />;
    }

    if (token === undefined) {
      return <ShowLoading />;
    }

    return (
      <IntegrationsConfigurationScreen1 appID={appID} secretToken={token} />
    );
  };

export default IntegrationsConfigurationScreen;
