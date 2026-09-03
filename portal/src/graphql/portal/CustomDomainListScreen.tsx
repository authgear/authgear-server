import React, { useCallback, useContext, useMemo, useState } from "react";
import cn from "classnames";
import { produce } from "immer";
import { useNavigate, useParams, useSearchParams } from "react-router-dom";
import { Context, FormattedMessage } from "../../intl";
import {
  Button,
  Dialog,
  DropdownMenu,
  Flex,
  Heading,
  IconButton as RadixIconButton,
  Text,
} from "@radix-ui/themes";
import {
  CheckCircledIcon,
  DotsVerticalIcon,
  ExclamationTriangleIcon,
  PlusIcon,
  TrashIcon,
} from "@radix-ui/react-icons";
import { Domain } from "./globalTypes.generated";
import { useDomainsQuery } from "./query/domainsQuery";
import { useCreateDomainMutation } from "./mutations/createDomainMutation";
import { useDeleteDomainMutation } from "./mutations/deleteDomainMutation";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import {
  ErrorParseRule,
  makeReasonErrorParseRule,
  parseAPIErrors,
  parseRawError,
} from "../../error/parse";
import Link from "../../Link";

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
import { TextField as RadixTextField } from "../../components/v2/TextField/TextField";
import FeatureDisabledMessageBar from "./FeatureDisabledMessageBar";
import { FormContainerBase } from "../../FormContainerBase";
import { nullishCoalesce, or_ } from "../../util/operators";
import { getHostFromOrigin, getOriginFromDomain } from "../../util/domain";
import { FormErrorMessageBar } from "../../FormErrorMessageBar";
import { SettingsSectionCard } from "../../components/v2/SettingsSectionCard/SettingsSectionCard";
import { CardTable } from "../../components/v2/CardTable/CardTable";
import { Callout } from "../../components/v2/Callout/Callout";
import { PrimaryButton } from "../../components/v2/Button/PrimaryButton/PrimaryButton";
import { SecondaryButton } from "../../components/v2/Button/SecondaryButton/SecondaryButton";

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

interface AddDomainDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

const AddDomainDialog: React.VFC<AddDomainDialogProps> =
  function AddDomainDialog(props) {
    const { open, onOpenChange } = props;
    const { renderToString } = useContext(Context);
    const { appID } = useParams() as { appID: string };

    const [newDomain, setNewDomain] = useState("");

    const {
      createDomain,
      loading: creatingDomain,
      error: createDomainError,
    } = useCreateDomainMutation(appID);

    const onClose = useCallback(() => {
      if (creatingDomain) {
        return;
      }
      setNewDomain("");
      onOpenChange(false);
    }, [creatingDomain, onOpenChange]);

    const onAddClick = useCallback(() => {
      createDomain(newDomain)
        .then((success) => {
          if (success) {
            setNewDomain("");
            onOpenChange(false);
          }
        })
        .catch(() => {});
    }, [createDomain, newDomain, onOpenChange]);

    const isModified = newDomain.trim() !== "";

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
      <Dialog.Root
        open={open}
        onOpenChange={(nextOpen) => {
          if (!nextOpen) {
            onClose();
            return;
          }
          onOpenChange(nextOpen);
        }}
      >
        <Dialog.Content maxWidth="400px" size="3">
          <Dialog.Title>
            <FormattedMessage id="CustomDomainListScreen.add-domain.dialog.title" />
          </Dialog.Title>
          <form
            className={styles.addDomainForm}
            onSubmit={(ev) => {
              ev.preventDefault();
              if (!isModified || creatingDomain) {
                return;
              }
              onAddClick();
            }}
          >
            <RadixTextField
              size="2"
              label={
                <FormattedMessage id="CustomDomainListScreen.add-domain.dialog.domain.label" />
              }
              placeholder={renderToString(
                "CustomDomainListScreen.domain-list.add-domain.placeholder"
              )}
              value={newDomain}
              onChange={(e) => {
                setNewDomain(e.currentTarget.value);
              }}
              error={
                errors.length > 0 ? (
                  <ErrorRenderer errors={errors} />
                ) : undefined
              }
            />
            <Flex gap="3" mt="4" justify="end">
              <SecondaryButton
                size="2"
                text={<FormattedMessage id="cancel" />}
                onClick={onClose}
                disabled={creatingDomain}
              />
              <PrimaryButton
                type="submit"
                size="2"
                disabled={!isModified}
                loading={creatingDomain}
                text={<FormattedMessage id="add" />}
              />
            </Flex>
          </form>
        </Dialog.Content>
      </Dialog.Root>
    );
  };

interface DomainRowProps {
  item: DomainListItem;
  onDeleteClick: (domainID: string, domain: string) => void;
  onDomainActivate: (urlOrigin: string, cookieDomain: string) => void;
}

