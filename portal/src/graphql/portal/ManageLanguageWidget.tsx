import React, { useCallback, useContext, useMemo, useState } from "react";
import cn from "classnames";
import {
  ActionButton,
  Checkbox,
  Label,
  IconButton,
  DefaultButton,
  Dialog,
  DialogFooter,
  DirectionalHint,
  Dropdown,
  IDialogProps,
  IDropdownOption,
  IListProps,
  ITooltipProps,
  List,
  PrimaryButton,
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
  className?: string;
  selectOnly: boolean;

  // The supported languages.
  supportedLanguages: LanguageTag[];
  onChangeSupportedLanguages?: (newSupportedLanguages: LanguageTag[]) => void;

  // The selected language.
  selectedLanguage: LanguageTag;
  onChangeSelectedLanguage: (newSelectedLanguage: LanguageTag) => void;

  // The fallback language.
  fallbackLanguage: LanguageTag;
  onChangeFallbackLanguage?: (newFallbackLanguage: LanguageTag) => void;
}

interface ManageLanguageWidgetDialogProps {
  presented: boolean;
  onDismiss: () => void;
  fallbackLanguage: LanguageTag;
  supportedLanguages: LanguageTag[];
  onChangeSupportedLanguages?: (newSupportedLanguages: LanguageTag[]) => void;
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
    onChangeSupportedLanguages?.(newSupportedLanguages);
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
    className,
    selectOnly,
    supportedLanguages,
    onChangeSupportedLanguages,
    selectedLanguage,
    onChangeSelectedLanguage,
    fallbackLanguage,
  } = props;

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

  const templateLocaleOptions = useMemo(() => {
    const options = [];
    for (const locale of supportedLanguages) {
      options.push({
        key: locale,
        text: displayTemplateLocale(locale),
        data: {
          isFallbackLanguage: fallbackLanguage === locale,
        },
      });
    }

    return options;
  }, [supportedLanguages, displayTemplateLocale, fallbackLanguage]);

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
      return (
        <Text>
          <FormattedMessage
            id="ManageLanguageWidget.dropdown-title"
            values={{
              LANG: option?.text ?? "",
              IS_FALLBACK: String(option?.data.isFallbackLanguage ?? false),
            }}
          />
        </Text>
      );
    },
    []
  );

  const onRenderTitle: IRenderFunction<IDropdownOption[]> = useCallback(
    (options?: IDropdownOption[]) => {
      const option = options?.[0];
      return (
        <Text>
          <FormattedMessage
            id="ManageLanguageWidget.dropdown-title"
            values={{
              LANG: option?.text ?? "",
              IS_FALLBACK: String(option?.data.isFallbackLanguage ?? false),
            }}
          />
        </Text>
      );
    },
    []
  );

  return (
    <>
      <ManageLanguageWidgetDialog
        presented={isDialogPresented}
        onDismiss={dismissDialog}
        supportedLanguages={supportedLanguages}
        fallbackLanguage={fallbackLanguage}
        onChangeSupportedLanguages={onChangeSupportedLanguages}
      />
      <div className={cn(className, styles.root)}>
        <Label className={styles.titleLabel}>
          <FormattedMessage id="ManageLanguageWidget.title" />
        </Label>
        <Dropdown
          id="language-widget"
          className={styles.dropdown}
          options={templateLocaleOptions}
          onChange={onChangeTemplateLocale}
          selectedKey={selectedLanguage}
          onRenderTitle={onRenderTitle}
          onRenderOption={onRenderOption}
        />
        {!selectOnly && (
          <IconButton
            iconProps={{
              iconName: "Settings",
            }}
            onClick={presentDialog}
          />
        )}
      </div>
    </>
  );
};

export default ManageLanguageWidget;
