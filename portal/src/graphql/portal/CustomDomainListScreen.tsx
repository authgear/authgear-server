import React, {
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import cn from "classnames";
import { produce } from "immer";
import { useNavigate, useParams, useSearchParams } from "react-router-dom";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import {
  DetailsList,
  Dialog,
  DialogFooter,
  IColumn,
  IDetailsListProps,
  IDialogProps,
  MessageBar,
  MessageBarType,
  SelectionMode,
  Separator,
  Text,
  VerticalDivider,
} from "@fluentui/react";
import { Domain } from "./globalTypes.generated";
import { useDomainsQuery } from "./query/domainsQuery";
import { useCreateDomainMutation } from "./mutations/createDomainMutation";
import { useDeleteDomainMutation } from "./mutations/deleteDomainMutation";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import ButtonWithLoading from "../../ButtonWithLoading";
import DefaultButton from "../../DefaultButton";
import ActionButton from "../../ActionButton";
import { useTextField } from "../../hook/useInput";
import {
  ErrorParseRule,
  makeReasonErrorParseRule,
  parseAPIErrors,
  parseRawError,
} from "../../error/parse";
import ErrorDialog from "../../error/ErrorDialog";
import { useSystemConfig } from "../../context/SystemConfigContext";
import NavBreadcrumb, { BreadcrumbItem } from "../../NavBreadcrumb";

import styles from "./CustomDomainListScreen.module.css";
import { CustomDomainFeatureConfig, PortalAPIAppConfig } from "../../types";
import { clearEmptyObject } from "../../util/misc";
import {
  AppConfigFormModel,
  useAppConfigForm,
} from "../../hook/useAppConfigForm";
import { useAppFeatureConfigQuery } from "./query/appFeatureConfigQuery";
import ScreenContent from "../../ScreenContent";
import ErrorRenderer from "../../ErrorRenderer";
import ScreenLayoutScrollView from "../../ScreenLayoutScrollView";
import TextField from "../../TextField";
import FeatureDisabledMessageBar from "./FeatureDisabledMessageBar";
import WidgetTitle from "../../WidgetTitle";
import { useId } from "../../hook/useId";
import {
  FormContainerBase,
  useFormContainerBaseContext,
} from "../../FormContainerBase";
import { nullishCoalesce, or_ } from "../../util/operators";
import { FormErrorMessageBar } from "../../FormErrorMessageBar";
import PrimaryButton from "../../PrimaryButton";
import FormTextField from "../../FormTextField";

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
  cookieDomain: string;
  urlOrigin: string;
  isVerified: boolean;
  isCustom: boolean;
  isPublicOrigin: boolean;
}

interface FormState {
  publicOrigin: string;
  cookieDomain?: string;
}

function constructFormState(config: PortalAPIAppConfig): FormState {
  return {
    publicOrigin: config.http?.public_origin ?? "",
    cookieDomain: config.http?.cookie_domain,
  };
}

function constructConfig(
  config: PortalAPIAppConfig,
  _initialState: FormState,
  currentState: FormState
): PortalAPIAppConfig {
  return produce(config, (config) => {
    config.http ??= {};
    config.http.public_origin = currentState.publicOrigin;
    config.http.cookie_domain = currentState.cookieDomain;
    clearEmptyObject(config);
  });
}

interface RedirectURLFormState {
  postLoginURL: string;
  postLogoutURL: string;
}

