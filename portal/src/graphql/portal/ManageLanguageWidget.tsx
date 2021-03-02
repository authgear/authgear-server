import React, { useCallback, useContext, useMemo, useState } from "react";
import {
  ActionButton,
  Checkbox,
  DefaultButton,
  Dialog,
  DialogFooter,
  DirectionalHint,
  Dropdown,
  IContextualMenuItem,
  IContextualMenuProps,
  IDialogProps,
  IDropdownOption,
  IListProps,
  ITooltipProps,
  List,
  PrimaryButton,
  Stack,
  Text,
  TooltipHost,
  VerticalDivider,
  IRenderFunction,
} from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";

import { useSystemConfig } from "../../context/SystemConfigContext";
import { useCheckbox } from "../../hook/useInput";
import { LanguageTag } from "../../util/resource";

import styles from "./ManageLanguageWidget.module.scss";

interface ManageLanguageWidgetProps {
  // The supported languages.
  supportedLanguages: LanguageTag[];
  onChangeSupportedLanguages: (newSupportedLanguages: LanguageTag[]) => void;

  // The selected language.
  selectedLanguage: LanguageTag;
  onChangeSelectedLanguage: (newSelectedLanguage: LanguageTag) => void;

  // The fallback language.
  fallbackLanguage: LanguageTag;
  onChangeFallbackLanguage: (newFallbackLanguage: LanguageTag) => void;
}

interface ManageLanguageWidgetDialogProps {
  presented: boolean;
  onDismiss: () => void;
  fallbackLanguage: LanguageTag;
  supportedLanguages: LanguageTag[];
  onChangeSupportedLanguages: (newSupportedLanguages: LanguageTag[]) => void;
}

interface TemplateLocaleListItemProps {
  language: LanguageTag;
  checked: boolean;
  onSelectItem: (locale: LanguageTag) => void;
}

interface SelectedTemplateLocaleItemProps {
  language: LanguageTag;
  onRemove: (locale: LanguageTag) => void;
  isFallbackLanguage: boolean;
}

const DIALOG_STYLES = {
  main: {
    maxWidth: "none !important",
    width: "80vw !important",
    minWidth: "500px !important",
  },
};

function getLanguageLocaleKey(locale: LanguageTag) {
  return `Locales.${locale}`;
}

const TemplateLocaleListItem: React.FC<TemplateLocaleListItemProps> = function TemplateLocaleListItem(
  props: TemplateLocaleListItemProps
) {
  const { language, checked, onSelectItem } = props;

  const { onChange } = useCheckbox(() => {
    onSelectItem(language);
  });

  return (
    <div className={styles.dialogLocaleListItem}>
      <Checkbox checked={checked} disabled={checked} onChange={onChange} />
      <Text className={styles.dialogLocaleListItemText}>
        <FormattedMessage id={getLanguageLocaleKey(language)} />
      </Text>
    </div>
  );
};

const SelectedTemplateLocaleItem: React.FC<SelectedTemplateLocaleItemProps> = function SelectedTemplateLocaleItem(
  props: SelectedTemplateLocaleItemProps
) {
  const { language, onRemove, isFallbackLanguage } = props;
  const { themes } = useSystemConfig();

  const tooltipProps: ITooltipProps = useMemo(() => {
    return {
      onRenderContent: () => (
        <div className={styles.tooltip}>
          <Text className={styles.tooltipMessage}>
            <FormattedMessage id="ManageLanguageWidget.warning.remove-default-language" />
          </Text>
        </div>
      ),
    };
  }, []);

  const onDeleteClicked = useCallback(() => {
    onRemove(language);
  }, [language, onRemove]);

  return (
    <div className={styles.dialogSelectedItem}>
      <Text>
        <FormattedMessage id={getLanguageLocaleKey(language)} />
      </Text>
      <TooltipHost
        hidden={!isFallbackLanguage}
        tooltipProps={tooltipProps}
        directionalHint={DirectionalHint.bottomCenter}
      >
        <ActionButton
          iconProps={{ iconName: "Delete" }}
          theme={themes.destructive}
          onClick={onDeleteClicked}
          disabled={isFallbackLanguage}
        />
      </TooltipHost>
    </div>
  );
};

interface TemplateLocaleListItemProps {
  language: LanguageTag;
  checked: boolean;
  onSelectItem: (locale: LanguageTag) => void;
}

