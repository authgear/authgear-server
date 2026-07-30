import React, { useCallback, useContext, useMemo, useRef, useState } from "react";
import { useParams } from "react-router-dom";
import cn from "classnames";
import { Dialog, Flex, Switch, Text } from "@radix-ui/themes";
import { Cross2Icon, Pencil1Icon } from "@radix-ui/react-icons";
import { Context, FormattedMessage } from "../../intl";
import { produce } from "immer";
import FormContainer from "../../FormContainer";
import {
  AppConfigFormModel,
  useAppConfigForm,
} from "../../hook/useAppConfigForm";
import ScreenContent from "../../ScreenContent";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import UserProfileAttributesList, {
  UserProfileAttributesListItem,
  UserProfileAttributesListSection,
  ItemComponentProps,
} from "../../UserProfileAttributesList";
import {
  PortalAPIAppConfig,
  StandardAttributesAccessControlConfig,
  UserProfileAttributesAccessControl,
} from "../../types";
import { parseJSONPointer } from "../../util/jsonpointer";
import { SaveFunctionBar } from "../../components/v2/SaveFunctionBar/SaveFunctionBar";
import { PrimaryButton } from "../../components/v2/Button/PrimaryButton/PrimaryButton";
import { SecondaryButton } from "../../components/v2/Button/SecondaryButton/SecondaryButton";
import {
  TextField,
  TextFieldIcon,
} from "../../components/v2/TextField/TextField";
import { useFormContainerBaseContext } from "../../FormContainerBase";
import styles from "./StandardAttributesConfigurationScreen.module.css";
import ExternalLink from "../../ExternalLink";

// Identity attributes are always required and cannot be disabled.
const REQUIRED_POINTERS = new Set([
  "/email",
  "/phone_number",
  "/preferred_username",
]);

const DEFAULT_ACCESS_CONTROL: UserProfileAttributesAccessControl = {
  portal_ui: "readwrite",
  bearer: "readonly",
  end_user: "readwrite",
};

// The config has no on/off flag for a standard attribute, so a disabled
// attribute is stored as hidden from every party.
const DISABLED_ACCESS_CONTROL: UserProfileAttributesAccessControl = {
  portal_ui: "hidden",
  bearer: "hidden",
  end_user: "hidden",
};

function isDisabled(
  accessControl: UserProfileAttributesAccessControl
): boolean {
  return (
    accessControl.portal_ui === "hidden" &&
    accessControl.bearer === "hidden" &&
    accessControl.end_user === "hidden"
  );
}

interface FormState {
  standardAttributesItems: StandardAttributesAccessControlConfig[];
}

interface StandardAttributesConfigurationScreenContentProps {
  form: AppConfigFormModel<FormState>;
}

const standardAttributeSections: UserProfileAttributesListSection[] = [
  {
    key: "identity",
    titleMessageId:
      "StandardAttributesConfigurationScreen.section.identity-attributes",
    pointers: ["/email", "/phone_number", "/preferred_username"],
  },
  {
    key: "name",
    titleMessageId:
      "StandardAttributesConfigurationScreen.section.name-attributes",
    pointers: [
      "/name",
      "/given_name",
      "/family_name",
      "/middle_name",
      "/nickname",
    ],
  },
  {
    key: "profile",
    titleMessageId:
      "StandardAttributesConfigurationScreen.section.profile-attributes",
    pointers: [
      "/profile",
      "/picture",
      "/website",
      "/gender",
      "/birthdate",
      "/address",
    ],
  },
  {
    key: "local-preferences",
    titleMessageId:
      "StandardAttributesConfigurationScreen.section.local-preferences-attributes",
    pointers: ["/zoneinfo", "/locale"],
  },
];

const naturalOrder = standardAttributeSections.flatMap(
  (section) => section.pointers
);

function constructFormState(config: PortalAPIAppConfig): FormState {
  const items = config.user_profile?.standard_attributes?.access_control ?? [];
  const listedItems = items.filter(
    (a) =>
      naturalOrder.indexOf(a.pointer) >= 0 &&
      (REQUIRED_POINTERS.has(a.pointer) || !isDisabled(a.access_control))
  );
  listedItems.sort((a, b) => {
    const ia = naturalOrder.indexOf(a.pointer);
    const ib = naturalOrder.indexOf(b.pointer);
    return ia - ib;
  });
  return {
    standardAttributesItems: listedItems,
  };
}