function constructRedirectURLFormState(
  config: PortalAPIAppConfig
): RedirectURLFormState {
  return {
    postLoginURL: config.ui?.default_redirect_uri ?? "",
    postLogoutURL: config.ui?.default_post_logout_redirect_uri ?? "",
  };
}
function constructConfigFromRedirectURLFormState(
  config: PortalAPIAppConfig,
  _initialState: RedirectURLFormState,
  currentState: RedirectURLFormState
): PortalAPIAppConfig {
  return produce(config, (draft) => {
    draft.ui ??= {};
    draft.ui.default_redirect_uri = currentState.postLoginURL || undefined;
    draft.ui.default_post_logout_redirect_uri =
      currentState.postLogoutURL || undefined;
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

const AddDomainSection: React.VFC = function AddDomainSection() {
  const { renderToString } = useContext(Context);
  const { appID } = useParams() as { appID: string };

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
      makeReasonErrorParseRule(
        "DuplicatedDomain",
        "CustomDomainListScreen.add-domain.duplicated-error"
      ),
      makeReasonErrorParseRule(
        "InvalidDomain",
        "CustomDomainListScreen.add-domain.invalid-error"
      ),
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

  return (
    <form className={styles.addDomain} onSubmit={onAddClick}>
      <TextField
        className={styles.addDomainField}
        placeholder={renderToString(
          "CustomDomainListScreen.domain-list.add-domain.placeholder"
        )}
        value={newDomain}
        onChange={onNewDomainChange}
        errorMessage={
          errors.length > 0 ? <ErrorRenderer errors={errors} /> : undefined
        }
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
  cookieDomain: string;
  urlOrigin: string;
  isCustomDomain: boolean;
  isVerified: boolean;
  isPublicOrigin: boolean;
  onDeleteClick: (domainID: string, domain: string) => void;
  onDomainActivate: (urlOrigin: string, cookieDomain: string) => void;
}

const DomainListActionButtons: React.VFC<DomainListActionButtonsProps> =
  // eslint-disable-next-line complexity
  function DomainListActionButtons(props) {
    const {
      domainID,
      domain,
      cookieDomain,
      urlOrigin,
      isCustomDomain,
      isVerified,
      isPublicOrigin,
      onDeleteClick: onDeleteClickProps,
      onDomainActivate: onDomainActivateProps,
    } = props;

    const { themes } = useSystemConfig();
    const navigate = useNavigate();

    const showDelete = domainID && isCustomDomain && !isPublicOrigin;
    const showVerify = domainID && !isVerified;
    const showActivate = domainID && isVerified && !isPublicOrigin;

    const onActivateClick = useCallback(() => {
      onDomainActivateProps(urlOrigin, cookieDomain);
    }, [urlOrigin, cookieDomain, onDomainActivateProps]);

    const onVerifyClicked = useCallback(() => {
      navigate(`./${domainID}/verify`);
    }, [domainID, navigate]);

    const onDeleteClick = useCallback(() => {
      if (!domainID) {
        return;
      }
      onDeleteClickProps(domainID, domain);
    }, [domainID, domain, onDeleteClickProps]);

    if (!showActivate && !showDelete && !showVerify) {
      return (
        <section className={styles.actionButtonContainer}>
          <Text>---</Text>
        </section>
      );
    }

    const buttonNodes: React.ReactNode[] = [];
    if (showActivate) {
      buttonNodes.push(
        <ActionButton
          key={`${domainID}-domain-set-public-origin`}
          className={styles.actionButton}
          theme={themes.actionButton}
          onClick={onActivateClick}
          text={<FormattedMessage id="activate" />}
        />
      );
    }

    if (showVerify) {
      buttonNodes.push(
        <ActionButton
          key={`${domainID}-domain-verify`}
          className={styles.actionButton}
          theme={themes.actionButton}
          onClick={onVerifyClicked}
          text={<FormattedMessage id="verify" />}
        />
      );
    }

    if (showDelete) {
      buttonNodes.push(
        <ActionButton
          key={`${domainID}-domain-delete`}
          className={styles.actionButton}
          theme={themes.destructive}
          onClick={onDeleteClick}
          text={<FormattedMessage id="delete" />}
        />
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

const DeleteDomainDialog: React.VFC<DeleteDomainDialogProps> =
  function DeleteDomainDialog(props: DeleteDomainDialogProps) {
    const { domain, domainID, visible, dismissDialog } = props;
    const { appID } = useParams() as { appID: string };
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
      makeReasonErrorParseRule(
        "Forbidden",
        "CustomDomainListScreen.delete-domain-dialog.forbidden-error"
      ),
    ];

    const dialogContentProps: IDialogProps["dialogContentProps"] =
      useMemo(() => {
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
              text={<FormattedMessage id="cancel" />}
            />
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

const UpdatePublicOriginDialog: React.VFC<UpdatePublicOriginDialogProps> =
  function UpdatePublicOriginDialog(props: UpdatePublicOriginDialogProps) {
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

    const dialogContentProps: IDialogProps["dialogContentProps"] =
      useMemo(() => {
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
              text={<FormattedMessage id="cancel" />}
            />
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

interface RedirectURLTextFieldProps {
  className?: string;
  fieldName: string;
  label: NonNullable<React.ReactNode>;
  description: NonNullable<React.ReactNode>;
  value: string;
  onChangeValue: (value: string) => void;
}
const RedirectURLTextField: React.VFC<RedirectURLTextFieldProps> =
  function RedirectURLTextField(props) {
    const { fieldName, className, label, description, value, onChangeValue } =
      props;
    const id = useId();
    const onChange = useCallback(
      (_e: React.FormEvent<any>, value?: string) => {
        onChangeValue(value ?? "");
      },
      [onChangeValue]
    );
    return (
      <div className={className}>
        <label htmlFor={id}>{label}</label>
        <Text className={cn("mt-2.5")} block={true}>
          {description}
        </Text>
        <FormTextField
          id={id}
          fieldName={fieldName}
          parentJSONPointer="/ui"
          className={cn("mt-2.5")}
          value={value}
          onChange={onChange}
        />
      </div>
    );
  };

interface RedirectURLFormProps {
  className?: string;
  redirectURLForm: AppConfigFormModel<RedirectURLFormState>;
}
const RedirectURLForm: React.VFC<RedirectURLFormProps> =
  function RedirectURLForm(props) {
    const { className, redirectURLForm } = props;

    const { canSave, onSubmit } = useFormContainerBaseContext();

    const onChangePostLoginURL = useCallback(
      (url: string) => {
        redirectURLForm.setState((prev) =>
          produce(prev, (draft) => {
            draft.postLoginURL = url;
          })
        );
      },
      [redirectURLForm]
    );

    const onChangePostLogoutURL = useCallback(
      (url: string) => {
        redirectURLForm.setState((prev) =>
          produce(prev, (draft) => {
            draft.postLogoutURL = url;
          })
        );
      },
      [redirectURLForm]
    );

    return (
      <form className={className} onSubmit={onSubmit}>
        <WidgetTitle>
          <FormattedMessage id="CustomDomainListScreen.redirectURLSection.title" />
        </WidgetTitle>
        <RedirectURLTextField
          className={cn("mt-4")}
          fieldName="default_redirect_uri"
          label={
            <FormattedMessage id="CustomDomainListScreen.redirectURLSection.input.postLoginURL.label" />
          }
          description={
            <FormattedMessage id="CustomDomainListScreen.redirectURLSection.input.postLoginURL.description" />
          }
          value={redirectURLForm.state.postLoginURL}
          onChangeValue={onChangePostLoginURL}
        />
        <RedirectURLTextField
          className={cn("mt-4")}
          fieldName="default_post_logout_redirect_uri"
          label={
            <FormattedMessage id="CustomDomainListScreen.redirectURLSection.input.postLogoutURL.label" />
          }
          description={
            <FormattedMessage id="CustomDomainListScreen.redirectURLSection.input.postLogoutURL.description" />
          }
          value={redirectURLForm.state.postLogoutURL}
          onChangeValue={onChangePostLogoutURL}
        />
        <PrimaryButton
          className={cn("mt-12")}
          type="submit"
          disabled={!canSave}
          text={<FormattedMessage id="save" />}
        ></PrimaryButton>
      </form>
    );
  };

interface CustomDomainListContentProps {
  domains: Domain[];
  appConfigForm: AppConfigFormModel<FormState>;
  redirectURLForm: AppConfigFormModel<RedirectURLFormState>;
  featureConfig?: CustomDomainFeatureConfig;
}

const CustomDomainListContent: React.VFC<CustomDomainListContentProps> =
  function CustomDomainListContent(props) {
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
      featureConfig,
      redirectURLForm,
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
    const [deleteDomainDialogVisible, setConfirmDeleteDomainDialogVisible] =
      useState(false);
    const [deleteDomainDialogData, setDeleteDomainDialogData] =
      useState<DeleteDomainDialogData>({ domainID: "", domain: "" });

    const domainListColumns: IColumn[] = useMemo(() => {
      return makeDomainListColumn(renderToString);
    }, [renderToString]);

    const savedPublicOriginRef = useRef<string>(state.publicOrigin);
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
          cookieDomain: domain.cookieDomain,
          urlOrigin: urlOrigin,
          isVerified: domain.isVerified,
          isCustom: domain.isCustom,
          isPublicOrigin: isPublicOrigin,
        };
      });
      const found = list.find((domain) => domain.isPublicOrigin);

      if (!found) {
        // cannot found a domain that match the public origin
        // should only happen in local development
        list.unshift({
          domain:
            getHostFromOrigin(prevSavedPublicOrigin) || prevSavedPublicOrigin,
          cookieDomain: "",
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

    const onDomainActivate = useCallback(
      (urlOrigin: string, cookieDomain: string) => {
        // set cookieDomain to the domain's cookieDomain
        setState((state) => ({
          ...state,
          publicOrigin: urlOrigin,
          cookieDomain: cookieDomain === "" ? undefined : cookieDomain,
        }));
      },
      [setState]
    );

    const dismissDeleteDomainDialog = useCallback(() => {
      setConfirmDeleteDomainDialogVisible(false);
    }, []);

    const confirmUpdatePublicOrigin = useCallback(() => {
      // save app config form
      save().catch(() => {});
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
                cookieDomain={item.cookieDomain}
                urlOrigin={item.urlOrigin}
                isVerified={item.isVerified}
                isCustomDomain={item.isCustom}
                isPublicOrigin={item.isPublicOrigin}
                onDeleteClick={onDeleteClick}
                onDomainActivate={onDomainActivate}
              />
            );
          default:
            return null;
        }
      },
      [onDeleteClick, onDomainActivate]
    );

    const customDomainDisabled = useMemo(() => {
      return featureConfig?.disabled ?? false;
    }, [featureConfig]);

    const renderDomainListHeader = useCallback<
      Required<IDetailsListProps>["onRenderDetailsHeader"]
    >(
      (props, defaultRenderer) => {
        const defaultHeaderNode = defaultRenderer?.(props) ?? null;
        return (
          <>
            {defaultHeaderNode}
            {!customDomainDisabled ? <AddDomainSection /> : null}
          </>
        );
      },
      [customDomainDisabled]
    );

    return (
      <ScreenLayoutScrollView>
        <ScreenContent>
          <NavBreadcrumb className={styles.widget} items={navBreadcrumbItems} />
          <div className={cn(styles.widget)}>
            <Text block={true}>
              <FormattedMessage id="CustomDomainListScreen.desc" />
            </Text>
            {customDomainDisabled ? (
              <FeatureDisabledMessageBar messageID="FeatureConfig.custom-domain.disabled" />
            ) : null}
            <DetailsList
              columns={domainListColumns}
              items={domainListItems}
              selectionMode={SelectionMode.none}
              onRenderItemColumn={renderDomainListColumn}
              onRenderDetailsHeader={renderDomainListHeader}
            />
          </div>
          <Separator className={cn(styles.widget)} />
          <RedirectURLForm
            className={cn(styles.widget)}
            redirectURLForm={redirectURLForm}
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
        </ScreenContent>
      </ScreenLayoutScrollView>
    );
  };

const CustomDomainListScreen: React.VFC = function CustomDomainListScreen() {
  const { appID } = useParams() as { appID: string };
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();

  const {
    domains,
    loading: fetchingDomains,
    error: fetchDomainsError,
    refetch: refetchDomains,
  } = useDomainsQuery(appID);

  const form = useAppConfigForm({ appID, constructFormState, constructConfig });
  const redirectURLForm = useAppConfigForm({
    appID,
    constructFormState: constructRedirectURLFormState,
    constructConfig: constructConfigFromRedirectURLFormState,
  });

  const featureConfig = useAppFeatureConfigQuery(appID);

  const isVerifySuccessMessageVisible = useMemo(() => {
    const verify = searchParams.get("verify");
    return verify === "success";
  }, [searchParams]);

  const dismissVerifySuccessMessageBar = useCallback(() => {
    navigate(".", { replace: true });
  }, [navigate]);

  const isloading = or_(
    fetchingDomains,
    form.isLoading,
    featureConfig.loading,
    redirectURLForm.isLoading
  );

  const error = nullishCoalesce(
    fetchDomainsError,
    featureConfig.error,
    form.loadError,
    redirectURLForm.loadError
  );

  const retry = useCallback(() => {
    refetchDomains().catch((e) => console.error(e));
    featureConfig.refetch().catch((e) => console.error(e));
    form.reload();
    redirectURLForm.reload();
  }, [featureConfig, refetchDomains, form, redirectURLForm]);

  if (isloading) {
    return <ShowLoading />;
  }

  if (error) {
    return <ShowError error={error} onRetry={retry} />;
  }

  return (
    <>
      <FormContainerBase form={redirectURLForm}>
        {isVerifySuccessMessageVisible ? (
          <MessageBar
            messageBarType={MessageBarType.success}
            onDismiss={dismissVerifySuccessMessageBar}
          >
            <FormattedMessage id="CustomDomainListScreen.verify-success-message" />
          </MessageBar>
        ) : null}
        <FormErrorMessageBar></FormErrorMessageBar>
        <CustomDomainListContent
          domains={domains ?? []}
          appConfigForm={form}
          redirectURLForm={redirectURLForm}
          featureConfig={featureConfig.effectiveFeatureConfig?.custom_domain}
        />
      </FormContainerBase>
    </>
  );
};

export default CustomDomainListScreen;
