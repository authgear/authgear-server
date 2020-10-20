import React, {
  useMemo,
  useContext,
  useCallback,
  useState,
  useEffect,
} from "react";
import { useParams, useNavigate, useSearchParams } from "react-router-dom";
import cn from "classnames";
import produce from "immer";
import { FormattedMessage, Context } from "@oursky/react-messageformat";
import {
  Text,
  DetailsList,
  IColumn,
  Stack,
  SelectionMode,
  IDetailsListProps,
  ActionButton,
  VerticalDivider,
  TextField,
  Dialog,
  IDialogProps,
  DialogFooter,
  DefaultButton,
  MessageBar,
  MessageBarType,
  Dropdown,
} from "@fluentui/react";

import { useAppConfigQuery } from "./query/appConfigQuery";
import { Domain, useDomainsQuery } from "./query/domainsQuery";
import { useUpdateAppConfigMutation } from "./mutations/updateAppConfigMutation";
import { useCreateDomainMutation } from "./mutations/createDomainMutation";
import { useDeleteDomainMutation } from "./mutations/deleteDomainMutation";
import { PortalAPIApp, PortalAPIAppConfig } from "../../types";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import ButtonWithLoading from "../../ButtonWithLoading";
import {
  ModifiedIndicatorPortal,
  ModifiedIndicatorWrapper,
} from "../../ModifiedIndicatorPortal";
import { actionButtonTheme, destructiveTheme } from "../../theme";
import { useDropdown, useTextField } from "../../hook/useInput";
import {
  GenericErrorHandlingRule,
  useGenericError,
} from "../../error/useGenericError";
import ErrorDialog from "../../error/ErrorDialog";
import { clearEmptyObject } from "../../util/misc";

import styles from "./DNSConfigurationScreen.module.scss";

interface DNSConfigurationProps {
  domains: Domain[];
  rawAppConfig: PortalAPIAppConfig | null;
  effectiveAppConfig: PortalAPIAppConfig | null;
}

interface PublicOriginConfigurationProps {
  rawAppConfig: PortalAPIAppConfig | null;
  effectiveAppConfig: PortalAPIAppConfig | null;
  verifiedDomains: Domain[];
}

interface DomainListItem {
  id: string;
  domain: string;
  isVerified: boolean;
  isCustom: boolean;
}

interface DomainListActionButtonsProps {
  domainID: string;
  domain: string;
  isCustomDomain: boolean;
  isVerified: boolean;
  onDeleteClick: (domainID: string, domain: string) => void;
}

interface DeleteDomainDialogData {
  domainID: string;
  domain: string;
}

interface DeleteDomainDialogProps extends Partial<DeleteDomainDialogData> {
  visible: boolean;
  dismissDialog: () => void;
}

