import React, { useCallback, useContext, useMemo, useState } from "react";
import { useParams } from "react-router-dom";
import {
  ActionButton,
  Checkbox,
  DefaultButton,
  Dialog,
  DialogFooter,
  DirectionalHint,
  Dropdown,
  IDialogProps,
  IListProps,
  ITooltipProps,
  List,
  ScrollablePane,
  Stack,
  Text,
  TooltipHost,
  VerticalDivider,
} from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";

import { useRemoveTemplateLocalesMutation } from "./mutations/updateAppTemplatesMutation";
import ButtonWithLoading from "../../ButtonWithLoading";
import ErrorDialog from "../../error/ErrorDialog";
import { useSystemConfig } from "../../context/SystemConfigContext";
import { useCheckbox, useDropdown } from "../../hook/useInput";
import { TemplateLocale } from "../../templates";

import styles from "./TemplateLocaleManagement.module.scss";

type TemplateLocaleUpdater = (locale: TemplateLocale) => void;

interface TemplateLocaleManagementProps {
  configuredTemplateLocales: TemplateLocale[];
  templateLocale: TemplateLocale;
  initialDefaultTemplateLocale: TemplateLocale;
  defaultTemplateLocale: TemplateLocale;
  onTemplateLocaleSelected: TemplateLocaleUpdater;
  onDefaultTemplateLocaleSelected: TemplateLocaleUpdater;
  pendingTemplateLocales: TemplateLocale[];
  onPendingTemplateLocalesChange: (locales: TemplateLocale[]) => void;
}

interface TemplateLocaleManagementDialogProps {
  defaultTemplateLocale: TemplateLocale;
  presented: boolean;
  onDismiss: () => void;
  configuredTemplateLocales: TemplateLocale[];
  pendingTemplateLocales: TemplateLocale[];
  onPendingTemplateLocalesChange: (locales: TemplateLocale[]) => void;
  onTemplateLocaleDeleted: (
    configuredLocales: TemplateLocale[],
    pendingLocales: TemplateLocale[]
  ) => void;
}

interface TemplateLocaleListItemProps {
  locale: TemplateLocale;
  onItemSelected: (locale: TemplateLocale, checked: boolean) => void;
}

interface SelectedTemplateLocaleItemProps {
  locale: TemplateLocale;
  onItemRemoved: (locale: TemplateLocale) => void;
  isDefaultLocale: boolean;
}

const DIALOG_STYLES = {
  main: {
    maxWidth: "none !important",
    width: "80vw !important",
    minWidth: "500px !important",
  },
};

function getLanguageLocaleKey(locale: TemplateLocale) {
  return `Locales.${locale}`;
}

const TemplateLocaleListItem: React.FC<TemplateLocaleListItemProps> = function TemplateLocaleListItem(
  props: TemplateLocaleListItemProps
) {
  const { locale, onItemSelected } = props;

  const { onChange } = useCheckbox((checked) => {
    onItemSelected(locale, checked);
  });

  return (
    <div className={styles.dialogLocaleListItem}>
      <Checkbox onChange={onChange} />
      <Text className={styles.dialogLocaleListItemText}>
        <FormattedMessage id={getLanguageLocaleKey(locale)} />
      </Text>
    </div>
  );
};

const SelectedTemplateLocaleItem: React.FC<SelectedTemplateLocaleItemProps> = function SelectedTemplateLocaleItem(
  props: SelectedTemplateLocaleItemProps
) {
  const { locale, onItemRemoved, isDefaultLocale } = props;
  const { themes } = useSystemConfig();

  const tooltipProps: ITooltipProps = useMemo(() => {
    return {
      onRenderContent: () => (
        <div className={styles.tooltip}>
          <Text className={styles.tooltipMessage}>
            <FormattedMessage id="TemplateLocaleManagementDialog.cannot-remove-default-language-error" />
          </Text>
        </div>
      ),
    };
  }, []);

  const onDeleteClicked = useCallback(() => {
    onItemRemoved(locale);
  }, [locale, onItemRemoved]);

  return (
    <div className={styles.dialogSelectedItem}>
      <Text>
        <FormattedMessage id={getLanguageLocaleKey(locale)} />
      </Text>
      <TooltipHost
        hidden={!isDefaultLocale}
        tooltipProps={tooltipProps}
        directionalHint={DirectionalHint.bottomCenter}
      >
        <ActionButton
          iconProps={{ iconName: "Delete" }}
          theme={themes.destructive}
          onClick={onDeleteClicked}
          disabled={isDefaultLocale}
        />
      </TooltipHost>
    </div>
  );
};

