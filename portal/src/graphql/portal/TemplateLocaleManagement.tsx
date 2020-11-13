import React, { useCallback, useContext, useMemo, useState } from "react";
import { useParams } from "react-router-dom";
import {
  ActionButton,
  Checkbox,
  DefaultButton,
  Dialog,
  DialogFooter,
  Dropdown,
  IDialogProps,
  IListProps,
  List,
  ScrollablePane,
  Stack,
  Text,
  VerticalDivider,
} from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";

import { useRemoveTemplateLocalesMutation } from "./mutations/updateAppTemplatesMutation";
import ButtonWithLoading from "../../ButtonWithLoading";
import ErrorDialog from "../../error/ErrorDialog";
import { useSystemConfig } from "../../context/SystemConfigContext";
import { useCheckbox, useDropdown } from "../../hook/useInput";
import { getConfiguredLocales, TemplateLocale } from "../../templates";

import styles from "./TemplateLocaleManagement.module.scss";

type TemplateLocaleUpdater = (locale: TemplateLocale) => void;

interface TemplateLocaleManagementProps {
  resourcePaths: string[];
  templateLocale: TemplateLocale;
  defaultTemplateLocale: TemplateLocale;
  onTemplateLocaleSelected: TemplateLocaleUpdater;
  onDefaultTemplateLocaleSelected: TemplateLocaleUpdater;
  pendingTemplateLocales: TemplateLocale[];
  onPendingTemplateLocalesChange: (locales: TemplateLocale[]) => void;
}

interface TemplateLocaleManagementDialogProps {
  resourcePaths: string[];
  presented: boolean;
  onDismiss: () => void;
  configuredTemplateLocales: TemplateLocale[];
  pendingTemplateLocales: TemplateLocale[];
  onPendingTemplateLocalesChange: (locales: TemplateLocale[]) => void;
}

interface TemplateLocaleListItemProps {
  locale: TemplateLocale;
  onItemSelected: (locale: TemplateLocale, checked: boolean) => void;
}

interface SelectedTemplateLocaleItemProps {
  locale: TemplateLocale;
  onItemRemoved: (locale: TemplateLocale) => void;
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
  const { locale, onItemRemoved } = props;
  const { themes } = useSystemConfig();

  const onDeleteClicked = useCallback(() => {
    onItemRemoved(locale);
  }, [locale, onItemRemoved]);

  return (
    <div className={styles.dialogSelectedItem}>
      <Text>
        <FormattedMessage id={getLanguageLocaleKey(locale)} />
      </Text>
      <ActionButton
        iconProps={{ iconName: "Delete" }}
        theme={themes.destructive}
        onClick={onDeleteClicked}
      />
    </div>
  );
};

const TemplateLocaleManagementDialog: React.FC<TemplateLocaleManagementDialogProps> = function TemplateLocaleManagementDialog(
  props: TemplateLocaleManagementDialogProps
) {
  const {
    resourcePaths,
    presented,
    onDismiss,
    configuredTemplateLocales,
    pendingTemplateLocales,
    onPendingTemplateLocalesChange,
  } = props;

  const { supportedResourceLocales } = useSystemConfig();
  const { appID } = useParams();
  const { renderToString } = useContext(Context);

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
        />
      );
    },
    [onSelctedTemplateLocaleRemoved]
  );

  const onCancel = useCallback(() => {
    setSelectedLocales(initialSelectedLocales);
    onDismiss();
  }, [onDismiss, initialSelectedLocales]);

  const onApplyClick = useCallback(() => {
    if (selectedLocales.length === 0) {
      setLocalErrorMessage(
        renderToString(
          "TemplateLocaleManagementDialog.cannot-remove-last-language-error"
        )
      );
      return;
    }
    const selectedLocaleSet = new Set(selectedLocales);
    const configuredLocaleSet = new Set(configuredTemplateLocales);
    const removedLocales = configuredTemplateLocales.filter(
      (locale) => !selectedLocaleSet.has(locale)
    );
    // NOTE: cannot remove all configured locales
    if (removedLocales.length === configuredTemplateLocales.length) {
      setLocalErrorMessage(
        renderToString(
          "TemplateLocaleManagementDialog.cannot-remove-last-language-error"
        )
      );
      return;
    }

    const updatedPendingTemplateLocales = selectedLocales.filter(
      (locale) => !configuredLocaleSet.has(locale)
    );

    if (removedLocales.length > 0) {
      removeTemplateLocales(resourcePaths, removedLocales)
        .then(() => {
          onPendingTemplateLocalesChange(updatedPendingTemplateLocales);
          onDismiss();
        })
        .catch(() => {});
    } else {
      onPendingTemplateLocalesChange(updatedPendingTemplateLocales);
      onDismiss();
    }
  }, [
    renderToString,
    resourcePaths,
    selectedLocales,
    configuredTemplateLocales,
    onPendingTemplateLocalesChange,
    removeTemplateLocales,
    onDismiss,
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
    resourcePaths,
    templateLocale,
    defaultTemplateLocale,
    onTemplateLocaleSelected,
    onDefaultTemplateLocaleSelected,
    pendingTemplateLocales,
    onPendingTemplateLocalesChange,
  } = props;

  const { themes } = useSystemConfig();
  const { renderToString } = useContext(Context);

  const configuredTemplateLocales = useMemo(() => {
    return getConfiguredLocales(resourcePaths);
  }, [resourcePaths]);

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
      if (locale === defaultTemplateLocale) {
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
      defaultTemplateLocale,
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

  return (
    <section className={styles.templateLocaleManagement}>
      <TemplateLocaleManagementDialog
        resourcePaths={resourcePaths}
        presented={isDialogPresented}
        configuredTemplateLocales={configuredTemplateLocales}
        pendingTemplateLocales={pendingTemplateLocales}
        onPendingTemplateLocalesChange={onPendingTemplateLocalesChange}
        onDismiss={dismissDialog}
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
        />
        <Dropdown
          className={styles.dropdown}
          label={renderToString(
            "TemplatesConfigurationScreen.default-template-locale-dropdown"
          )}
          options={defaultTemplateLocaleOptions}
          onChange={onDefaultTemplateLocaleChange}
        />
        <ActionButton theme={themes.actionButton} onClick={presentDialog}>
          <FormattedMessage id="TemplatesConfigurationScreen.manage-template-locale" />
        </ActionButton>
      </Stack>
      <ButtonWithLoading
        loading={false}
        labelId="save"
        loadingLabelId="saving"
      />
    </section>
  );
};

export default TemplateLocaleManagement;