const DomainRow: React.VFC<DomainRowProps> = function DomainRow(props) {
  const { item, onDeleteClick, onDomainActivate } = props;
  const { renderToString } = useContext(Context);
  const navigate = useNavigate();

  const showDelete = Boolean(item.id && item.isCustom && !item.isPublicOrigin);
  const showVerify = Boolean(item.id && !item.isVerified);
  const showActivate = Boolean(
    item.id && item.isVerified && !item.isPublicOrigin
  );
  const hasActions = showActivate || showDelete || showVerify;

  const statusNode = (() => {
    if (item.isPublicOrigin) {
      return (
        <span className={styles.statusVerified}>
          <CheckCircledIcon width="1rem" height="1rem" />
          <FormattedMessage id="CustomDomainListScreen.domain-list.status.active" />
        </span>
      );
    }
    if (item.isVerified) {
      return (
        <span className={styles.statusVerified}>
          <CheckCircledIcon width="1rem" height="1rem" />
          <FormattedMessage id="CustomDomainListScreen.domain-list.status.verified" />
        </span>
      );
    }
    return (
      <span className={styles.statusPending}>
        <ExclamationTriangleIcon width="1rem" height="1rem" />
        <FormattedMessage id="CustomDomainListScreen.domain-list.status.not-verified" />
      </span>
    );
  })();

  return (
    <CardTable.Row>
      <CardTable.Cell className={styles.colDomain}>
        <Text size="2" className={styles.domainName}>
          {item.domain}
        </Text>
      </CardTable.Cell>
      <CardTable.Cell className={styles.colStatus}>{statusNode}</CardTable.Cell>
      <CardTable.Cell className={styles.colActions}>
        {hasActions ? (
          <DropdownMenu.Root>
            <DropdownMenu.Trigger>
              <RadixIconButton
                className={styles.rowActionsButton}
                variant="soft"
                color="gray"
                size="2"
                aria-label={renderToString(
                  "CustomDomainListScreen.domain-list.row-actions"
                )}
              >
                <DotsVerticalIcon width="1rem" height="1rem" />
              </RadixIconButton>
            </DropdownMenu.Trigger>
            <DropdownMenu.Content align="end">
              {showActivate ? (
                <DropdownMenu.Item
                  onSelect={() => {
                    onDomainActivate(item.urlOrigin, item.cookieDomain);
                  }}
                >
                  <FormattedMessage id="activate" />
                </DropdownMenu.Item>
              ) : null}
              {showVerify ? (
                <DropdownMenu.Item
                  onSelect={() => {
                    if (item.id) {
                      navigate(`./${item.id}/verify`);
                    }
                  }}
                >
                  <FormattedMessage id="verify" />
                </DropdownMenu.Item>
              ) : null}
              {showDelete && (showActivate || showVerify) ? (
                <DropdownMenu.Separator />
              ) : null}
              {showDelete ? (
                <DropdownMenu.Item
                  color="red"
                  onSelect={() => {
                    if (item.id) {
                      onDeleteClick(item.id, item.domain);
                    }
                  }}
                >
                  <TrashIcon />
                  <FormattedMessage id="delete" />
                </DropdownMenu.Item>
              ) : null}
            </DropdownMenu.Content>
          </DropdownMenu.Root>
        ) : null}
      </CardTable.Cell>
    </CardTable.Row>
  );
};

interface DomainActionErrorDialogProps {
  error: unknown;
  rules?: ErrorParseRule[];
  fallbackErrorMessageID: string;
}