const ManageLanguageWidgetDialog: React.FC<ManageLanguageWidgetDialogProps> = function ManageLanguageWidgetDialog(
  props: ManageLanguageWidgetDialogProps
) {
  const {
    presented,
    onDismiss,
    fallbackLanguage,
    supportedLanguages,
    onChangeSupportedLanguages,
  } = props;

  const { supportedResourceLocales } = useSystemConfig();

  const [newSupportedLanguages, setNewSupportedLanguages] = useState<
    LanguageTag[]
  >(supportedLanguages);

  const onAddTemplateLocale = useCallback((locale: LanguageTag) => {
    setNewSupportedLanguages((prev) => {
      const idx = prev.findIndex((item) => item === locale);
      // Already present
      if (idx >= 0) {
        return prev;
      }
      return [...prev, locale];
    });
  }, []);

  const listItems = useMemo(() => {
    const items: TemplateLocaleListItemProps[] = [];
    for (const language of supportedResourceLocales) {
      items.push({
        language,
        checked: newSupportedLanguages.includes(language),
        onSelectItem: onAddTemplateLocale,
      });
    }
    return items;
  }, [onAddTemplateLocale, supportedResourceLocales, newSupportedLanguages]);

  const onRemoveTemplateLocale = useCallback((locale: LanguageTag) => {
    setNewSupportedLanguages((prev) => {
      return prev.filter((item) => item !== locale);
    });
  }, []);

  const renderLocaleListItemCell = useCallback<
    Required<IListProps<TemplateLocaleListItemProps>>["onRenderCell"]
  >((item?: TemplateLocaleListItemProps) => {
    if (item == null) {
      return null;
    }
    const { language, checked, onSelectItem } = item;
    return (
      <TemplateLocaleListItem
        language={language}
        checked={checked}
        onSelectItem={onSelectItem}
      />
    );
  }, []);

  const renderSelectedLocaleItemCell = useCallback<
    Required<IListProps<LanguageTag>>["onRenderCell"]
  >(
    (language) => {
      if (language == null) {
        return null;
      }
      return (
        <SelectedTemplateLocaleItem
          language={language}
          onRemove={onRemoveTemplateLocale}
          isFallbackLanguage={language === fallbackLanguage}
        />
      );
    },
    [onRemoveTemplateLocale, fallbackLanguage]
  );

  const onCancel = useCallback(() => {
    onDismiss();
  }, [onDismiss]);

  const onApplyClick = useCallback(() => {
    onChangeSupportedLanguages(newSupportedLanguages);
    onDismiss();
  }, [onChangeSupportedLanguages, newSupportedLanguages, onDismiss]);

  const modalProps = useMemo<IDialogProps["modalProps"]>(() => {
    return {
      isBlocking: true,
      topOffsetFixed: true,
      onDismissed: () => {
        setNewSupportedLanguages(supportedLanguages);
      },
    };
  }, [supportedLanguages]);

  return (
    <Dialog
      hidden={!presented}
      onDismiss={onCancel}
      title={
        <FormattedMessage id="ManageLanguageWidget.add-or-remove-languages" />
      }
      modalProps={modalProps}
      styles={DIALOG_STYLES}
    >
      <Text className={styles.dialogDesc}>
        <FormattedMessage id="ManageLanguageWidget.default-language-description" />
      </Text>
      <div className={styles.dialogContent}>
        <section className={styles.dialogColumn}>
          <Text className={styles.dialogColumnHeader}>
            <FormattedMessage id="ManageLanguageWidget.all-languages" />
          </Text>
          <section className={styles.dialogListWrapper}>
            <List items={listItems} onRenderCell={renderLocaleListItemCell} />
          </section>
        </section>
        <VerticalDivider className={styles.dialogDivider} />
        <section className={styles.dialogColumn}>
          <Text className={styles.dialogColumnHeader}>
            <FormattedMessage id="ManageLanguageWidget.app-languages" />
          </Text>
          <section className={styles.dialogListWrapper}>
            <List
              items={newSupportedLanguages}
              onRenderCell={renderSelectedLocaleItemCell}
            />
          </section>
        </section>
      </div>
      <DialogFooter>
        <DefaultButton onClick={onCancel}>
          <FormattedMessage id="cancel" />
        </DefaultButton>
        <PrimaryButton onClick={onApplyClick}>
          <FormattedMessage id="apply" />
        </PrimaryButton>
      </DialogFooter>
    </Dialog>
  );
};

