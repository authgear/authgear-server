import React, {
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import cn from "classnames";
import produce from "immer";
import { useNavigate, useParams, useSearchParams } from "react-router-dom";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import {
  ActionButton,
  DefaultButton,
  DetailsList,
  Dialog,
  DialogFooter,
  IColumn,
  IDetailsListProps,
  IDialogProps,
  MessageBar,
  MessageBarType,
  SelectionMode,
  Text,
  TextField,
  VerticalDivider,
} from "@fluentui/react";
import { Domain, useDomainsQuery } from "./query/domainsQuery";
import { useCreateDomainMutation } from "./mutations/createDomainMutation";
import { useDeleteDomainMutation } from "./mutations/deleteDomainMutation";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import ButtonWithLoading from "../../ButtonWithLoading";
import { useTextField } from "../../hook/useInput";
import {
  ErrorParseRule,
  parseAPIErrors,
  parseRawError,
  renderErrors,
} from "../../error/parse";
import ErrorDialog from "../../error/ErrorDialog";
import { useSystemConfig } from "../../context/SystemConfigContext";
import NavBreadcrumb, { BreadcrumbItem } from "../../NavBreadcrumb";

import styles from "./CustomDomainListScreen.module.scss";
import { PortalAPIAppConfig } from "../../types";
import { clearEmptyObject } from "../../util/misc";
import {
  AppConfigFormModel,
  useAppConfigForm,
} from "../../hook/useAppConfigForm";

function getOriginFromDomain(domain: string): string {
  // assume domain has no scheme
  // use https scheme
  return `https://${domain}`;
}

function getHostFromOrigin(urlOrigin: string): string {
  try {
    return new URL(urlOrigin).host;
  } catch (_: unknown) {
    return "";
  }
}

interface DomainListItem {
  id?: string;
  domain: string;
  urlOrigin: string;
  isVerified: boolean;
  isCustom: boolean;
  isPublicOrigin: boolean;
}

interface FormState {
  publicOrigin: string;
}

function constructFormState(config: PortalAPIAppConfig): FormState {
  return {
    publicOrigin: config.http?.public_origin ?? "",
  };
}

function constructConfig(
  config: PortalAPIAppConfig,
  initialState: FormState,
  currentState: FormState
): PortalAPIAppConfig {
  return produce(config, (config) => {
    config.http ??= {};
    if (currentState.publicOrigin !== initialState.publicOrigin) {
      config.http.public_origin = currentState.publicOrigin;
    }
    clearEmptyObject(config);
  });
}

function makeDomainListColumn(renderToString: (messageID: string) => string) {
  return [
    {
      key: "domain",
      name: renderToString("CustomDomainListScreen.domain-list.header.domain"),
      minWidth: 250,
      className: styles.domainListColumn,
    },
    {
      key: "isVerified",
      name: renderToString("CustomDomainListScreen.domain-list.header.status"),
      minWidth: 100,
      className: styles.domainListColumn,
    },
    {
      key: "action",
      name: renderToString("action"),
      minWidth: 150,
      className: styles.domainListColumn,
    },
  ];
}

const AddDomainSection: React.FC = function AddDomainSection() {
  const { renderToString } = useContext(Context);
  const { appID } = useParams();

  const [newDomain, setNewDomain] = useState("");
  const { onChange: onNewDomainChange } = useTextField((value) => {
    setNewDomain(value);
  });

  const {
    createDomain,
    loading: creatingDomain,
    error: createDomainError,
  } = useCreateDomainMutation(appID);

  const onAddClick = useCallback(
    (ev: React.FormEvent) => {
      ev.preventDefault();
      ev.stopPropagation();

      createDomain(newDomain)
        .then((success) => {
          if (success) {
            onNewDomainChange(null, "");
          }
        })
        .catch(() => {});
    },
    [createDomain, newDomain, onNewDomainChange]
  );

  const isModified = useMemo(() => {
    return newDomain !== "";
  }, [newDomain]);

  const errorRules: ErrorParseRule[] = useMemo(() => {
    return [
      {
        errorMessageID: "CustomDomainListScreen.add-domain.duplicated-error",
        reason: "DuplicatedDomain",
      },
      {
        errorMessageID: "CustomDomainListScreen.add-domain.invalid-error",
        reason: "InvalidDomain",
      },
    ];
  }, []);

  const errors = useMemo(() => {
    const apiErrors = parseRawError(createDomainError);
    const { topErrors } = parseAPIErrors(
      apiErrors,
      [],
      errorRules,
      "CustomDomainListScreen.add-domain.generic-error"
    );
    return topErrors;
  }, [createDomainError, errorRules]);
  const errorMessage = renderErrors(null, errors, renderToString);

  return (
    <form className={styles.addDomain} onSubmit={onAddClick}>
      <TextField
        className={styles.addDomainField}
        placeholder={renderToString(
          "CustomDomainListScreen.domain-list.add-domain.placeholder"
        )}
        value={newDomain}
        onChange={onNewDomainChange}
        errorMessage={errorMessage}
      />
      <ButtonWithLoading
        type="submit"
        className={styles.addDomainButton}
        disabled={!isModified}
        iconProps={{ iconName: "CircleAdditionSolid" }}
        loading={creatingDomain}
        labelId="add"
      />
    </form>
  );
};

interface DomainListActionButtonsProps {
  domainID?: string;
  domain: string;
  urlOrigin: string;
  isCustomDomain: boolean;
  isVerified: boolean;
  isPublicOrigin: boolean;
  onDeleteClick: (domainID: string, domain: string) => void;
  onSetPublicOriginClick: (urlOrigin: string) => void;
}

// eslint-disable-next-line complexity
const DomainListActionButtons: React.FC<DomainListActionButtonsProps> = function DomainListActionButtons(
  props
) {
  const {
    domainID,
    domain,
    urlOrigin,
    isCustomDomain,
    isVerified,
    isPublicOrigin,
    onDeleteClick: onDeleteClickProps,
    onSetPublicOriginClick: onSetPublicOriginClickProps,
  } = props;

  const { themes } = useSystemConfig();
  const navigate = useNavigate();

  const showDelete = domainID && isCustomDomain && !isPublicOrigin;
  const showVerify = domainID && !isVerified;
  const showSetPublicOrigin = domainID && isVerified && !isPublicOrigin;

  const onSetPublicOriginClick = useCallback(() => {
    onSetPublicOriginClickProps(urlOrigin);
  }, [urlOrigin, onSetPublicOriginClickProps]);

  const onVerifyClicked = useCallback(() => {
    navigate(`./${domainID}/verify`);
  }, [domainID, navigate]);

  const onDeleteClick = useCallback(() => {
    if (!domainID) {
      return;
    }
    onDeleteClickProps(domainID, domain);
  }, [domainID, domain, onDeleteClickProps]);

  if (!showSetPublicOrigin && !showDelete && !showVerify) {
    return (
      <section className={styles.actionButtonContainer}>
        <Text>---</Text>
      </section>
    );
  }

  const buttonNodes: React.ReactNode[] = [];
  if (showSetPublicOrigin) {
    buttonNodes.push(
      <ActionButton
        key={`${domainID}-domain-set-public-origin`}
        className={styles.actionButton}
        theme={themes.actionButton}
        onClick={onSetPublicOriginClick}
      >
        <FormattedMessage id="activate" />
      </ActionButton>
    );
  }

  if (showVerify) {
    buttonNodes.push(
      <ActionButton
        key={`${domainID}-domain-verify`}
        className={styles.actionButton}
        theme={themes.actionButton}
        onClick={onVerifyClicked}
      >
        <FormattedMessage id="verify" />
      </ActionButton>
    );
  }

  if (showDelete) {
    buttonNodes.push(
      <ActionButton
        key={`${domainID}-domain-delete`}
        className={styles.actionButton}
        theme={themes.destructive}
        onClick={onDeleteClick}
      >
        <FormattedMessage id="delete" />
      </ActionButton>
    );
  }

  return (
    <section className={styles.actionButtonContainer}>
      {buttonNodes.map((node, idx) => {
        if (idx !== 0) {
          return [
            <VerticalDivider
              key={`${domainID}-domain-action-divider-${idx}`}
              className={styles.divider}
            />,
            node,
          ];
        }
        return node;
      })}
    </section>
  );
};

interface DeleteDomainDialogProps {
  domainID: string;
  domain: string;
  visible: boolean;
  dismissDialog: () => void;
}

const DeleteDomainDialog: React.FC<DeleteDomainDialogProps> = function DeleteDomainDialog(
  props: DeleteDomainDialogProps
) {
  const { domain, domainID, visible, dismissDialog } = props;
  const { appID } = useParams();
  const { renderToString } = useContext(Context);
  const { themes } = useSystemConfig();

  const {
    deleteDomain,
    loading: deletingDomain,
    error: deleteDomainError,
  } = useDeleteDomainMutation(appID);

  const onConfirmClick = useCallback(() => {
    deleteDomain(domainID)
      .catch(() => {})
      .finally(() => {
        dismissDialog();
      });
  }, [domainID, deleteDomain, dismissDialog]);

  const errorRules: ErrorParseRule[] = [
    {
      reason: "Forbidden",
      errorMessageID:
        "CustomDomainListScreen.delete-domain-dialog.forbidden-error",
    },
  ];

  const dialogContentProps: IDialogProps["dialogContentProps"] = useMemo(() => {
    return {
      title: (
        <FormattedMessage id="CustomDomainListScreen.delete-domain-dialog.title" />
      ),
      subText: renderToString(
        "CustomDomainListScreen.delete-domain-dialog.message",
        { domain }
      ),
    };
  }, [renderToString, domain]);

  return (
    <>
      <Dialog
        hidden={!visible}
        dialogContentProps={dialogContentProps}
        modalProps={{ isBlocking: deletingDomain }}
        onDismiss={dismissDialog}
      >
        <DialogFooter>
          <ButtonWithLoading
            theme={themes.destructive}
            loading={deletingDomain}
            onClick={onConfirmClick}
            disabled={!visible}
            labelId="confirm"
          />
          <DefaultButton
            onClick={dismissDialog}
            disabled={deletingDomain || !visible}
          >
            <FormattedMessage id="cancel" />
          </DefaultButton>
        </DialogFooter>
      </Dialog>
      <ErrorDialog
        error={deleteDomainError}
        rules={errorRules}
        fallbackErrorMessageID="CustomDomainListScreen.delete-domain-dialog.generic-error"
      />
    </>
  );
};

interface UpdatePublicOriginDialogProps {
  urlOrigin: string;
  visible: boolean;
  isSaving: boolean;
  updateError: unknown;
  onConfirmClick: () => void;
  dismissDialog: () => void;
}

const UpdatePublicOriginDialog: React.FC<UpdatePublicOriginDialogProps> = function UpdatePublicOriginDialog(
  props: UpdatePublicOriginDialogProps
) {
  const {
    urlOrigin,
    visible,
    isSaving,
    updateError,
    onConfirmClick,
    dismissDialog,
  } = props;
  const { renderToString } = useContext(Context);
  const { themes } = useSystemConfig();

  const dialogContentProps: IDialogProps["dialogContentProps"] = useMemo(() => {
    return {
      title: (
        <FormattedMessage id="CustomDomainListScreen.activate-domain-dialog.title" />
      ),
      subText: renderToString(
        "CustomDomainListScreen.activate-domain-dialog.message",
        { domain: getHostFromOrigin(urlOrigin) }
      ),
    };
  }, [renderToString, urlOrigin]);

  return (
    <>
      <Dialog
        hidden={!visible}
        dialogContentProps={dialogContentProps}
        modalProps={{ isBlocking: isSaving }}
        onDismiss={dismissDialog}
      >
        <DialogFooter>
          <ButtonWithLoading
            theme={themes.actionButton}
            loading={isSaving}
            onClick={onConfirmClick}
            disabled={!visible}
            labelId="confirm"
          />
          <DefaultButton
            onClick={dismissDialog}
            disabled={isSaving || !visible}
          >
            <FormattedMessage id="cancel" />
          </DefaultButton>
        </DialogFooter>
      </Dialog>
      <ErrorDialog
        error={updateError}
        rules={[]}
        fallbackErrorMessageID="CustomDomainListScreen.activate-domain-dialog.generic-error"
      />
    </>
  );
};

interface CustomDomainListContentProps {
  domains: Domain[];
  appConfigForm: AppConfigFormModel<FormState>;
}

const CustomDomainListContent: React.FC<CustomDomainListContentProps> = function CustomDomainListContent(
  props
) {
  const {
    domains,
    appConfigForm: {
      state,
      setState,
      isDirty,
      isUpdating,
      save,
      reset,
      updateError,
    },
  } = props;

  const { renderToString } = useContext(Context);

  const navBreadcrumbItems: BreadcrumbItem[] = useMemo(() => {
    return [
      {
        to: ".",
        label: <FormattedMessage id="CustomDomainListScreen.title" />,
      },
    ];
  }, []);

  interface DeleteDomainDialogData {
    domainID: string;
    domain: string;
  }
  const [
    deleteDomainDialogVisible,
    setConfirmDeleteDomainDialogVisible,
  ] = useState(false);
  const [
    deleteDomainDialogData,
    setDeleteDomainDialogData,
  ] = useState<DeleteDomainDialogData>({ domainID: "", domain: "" });

  const domainListColumns: IColumn[] = useMemo(() => {
    return makeDomainListColumn(renderToString);
  }, [renderToString]);

  const savedPublicOriginRef = useRef<string>("");
  useEffect(() => {
    if (!isDirty) {
      savedPublicOriginRef.current = state.publicOrigin;
    }
  }, [isDirty, state]);
  const prevSavedPublicOrigin = savedPublicOriginRef.current;

  const domainListItems: DomainListItem[] = useMemo(() => {
    const list: DomainListItem[] = domains.map((domain) => {
      const urlOrigin = getOriginFromDomain(domain.domain);
      const isPublicOrigin = urlOrigin === prevSavedPublicOrigin;
      return {
        id: domain.id,
        domain: domain.domain,
        urlOrigin: urlOrigin,
        isVerified: domain.isVerified,
        isCustom: domain.isCustom,
        isPublicOrigin: isPublicOrigin,
      };
    });
    const found = list.find((domain) => domain.isPublicOrigin);

    if (!found) {
      list.unshift({
        domain:
          getHostFromOrigin(prevSavedPublicOrigin) || prevSavedPublicOrigin,
        urlOrigin: prevSavedPublicOrigin,
        isCustom: false,
        isVerified: false,
        isPublicOrigin: true,
      });
    }
    return list;
  }, [domains, prevSavedPublicOrigin]);

  const onDeleteClick = useCallback((domainID: string, domain: string) => {
    setDeleteDomainDialogData({
      domainID,
      domain,
    });
    setConfirmDeleteDomainDialogVisible(true);
  }, []);

  const onSetPublicOriginClick = useCallback(
    (urlOrigin: string) => {
      setState((state) => ({
        ...state,
        publicOrigin: urlOrigin,
      }));
    },
    [setState]
  );

  const dismissDeleteDomainDialog = useCallback(() => {
    setConfirmDeleteDomainDialogVisible(false);
  }, []);

  const confirmUpdatePublicOrigin = useCallback(() => {
    // save app config form
    save();
  }, [save]);

  const dismissUpdatePublicOriginDialog = useCallback(() => {
    // reset app config form
    reset();
  }, [reset]);

  const renderDomainListColumn = useCallback<
    Required<IDetailsListProps>["onRenderItemColumn"]
  >(
    (item: DomainListItem, _index?: number, column?: IColumn) => {
      switch (column?.key) {
        case "domain":
          return <span>{item.domain}</span>;
        case "isVerified": {
          if (item.isPublicOrigin) {
            return (
              <span className={styles.activeStatus}>
                <FormattedMessage id="CustomDomainListScreen.domain-list.status.active" />
              </span>
            );
          }
          if (item.isVerified) {
            return (
              <span>
                <FormattedMessage id="CustomDomainListScreen.domain-list.status.verified" />
              </span>
            );
          }
          return (
            <span>
              <FormattedMessage id="CustomDomainListScreen.domain-list.status.not-verified" />
            </span>
          );
        }
        case "action":
          return (
            <DomainListActionButtons
              domainID={item.id}
              domain={item.domain}
              urlOrigin={item.urlOrigin}
              isVerified={item.isVerified}
              isCustomDomain={item.isCustom}
              isPublicOrigin={item.isPublicOrigin}
              onDeleteClick={onDeleteClick}
              onSetPublicOriginClick={onSetPublicOriginClick}
            />
          );
        default:
          return null;
      }
    },
    [onDeleteClick, onSetPublicOriginClick]
  );

  const renderDomainListHeader = useCallback<
    Required<IDetailsListProps>["onRenderDetailsHeader"]
  >((props, defaultRenderer) => {
    const defaultHeaderNode = defaultRenderer?.(props) ?? null;
    return (
      <>
        {defaultHeaderNode}
        <AddDomainSection />
      </>
    );
  }, []);

  return (
    <div className={styles.content}>
      <NavBreadcrumb items={navBreadcrumbItems} />
      <Text className={styles.description}>
        <FormattedMessage id="CustomDomainListScreen.desc" />
      </Text>
      <DetailsList
        columns={domainListColumns}
        items={domainListItems}
        selectionMode={SelectionMode.none}
        onRenderItemColumn={renderDomainListColumn}
        onRenderDetailsHeader={renderDomainListHeader}
      />

      <DeleteDomainDialog
        domain={deleteDomainDialogData.domain}
        domainID={deleteDomainDialogData.domainID}
        visible={deleteDomainDialogVisible}
        dismissDialog={dismissDeleteDomainDialog}
      />
      {/* UpdatePublicOriginDialog depends on app config form state */}
      <UpdatePublicOriginDialog
        urlOrigin={state.publicOrigin}
        visible={isDirty}
        isSaving={isUpdating}
        updateError={updateError}
        onConfirmClick={confirmUpdatePublicOrigin}
        dismissDialog={dismissUpdatePublicOriginDialog}
      />
    </div>
  );
};

const CustomDomainListScreen: React.FC = function CustomDomainListScreen() {
  const { appID } = useParams();
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();

  const {
    domains,
    loading: fetchingDomains,
    error: fetchDomainsError,
    refetch: refetchDomains,
  } = useDomainsQuery(appID);

  const form = useAppConfigForm(appID, constructFormState, constructConfig);

  const isVerifySuccessMessageVisible = useMemo(() => {
    const verify = searchParams.get("verify");
    return verify === "success";
  }, [searchParams]);

  const dismissVerifySuccessMessageBar = useCallback(() => {
    navigate(".", { replace: true });
  }, [navigate]);

  if (fetchingDomains || form.isLoading) {
    return <ShowLoading />;
  }

  if (fetchDomainsError) {
    return <ShowError error={fetchDomainsError} onRetry={refetchDomains} />;
  }

  if (form.loadError) {
    return <ShowError error={form.loadError} onRetry={form.reload} />;
  }

  return (
    <main>
      <MessageBar
        className={cn(
          styles.verifySuccessMessageBar,
          isVerifySuccessMessageVisible && styles.visible
        )}
        messageBarType={MessageBarType.success}
        onDismiss={dismissVerifySuccessMessageBar}
      >
        <FormattedMessage id="CustomDomainListScreen.verify-success-message" />
      </MessageBar>
      <CustomDomainListContent domains={domains ?? []} appConfigForm={form} />
    </main>
  );
};

export default CustomDomainListScreen;
