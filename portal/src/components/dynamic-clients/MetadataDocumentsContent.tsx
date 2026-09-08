import React, {
  useCallback,
  useContext,
  useMemo,
  useRef,
  useState,
} from "react";
import cn from "classnames";
import { Flex, RadioGroup, Text } from "@radix-ui/themes";
import { QuestionMarkCircledIcon } from "@radix-ui/react-icons";
import { produce } from "immer";
import { Context, FormattedMessage } from "../../intl";
import ExternalLink from "../../ExternalLink";
import { useFormContainerBaseContext } from "../../FormContainerBase";
import { PortalAPIAppConfig } from "../../types";
import { clearEmptyObject } from "../../util/misc";
import { parseIntegerAllowLeadingZeros } from "../../util/input";
import { AppConfigFormModel } from "../../hook/useAppConfigForm";
import { SettingsSectionCard } from "../v2/SettingsSectionCard/SettingsSectionCard";
import { Toggle } from "../v2/Toggle/Toggle";
import { TextField } from "../v2/TextField/TextField";
import { TextFieldList } from "../v2/TextFieldList/TextFieldList";
import { ConfirmationDialog } from "../v2/ConfirmationDialog/ConfirmationDialog";
import { SaveFunctionBar } from "../v2/SaveFunctionBar/SaveFunctionBar";
import { useCalloutToast } from "../v2/Callout/Callout";
import { Tooltip } from "../v2/Tooltip/Tooltip";
import styles from "./MetadataDocumentsContent.module.css";

// JSON pointers of the objects holding the fields, so the config schema's
// validation errors -- "minimum" on a lifetime, "pattern" on a domain --
// bind to the field that caused them instead of only reaching the generic
// error bar. TextFieldList appends the item index itself.
// A save the admin triggered themselves needs only a glance to confirm.
const SAVED_TOAST_DURATION_MS = 2000;

const CIMD_JSON_POINTER = "/oauth/client_id_metadata_document";
const CIMD_CLIENT_CONFIG_JSON_POINTER =
  "/oauth/client_id_metadata_document/client_config";

// Whether the CIMD allowlist restricts anything. An absent or empty
// allowed_domains means "any domain", not "no domain" (docs/specs/cimd.md
// § Domain Trust) -- the form makes that an explicit choice so an admin
// never has to infer the permissive reading from an empty field.
export type CIMDDomainMode = "any" | "list";

// Only this mechanism's fields live here. The tab that renders this owns the
// form, so its save bar commits CIMD and nothing else -- not the sibling DCR
// tab's settings, and not the client list.
export interface FormState {
  cimdEnabled: boolean;
  cimdDomainMode: CIMDDomainMode;
  cimdAllowedDomains: string[];
  cimdAccessTokenLifetimeSeconds: number | undefined;
  cimdRefreshTokenLifetimeSeconds: number | undefined;
  cimdRefreshTokenIdleTimeoutEnabled: boolean;
  cimdRefreshTokenIdleTimeoutSeconds: number | undefined;
}

export function constructFormState(config: PortalAPIAppConfig): FormState {
  const cimd = config.oauth?.client_id_metadata_document;
  const cimdAllowedDomains = [...(cimd?.allowed_domains ?? [])];
  return {
    cimdEnabled: cimd?.enabled ?? false,
    cimdDomainMode: cimdAllowedDomains.length > 0 ? "list" : "any",
    cimdAllowedDomains,
    cimdAccessTokenLifetimeSeconds:
      cimd?.client_config?.access_token_lifetime_seconds,
    cimdRefreshTokenLifetimeSeconds:
      cimd?.client_config?.refresh_token_lifetime_seconds,
    cimdRefreshTokenIdleTimeoutEnabled:
      cimd?.client_config?.refresh_token_idle_timeout_enabled ?? true,
    cimdRefreshTokenIdleTimeoutSeconds:
      cimd?.client_config?.refresh_token_idle_timeout_seconds,
  };
}

