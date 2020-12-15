import React, { useCallback, useContext, useMemo, useState } from "react";
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

interface DomainListItem {
  id: string;
  domain: string;
  isVerified: boolean;
  isCustom: boolean;
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
  domainID: string;
  domain: string;
  isCustomDomain: boolean;
  isVerified: boolean;
  onDeleteClick: (domainID: string, domain: string) => void;
}

const DomainListActionButtons: React.FC<DomainListActionButtonsProps> = function DomainListActionButtons(
  props
) {
  const {
    domainID,
    domain,
    isCustomDomain,
    isVerified,
    onDeleteClick: onDeleteClickProps,
  } = props;

  const { themes } = useSystemConfig();
  const navigate = useNavigate();

  const showDelete = isCustomDomain;
  const showVerify = !isVerified;

  const onVerifyClicked = useCallback(() => {
    navigate(`./${domainID}/verify`);
  }, [domainID, navigate]);

  const onDeleteClick = useCallback(() => {
    onDeleteClickProps(domainID, domain);
  }, [domainID, domain, onDeleteClickProps]);

  if (!showDelete && !showVerify) {
    return (
      <section className={styles.actionButtonContainer}>
        <Text>---</Text>
      </section>
    );
  }

  return (
    <section className={styles.actionButtonContainer}>
      {showVerify && (
        <ActionButton
          className={styles.actionButton}
          theme={themes.actionButton}
          onClick={onVerifyClicked}
        >
          <FormattedMessage id="verify" />
        </ActionButton>
      )}
      {showVerify && showDelete && (
        <VerticalDivider className={styles.divider} />
      )}
      {showDelete && (
        <ActionButton
          className={styles.actionButton}
          theme={themes.destructive}
          onClick={onDeleteClick}
        >
          <FormattedMessage id="delete" />
        </ActionButton>
      )}
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

interface CustomDomainListContentProps {
  domains: Domain[];
}

const CustomDomainListContent: React.FC<CustomDomainListContentProps> = function CustomDomainListContent(
  props
) {
  const { domains } = props;

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

  const domainListItems: DomainListItem[] = useMemo(() => {
    return domains.map((domain) => ({
      id: domain.id,
      domain: domain.domain,
      isVerified: domain.isVerified,
      isCustom: domain.isCustom,
    }));
  }, [domains]);

  const onDeleteClick = useCallback((domainID: string, domain: string) => {
    setDeleteDomainDialogData({
      domainID,
      domain,
    });
    setConfirmDeleteDomainDialogVisible(true);
  }, []);

  const dismissDeleteDomainDialog = useCallback(() => {
    setConfirmDeleteDomainDialogVisible(false);
  }, []);

  const renderDomainListColumn = useCallback<
    Required<IDetailsListProps>["onRenderItemColumn"]
  >(
    (item: DomainListItem, _index?: number, column?: IColumn) => {
      switch (column?.key) {
        case "domain":
          return <span>{item.domain}</span>;
        case "isVerified": {
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
              isVerified={item.isVerified}
              isCustomDomain={item.isCustom}
              onDeleteClick={onDeleteClick}
            />
          );
        default:
          return null;
      }
    },
    [onDeleteClick]
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
    <div>
      <NavBreadcrumb className={styles.header} items={navBreadcrumbItems} />
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

  const isVerifySuccessMessageVisible = useMemo(() => {
    const verify = searchParams.get("verify");
    return verify === "success";
  }, [searchParams]);

  const dismissVerifySuccessMessageBar = useCallback(() => {
    navigate(".", { replace: true });
  }, [navigate]);

  if (fetchingDomains) {
    return <ShowLoading />;
  }

  if (fetchDomainsError) {
    return <ShowError error={fetchDomainsError} onRetry={refetchDomains} />;
  }

  return (
    <main className={styles.root}>
      {isVerifySuccessMessageVisible && (
        <MessageBar
          className={styles.verifySuccessMessageBar}
          messageBarType={MessageBarType.success}
          onDismiss={dismissVerifySuccessMessageBar}
        >
          <FormattedMessage id="CustomDomainListScreen.verify-success-message" />
        </MessageBar>
      )}
      <CustomDomainListContent domains={domains ?? []} />
    </main>
  );
};

export default CustomDomainListScreen;
