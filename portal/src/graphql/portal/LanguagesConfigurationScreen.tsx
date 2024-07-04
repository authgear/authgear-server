import React, {
  PropsWithChildren,
  useCallback,
  useContext,
  useMemo,
  useState,
} from "react";
import { useParams } from "react-router-dom";
import { Checkbox, IDropdownOption, Label, List, Text } from "@fluentui/react";
import { produce } from "immer";
import cn from "classnames";
import {
  Context as MFContext,
  FormattedMessage,
} from "@oursky/react-messageformat";

import FormContainer from "../../FormContainer";
import HorizontalDivider from "../../HorizontalDivider";
import ScreenContent from "../../ScreenContent";
import ScreenTitle from "../../ScreenTitle";
import { SearchableDropdown } from "../../components/common/SearchableDropdown";
import WidgetTitle from "../../WidgetTitle";
import WidgetDescription from "../../WidgetDescription";

import { PortalAPIAppConfig } from "../../types";
import { clearEmptyObject } from "../../util/misc";
import { useAppConfigForm } from "../../hook/useAppConfigForm";
import { useSystemConfig } from "../../context/SystemConfigContext";

import { LanguageTag } from "../../util/resource";

import styles from "./LanguagesConfigurationScreen.module.css";
import WidgetSubtitle from "../../WidgetSubtitle";

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
  availableLanguages: LanguageTag[],
  language: LanguageTag,
  checked: boolean
): (s: ConfigFormState) => ConfigFormState {
  return (state) => {
    const supportedLanguageSet = new Set(state.supportedLanguages);
    if (checked) {
      supportedLanguageSet.add(language);
    } else {
      supportedLanguageSet.delete(language);
    }
    const supportedLanguages = availableLanguages.filter((lang) =>
      supportedLanguageSet.has(lang)
    );
    return {
      ...state,
      supportedLanguages,
    };
  };
}

function updatePrimaryLanguage(
  availableLanguages: LanguageTag[],
  primaryLanguage: LanguageTag
): (s: ConfigFormState) => ConfigFormState {
  return (state) => {
    return toggleSupportedLanguage(
      availableLanguages,
      primaryLanguage,
      true
    )({ ...state, fallbackLanguage: primaryLanguage });
  };
}

interface SectionProps {
  className?: string;
}
const Section: React.VFC<PropsWithChildren<SectionProps>> = function Section(
  props
) {
  const { className, children } = props;
  return <div className={cn("space-y-4", className)}>{children}</div>;
};

interface SelectPrimaryLanguageWidgetProps {
  className?: string;
  availableLanguages: LanguageTag[];
  primaryLanguage: LanguageTag;
  onChangePrimaryLanguage: (language: LanguageTag) => void;
}
const SelectPrimaryLanguageSection: React.VFC<SelectPrimaryLanguageWidgetProps> =
  function SelectPrimaryLanguageSection(props) {
    const {
      className,
      availableLanguages,
      primaryLanguage,
      onChangePrimaryLanguage,
    } = props;

    const { getLanguageDisplayText } = useContext(PageContext);

    const [searchValue, setSearchValue] = useState("");
    const dropdownOptions: IDropdownOption[] = useMemo(() => {
      const filteredLanguages = availableLanguages.filter(
        (lang) =>
          lang.toLowerCase().includes(searchValue.toLowerCase()) ||
          getLanguageDisplayText(lang)
            .toLowerCase()
            .includes(searchValue.toLowerCase())
      );
      return filteredLanguages.map((lang) => ({
        key: lang,
        text: getLanguageDisplayText(lang),
      }));
    }, [availableLanguages, searchValue, getLanguageDisplayText]);

    const selectedOption = useMemo(() => {
      return dropdownOptions.find((option) => option.key === primaryLanguage);
    }, [dropdownOptions, primaryLanguage]);

    const onChange = useCallback(
      (_e: unknown, option?: IDropdownOption) => {
        const key = option?.key as string | null;
        if (key) {
          onChangePrimaryLanguage(key);
        }
      },
      [onChangePrimaryLanguage]
    );

    return (
      <Section className={className}>
        <WidgetTitle>
          <FormattedMessage id="LanguagesConfigurationScreen.selectPrimaryLanguageWidget.title" />
        </WidgetTitle>
        <WidgetDescription>
          <FormattedMessage id="LanguagesConfigurationScreen.selectPrimaryLanguageWidget.description" />
        </WidgetDescription>
        <Label>
          <FormattedMessage id="LanguagesConfigurationScreen.selectPrimaryLanguageWidget.dropdown.label" />
          <SearchableDropdown
            className={cn("mt-1")}
            options={dropdownOptions}
            onChange={onChange}
            selectedItem={selectedOption}
            searchValue={searchValue}
            onSearchValueChange={setSearchValue}
            onRenderCaretDown={undefined} // we want to use the default arrow down icon
          />
        </Label>
      </Section>
    );
  };