const ManageLanguageWidget: React.FC<ManageLanguageWidgetProps> = function ManageLanguageWidget(
  props: ManageLanguageWidgetProps
) {
  const {
    supportedLanguages,
    onChangeSupportedLanguages,

    selectedLanguage,
    onChangeSelectedLanguage,

    fallbackLanguage,
    onChangeFallbackLanguage,
  } = props;

  const { themes } = useSystemConfig();

  const { renderToString } = useContext(Context);

  const [isDialogPresented, setIsDialogPresented] = useState(false);

  const displayTemplateLocale = useCallback(
    (locale: LanguageTag) => {
      return renderToString(getLanguageLocaleKey(locale));
    },
    [renderToString]
  );

  const presentDialog = useCallback(() => {
    setIsDialogPresented(true);
  }, []);

  const dismissDialog = useCallback(() => {
    setIsDialogPresented(false);
  }, []);

  const subMenuPropsItems: IContextualMenuItem[] = [];
  for (const lang of supportedLanguages) {
    subMenuPropsItems.push({
      key: lang,
      text: displayTemplateLocale(lang),
      iconProps:
        lang === fallbackLanguage
          ? {
              iconName: "CheckMark",
            }
          : undefined,
      onClick: (
        e?: React.SyntheticEvent<HTMLElement>,
        item?: IContextualMenuItem
      ) => {
        // Do not prevent default so that the menu is dismissed.
        // e?.preventDefault();
        e?.stopPropagation();
        if (item != null) {
          onChangeFallbackLanguage(item.key);
        }
      },
    });
  }

  const menuItems: IContextualMenuItem[] = [
    {
      key: "change-default-language",
      text: renderToString("ManageLanguageWidget.change-default-language"),
      subMenuProps: {
        items: subMenuPropsItems,
      },
    },
    {
      key: "add-or-remove-languages",
      text: renderToString("ManageLanguageWidget.add-or-remove-languages"),
      onClick: (e?: React.SyntheticEvent<HTMLElement>) => {
        e?.preventDefault();
        e?.stopPropagation();
        presentDialog();
      },
    },
  ];

  const menuProps: IContextualMenuProps = {
    items: menuItems,
  };

  const templateLocaleOptions = useMemo(() => {
    const options = [];
    for (const locale of supportedLanguages) {
      options.push({
        key: locale,
        text: displayTemplateLocale(locale),
        data: {
          invalid: false,
        },
      });
    }

    return options;
  }, [supportedLanguages, displayTemplateLocale]);

  const onChangeTemplateLocale = useCallback(
    (_e: unknown, option?: IDropdownOption) => {
      if (option != null) {
        onChangeSelectedLanguage(option.key.toString());
      }
    },
    [onChangeSelectedLanguage]
  );

  const onRenderOption: IRenderFunction<IDropdownOption> = useCallback(
    (option?: IDropdownOption) => {
      const invalid = option?.data?.invalid === true;
      return (
        <Text
          styles={
            invalid
              ? {
                  root: {
                    color: themes.main.semanticColors.errorText,
                  },
                }
              : undefined
          }
        >
          {option?.text ?? ""}
        </Text>
      );
    },
    [themes]
  );

  const onRenderTitle: IRenderFunction<IDropdownOption[]> = useCallback(
    (options?: IDropdownOption[]) => {
      const option = options?.[0];
      const invalid = option?.data?.invalid === true;
      return (
        <Text
          styles={
            invalid
              ? {
                  root: {
                    color: themes.main.semanticColors.errorText,
                  },
                }
              : undefined
          }
        >
          {option?.text ?? ""}
        </Text>
      );
    },
    [themes]
  );

  return (
    <section className={styles.templateLocaleManagement}>
      <ManageLanguageWidgetDialog
        presented={isDialogPresented}
        onDismiss={dismissDialog}
        supportedLanguages={supportedLanguages}
        fallbackLanguage={fallbackLanguage}
        onChangeSupportedLanguages={onChangeSupportedLanguages}
      />
      <Stack
        className={styles.inputContainer}
        verticalAlign="start"
        horizontal={true}
        tokens={{ childrenGap: 10 }}
      >
        <Dropdown
          className={styles.dropdown}
          label={renderToString("ManageLanguageWidget.title")}
          options={templateLocaleOptions}
          onChange={onChangeTemplateLocale}
          selectedKey={selectedLanguage}
          onRenderTitle={onRenderTitle}
          onRenderOption={onRenderOption}
        />
        <DefaultButton className={styles.contextualMenu} menuProps={menuProps}>
          <FormattedMessage id="ManageLanguageWidget.manage-languages" />
        </DefaultButton>
      </Stack>
    </section>
  );
};

export default ManageLanguageWidget;
