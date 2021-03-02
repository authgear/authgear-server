import React, {
  useCallback,
  useContext,
  useMemo,
  useState,
  useEffect,
} from "react";
import cn from "classnames";
import {
  Checkbox,
  Label,
  IconButton,
  DefaultButton,
  SearchBox,
  Link,
  Dialog,
  DialogFooter,
  Dropdown,
  IDialogProps,
  IDropdownOption,
  IListProps,
  List,
  PrimaryButton,
  Text,
  IRenderFunction,
} from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";

import { useSystemConfig } from "../../context/SystemConfigContext";
import { LanguageTag } from "../../util/resource";
import { useExactKeywordSearch } from "../../util/search";

import styles from "./ManageLanguageWidget.module.scss";

interface ManageLanguageWidgetProps {
  className?: string;
  selectOnly: boolean;

  // The supported languages.
  supportedLanguages: LanguageTag[];

  // The selected language.
  selectedLanguage: LanguageTag;
  onChangeSelectedLanguage: (newSelectedLanguage: LanguageTag) => void;

  // The fallback language.
  fallbackLanguage: LanguageTag;

  onChangeLanguages?: (
    supportedLanguages: LanguageTag[],
    fallbackLanguage: LanguageTag
  ) => void;
}

interface ManageLanguageWidgetDialogProps {
  presented: boolean;
  onDismiss: () => void;
  fallbackLanguage: LanguageTag;
  supportedLanguages: LanguageTag[];
  onChangeLanguages?: (
    supportedLanguages: LanguageTag[],
    fallbackLanguage: LanguageTag
  ) => void;
}

interface CellProps {
  language: LanguageTag;
  fallbackLanguage: LanguageTag;
  checked: boolean;
  onChange: (locale: LanguageTag) => void;
  onClickSetAsFallback: (locale: LanguageTag) => void;
}

const DIALOG_STYLES = {
  main: {
    minWidth: "533px !important",
  },
};

function getLanguageLocaleKey(locale: LanguageTag) {
  return `Locales.${locale}`;
}

const Cell: React.FC<CellProps> = function Cell(props: CellProps) {
  const {
    language,
    checked,
    fallbackLanguage,
    onChange: onChangeProp,
    onClickSetAsFallback: onClickSetAsFallbackProp,
  } = props;
  const disabled = language === fallbackLanguage;

  const { renderToString } = useContext(Context);

  const onChange = useCallback(() => {
    onChangeProp(language);
  }, [language, onChangeProp]);

  const onClickSetAsFallback = useCallback(() => {
    onClickSetAsFallbackProp(language);
  }, [language, onClickSetAsFallbackProp]);

  return (
    <div className={styles.cellRoot}>
      <Checkbox
        className={styles.cellCheckbox}
        checked={checked}
        disabled={disabled}
        onChange={onChange}
      />
      <Text className={styles.cellText}>
        <FormattedMessage
          id="ManageLanguageWidget.language-label"
          values={{
            LANG: renderToString(getLanguageLocaleKey(language)),
            IS_FALLBACK: String(fallbackLanguage === language),
          }}
        />
      </Text>
      {checked && !disabled && (
        <Link onClick={onClickSetAsFallback}>
          <FormattedMessage id="ManageLanguageWidget.set-as-default" />
        </Link>
      )}
    </div>
  );
};

const ManageLanguageWidgetDialog: React.FC<ManageLanguageWidgetDialogProps> = function ManageLanguageWidgetDialog(
  props: ManageLanguageWidgetDialogProps
) {
  const {
    presented,
    onDismiss,
    fallbackLanguage,
    supportedLanguages,
    onChangeLanguages,
  } = props;

  const { renderToString } = useContext(Context);

  const { supportedResourceLocales } = useSystemConfig();

  const originalItems = useMemo(() => {
    return supportedResourceLocales.map((a) => {
      return {
        language: a,
        text: renderToString(getLanguageLocaleKey(a)),
      };
    });
  }, [supportedResourceLocales, renderToString]);

  const [newSupportedLanguages, setNewSupportedLanguages] = useState<
    LanguageTag[]
  >(supportedLanguages);

  const [newFallbackLanguage, setNewFallbackLanguage] = useState<LanguageTag>(
    fallbackLanguage
  );

  const [searchString, setSearchString] = useState<string>("");
  const { search } = useExactKeywordSearch(originalItems, ["text"]);
  const filteredItems = useMemo(() => {
    return search(searchString);
  }, [search, searchString]);

  const onSearch = useCallback((_e, value?: string) => {
    if (value == null) {
      return;
    }
    setSearchString(value);
  }, []);

  useEffect(() => {
    if (presented) {
      setNewSupportedLanguages(supportedLanguages);
      setNewFallbackLanguage(fallbackLanguage);
      setSearchString("");
    }
  }, [presented, supportedLanguages, fallbackLanguage]);

  const onToggleLanguage = useCallback((locale: LanguageTag) => {
    setNewSupportedLanguages((prev) => {
      const idx = prev.findIndex((item) => item === locale);
      if (idx >= 0) {
        return prev.filter((item) => item !== locale);
      }
      return [...prev, locale];
    });
  }, []);

  const onClickSetAsFallback = useCallback((locale: LanguageTag) => {
    setNewFallbackLanguage(locale);
  }, []);

  const listItems = useMemo(() => {
    const items: CellProps[] = [];
    for (const listItem of filteredItems) {
      const { language } = listItem;
      items.push({
        language,
        checked: newSupportedLanguages.includes(language),
        fallbackLanguage: newFallbackLanguage,
        onChange: onToggleLanguage,
        onClickSetAsFallback,
      });
    }
    return items;
  }, [
    onToggleLanguage,
    onClickSetAsFallback,
    newSupportedLanguages,
    newFallbackLanguage,
    filteredItems,
  ]);

  const renderLocaleListItemCell = useCallback<
    Required<IListProps<CellProps>>["onRenderCell"]
  >((item?: CellProps) => {
    if (item == null) {
      return null;
    }
    return <Cell {...item} />;
  }, []);

  const onCancel = useCallback(() => {
    onDismiss();
  }, [onDismiss]);

  const onApplyClick = useCallback(() => {
    onChangeLanguages?.(newSupportedLanguages, newFallbackLanguage);
    onDismiss();
  }, [
    onChangeLanguages,
    newSupportedLanguages,
    newFallbackLanguage,
    onDismiss,
  ]);

  const modalProps = useMemo<IDialogProps["modalProps"]>(() => {
    return {
      isBlocking: true,
      topOffsetFixed: true,
      onDismissed: () => {
        setNewSupportedLanguages(supportedLanguages);
        setNewFallbackLanguage(fallbackLanguage);
      },
    };
  }, [supportedLanguages, fallbackLanguage]);

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
      <SearchBox
        className={styles.searchBox}
        placeholder={renderToString("search")}
        onChange={onSearch}
      />
      <Text variant="small" className={styles.dialogColumnHeader}>
        <FormattedMessage id="ManageLanguageWidget.languages" />
      </Text>
      <div className={styles.dialogListWrapper}>
        <List items={listItems} onRenderCell={renderLocaleListItemCell} />
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
    selectedLanguage,
    onChangeSelectedLanguage,
    fallbackLanguage,
    onChangeLanguages,
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
            id="ManageLanguageWidget.language-label"
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
            id="ManageLanguageWidget.language-label"
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
        onChangeLanguages={onChangeLanguages}
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