interface SupportedLanguageOption {
  key: LanguageTag;
  selected: boolean;
  disabled: boolean;
}

interface SupportedLanguageCheckboxProps {
  language: LanguageTag;
  selected: boolean;
  disabled: boolean;
  onToggleSupportedLanguage: (lang: LanguageTag, checked: boolean) => void;
}
const SupportedLanguageCheckbox: React.VFC<SupportedLanguageCheckboxProps> =
  function SupportedLanguageCheckbox(props) {
    const { language, disabled, selected, onToggleSupportedLanguage } = props;
    const { getLanguageDisplayText } = useContext(PageContext);
    const onChange = useCallback(
      (_e: unknown, checked?: boolean) => {
        onToggleSupportedLanguage(language, checked === true);
      },
      [language, onToggleSupportedLanguage]
    );
    return (
      <div className={cn("flex", "items-center")}>
        <Checkbox disabled={disabled} checked={selected} onChange={onChange} />
        <Text className={cn("ml-1")}>{getLanguageDisplayText(language)}</Text>
      </div>
    );
  };

interface BuiltInTranslationSectionProps {
  primaryLanguage: LanguageTag;
  builtinLanguages: LanguageTag[];
  supportedLanguages: LanguageTag[];
  onToggleSupportedLanguage: (lang: LanguageTag, selected: boolean) => void;
}
const BuiltInTranslationSection: React.VFC<BuiltInTranslationSectionProps> =
  function BuiltInTranslationSection(props) {
    const {
      primaryLanguage,
      builtinLanguages,
      supportedLanguages,
      onToggleSupportedLanguage,
    } = props;
    const options = useMemo<SupportedLanguageOption[]>(() => {
      const supportedLanguageSet = new Set(supportedLanguages);
      return builtinLanguages.map((lang) => ({
        key: lang,
        selected: supportedLanguageSet.has(lang),
        disabled: lang === primaryLanguage,
      }));
    }, [primaryLanguage, builtinLanguages, supportedLanguages]);

    return (
      <Section>
        <WidgetSubtitle>
          <FormattedMessage id="LanguagesConfigurationScreen.builtInTranslation.title" />
        </WidgetSubtitle>
        <WidgetDescription>
          <FormattedMessage id="LanguagesConfigurationScreen.builtInTranslation.description" />
        </WidgetDescription>
        <ul className={cn("block", "list-none", "space-y-4", "pt-2")}>
          {options.map((option) => (
            <li key={option.key} className={cn("flex", "items-center")}>
              <SupportedLanguageCheckbox
                language={option.key}
                disabled={option.disabled}
                selected={option.selected}
                onToggleSupportedLanguage={onToggleSupportedLanguage}
              />
            </li>
          ))}
        </ul>
      </Section>
    );
  };

interface CustomTranslationSectionProps {
  primaryLanguage: LanguageTag;
  builtinLanguages: LanguageTag[];
  availableLanguages: LanguageTag[];
  supportedLanguages: LanguageTag[];
  onToggleSupportedLanguage: (lang: LanguageTag, selected: boolean) => void;
}
const CustomTranslationSection: React.VFC<CustomTranslationSectionProps> =
  function CustomTranslationSection(props) {
    const {
      primaryLanguage,
      builtinLanguages,
      availableLanguages,
      supportedLanguages,
      onToggleSupportedLanguage,
    } = props;

    const options = useMemo<SupportedLanguageOption[]>(() => {
      const builtinLanguageSet = new Set(builtinLanguages);
      const supportedLanguageSet = new Set(supportedLanguages);
      return availableLanguages
        .filter((lang) => !builtinLanguageSet.has(lang))
        .map((lang) => ({
          key: lang,
          selected: supportedLanguageSet.has(lang),
          disabled: lang === primaryLanguage,
        }));
    }, [
      primaryLanguage,
      builtinLanguages,
      availableLanguages,
      supportedLanguages,
    ]);

    const renderListItem = useCallback(
      (option?: SupportedLanguageOption) => {
        if (option == null) {
          return null;
        }
        return (
          <SupportedLanguageCheckbox
            language={option.key}
            disabled={option.disabled}
            selected={option.selected}
            onToggleSupportedLanguage={onToggleSupportedLanguage}
          />
        );
      },
      [onToggleSupportedLanguage]
    );

    return (
      <Section>
        <WidgetSubtitle>
          <FormattedMessage id="LanguagesConfigurationScreen.customTranslation.title" />
        </WidgetSubtitle>
        <WidgetDescription>
          <FormattedMessage id="LanguagesConfigurationScreen.customTranslation.description" />
        </WidgetDescription>
        <div className={styles.customTranslationListContainer}>
          <List items={options} onRenderCell={renderListItem}></List>
        </div>
      </Section>
    );
  };