function constructConfig(
  rawConfig: PortalAPIAppConfig,
  _initialState: FormState,
  currentState: FormState,
  effectiveConfig: PortalAPIAppConfig
): PortalAPIAppConfig {
  const modifiedEffectiveConfig = produce(
    effectiveConfig,
    (effectiveConfig) => {
      effectiveConfig.user_profile ??= {};
      effectiveConfig.user_profile.standard_attributes ??= {};
      const itemByPointer = new Map(
        currentState.standardAttributesItems.map((item) => [item.pointer, item])
      );
      for (const accessControl of effectiveConfig.user_profile
        .standard_attributes.access_control ?? []) {
        const item = itemByPointer.get(accessControl.pointer);
        if (item != null) {
          accessControl.access_control = item.access_control;
        } else if (naturalOrder.indexOf(accessControl.pointer) >= 0) {
          accessControl.access_control = DISABLED_ACCESS_CONTROL;
        }
      }
    }
  );

  const accessControl =
    modifiedEffectiveConfig.user_profile?.standard_attributes?.access_control;
  return produce(rawConfig, (rawConfig) => {
    rawConfig.user_profile ??= {};
    rawConfig.user_profile.standard_attributes ??= {};
    rawConfig.user_profile.standard_attributes.access_control = accessControl;
  });
}

interface EditAttributesDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  enabledPointers: Set<string>;
  onTogglePointer: (pointer: string, enabled: boolean) => void;
  onApply: () => void;
  onCancel: () => void;
}

function EditAttributesDialog({
  open,
  onOpenChange,
  enabledPointers,
  onTogglePointer,
  onApply,
  onCancel,
}: EditAttributesDialogProps) {
  const { renderToString } = useContext(Context);

  return (
    <Dialog.Root open={open} onOpenChange={onOpenChange}>
      <Dialog.Content maxWidth="560px" size="3">
        <Dialog.Title>
          <FormattedMessage id="StandardAttributesConfigurationScreen.edit-attributes.dialog.title" />
        </Dialog.Title>
        <Dialog.Description size="2" mb="4">
          <FormattedMessage id="StandardAttributesConfigurationScreen.edit-attributes.dialog.description" />
        </Dialog.Description>

        <div className={styles.editDialogTable}>
          {/* Column header */}
          <div className={cn(styles.editDialogHeaderRow)}>
            <Text size="2" weight="medium">
              <FormattedMessage id="StandardAttributesConfigurationScreen.edit-attributes.dialog.column.attribute" />
            </Text>
            <Text size="2" weight="medium">
              <FormattedMessage id="StandardAttributesConfigurationScreen.edit-attributes.dialog.column.enabled" />
            </Text>
          </div>

          {/* Attribute rows grouped by section */}
          {standardAttributeSections.map((section) => (
            <React.Fragment key={section.key}>
              <div className={styles.editDialogSectionHeader}>
                <Text size="1" weight="medium" color="gray">
                  <FormattedMessage id={section.titleMessageId} />
                </Text>
              </div>
              {section.pointers.map((pointer) => {
                const fieldName = parseJSONPointer(pointer)[0];
                const required = REQUIRED_POINTERS.has(pointer);
                const checked = required || enabledPointers.has(pointer);
                return (
                  <div key={pointer} className={styles.editDialogRow}>
                    <div className={styles.editDialogRowLabel}>
                      <Text as="p" size="2">
                        {renderToString("standard-attribute." + fieldName)}
                      </Text>
                      {required ? (
                        <Text as="p" size="1" color="gray">
                          <FormattedMessage id="StandardAttributesConfigurationScreen.edit-attributes.dialog.required" />
                        </Text>
                      ) : null}
                    </div>
                    <Switch
                      checked={checked}
                      disabled={required}
                      onCheckedChange={(c) => onTogglePointer(pointer, c)}
                    />
                  </div>
                );
              })}
            </React.Fragment>
          ))}
        </div>

        <Flex gap="3" justify="end" mt="4">
          <SecondaryButton
            size="2"
            text={<FormattedMessage id="cancel" />}
            onClick={onCancel}
          />
          <PrimaryButton
            size="2"
            text={
              <FormattedMessage id="StandardAttributesConfigurationScreen.edit-attributes.dialog.apply" />
            }
            onClick={onApply}
          />
        </Flex>
      </Dialog.Content>
    </Dialog.Root>
  );
}