function makeDomainListColumn(renderToString: (messageID: string) => string) {
  return [
    {
      key: "domain",
      name: renderToString("DNSConfigurationScreen.domain-list.header.domain"),
      minWidth: 250,
      className: styles.domainListColumn,
    },
    {
      key: "isVerified",
      name: renderToString("DNSConfigurationScreen.domain-list.header.status"),
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

function getPublicOriginFromDomain(domain: Domain): string {
  // assume domain has no scheme
  // use https scheme
  return `https://${domain.domain}`;
}

function savePublicOrigin(
  publicOrigin: string | undefined,
  rawAppConfig: PortalAPIAppConfig | null,
  updateAppConfig: (
    appConfig: PortalAPIAppConfig
  ) => Promise<PortalAPIApp | null>
): void {
  if (rawAppConfig == null) {
    return;
  }

  const newPublicOrigin =
    publicOrigin?.trim() !== "" ? publicOrigin : undefined;

  if (newPublicOrigin == null) {
    // required field, cannot save if field missing
    return;
  }

  const newAppConfig = produce(rawAppConfig, (draftConfig) => {
    draftConfig.http = draftConfig.http ?? {};
    draftConfig.http.public_origin = newPublicOrigin;

    clearEmptyObject(draftConfig);
  });

  updateAppConfig(newAppConfig).catch(() => {});
}

function getVerifiedDomains(domains: Domain[]): Domain[] {
  return domains.filter((domain) => domain.isVerified);
}

const PublicOriginConfiguration: React.FC<PublicOriginConfigurationProps> = function PublicOriginConfiguration(
  props: PublicOriginConfigurationProps
) {
  const { rawAppConfig, effectiveAppConfig, verifiedDomains } = props;
  const { appID } = useParams();

  const {
    updateAppConfig,
    loading: updatingAppConfig,
    error: updateAppConfigError,
  } = useUpdateAppConfigMutation(appID);

  const initialPublicOrigin = useMemo(() => {
    return effectiveAppConfig?.http?.public_origin ?? "";
  }, [effectiveAppConfig]);

  const [publicOrigin, setPublicOrigin] = useState(initialPublicOrigin);

  const isModified = useMemo(() => {
    return initialPublicOrigin !== publicOrigin;
  }, [initialPublicOrigin, publicOrigin]);

  const resetForm = useCallback(() => {
    setPublicOrigin(initialPublicOrigin);
  }, [initialPublicOrigin]);

  const publicOriginOptionKeys = useMemo(() => {
    const keys = verifiedDomains.map(getPublicOriginFromDomain);
    if (initialPublicOrigin !== "" && !keys.includes(initialPublicOrigin)) {
      keys.unshift(initialPublicOrigin);
    }
    return keys;
  }, [verifiedDomains, initialPublicOrigin]);

  const {
    options: publicOriginOptions,
    onChange: onPublicOriginChange,
  } = useDropdown(
    publicOriginOptionKeys,
    (option) => {
      setPublicOrigin(option);
    },
    publicOrigin
  );

  const onSaveClick = useCallback(() => {
    savePublicOrigin(publicOrigin, rawAppConfig, updateAppConfig);
  }, [publicOrigin, rawAppConfig, updateAppConfig]);

  const resetPublicOrigin = useCallback(() => {
    setPublicOrigin(initialPublicOrigin);
  }, [initialPublicOrigin]);

  // if selected public origin is deleted
  // reset to public origin in config
  useEffect(() => {
    if (!publicOriginOptionKeys.includes(publicOrigin)) {
      resetPublicOrigin();
    }
  }, [publicOrigin, publicOriginOptionKeys, resetPublicOrigin]);

  return (
    <section className={styles.publicOrigin}>
      <ModifiedIndicatorPortal resetForm={resetForm} isModified={isModified} />
      {updateAppConfigError && <ShowError error={updateAppConfigError} />}
      <Text
        as="h2"
        className={cn(
          styles.header,
          styles.subHeader,
          styles.publicOriginHeader
        )}
      >
        <FormattedMessage id="DNSConfigurationScreen.public-origin.header" />
      </Text>
      <Text className={styles.publicOriginDesc}>
        <FormattedMessage id="DNSConfigurationScreen.public-origin.desc" />
      </Text>
      <div className={styles.publicOriginInput}>
        <Dropdown
          className={styles.publicOriginField}
          options={publicOriginOptions}
          selectedKey={publicOrigin}
          onChange={onPublicOriginChange}
        />
        <ButtonWithLoading
          className={styles.savePublicOriginButton}
          disabled={!isModified}
          labelId="save"
          loadingLabelId="saving"
          loading={updatingAppConfig}
          onClick={onSaveClick}
        />
      </div>
    </section>
  );
};

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

  const errorRules: GenericErrorHandlingRule[] = useMemo(() => {
    return [
      {
        errorMessageID: "DNSConfigurationScreen.add-domain.duplicated-error",
        reason: "DuplicatedDomain",
      },
      {
        errorMessageID: "DNSConfigurationScreen.add-domain.invalid-error",
        reason: "InvalidDomain",
      },
    ];
  }, []);

  const { errorMessage: addDomainErrorMessage } = useGenericError(
    createDomainError,
    errorRules,
    "DNSConfigurationScreen.add-domain.generic-error"
  );

  return (
    <form className={styles.addDomain} onSubmit={onAddClick}>
      <TextField
        className={styles.addDomainField}
        placeholder={renderToString(
          "DNSConfigurationScreen.domain-list.add-domain.placeholder"
        )}
        value={newDomain}
        onChange={onNewDomainChange}
        errorMessage={addDomainErrorMessage}
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

const DomainListActionButtons: React.FC<DomainListActionButtonsProps> = function DomainListActionButtons(
  props: DomainListActionButtonsProps
) {
  const {
    domainID,
    domain,
    isCustomDomain,
    isVerified,
    onDeleteClick: onDeleteClickProps,
  } = props;

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
          theme={actionButtonTheme}
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
          theme={destructiveTheme}
          onClick={onDeleteClick}
        >
          <FormattedMessage id="delete" />
        </ActionButton>
      )}
    </section>
  );
};

const DeleteDomainDialog: React.FC<DeleteDomainDialogProps> = function DeleteDomainDialog(
  props: DeleteDomainDialogProps
) {
  const { domain, domainID, visible, dismissDialog } = props;
  const { appID } = useParams();
  const { renderToString } = useContext(Context);

  const {
    deleteDomain,
    loading: deletingDomain,
    error: deleteDomainError,
  } = useDeleteDomainMutation(appID);

  const onConfirmClick = useCallback(() => {
    deleteDomain(domainID!)
      .catch(() => {})
      .finally(() => {
        dismissDialog();
      });
  }, [domainID, deleteDomain, dismissDialog]);

  const errorRules: GenericErrorHandlingRule[] = [
    {
      reason: "Forbidden",
      errorMessageID:
        "DNSConfigurationScreen.delete-domain-dialog.forbidden-error",
    },
  ];

  const dialogContentProps: IDialogProps["dialogContentProps"] = useMemo(() => {
    return {
      title: (
        <FormattedMessage id="DNSConfigurationScreen.delete-domain-dialog.title" />
      ),
      subText: renderToString(
        "DNSConfigurationScreen.delete-domain-dialog.message",
        { domain: domain ?? "" }
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
            theme={destructiveTheme}
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
        fallbackErrorMessageID="DNSConfigurationScreen.delete-domain-dialog.generic-error"
      />
    </>
  );
};

const DNSConfiguration: React.FC<DNSConfigurationProps> = function DNSConfiguration(
  props: DNSConfigurationProps
) {
  const { domains, rawAppConfig, effectiveAppConfig } = props;

  const { renderToString } = useContext(Context);

  const [
    deleteDomainDialogVisible,
    setConfirmDeleteDomainDialogVisible,
  ] = useState(false);
  const [
    deleteDomainDialogData,
    setDeleteDomainDialogData,
  ] = useState<DeleteDomainDialogData | null>(null);

  const verifiedDomains = useMemo(() => {
    return getVerifiedDomains(domains);
  }, [domains]);

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
                <FormattedMessage id="DNSConfigurationScreen.domain-list.status.verified" />
              </span>
            );
          }
          return (
            <span>
              <FormattedMessage id="DNSConfigurationScreen.domain-list.status.not-verified" />
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
    <section className={styles.content}>
      <DeleteDomainDialog
        domain={deleteDomainDialogData?.domain}
        domainID={deleteDomainDialogData?.domainID}
        visible={deleteDomainDialogVisible}
        dismissDialog={dismissDeleteDomainDialog}
      />
      <PublicOriginConfiguration
        rawAppConfig={rawAppConfig}
        effectiveAppConfig={effectiveAppConfig}
        verifiedDomains={verifiedDomains}
      />
      <DetailsList
        columns={domainListColumns}
        items={domainListItems}
        selectionMode={SelectionMode.none}
        onRenderItemColumn={renderDomainListColumn}
        onRenderDetailsHeader={renderDomainListHeader}
      />
    </section>
  );
};

const DNSConfigurationScreen: React.FC = function DNSConfigurationScreen() {
  const { appID } = useParams();
  const [searchParams] = useSearchParams();

  const {
    effectiveAppConfig,
    rawAppConfig,
    loading: fetchingAppConfig,
    error: fetchAppConfigError,
    refetch: refetchAppConfig,
  } = useAppConfigQuery(appID);
  const {
    domains,
    loading: fetchingDomains,
    error: fetchDomainsError,
    refetch: refetchDomains,
  } = useDomainsQuery(appID);

  const initialVerifySuccessMessageBarVisible = useMemo(() => {
    const verify = searchParams.get("verify");
    return verify === "success";
  }, [searchParams]);

  const [
    verifySuccessMessageBarVisible,
    setVerifySuccessMessageBarVisible,
  ] = useState(initialVerifySuccessMessageBarVisible);

  const dismissVerifySuccessMessageBar = useCallback(() => {
    setVerifySuccessMessageBarVisible(false);
  }, []);

  if (fetchingAppConfig || fetchingDomains) {
    return <ShowLoading />;
  }

  if (fetchAppConfigError != null || fetchDomainsError != null) {
    return (
      <Stack>
        {fetchAppConfigError && (
          <ShowError error={fetchAppConfigError} onRetry={refetchAppConfig} />
        )}
        {fetchDomainsError && (
          <ShowError error={fetchDomainsError} onRetry={refetchDomains} />
        )}
      </Stack>
    );
  }
  return (
    <main className={styles.root}>
      {verifySuccessMessageBarVisible && (
        <MessageBar
          className={styles.verifySuccessMessageBar}
          messageBarType={MessageBarType.success}
          onDismiss={dismissVerifySuccessMessageBar}
        >
          <FormattedMessage id="DNSConfigurationScreen.verify-success-message" />
        </MessageBar>
      )}
      <ModifiedIndicatorWrapper className={styles.screen}>
        <Text className={cn(styles.header, styles.mainHeader)} as="h1">
          <FormattedMessage id="DNSConfigurationScreen.title" />
        </Text>
        <Text className={styles.desc}>
          <FormattedMessage id="DNSConfigurationScreen.desc" />
        </Text>
        <DNSConfiguration
          effectiveAppConfig={effectiveAppConfig}
          rawAppConfig={rawAppConfig}
          domains={domains ?? []}
        />
      </ModifiedIndicatorWrapper>
    </main>
  );
};

export default DNSConfigurationScreen;