interface SupportedLanguagesSectionProps {
  className?: string;
  primaryLanguage: LanguageTag;
  builtinLanguages: LanguageTag[];
  availableLanguages: LanguageTag[];
  supportedLanguages: LanguageTag[];
  onToggleSupportedLanguage: (lang: LanguageTag, selected: boolean) => void;
}
const SupportedLanguagesSection: React.VFC<SupportedLanguagesSectionProps> =
  function SupportedLanguagesSection(props) {
    const {
      className,
      primaryLanguage,
      builtinLanguages,
      availableLanguages,
      supportedLanguages,
      onToggleSupportedLanguage,
    } = props;
    return (
      <Section className={cn("space-y-8", className)}>
        <WidgetTitle>
          <FormattedMessage id="LanguagesConfigurationScreen.supportedLanguages.title" />
        </WidgetTitle>
        <BuiltInTranslationSection
          primaryLanguage={primaryLanguage}
          builtinLanguages={builtinLanguages}
          supportedLanguages={supportedLanguages}
          onToggleSupportedLanguage={onToggleSupportedLanguage}
        />
        <CustomTranslationSection
          primaryLanguage={primaryLanguage}
          builtinLanguages={builtinLanguages}
          availableLanguages={availableLanguages}
          supportedLanguages={supportedLanguages}
          onToggleSupportedLanguage={onToggleSupportedLanguage}
        />
      </Section>
    );
  };

const LanguagesConfigurationScreen: React.VFC =
  function LanguagesConfigurationScreen() {
    const { appID } = useParams() as { appID: string };
    const { renderToString } = useContext(MFContext);
    const { availableLanguages, builtinLanguages } = useSystemConfig();

    const appConfigForm = useAppConfigForm({
      appID,
      constructFormState,
      constructConfig,
    });

    const onChangePrimaryLanguage = useCallback(
      (primaryLanguage: string) => {
        appConfigForm.setState(
          updatePrimaryLanguage(availableLanguages, primaryLanguage)
        );
      },
      [appConfigForm, availableLanguages]
    );

    const onToggleSupportedLanguage = useCallback(
      (language: LanguageTag, checked: boolean) => {
        appConfigForm.setState(
          toggleSupportedLanguage(availableLanguages, language, checked)
        );
      },
      [appConfigForm, availableLanguages]
    );

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

    return (
      <PageContext.Provider value={pageContextValue}>
        <FormContainer form={appConfigForm} canSave={true}>
          <ScreenContent>
            <ScreenTitle className={cn("col-span-8", "tablet:col-span-full")}>
              <FormattedMessage id="LanguagesConfigurationScreen.title" />
            </ScreenTitle>
            <SelectPrimaryLanguageSection
              className={styles.pageSection}
              availableLanguages={sortedLanguages.availableLanguages}
              primaryLanguage={appConfigForm.state.fallbackLanguage}
              onChangePrimaryLanguage={onChangePrimaryLanguage}
            />
            <HorizontalDivider className={cn(styles.pageSection, "my-8")} />
            <SupportedLanguagesSection
              className={styles.pageSection}
              primaryLanguage={appConfigForm.state.fallbackLanguage}
              builtinLanguages={sortedLanguages.builtinLanguages}
              availableLanguages={sortedLanguages.availableLanguages}
              supportedLanguages={appConfigForm.state.supportedLanguages}
              onToggleSupportedLanguage={onToggleSupportedLanguage}
            />
          </ScreenContent>
        </FormContainer>
      </PageContext.Provider>
    );
  };

export default LanguagesConfigurationScreen;