export function constructConfig(
  config: PortalAPIAppConfig,
  _initialState: FormState,
  currentState: FormState
): PortalAPIAppConfig {
  const [newConfig, _] = produce(
    [config, currentState],
    ([config, currentState]) => {
      config.oauth ??= {};
      config.oauth.client_id_metadata_document ??= {};
      const cimd = config.oauth.client_id_metadata_document;

      if (currentState.cimdEnabled) {
        cimd.enabled = true;
      } else {
        delete cimd.enabled;
      }

      if (currentState.cimdDomainMode === "list") {
        // An empty list in "Only these domains" mode is never written as an
        // empty array: the server reads that as "any domain", which is the
        // opposite of what the admin picked. Writing a single empty entry
        // instead fails the config schema's minLength, so the save is
        // refused with the error bound to the offending field rather than
        // silently widening access.
        cimd.allowed_domains =
          currentState.cimdAllowedDomains.length > 0
            ? currentState.cimdAllowedDomains
            : [""];
      } else {
        delete cimd.allowed_domains;
      }

      cimd.client_config ??= {};
      const cimdClientConfig = cimd.client_config;

      if (currentState.cimdAccessTokenLifetimeSeconds != null) {
        cimdClientConfig.access_token_lifetime_seconds =
          currentState.cimdAccessTokenLifetimeSeconds;
      } else {
        delete cimdClientConfig.access_token_lifetime_seconds;
      }

      if (currentState.cimdRefreshTokenLifetimeSeconds != null) {
        cimdClientConfig.refresh_token_lifetime_seconds =
          currentState.cimdRefreshTokenLifetimeSeconds;
      } else {
        delete cimdClientConfig.refresh_token_lifetime_seconds;
      }

      if (currentState.cimdRefreshTokenIdleTimeoutEnabled) {
        // Absent means enabled — the server default.
        delete cimdClientConfig.refresh_token_idle_timeout_enabled;
      } else {
        cimdClientConfig.refresh_token_idle_timeout_enabled = false;
      }

      if (currentState.cimdRefreshTokenIdleTimeoutSeconds != null) {
        cimdClientConfig.refresh_token_idle_timeout_seconds =
          currentState.cimdRefreshTokenIdleTimeoutSeconds;
      } else {
        delete cimdClientConfig.refresh_token_idle_timeout_seconds;
      }

      clearEmptyObject(config);
    }
  );
  return newConfig;
}

export interface MetadataDocumentsContentProps {
  form: AppConfigFormModel<FormState>;
}

