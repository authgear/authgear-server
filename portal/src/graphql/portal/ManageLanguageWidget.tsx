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
import { TemplateLocale } from "../../templates";

import styles from "./ManageLanguageWidget.module.scss";

type TemplateLocaleUpdater = (locale: TemplateLocale) => void;

interface ManageLanguageWidgetProps {
  // The list of languages.
  templateLocales: TemplateLocale[];
  onChangeTemplateLocales: (locales: TemplateLocale[]) => void;

  // The selected language.
  templateLocale: TemplateLocale;
  onSelectTemplateLocale: TemplateLocaleUpdater;

  // The default language.
  defaultTemplateLocale: TemplateLocale;
  onSelectDefaultTemplateLocale: TemplateLocaleUpdater;

  invalidTemplateLocales: TemplateLocale[];
}

interface ManageLanguageWidgetDialogProps {
  presented: boolean;
  onDismiss: () => void;
  defaultTemplateLocale: TemplateLocale;
  templateLocales: TemplateLocale[];
  onChangeTemplateLocales: (locales: TemplateLocale[]) => void;
}

interface TemplateLocaleListItemProps {
  locale: TemplateLocale;
  checked: boolean;
  onSelectItem: (locale: TemplateLocale) => void;
}

interface SelectedTemplateLocaleItemProps {
  locale: TemplateLocale;
  onRemove: (locale: TemplateLocale) => void;
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
  const { locale, checked, onSelectItem } = props;

  const { onChange } = useCheckbox(() => {
    onSelectItem(locale);
  });

  return (
    <div className={styles.dialogLocaleListItem}>
      <Checkbox checked={checked} disabled={checked} onChange={onChange} />
      <Text className={styles.dialogLocaleListItemText}>
        <FormattedMessage id={getLanguageLocaleKey(locale)} />
      </Text>
    </div>
  );
};

const SelectedTemplateLocaleItem: React.FC<SelectedTemplateLocaleItemProps> = function SelectedTemplateLocaleItem(
  props: SelectedTemplateLocaleItemProps
) {
  const { locale, onRemove, isDefaultLocale } = props;
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
    onRemove(locale);
  }, [locale, onRemove]);

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

interface TemplateLocaleListItemProps {
  locale: TemplateLocale;
  checked: boolean;
  onSelectItem: (locale: TemplateLocale) => void;
}

const ManageLanguageWidgetDialog: React.FC<ManageLanguageWidgetDialogProps> = function ManageLanguageWidgetDialog(
  props: ManageLanguageWidgetDialogProps
) {
  const {
    presented,
    onDismiss,
    defaultTemplateLocale,
    templateLocales,
    onChangeTemplateLocales,
  } = props;

  const { supportedResourceLocales } = useSystemConfig();

  const [newLocales, setNewLocales] = useState<TemplateLocale[]>(
    templateLocales
  );

  const onAddTemplateLocale = useCallback((locale: TemplateLocale) => {
    setNewLocales((prev) => {
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
    for (const locale of supportedResourceLocales) {
      items.push({
        locale,
        checked:
          templateLocales.includes(locale) || newLocales.includes(locale),
        onSelectItem: onAddTemplateLocale,
      });
    }
    return items;
  }, [
    onAddTemplateLocale,
    supportedResourceLocales,
    templateLocales,
    newLocales,
  ]);

  const onRemoveTemplateLocale = useCallback((locale: TemplateLocale) => {
    setNewLocales((prev) => {
      return prev.filter((item) => item !== locale);
    });
  }, []);

  const renderLocaleListItemCell = useCallback<
    Required<IListProps<TemplateLocaleListItemProps>>["onRenderCell"]
  >((item?: TemplateLocaleListItemProps) => {
    if (item == null) {
      return null;
    }
    const { locale, checked, onSelectItem } = item;
    return (
      <TemplateLocaleListItem
        locale={locale}
        checked={checked}
        onSelectItem={onSelectItem}
      />
    );
  }, []);

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
          onRemove={onRemoveTemplateLocale}
          isDefaultLocale={locale === defaultTemplateLocale}
        />
      );
    },
    [onRemoveTemplateLocale, defaultTemplateLocale]
  );

  const onCancel = useCallback(() => {
    setNewLocales(templateLocales);
    onDismiss();
  }, [onDismiss, templateLocales]);

  const onApplyClick = useCallback(() => {
    onChangeTemplateLocales(newLocales);
    onDismiss();
  }, [onChangeTemplateLocales, newLocales, onDismiss]);

  const modalProps = useMemo<IDialogProps["modalProps"]>(() => {
    return {
      isBlocking: true,
      topOffsetFixed: true,
    };
  }, []);

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
              items={newLocales}
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
    templateLocales,
    onChangeTemplateLocales,
    templateLocale,
    onSelectTemplateLocale,
    defaultTemplateLocale,
    onSelectDefaultTemplateLocale,
    invalidTemplateLocales,
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

  const presentDialog = useCallback(() => {
    setIsDialogPresented(true);
  }, []);

  const dismissDialog = useCallback(() => {
    setIsDialogPresented(false);
  }, []);

  const subMenuPropsItems: IContextualMenuItem[] = [];
  for (const lang of templateLocales) {
    subMenuPropsItems.push({
      key: lang,
      text: displayTemplateLocale(lang),
      iconProps:
        lang === defaultTemplateLocale
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
          onSelectDefaultTemplateLocale(item.key);
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
    for (const locale of templateLocales) {
      options.push({
        key: locale,
        text: displayTemplateLocale(locale),
        data: {
          invalid: invalidTemplateLocales.includes(locale),
        },
      });
    }

    // Handle the case that selected locale is not in templateLocales.
    // This happens when the default locale does not have templates.
    if (!templateLocales.includes(templateLocale)) {
      options.push({
        key: templateLocale,
        text: displayTemplateLocale(templateLocale),
        hidden: true,
      });
    }

    return options;
  }, [
    templateLocales,
    templateLocale,
    displayTemplateLocale,
    invalidTemplateLocales,
  ]);

  const onChangeTemplateLocale = useCallback(
    (_e: unknown, option?: IDropdownOption) => {
      if (option != null) {
        onSelectTemplateLocale(option.key.toString());
      }
    },
    [onSelectTemplateLocale]
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
        templateLocales={templateLocales}
        defaultTemplateLocale={defaultTemplateLocale}
        onChangeTemplateLocales={onChangeTemplateLocales}
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
          selectedKey={templateLocale}
          onRenderTitle={onRenderTitle}
          onRenderOption={onRenderOption}
          errorMessage={
            invalidTemplateLocales.length > 0
              ? renderToString("ManageLanguageWidget.error.invalid-language")
              : undefined
          }
        />
        <DefaultButton className={styles.contextualMenu} menuProps={menuProps}>
          <FormattedMessage id="ManageLanguageWidget.manage-languages" />
        </DefaultButton>
      </Stack>
    </section>
  );
};

export default ManageLanguageWidget;