const DomainActionErrorDialog: React.VFC<DomainActionErrorDialogProps> =
  function DomainActionErrorDialog(props) {
    const { error, rules = [], fallbackErrorMessageID } = props;
    const [open, setOpen] = useState(false);

    const [prevError, setPrevError] = useState<unknown>(null);
    if (error !== prevError) {
      setPrevError(error);
      if (error != null) {
        setOpen(true);
      }
    }

    const errors = useMemo(() => {
      const apiErrors = parseRawError(error);
      const { topErrors } = parseAPIErrors(
        apiErrors,
        [],
        rules,
        fallbackErrorMessageID
      );
      return topErrors;
    }, [error, rules, fallbackErrorMessageID]);

    return (
      <Dialog.Root open={open} onOpenChange={setOpen}>
        <Dialog.Content maxWidth="400px" size="3">
          <Dialog.Title>
            <FormattedMessage id="error" />
          </Dialog.Title>
          <Dialog.Description size="2">
            <ErrorRenderer errors={errors} />
          </Dialog.Description>
          <Flex gap="3" mt="4" justify="end">
            <PrimaryButton
              size="2"
              onClick={() => {
                setOpen(false);
              }}
              text={<FormattedMessage id="ok" />}
            />
          </Flex>
        </Dialog.Content>
      </Dialog.Root>
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

    const errorRules: ErrorParseRule[] = useMemo(() => {
      return [
        makeReasonErrorParseRule(
          "Forbidden",
          "CustomDomainListScreen.delete-domain-dialog.forbidden-error"
        ),
      ];
    }, []);

    return (
      <>
        <Dialog.Root
          open={visible}
          onOpenChange={(open) => {
            if (!open && !deletingDomain) {
              dismissDialog();
            }
          }}
        >
          <Dialog.Content maxWidth="400px" size="3">
            <Dialog.Title>
              <FormattedMessage id="CustomDomainListScreen.delete-domain-dialog.title" />
            </Dialog.Title>
            <Dialog.Description size="2">
              <FormattedMessage
                id="CustomDomainListScreen.delete-domain-dialog.message"
                values={{ domain }}
              />
            </Dialog.Description>
            <Flex gap="3" mt="4" justify="end">
              <SecondaryButton
                size="2"
                onClick={dismissDialog}
                disabled={deletingDomain}
                text={<FormattedMessage id="cancel" />}
              />
              <Button
                size="2"
                variant="solid"
                color="red"
                loading={deletingDomain}
                disabled={!visible}
                onClick={onConfirmClick}
              >
                <FormattedMessage id="confirm" />
              </Button>
            </Flex>
          </Dialog.Content>
        </Dialog.Root>
        <DomainActionErrorDialog
          error={deleteDomainError}
          rules={errorRules}
          fallbackErrorMessageID="CustomDomainListScreen.delete-domain-dialog.generic-error"
        />
      </>
    );
  };

interface UpdatePublicOriginDialogData {
  urlOrigin: string;
}
interface UpdatePublicOriginDialogProps extends UpdatePublicOriginDialogData {
  visible: boolean;
  isSaving: boolean;
  updateError: unknown;
  onConfirmClick: () => void;
  dismissDialog: () => void;
}

const UpdatePublicOriginDialog: React.VFC<UpdatePublicOriginDialogProps> =
  function UpdatePublicOriginDialog(props: UpdatePublicOriginDialogProps) {
    const {
      visible,
      isSaving,
      updateError,
      urlOrigin,
      onConfirmClick: onConfirmClickProps,
      dismissDialog,
    } = props;

    const onConfirmClick = useCallback(() => {
      onConfirmClickProps();
    }, [onConfirmClickProps]);

    return (
      <>
        <Dialog.Root
          open={visible}
          onOpenChange={(open) => {
            if (!open && !isSaving) {
              dismissDialog();
            }
          }}
        >
          <Dialog.Content maxWidth="400px" size="3">
            <Dialog.Title>
              <FormattedMessage id="CustomDomainListScreen.activate-domain-dialog.title" />
            </Dialog.Title>
            <Dialog.Description size="2">
              <FormattedMessage
                id="CustomDomainListScreen.activate-domain-dialog.message"
                values={{ domain: urlOrigin }}
              />
            </Dialog.Description>
            <Flex gap="3" mt="4" justify="end">
              <SecondaryButton
                size="2"
                onClick={dismissDialog}
                disabled={isSaving}
                text={<FormattedMessage id="cancel" />}
              />
              <PrimaryButton
                size="2"
                loading={isSaving}
                disabled={!visible}
                onClick={onConfirmClick}
                text={<FormattedMessage id="confirm" />}
              />
            </Flex>
          </Dialog.Content>
        </Dialog.Root>
        <DomainActionErrorDialog
          error={updateError}
          fallbackErrorMessageID="CustomDomainListScreen.activate-domain-dialog.generic-error"
        />
      </>
    );
  };

interface CustomDomainListContentProps {
  domains: Domain[];
  appConfigForm: AppConfigFormModel<FormState>;
  featureConfig?: CustomDomainFeatureConfig;
}

const CustomDomainListContent: React.VFC<CustomDomainListContentProps> =
  function CustomDomainListContent(props) {
    const {
      domains,
      appConfigForm: {
        state,
        setState,
        getIsDirty,
        isUpdating,
        save,
        reset,
        updateError,
      },
      featureConfig,
    } = props;

    const { appID } = useParams() as { appID: string };
    const [isAddingDomain, setIsAddingDomain] = useState(false);

    interface DeleteDomainDialogData {
      domainID: string;
      domain: string;
    }
    const [deleteDomainDialogVisible, setConfirmDeleteDomainDialogVisible] =
      useState(false);
    const [deleteDomainDialogData, setDeleteDomainDialogData] =
      useState<DeleteDomainDialogData>({ domainID: "", domain: "" });

    const isDirty = useMemo(() => getIsDirty(), [getIsDirty]);

    const [prevSavedPublicOrigin, setPrevSavedPublicOrigin] = useState<string>(
      state.publicOrigin
    );
    if (!isDirty && prevSavedPublicOrigin !== state.publicOrigin) {
      setPrevSavedPublicOrigin(state.publicOrigin);
    }

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

    const customDomainDisabled = useMemo(() => {
      return featureConfig?.disabled ?? false;
    }, [featureConfig]);

    const onOpenAddDomain = useCallback(() => {
      setIsAddingDomain(true);
    }, []);

    return (
      <ScreenLayoutScrollView>
        <ScreenContent layout="list">
          <div className={cn(styles.widget, styles.pageHeader)}>
            <Heading
              as="h1"
              size="5"
              weight="bold"
              className={styles.pageTitle}
            >
              <FormattedMessage id="CustomDomainListScreen.title" />
            </Heading>
            <Text
              as="p"
              size="2"
              color="gray"
              className={styles.pageDescription}
            >
              <FormattedMessage id="CustomDomainListScreen.desc" />
            </Text>
          </div>

          {customDomainDisabled ? (
            <div className={styles.widget}>
              <FeatureDisabledMessageBar messageID="FeatureConfig.custom-domain.disabled" />
            </div>
          ) : null}

          <SettingsSectionCard
            className={styles.widget}
            contentClassName={styles.domainsCardContent}
            title={
              <span className={styles.domainsCardTitleBlock}>
                <FormattedMessage id="CustomDomainListScreen.domains-section.title" />
                <Text
                  as="span"
                  size="1"
                  color="gray"
                  className={styles.domainsCardTitleDescription}
                >
                  <FormattedMessage id="CustomDomainListScreen.domains-section.description" />
                </Text>
              </span>
            }
          >
            <CardTable>
              <CardTable.Header>
                <CardTable.HeaderCell className={styles.colDomain}>
                  <FormattedMessage id="CustomDomainListScreen.domain-list.header.domain" />
                </CardTable.HeaderCell>
                <CardTable.HeaderCell className={styles.colStatus}>
                  <FormattedMessage id="CustomDomainListScreen.domain-list.header.status" />
                </CardTable.HeaderCell>
                <CardTable.HeaderCell
                  className={styles.colActions}
                  aria-hidden={true}
                />
              </CardTable.Header>
              {domainListItems.map((item) => (
                <DomainRow
                  key={item.id ?? item.domain}
                  item={item}
                  onDeleteClick={onDeleteClick}
                  onDomainActivate={onDomainActivate}
                />
              ))}
            </CardTable>

            {!customDomainDisabled ? (
              <button
                type="button"
                className={styles.addDomainButton}
                onClick={onOpenAddDomain}
              >
                <PlusIcon width="1rem" height="1rem" />
                <FormattedMessage id="CustomDomainListScreen.add-domain" />
              </button>
            ) : null}

            <Callout
              type="info"
              showCloseButton={false}
              className={styles.endpointCallout}
              text={
                <FormattedMessage
                  id="CustomDomainListScreen.rediect-endpoint-direct-access.message"
                  values={{
                    // eslint-disable-next-line react/no-unstable-nested-components
                    reactRouterLink: (chunks: React.ReactNode) => (
                      <Link
                        className={styles.endpointCalloutLink}
                        to={`/project/${appID}/advanced/endpoint-direct-access`}
                      >
                        {chunks}
                      </Link>
                    ),
                  }}
                />
              }
            />
          </SettingsSectionCard>

          {isAddingDomain ? (
            <AddDomainDialog
              open={isAddingDomain}
              onOpenChange={setIsAddingDomain}
            />
          ) : null}
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
    featureConfig.isLoading
  );

  const error = nullishCoalesce(
    fetchDomainsError,
    featureConfig.loadError,
    form.loadError
  );

  const retry = useCallback(() => {
    refetchDomains().catch((e) => console.error(e));
    featureConfig.refetch().catch((e) => console.error(e));
    form.reload();
  }, [featureConfig, refetchDomains, form]);

  if (isloading) {
    return <ShowLoading />;
  }

  if (error) {
    return <ShowError error={error} onRetry={retry} />;
  }

  return (
    <>
      <FormContainerBase form={form}>
        {isVerifySuccessMessageVisible ? (
          <Callout
            type="success"
            onClose={dismissVerifySuccessMessageBar}
            text={
              <FormattedMessage id="CustomDomainListScreen.verify-success-message" />
            }
          />
        ) : null}
        <FormErrorMessageBar></FormErrorMessageBar>
        <CustomDomainListContent
          domains={domains ?? []}
          appConfigForm={form}
          featureConfig={featureConfig.effectiveFeatureConfig?.custom_domain}
        />
      </FormContainerBase>
    </>
  );
};

export default CustomDomainListScreen;