function ItemComponent(
  props: ItemComponentProps<StandardAttributesAccessControlConfig>
) {
  const { className, item } = props;
  const { pointer } = item;
  const fieldName = parseJSONPointer(pointer)[0];
  const messageId = "standard-attribute.description." + fieldName;

  const renderExternalLink = useCallback(
    (children: React.ReactNode) => (
      <ExternalLink href="https://en.wikipedia.org/wiki/List_of_tz_database_time_zones">
        {children}
      </ExternalLink>
    ),
    []
  );

  const values = useMemo(() => {
    if (messageId === "standard-attribute.description.zoneinfo") {
      return {
        externalLink: renderExternalLink,
      };
    }
    return {};
  }, [messageId, renderExternalLink]);

  return (
    <div className={className}>
      <Text as="p" size="2" weight="medium">
        <FormattedMessage id={"standard-attribute." + fieldName} />
      </Text>
      <Text as="p" size="1" color="gray">
        <FormattedMessage id={messageId} values={values} />
      </Text>
    </div>
  );
}

const StandardAttributesConfigurationScreenContent: React.VFC<StandardAttributesConfigurationScreenContentProps> =
  function StandardAttributesConfigurationScreenContent(props) {
    const { state, setState } = props.form;
    const { isDirty } = useFormContainerBaseContext();
    const { renderToString } = useContext(Context);
    const contentWidthAnchorRef = useRef<HTMLDivElement>(null);
    const [searchKeyword, setSearchKeyword] = useState("");

    // Edit Attributes dialog state
    const [editDialogOpen, setEditDialogOpen] = useState(false);
    const [draftEnabledPointers, setDraftEnabledPointers] = useState<Set<string>>(
      new Set()
    );

    const currentEnabledPointers = useMemo(
      () => new Set(state.standardAttributesItems.map((item) => item.pointer)),
      [state.standardAttributesItems]
    );

    const onOpenEditDialog = useCallback(() => {
      // Seed draft with currently enabled pointers (always include required ones)
      const draft = new Set(currentEnabledPointers);
      REQUIRED_POINTERS.forEach((p) => draft.add(p));
      setDraftEnabledPointers(draft);
      setEditDialogOpen(true);
    }, [currentEnabledPointers]);

    const onCloseEditDialog = useCallback(() => {
      setEditDialogOpen(false);
    }, []);

    const onEditDialogOpenChange = useCallback(
      (open: boolean) => {
        if (!open) setEditDialogOpen(false);
      },
      []
    );

    const onToggleDraftPointer = useCallback(
      (pointer: string, enabled: boolean) => {
        setDraftEnabledPointers((prev) => {
          const next = new Set(prev);
          if (enabled) {
            next.add(pointer);
          } else {
            next.delete(pointer);
          }
          return next;
        });
      },
      []
    );

    const onApplyEditDialog = useCallback(() => {
      setState((prev) => {
        const prevByPointer = new Map(
          prev.standardAttributesItems.map((item) => [item.pointer, item])
        );
        const newItems = naturalOrder
          .filter(
            (pointer) =>
              draftEnabledPointers.has(pointer) ||
              REQUIRED_POINTERS.has(pointer)
          )
          .map(
            (pointer): StandardAttributesAccessControlConfig =>
              prevByPointer.get(pointer) ?? {
                pointer,
                access_control: DEFAULT_ACCESS_CONTROL,
              }
          );
        return { ...prev, standardAttributesItems: newItems };
      });
      setEditDialogOpen(false);
    }, [setState, draftEnabledPointers]);

    const onChangeSearchKeyword = useCallback(
      (e: React.ChangeEvent<HTMLInputElement>) => {
        setSearchKeyword(e.currentTarget.value);
      },
      []
    );

    const onClearSearchKeyword = useCallback(() => {
      setSearchKeyword("");
    }, []);

    const filteredItems = useMemo(() => {
      const keyword = searchKeyword.trim().toLowerCase();
      if (keyword === "") {
        return state.standardAttributesItems;
      }
      return state.standardAttributesItems.filter((item) => {
        const fieldName = parseJSONPointer(item.pointer)[0];
        const name = renderToString(
          "standard-attribute." + fieldName
        ).toLowerCase();
        return (
          name.includes(keyword) || fieldName.toLowerCase().includes(keyword)
        );
      });
    }, [searchKeyword, state.standardAttributesItems, renderToString]);

    const onChangeItems = useCallback(
      (newItems: UserProfileAttributesListItem[]) => {
        setState((prev) => {
          const updatedByPointer = new Map(
            newItems.map((item) => [item.pointer, item])
          );
          return {
            ...prev,
            standardAttributesItems: prev.standardAttributesItems.map(
              (item) => updatedByPointer.get(item.pointer) ?? item
            ),
          };
        });
      },
      [setState]
    );

    return (
      <ScreenContent
        layout="list"
        className={cn(
          styles.screenContent,
          isDirty ? styles.contentWithSaveBar : null
        )}
      >
        <div ref={contentWidthAnchorRef} className={styles.widget}>
          <div className={styles.header}>
            <Text as="p" size="5" weight="bold" className={styles.pageTitle}>
              <FormattedMessage id="StandardAttributesConfigurationScreen.title" />
            </Text>
            <SecondaryButton
              size="2"
              text={
                <span className={styles.editButtonContent}>
                  <Pencil1Icon width="1rem" height="1rem" />
                  <FormattedMessage id="StandardAttributesConfigurationScreen.edit-attributes.button" />
                </span>
              }
              onClick={onOpenEditDialog}
            />
          </div>
        </div>
        <div className={cn(styles.widget, styles.toolbar)}>
          <div className={styles.searchField}>
            <TextField
              size="2"
              type="search"
              value={searchKeyword}
              placeholder={renderToString(
                "StandardAttributesConfigurationScreen.search.placeholder"
              )}
              iconStart={TextFieldIcon.MagnifyingGlass}
              onChange={onChangeSearchKeyword}
              suffixPlain={true}
              suffix={
                searchKeyword !== "" ? (
                  <button
                    type="button"
                    className={styles.searchClearButton}
                    aria-label={renderToString(
                      "StandardAttributesConfigurationScreen.clear-search"
                    )}
                    onClick={onClearSearchKeyword}
                  >
                    <Cross2Icon className={styles.searchClearIcon} />
                  </button>
                ) : undefined
              }
            />
          </div>
        </div>
        <div className={cn(styles.widget, styles.tableWidget)}>
          {filteredItems.length === 0 ? (
            <Text as="p" size="2" color="gray" className={styles.emptySearch}>
              <FormattedMessage id="SearchableDropdown.empty" />
            </Text>
          ) : (
            <UserProfileAttributesList
              items={filteredItems}
              onChangeItems={onChangeItems}
              ItemComponent={ItemComponent}
              sections={standardAttributeSections}
            />
          )}
        </div>
        <SaveFunctionBar anchorRef={contentWidthAnchorRef} />

        <EditAttributesDialog
          open={editDialogOpen}
          onOpenChange={onEditDialogOpenChange}
          enabledPointers={draftEnabledPointers}
          onTogglePointer={onToggleDraftPointer}
          onApply={onApplyEditDialog}
          onCancel={onCloseEditDialog}
        />
      </ScreenContent>
    );
  };

const StandardAttributesConfigurationScreen: React.VFC =
  function StandardAttributesConfigurationScreen() {
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

    return (
      <FormContainer
        form={form}
        hideFooterComponent={true}
        canSave={true}
      >
        <StandardAttributesConfigurationScreenContent form={form} />
      </FormContainer>
    );
  };

export default StandardAttributesConfigurationScreen;
