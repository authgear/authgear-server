import React, {
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import { useParams } from "react-router-dom";
import {
  DropdownMenu,
  IconButton as RadixIconButton,
  Select,
  Text,
} from "@radix-ui/themes";
import {
  DotsVerticalIcon,
  PlusIcon,
  TrashIcon,
} from "@radix-ui/react-icons";
import { produce } from "immer";
import cn from "classnames";
import { Context as MFContext, FormattedMessage } from "../../intl";

import FormContainer from "../../FormContainer";
import ScreenContent from "../../ScreenContent";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import { FormField } from "../../components/v2/FormField/FormField";
import { SaveFunctionBar } from "../../components/v2/SaveFunctionBar/SaveFunctionBar";
import { SettingsSectionCard } from "../../components/v2/SettingsSectionCard/SettingsSectionCard";
import { Badge } from "../../components/v2/Badge/Badge";
import { Tooltip } from "../../components/v2/Tooltip/Tooltip";
import {
  TextField,
  TextFieldIcon,
} from "../../components/v2/TextField/TextField";
import { useFormContainerBaseContext } from "../../FormContainerBase";

import { PortalAPIAppConfig } from "../../types";
import { clearEmptyObject } from "../../util/misc";
import {
  AppConfigFormModel,
  useAppConfigForm,
} from "../../hook/useAppConfigForm";
import { useSystemConfig } from "../../context/SystemConfigContext";

import { LanguageTag } from "../../util/resource";

import styles from "./LanguagesConfigurationScreen.module.css";

interface PageContextValue {
  getLanguageDisplayText: (lang: LanguageTag) => string;
}
const PageContext = React.createContext<PageContextValue>(null as any);

interface ConfigFormState {
  supportedLanguages: string[];
  fallbackLanguage: string;
}

function constructFormState(config: PortalAPIAppConfig): ConfigFormState {
  const fallbackLanguage = config.localization?.fallback_language ?? "en";
  return {
    fallbackLanguage,
    supportedLanguages: config.localization?.supported_languages ?? [
      fallbackLanguage,
    ],
  };
}

function constructConfig(
  config: PortalAPIAppConfig,
  _initialState: ConfigFormState,
  currentState: ConfigFormState
): PortalAPIAppConfig {
  return produce(config, (config) => {
    config.localization = config.localization ?? {};
    config.localization.fallback_language = currentState.fallbackLanguage;
    config.localization.supported_languages = currentState.supportedLanguages;
    clearEmptyObject(config);
  });
}

function toggleSupportedLanguage(
  language: LanguageTag,
  checked: boolean
): (s: ConfigFormState) => ConfigFormState {
  return (state) => {
    if (checked) {
      if (state.supportedLanguages.includes(language)) {
        return state;
      }
      // Append newly added languages to the end so the table grows downward.
      return {
        ...state,
        supportedLanguages: [...state.supportedLanguages, language],
      };
    }
    return {
      ...state,
      supportedLanguages: state.supportedLanguages.filter(
        (lang) => lang !== language
      ),
    };
  };
}

function updatePrimaryLanguage(
  primaryLanguage: LanguageTag
): (s: ConfigFormState) => ConfigFormState {
  return (state) => {
    return toggleSupportedLanguage(
      primaryLanguage,
      true
    )({ ...state, fallbackLanguage: primaryLanguage });
  };
}

interface SelectPrimaryLanguageSectionProps {
  supportedLanguages: LanguageTag[];
  primaryLanguage: LanguageTag;
  onChangePrimaryLanguage: (language: LanguageTag) => void;
}
const SelectPrimaryLanguageSection: React.VFC<SelectPrimaryLanguageSectionProps> =
  function SelectPrimaryLanguageSection(props) {
    const {
      supportedLanguages,
      primaryLanguage,
      onChangePrimaryLanguage,
    } = props;

    const { getLanguageDisplayText } = useContext(PageContext);

    const options = useMemo(() => {
      return supportedLanguages.map((lang) => ({
        value: lang,
        label: getLanguageDisplayText(lang),
      }));
    }, [supportedLanguages, getLanguageDisplayText]);

    return (
      <SettingsSectionCard
        className={styles.widget}
        contentClassName="gap-4"
        title={
          <FormattedMessage id="LanguagesConfigurationScreen.selectPrimaryLanguageWidget.title" />
        }
      >
        <Text as="p" size="2" color="gray" className={styles.sectionDescription}>
          <FormattedMessage id="LanguagesConfigurationScreen.selectPrimaryLanguageWidget.description" />
        </Text>
        <FormField
          size="2"
          labelSize="2"
          labelSpace="1"
          label={
            <FormattedMessage id="LanguagesConfigurationScreen.selectPrimaryLanguageWidget.dropdown.label" />
          }
        >
          <Select.Root
            value={primaryLanguage}
            onValueChange={onChangePrimaryLanguage}
          >
            <Select.Trigger
              variant="surface"
              className={styles.languageSelectTrigger}
            />
            <Select.Content>
              {options.map((opt) => (
                <Select.Item key={opt.value} value={opt.value}>
                  {opt.label}
                </Select.Item>
              ))}
            </Select.Content>
          </Select.Root>
        </FormField>
      </SettingsSectionCard>
    );
  };

interface SupportedLanguagesListProps {
  primaryLanguage: LanguageTag;
  builtinLanguages: LanguageTag[];
  availableLanguages: LanguageTag[];
  supportedLanguages: LanguageTag[];
  onToggleSupportedLanguage: (lang: LanguageTag, selected: boolean) => void;
}
const SupportedLanguagesList: React.VFC<SupportedLanguagesListProps> =
  function SupportedLanguagesList(props) {
    const {
      primaryLanguage,
      builtinLanguages,
      availableLanguages,
      supportedLanguages,
      onToggleSupportedLanguage,
    } = props;
    const { renderToString } = useContext(MFContext);
    const { getLanguageDisplayText } = useContext(PageContext);
    const addLanguageControlRef = useRef<HTMLDivElement>(null);
    const searchInputContainerRef = useRef<HTMLDivElement>(null);

    const [searchKeyword, setSearchKeyword] = useState("");
    const [isAddingLanguage, setIsAddingLanguage] = useState(false);

    const builtinLanguageSet = useMemo(
      () => new Set(builtinLanguages),
      [builtinLanguages]
    );
    const availableLanguageSet = useMemo(
      () => new Set(availableLanguages),
      [availableLanguages]
    );
    const supportedLanguageSet = useMemo(
      () => new Set(supportedLanguages),
      [supportedLanguages]
    );

    const selectedLanguages = useMemo(() => {
      const languages = supportedLanguages.filter((lang) =>
        availableLanguageSet.has(lang)
      );
      const english = languages.filter((lang) => lang === "en");
      const rest = languages
        .filter((lang) => lang !== "en")
        .sort((a, b) =>
          getLanguageDisplayText(a).localeCompare(getLanguageDisplayText(b))
        );
      return [...english, ...rest];
    }, [supportedLanguages, availableLanguageSet, getLanguageDisplayText]);

    const candidateLanguages = useMemo(() => {
      const keyword = searchKeyword.trim().toLowerCase();
      return availableLanguages
        .filter((lang) => !supportedLanguageSet.has(lang))
        .filter((lang) => {
          if (keyword === "") {
            return true;
          }
          const label = getLanguageDisplayText(lang).toLowerCase();
          return (
            label.includes(keyword) || lang.toLowerCase().includes(keyword)
          );
        });
    }, [
      availableLanguages,
      supportedLanguageSet,
      searchKeyword,
      getLanguageDisplayText,
    ]);

    const onChangeSearchKeyword = useCallback(
      (e: React.ChangeEvent<HTMLInputElement>) => {
        setSearchKeyword(e.currentTarget.value);
      },
      []
    );

    const onCloseAddLanguages = useCallback(() => {
      setIsAddingLanguage(false);
      setSearchKeyword("");
    }, []);

    const onOpenAddLanguages = useCallback(() => {
      setIsAddingLanguage(true);
      window.setTimeout(() => {
        const input =
          searchInputContainerRef.current?.querySelector("input");
        input?.focus();
      }, 0);
    }, []);

    const onSelectLanguage = useCallback(
      (lang: LanguageTag) => {
        onToggleSupportedLanguage(lang, true);
        onCloseAddLanguages();
      },
      [onToggleSupportedLanguage, onCloseAddLanguages]
    );

    const onRemoveLanguage = useCallback(
      (lang: LanguageTag) => {
        if (lang === primaryLanguage || lang === "en") {
          return;
        }
        onToggleSupportedLanguage(lang, false);
      },
      [onToggleSupportedLanguage, primaryLanguage]
    );

    const allLanguagesAdded = availableLanguages.every((lang) =>
      supportedLanguageSet.has(lang)
    );

    const showSearch = isAddingLanguage && !allLanguagesAdded;

    useEffect(() => {
      if (!showSearch) {
        return;
      }
      const onPointerDown = (event: PointerEvent) => {
        const root = addLanguageControlRef.current;
        if (root != null && !root.contains(event.target as Node)) {
          onCloseAddLanguages();
        }
      };
      document.addEventListener("pointerdown", onPointerDown);
      return () => {
        document.removeEventListener("pointerdown", onPointerDown);
      };
    }, [showSearch, onCloseAddLanguages]);

    return (
      <div className={styles.supportedLanguagesList}>
        <Text as="p" size="2" color="gray" className={styles.sectionDescription}>
          <FormattedMessage id="LanguagesConfigurationScreen.supportedLanguages.description" />
        </Text>

        <div className={styles.addLanguageToolbar}>
          <div
            ref={addLanguageControlRef}
            className={styles.addLanguageControl}
          >
            <div
              className={cn(
                showSearch && styles.addLanguageButtonWrapHidden
              )}
            >
              <button
                type="button"
                className={styles.addLanguageTrigger}
                disabled={allLanguagesAdded}
                onClick={onOpenAddLanguages}
              >
                <span className={styles.addLanguageButtonContent}>
                  <PlusIcon width="1rem" height="1rem" />
                  <FormattedMessage id="LanguagesConfigurationScreen.supportedLanguages.add-languages" />
                </span>
              </button>
            </div>
            {showSearch ? (
              <>
                <div
                  ref={searchInputContainerRef}
                  className={styles.addLanguageSearch}
                >
                  <TextField
                    size="2"
                    type="search"
                    value={searchKeyword}
                    placeholder={renderToString(
                      "LanguagesConfigurationScreen.supportedLanguages.search.placeholder"
                    )}
                    iconStart={TextFieldIcon.MagnifyingGlass}
                    onChange={onChangeSearchKeyword}
                  />
                </div>
                <div className={styles.addLanguageDropdown}>
                  <div className={styles.addLanguageDropdownList}>
                    {candidateLanguages.length === 0 ? (
                      <Text
                        as="p"
                        size="2"
                        color="gray"
                        className={styles.emptySuggestions}
                      >
                        <FormattedMessage id="SearchableDropdown.empty" />
                      </Text>
                    ) : (
                      candidateLanguages.map((lang) => {
                        const isCustom = !builtinLanguageSet.has(lang);
                        return (
                          <button
                            key={lang}
                            type="button"
                            className={styles.languageSuggestionItem}
                            onClick={() => {
                              onSelectLanguage(lang);
                            }}
                          >
                            <span className={styles.languageSuggestionItemMain}>
                              <Text as="span" size="2">
                                {getLanguageDisplayText(lang)}
                              </Text>
                              {isCustom ? (
                                <Badge
                                  size="1"
                                  variant="neutral"
                                  text={
                                    <FormattedMessage id="LanguagesConfigurationScreen.custom-language-badge" />
                                  }
                                />
                              ) : null}
                            </span>
                            <Text as="span" size="1" color="gray">
                              {lang}
                            </Text>
                          </button>
                        );
                      })
                    )}
                  </div>
                </div>
              </>
            ) : null}
          </div>
        </div>

        <div className={styles.languagesTableWrapper}>
          <div className={styles.languagesTable}>
            <div className={styles.languagesTableHeader}>
              <div className={styles.languagesTableHeaderCellLanguage}>
                <FormattedMessage id="LanguagesConfigurationScreen.supportedLanguages.column.language" />
              </div>
              <div
                className={styles.languagesTableHeaderCellActions}
                aria-hidden={true}
              />
            </div>
            {selectedLanguages.length === 0 ? (
              <div className={styles.languagesTableEmptyRow}>
                <Text as="p" size="2" color="gray">
                  <FormattedMessage id="LanguagesConfigurationScreen.supportedLanguages.empty" />
                </Text>
              </div>
            ) : (
              selectedLanguages.map((lang) => {
                const isCustom = !builtinLanguageSet.has(lang);
                const canDelete =
                  lang !== primaryLanguage && lang !== "en";
                return (
                  <div key={lang} className={styles.languagesTableRow}>
                    <div className={styles.languagesTableCellLanguage}>
                      <div className={styles.languagesTableCellLanguageInner}>
                        <Text
                          size="2"
                          className={styles.languagesTableCellLanguageText}
                        >
                          {getLanguageDisplayText(lang)}
                        </Text>
                        {isCustom ? (
                          <Badge
                            size="1"
                            variant="neutral"
                            text={
                              <FormattedMessage id="LanguagesConfigurationScreen.custom-language-badge" />
                            }
                          />
                        ) : null}
                      </div>
                    </div>
                    <div className={styles.languagesTableCellActions}>
                      <DropdownMenu.Root>
                        <DropdownMenu.Trigger>
                          <RadixIconButton
                            className={styles.rowActionsButton}
                            variant="soft"
                            color="gray"
                            size="2"
                            aria-label={renderToString(
                              "LanguagesConfigurationScreen.supportedLanguages.row-actions"
                            )}
                          >
                            <DotsVerticalIcon width="1rem" height="1rem" />
                          </RadixIconButton>
                        </DropdownMenu.Trigger>
                        <DropdownMenu.Content align="end">
                          {lang === "en" ? (
                            <Tooltip
                              content={
                                <FormattedMessage id="LanguagesConfigurationScreen.cannot-remove-default-language" />
                              }
                            >
                              <span className={styles.disabledDeleteTooltipTarget}>
                                <DropdownMenu.Item
                                  color="red"
                                  disabled={true}
                                  onSelect={(e) => {
                                    e.preventDefault();
                                  }}
                                >
                                  <TrashIcon />
                                  <FormattedMessage id="delete" />
                                </DropdownMenu.Item>
                              </span>
                            </Tooltip>
                          ) : (
                            <DropdownMenu.Item
                              color="red"
                              disabled={!canDelete}
                              onSelect={() => {
                                if (canDelete) {
                                  onRemoveLanguage(lang);
                                }
                              }}
                            >
                              <TrashIcon />
                              <FormattedMessage id="delete" />
                            </DropdownMenu.Item>
                          )}
                        </DropdownMenu.Content>
                      </DropdownMenu.Root>
                    </div>
                  </div>
                );
              })
            )}
          </div>
        </div>
      </div>
    );
  };

interface SupportedLanguagesSectionProps {
  primaryLanguage: LanguageTag;
  builtinLanguages: LanguageTag[];
  availableLanguages: LanguageTag[];
  supportedLanguages: LanguageTag[];
  onToggleSupportedLanguage: (lang: LanguageTag, selected: boolean) => void;
  className?: string;
}
const SupportedLanguagesSection: React.VFC<SupportedLanguagesSectionProps> =
  function SupportedLanguagesSection(props) {
    const {
      primaryLanguage,
      builtinLanguages,
      availableLanguages,
      supportedLanguages,
      onToggleSupportedLanguage,
      className,
    } = props;
    return (
      <SettingsSectionCard
        className={cn(styles.widget, className)}
        contentClassName="gap-4"
        title={
          <FormattedMessage id="LanguagesConfigurationScreen.supportedLanguages.title" />
        }
      >
        <SupportedLanguagesList
          primaryLanguage={primaryLanguage}
          builtinLanguages={builtinLanguages}
          availableLanguages={availableLanguages}
          supportedLanguages={supportedLanguages}
          onToggleSupportedLanguage={onToggleSupportedLanguage}
        />
      </SettingsSectionCard>
    );
  };

interface LanguagesConfigurationContentProps {
  form: AppConfigFormModel<ConfigFormState>;
  availableLanguages: LanguageTag[];
  builtinLanguages: LanguageTag[];
}

const LanguagesConfigurationContent: React.VFC<LanguagesConfigurationContentProps> =
  function LanguagesConfigurationContent(props) {
    const { form, availableLanguages, builtinLanguages } = props;
    const { state, setState } = form;
    const { isDirty } = useFormContainerBaseContext();
    const contentWidthAnchorRef = useRef<HTMLDivElement>(null);

    const onChangePrimaryLanguage = useCallback(
      (primaryLanguage: string) => {
        setState(updatePrimaryLanguage(primaryLanguage));
      },
      [setState]
    );

    const onToggleSupportedLanguage = useCallback(
      (language: LanguageTag, checked: boolean) => {
        setState(toggleSupportedLanguage(language, checked));
      },
      [setState]
    );

    return (
      <ScreenContent className={cn(isDirty ? styles.contentWithSaveBar : null)}>
        <div
          ref={contentWidthAnchorRef}
          className={cn(styles.widget, styles.pageHeader)}
        >
          <Text as="p" size="5" weight="bold" className={styles.pageTitle}>
            <FormattedMessage id="LanguagesConfigurationScreen.title" />
          </Text>
        </div>
        <SelectPrimaryLanguageSection
          primaryLanguage={state.fallbackLanguage}
          supportedLanguages={state.supportedLanguages}
          onChangePrimaryLanguage={onChangePrimaryLanguage}
        />
        <SupportedLanguagesSection
          className={isDirty ? styles.settingsCardSaveBarClearance : undefined}
          primaryLanguage={state.fallbackLanguage}
          builtinLanguages={builtinLanguages}
          availableLanguages={availableLanguages}
          supportedLanguages={state.supportedLanguages}
          onToggleSupportedLanguage={onToggleSupportedLanguage}
        />
        <SaveFunctionBar anchorRef={contentWidthAnchorRef} />
      </ScreenContent>
    );
  };

const LanguagesConfigurationScreen: React.VFC =
  function LanguagesConfigurationScreen() {
    const { appID } = useParams() as { appID: string };
    const { renderToString } = useContext(MFContext);
    const { availableLanguages, builtinLanguages } = useSystemConfig();

    const form = useAppConfigForm({
      appID,
      constructFormState,
      constructConfig,
    });

    const pageContextValue = useMemo<PageContextValue>(() => {
      return {
        getLanguageDisplayText: (lang: LanguageTag) =>
          renderToString(`Locales.${lang}`),
      };
    }, [renderToString]);

    const sortedLanguages = useMemo(() => {
      const sortLanguage = (a: LanguageTag, b: LanguageTag) => {
        return pageContextValue
          .getLanguageDisplayText(a)
          .localeCompare(pageContextValue.getLanguageDisplayText(b));
      };
      return {
        availableLanguages: [...availableLanguages].sort(sortLanguage),
        builtinLanguages: [...builtinLanguages].sort(sortLanguage),
      };
    }, [pageContextValue, availableLanguages, builtinLanguages]);

    if (form.isLoading) {
      return <ShowLoading />;
    }

    if (form.loadError) {
      return <ShowError error={form.loadError} onRetry={form.reload} />;
    }

    return (
      <PageContext.Provider value={pageContextValue}>
        <FormContainer form={form} hideFooterComponent={true}>
          <LanguagesConfigurationContent
            form={form}
            availableLanguages={sortedLanguages.availableLanguages}
            builtinLanguages={sortedLanguages.builtinLanguages}
          />
        </FormContainer>
      </PageContext.Provider>
    );
  };

export default LanguagesConfigurationScreen;