const TemplateLocaleManagementDialog: React.FC<TemplateLocaleManagementDialogProps> = function TemplateLocaleManagementDialog(
  props: TemplateLocaleManagementDialogProps
) {
  const {
    presented,
    onDismiss,
    configuredTemplateLocales,
    pendingTemplateLocales,
    onPendingTemplateLocalesChange,
    onTemplateLocaleDeleted,
    defaultTemplateLocale,
  } = props;

  const { supportedResourceLocales } = useSystemConfig();
  const { appID } = useParams();

  const {
    removeTemplateLocales,
    loading: removingTemplateLoacles,
    error: removeTemplateLocalesError,
  } = useRemoveTemplateLocalesMutation(appID);

  const [localErrorMessage, setLocalErrorMessage] = useState<
    string | undefined
  >();

  const initialSelectedLocales = useMemo(() => {
    return configuredTemplateLocales.concat(pendingTemplateLocales);
  }, [configuredTemplateLocales, pendingTemplateLocales]);
  const [selectedLocales, setSelectedLocales] = useState<TemplateLocale[]>(
    initialSelectedLocales
  );

  const onTemplateLocaleListItemSelected = useCallback(
    (locale: TemplateLocale, checked: boolean) => {
      setSelectedLocales((prev) => {
        const modifiedIndex = prev.findIndex((item) => item === locale);
        if (checked && modifiedIndex < 0) {
          return [...prev, locale];
        }
        if (!checked && modifiedIndex >= 0) {
          const updated = [...prev];
          updated.splice(modifiedIndex, 1);
          return updated;
        }
        return prev;
      });
    },
    []
  );
  const onSelctedTemplateLocaleRemoved = useCallback(
    (locale: TemplateLocale) => {
      setSelectedLocales((prev) => {
        return prev.filter((item) => item !== locale);
      });
    },
    []
  );

  const renderLocaleListItemCell = useCallback<
    Required<IListProps<TemplateLocale>>["onRenderCell"]
  >(
    (locale) => {
      if (locale == null) {
        return null;
      }
      return (
        <TemplateLocaleListItem
          locale={locale}
          onItemSelected={onTemplateLocaleListItemSelected}
        />
      );
    },
    [onTemplateLocaleListItemSelected]
  );

  const renderSelectedLocaleItemCell = useCallback<
    Required<IListProps<TemplateLocale>>["onRenderCell"]
  >(
    (locale) => {
      if (locale == null) {
        return null;
      }
      return (
        <SelectedTemplateLocaleItem
          locale={locale}
          onItemRemoved={onSelctedTemplateLocaleRemoved}
          isDefaultLocale={locale === defaultTemplateLocale}
        />
      );
    },
    [onSelctedTemplateLocaleRemoved, defaultTemplateLocale]
  );

  const onCancel = useCallback(() => {
    setSelectedLocales(initialSelectedLocales);
    onDismiss();
  }, [onDismiss, initialSelectedLocales]);

  const onApplyClick = useCallback(() => {
    const selectedLocaleSet = new Set(selectedLocales);
    const configuredLocaleSet = new Set(configuredTemplateLocales);
    const removedLocales = configuredTemplateLocales.filter(
      (locale) => !selectedLocaleSet.has(locale)
    );
    if (removedLocales.includes(defaultTemplateLocale)) {
      setLocalErrorMessage(
        "TemplateLocaleManagementDialog.cannot-remove-default-language-error"
      );
      return;
    }
    const updatedConfiguredTemplateLocales = configuredTemplateLocales.filter(
      (locale) => selectedLocaleSet.has(locale)
    );

    const updatedPendingTemplateLocales = selectedLocales.filter(
      (locale) => !configuredLocaleSet.has(locale)
    );

    if (removedLocales.length > 0) {
      removeTemplateLocales(removedLocales)
        .then(() => {
          onPendingTemplateLocalesChange(updatedPendingTemplateLocales);
          onTemplateLocaleDeleted(
            updatedConfiguredTemplateLocales,
            updatedPendingTemplateLocales
          );
          onDismiss();
        })
        .catch(() => {});
    } else {
      onPendingTemplateLocalesChange(updatedPendingTemplateLocales);
      onDismiss();
    }
  }, [
    defaultTemplateLocale,
    selectedLocales,
    configuredTemplateLocales,
    onPendingTemplateLocalesChange,
    removeTemplateLocales,
    onDismiss,
    onTemplateLocaleDeleted,
  ]);

  const modalProps = useMemo<IDialogProps["modalProps"]>(() => {
    return {
      isBlocking: true,
      topOffsetFixed: true,
    };
  }, []);

  return (
    <>
      <ErrorDialog
        error={removeTemplateLocalesError}
        rules={[]}
        errorMessage={localErrorMessage}
        fallbackErrorMessageID="TemplateLocaleManagementDialog.apply-error"
      />
      <Dialog
        hidden={!presented}
        onDismiss={onDismiss}
        title={<FormattedMessage id="TemplateLocaleManagementDialog.title" />}
        modalProps={modalProps}
        styles={DIALOG_STYLES}
      >
        <Text className={styles.dialogDesc}>
          <FormattedMessage id="TemplateLocaleManagementDialog.desc" />
        </Text>
        <div className={styles.dialogContent}>
          <section className={styles.dialogColumn}>
            <Text className={styles.dialogColumnHeader}>
              <FormattedMessage id="TemplateLocaleManagementDialog.supported-resource-locales-header" />
            </Text>
            <section className={styles.dialogListWrapper}>
              <ScrollablePane>
                <List
                  items={supportedResourceLocales}
                  onRenderCell={renderLocaleListItemCell}
                />
              </ScrollablePane>
            </section>
          </section>
          <VerticalDivider className={styles.dialogDivider} />
          <section className={styles.dialogColumn}>
            <Text className={styles.dialogColumnHeader}>
              <FormattedMessage id="TemplateLocaleManagementDialog.selected-template-locales-header" />
            </Text>
            <section className={styles.dialogListWrapper}>
              <ScrollablePane>
                <List
                  items={selectedLocales}
                  onRenderCell={renderSelectedLocaleItemCell}
                />
              </ScrollablePane>
            </section>
          </section>
        </div>
        <DialogFooter>
          <DefaultButton onClick={onCancel}>
            <FormattedMessage id="cancel" />
          </DefaultButton>
          <ButtonWithLoading
            loading={removingTemplateLoacles}
            onClick={onApplyClick}
            labelId="apply"
            loadingLabelId="applying"
          />
        </DialogFooter>
      </Dialog>
    </>
  );
};