export const MetadataDocumentsContent: React.VFC<MetadataDocumentsContentProps> =
  function MetadataDocumentsContent({ form }) {
    const { state, setState, isUpdating, effectiveConfig } = form;
    const { showToast } = useCalloutToast();
    const { renderToString } = useContext(Context);
    const { getIsDirty } = useFormContainerBaseContext();
    const isDirty = useMemo(() => getIsDirty(), [getIsDirty]);
    const anchorRef = useRef<HTMLDivElement>(null);

    // Every client_config field is resolved server-side by
    // OAuthDynamicClientTokenLifetimesConfig.SetDefaults(), so the effective
    // config carries concrete values even when authgear.yaml omits the whole
    // section. Read the placeholders from there rather than restating the Go
    // defaults here, where they would drift silently.
    const effectiveClientConfig =
      effectiveConfig.oauth?.client_id_metadata_document?.client_config;

    const [isAnyDomainConfirmationVisible, setIsAnyDomainConfirmationVisible] =
      useState(false);

    const enabled = state.cimdEnabled;

    // Saved immediately, matching the same switch on the DCR tab. CIMD has no
    // counterpart to that tab's initial access tokens -- nothing here acts
    // before a save -- so this is for consistency between two otherwise
    // identical tabs rather than to close a hole.
    const onEnabledChange = useCallback(
      (checked: boolean) => {
        form
          .saveWith((prev) =>
            // Turning it off starts from the last saved state rather than the
            // pending one: an unfinished edit elsewhere on the tab -- a seeded
            // empty domain row, say -- would otherwise fail validation and
            // take the switch down with it, leaving CIMD enabled on the
            // server while the switch reads off. Turning it on carries
            // pending edits, which is what the toast reports.
            checked
              ? { ...prev, cimdEnabled: true }
              : { ...form.initialState, cimdEnabled: false }
          )
          .then(() => {
            showToast({
              type: "success",
              text: <FormattedMessage id="changes-saved" />,
              duration: SAVED_TOAST_DURATION_MS,
            });
          })
          // performSave rethrows, and the form's error bar renders it.
          .catch(() => {});
      },
      [form, showToast]
    );

    const setDomainMode = useCallback(
      (mode: CIMDDomainMode) => {
        setState((prev) => ({
          ...prev,
          cimdDomainMode: mode,
          // Entering "Only these domains" with nothing listed would describe
          // a restriction that does not exist, so seed one empty row: the
          // admin sees where the domain goes, and an unfilled row fails
          // validation on save instead of quietly meaning "any domain".
          cimdAllowedDomains:
            mode === "list" && prev.cimdAllowedDomains.length === 0
              ? [""]
              : prev.cimdAllowedDomains,
        }));
      },
      [setState]
    );

    const onDomainModeChange = useCallback(
      (mode: CIMDDomainMode) => {
        if (mode === "any" && state.cimdDomainMode === "list") {
          // Widening from a restricted allowlist to "any domain" is the
          // permissive act on this card -- confirm it before reflecting it in
          // the form state.
          setIsAnyDomainConfirmationVisible(true);
          return;
        }
        setDomainMode(mode);
      },
      [state.cimdDomainMode, setDomainMode]
    );

    const onConfirmAnyDomain = useCallback(() => {
      setIsAnyDomainConfirmationVisible(false);
      setDomainMode("any");
    }, [setDomainMode]);

    const onCancelAnyDomain = useCallback(() => {
      setIsAnyDomainConfirmationVisible(false);
    }, []);

    const setAllowedDomains = useCallback(
      (domains: string[]) => {
        setState((prev) => ({
          ...prev,
          cimdAllowedDomains: domains,
        }));
      },
      [setState]
    );

    // TextFieldList hands back the CURRENT list, so each callback computes the
    // updated one itself.
    const onDomainAdd = useCallback(
      (list: string[], item: string) => {
        setAllowedDomains([...list, item]);
      },
      [setAllowedDomains]
    );

    const onDomainChange = useCallback(
      (list: string[], index: number, item: string) => {
        setAllowedDomains(list.map((value, i) => (i === index ? item : value)));
      },
      [setAllowedDomains]
    );

    const onDomainDelete = useCallback(
      (list: string[], index: number) => {
        setAllowedDomains(list.filter((_, i) => i !== index));
      },
      [setAllowedDomains]
    );

    const onAccessTokenLifetimeChange = useCallback(
      (e: React.ChangeEvent<HTMLInputElement>) => {
        const value = e.target.value;
        setState((prev) => ({
          ...prev,
          cimdAccessTokenLifetimeSeconds: parseIntegerAllowLeadingZeros(value),
        }));
      },
      [setState]
    );

    const onRefreshTokenLifetimeChange = useCallback(
      (e: React.ChangeEvent<HTMLInputElement>) => {
        const value = e.target.value;
        setState((prev) => ({
          ...prev,
          cimdRefreshTokenLifetimeSeconds: parseIntegerAllowLeadingZeros(value),
        }));
      },
      [setState]
    );

    const onRefreshTokenIdleTimeoutEnabledChange = useCallback(
      (checked: boolean) => {
        setState((prev) => ({
          ...prev,
          cimdRefreshTokenIdleTimeoutEnabled: checked,
        }));
      },
      [setState]
    );

    const onRefreshTokenIdleTimeoutChange = useCallback(
      (e: React.ChangeEvent<HTMLInputElement>) => {
        const value = e.target.value;
        setState((prev) => ({
          ...prev,
          cimdRefreshTokenIdleTimeoutSeconds:
            parseIntegerAllowLeadingZeros(value),
        }));
      },
      [setState]
    );

    return (
      <>
        <div
          ref={anchorRef}
          className={cn(
            styles.widget,
            styles.content,
            isDirty && styles.contentWithSaveBar
          )}
        >
          <SettingsSectionCard
            contentClassName="gap-4"
            title={
              <FormattedMessage id="MetadataDocumentsScreen.enable.title" />
            }
          >
            <Text as="p" size="2" color="gray">
              <FormattedMessage
                id="CIMDSection.enable.description"
                values={{
                  // eslint-disable-next-line react/no-unstable-nested-components
                  mcpLink: (chunks: React.ReactNode) => (
                    <ExternalLink href="https://docs.authgear.com/get-started/auth-for-mcp">
                      {chunks}
                    </ExternalLink>
                  ),
                }}
              />
            </Text>
            <Toggle
              checked={enabled}
              disabled={isUpdating}
              onCheckedChange={onEnabledChange}
              text={<FormattedMessage id="CIMDSection.enable.toggle.label" />}
            />
          </SettingsSectionCard>

          {enabled ? (
            <SettingsSectionCard
              contentClassName="gap-4"
              title={
                <FormattedMessage id="CIMDSection.trusted-domains.title" />
              }
              description={
                <FormattedMessage id="CIMDSection.trusted-domains.note" />
              }
            >
              <RadioGroup.Root
                value={state.cimdDomainMode}
                onValueChange={onDomainModeChange}
                disabled={isUpdating}
              >
                <Flex direction="column" gap="3">
                  <Text as="label" size="2">
                    <Flex gap="2" align="start">
                      <RadioGroup.Item value="any" />
                      <FormattedMessage id="CIMDSection.trusted-domains.any.title" />
                    </Flex>
                  </Text>

                  <div className="flex flex-col gap-2">
                    <Text as="label" size="2">
                      <Flex gap="2" align="start">
                        <RadioGroup.Item value="list" />
                        <FormattedMessage id="CIMDSection.trusted-domains.list.title" />
                      </Flex>
                    </Text>
                    {state.cimdDomainMode === "list" ? (
                      <div className={styles.fieldInOption}>
                        <TextFieldList
                          parentJSONPointer={CIMD_JSON_POINTER}
                          fieldName="allowed_domains"
                          list={state.cimdAllowedDomains}
                          onListItemAdd={onDomainAdd}
                          onListItemChange={onDomainChange}
                          onListItemDelete={onDomainDelete}
                          addButtonLabelMessageID="CIMDSection.trusted-domains.add"
                          deleteButtonAriaLabel={renderToString(
                            "CIMDSection.trusted-domains.delete"
                          )}
                          placeholder={renderToString(
                            "CIMDSection.trusted-domains.placeholder"
                          )}
                          itemErrorMessageID="CIMDSection.trusted-domains.invalid"
                          label={
                            <span className={styles.labelWithHint}>
                              <FormattedMessage id="CIMDSection.trusted-domains.label" />
                              <Tooltip
                                content={renderToString(
                                  "CIMDSection.trusted-domains.wildcard-hint"
                                )}
                              >
                                <QuestionMarkCircledIcon
                                  className={styles.hintIcon}
                                  width="1rem"
                                  height="1rem"
                                />
                              </Tooltip>
                            </span>
                          }
                          disabled={isUpdating}
                          minItem={1}
                        />
                      </div>
                    ) : null}
                  </div>
                </Flex>
              </RadioGroup.Root>
            </SettingsSectionCard>
          ) : null}

          {enabled ? (
            <SettingsSectionCard
              contentClassName="gap-4"
              title={
                <FormattedMessage id="DynamicClientsTab.client-config.title" />
              }
              description={
                <FormattedMessage id="CIMDSection.client-config.description" />
              }
            >
              <TextField
                size="2"
                labelSize="2"
                type="text"
                label={
                  <FormattedMessage id="DynamicClientsTab.access-token-lifetime.label" />
                }
                parentJSONPointer={CIMD_CLIENT_CONFIG_JSON_POINTER}
                fieldName="access_token_lifetime_seconds"
                placeholder={effectiveClientConfig?.access_token_lifetime_seconds?.toFixed(
                  0
                )}
                value={state.cimdAccessTokenLifetimeSeconds?.toFixed(0) ?? ""}
                onChange={onAccessTokenLifetimeChange}
              />
              <TextField
                size="2"
                labelSize="2"
                type="text"
                label={
                  <FormattedMessage id="DynamicClientsTab.refresh-token-lifetime.label" />
                }
                parentJSONPointer={CIMD_CLIENT_CONFIG_JSON_POINTER}
                fieldName="refresh_token_lifetime_seconds"
                placeholder={effectiveClientConfig?.refresh_token_lifetime_seconds?.toFixed(
                  0
                )}
                value={state.cimdRefreshTokenLifetimeSeconds?.toFixed(0) ?? ""}
                onChange={onRefreshTokenLifetimeChange}
              />
              <div className="flex flex-col gap-1">
                <Toggle
                  checked={state.cimdRefreshTokenIdleTimeoutEnabled}
                  disabled={isUpdating}
                  onCheckedChange={onRefreshTokenIdleTimeoutEnabledChange}
                  text={
                    <FormattedMessage id="EditOAuthClientForm.refresh-token-idle-timeout-enabled.label" />
                  }
                />
                <Text as="p" size="1" color="gray">
                  <FormattedMessage id="EditOAuthClientForm.refresh-token-idle-timeout-enabled.description" />
                </Text>
              </div>
              <TextField
                size="2"
                labelSize="2"
                type="text"
                disabled={!state.cimdRefreshTokenIdleTimeoutEnabled}
                label={
                  <FormattedMessage id="DynamicClientsTab.refresh-token-idle-timeout.label" />
                }
                parentJSONPointer={CIMD_CLIENT_CONFIG_JSON_POINTER}
                fieldName="refresh_token_idle_timeout_seconds"
                placeholder={effectiveClientConfig?.refresh_token_idle_timeout_seconds?.toFixed(
                  0
                )}
                value={
                  state.cimdRefreshTokenIdleTimeoutSeconds?.toFixed(0) ?? ""
                }
                onChange={onRefreshTokenIdleTimeoutChange}
              />
            </SettingsSectionCard>
          ) : null}

          <ConfirmationDialog
            open={isAnyDomainConfirmationVisible}
            onOpenChange={setIsAnyDomainConfirmationVisible}
            title={
              <FormattedMessage id="CIMDSection.any-domain.confirm.title" />
            }
            description={
              <FormattedMessage id="CIMDSection.any-domain.confirm.description" />
            }
            confirmText={
              <FormattedMessage id="CIMDSection.any-domain.confirm.confirm" />
            }
            cancelText={<FormattedMessage id="cancel" />}
            onConfirm={onConfirmAnyDomain}
            onCancel={onCancelAnyDomain}
          />

          <SaveFunctionBar anchorRef={anchorRef} />
        </div>
      </>
    );
  };