const TemplateLocaleManagement: React.FC<TemplateLocaleManagementProps> = function TemplateLocaleManagement(
  props: TemplateLocaleManagementProps
) {
  const {
    configuredTemplateLocales,
    templateLocale,
    initialDefaultTemplateLocale,
    defaultTemplateLocale,
    onTemplateLocaleSelected,
    onDefaultTemplateLocaleSelected,
    pendingTemplateLocales,
    onPendingTemplateLocalesChange,
  } = props;

  const { themes } = useSystemConfig();
  const { renderToString } = useContext(Context);

  const [isDialogPresented, setIsDialogPresented] = useState(false);

  const displayTemplateLocale = useCallback(
    (locale: TemplateLocale) => {
      return renderToString(getLanguageLocaleKey(locale));
    },
    [renderToString]
  );

  const templateLocaleList = useMemo(() => {
    return configuredTemplateLocales.concat(pendingTemplateLocales);
  }, [configuredTemplateLocales, pendingTemplateLocales]);

  const displayTemplateLocaleOption = useCallback(
    (locale: TemplateLocale) => {
      const localeDisplayText = displayTemplateLocale(locale);
      if (locale === initialDefaultTemplateLocale) {
        return renderToString(
          "TemplatesConfigurationScreen.default-template-locale",
          { locale: localeDisplayText }
        );
      } else if (pendingTemplateLocales.includes(locale)) {
        return renderToString(
          "TemplatesConfigurationScreen.pending-template-locale",
          { locale: localeDisplayText }
        );
      }
      return localeDisplayText;
    },
    [
      initialDefaultTemplateLocale,
      pendingTemplateLocales,
      displayTemplateLocale,
      renderToString,
    ]
  );

  const {
    options: templateLocaleOptions,
    onChange: onTemplateLocaleChange,
  } = useDropdown(
    templateLocaleList,
    onTemplateLocaleSelected,
    templateLocale,
    displayTemplateLocaleOption
  );

  const {
    options: defaultTemplateLocaleOptions,
    onChange: onDefaultTemplateLocaleChange,
  } = useDropdown(
    configuredTemplateLocales,
    onDefaultTemplateLocaleSelected,
    defaultTemplateLocale,
    displayTemplateLocale
  );

  const presentDialog = useCallback(() => {
    setIsDialogPresented(true);
  }, []);

  const dismissDialog = useCallback(() => {
    setIsDialogPresented(false);
  }, []);

  const onTemplateLocaleDeleted = useCallback(
    (configuredLocales: TemplateLocale[], pendingLocales: TemplateLocale[]) => {
      // Check if selected is deleted
      const oldLocaleList = [...templateLocaleList];
      const localeList = configuredLocales.concat(pendingLocales);
      if (!localeList.includes(templateLocale)) {
        const prevOptionIndex = oldLocaleList.findIndex(
          (locale) => locale === templateLocale
        );
        if (prevOptionIndex !== -1) {
          // find element from old locale list which exists from
          // updated list from previous selected item
          for (let i = 0; i < oldLocaleList.length; i++) {
            const currIndex = (prevOptionIndex + i) % oldLocaleList.length;
            const currElem = oldLocaleList[currIndex];
            if (localeList.includes(currElem)) {
              onTemplateLocaleSelected(currElem);
              break;
            }
          }
        } else {
          onTemplateLocaleSelected(localeList[0]);
        }
      }
    },
    [onTemplateLocaleSelected, templateLocale, templateLocaleList]
  );

  return (
    <section className={styles.templateLocaleManagement}>
      <TemplateLocaleManagementDialog
        presented={isDialogPresented}
        configuredTemplateLocales={configuredTemplateLocales}
        pendingTemplateLocales={pendingTemplateLocales}
        onPendingTemplateLocalesChange={onPendingTemplateLocalesChange}
        onDismiss={dismissDialog}
        onTemplateLocaleDeleted={onTemplateLocaleDeleted}
        defaultTemplateLocale={defaultTemplateLocale}
      />
      <Stack
        className={styles.inputContainer}
        verticalAlign="end"
        horizontal={true}
        tokens={{ childrenGap: 10 }}
      >
        <Dropdown
          className={styles.dropdown}
          label={renderToString(
            "TemplatesConfigurationScreen.template-locale-dropdown"
          )}
          options={templateLocaleOptions}
          onChange={onTemplateLocaleChange}
          selectedKey={templateLocale}
        />
        <Dropdown
          className={styles.dropdown}
          label={renderToString(
            "TemplatesConfigurationScreen.default-template-locale-dropdown"
          )}
          options={defaultTemplateLocaleOptions}
          onChange={onDefaultTemplateLocaleChange}
          selectedKey={defaultTemplateLocale}
        />
        <ActionButton theme={themes.actionButton} onClick={presentDialog}>
          <FormattedMessage id="TemplatesConfigurationScreen.manage-template-locale" />
        </ActionButton>
      </Stack>
    </section>
  );
};

export default TemplateLocaleManagement;
