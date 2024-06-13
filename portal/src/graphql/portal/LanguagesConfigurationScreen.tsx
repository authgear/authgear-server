import React, {
  PropsWithChildren,
  useCallback,
  useMemo,
  useState,
} from "react";
import { useParams } from "react-router-dom";
import { IDropdownOption, Label } from "@fluentui/react";
import { produce } from "immer";
import cn from "classnames";
import { FormattedMessage } from "@oursky/react-messageformat";

import { PortalAPIAppConfig } from "../../types";
import { clearEmptyObject } from "../../util/misc";
import { useAppConfigForm } from "../../hook/useAppConfigForm";
import FormContainer from "../../FormContainer";
import ScreenContent from "../../ScreenContent";
import ScreenTitle from "../../ScreenTitle";
import { useSystemConfig } from "../../context/SystemConfigContext";
import WidgetTitle from "../../WidgetTitle";
import WidgetDescription from "../../WidgetDescription";
import { SearchableDropdown } from "../../components/common/SearchableDropdown";

import styles from "./LanguagesConfigurationScreen.module.css";

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

interface SectionProps {
  className?: string;
}
const Section: React.VFC<PropsWithChildren<SectionProps>> = function Section(
  props
) {
  const { className, children } = props;
  return (
    <div className={cn("flex", "flex-col", "gap-y-4", className)}>
      {children}
    </div>
  );
};

interface SelectPrimaryLanguageWidgetProps {
  className?: string;
  availableLanguages: string[];
  primaryLanguage: string;
  onChangePrimaryLanguage: (language: string) => void;
}
const SelectPrimaryLanguageSection: React.VFC<SelectPrimaryLanguageWidgetProps> =
  function SelectPrimaryLanguageSection(props) {
    const {
      className,
      availableLanguages,
      primaryLanguage,
      onChangePrimaryLanguage,
    } = props;

    const [searchValue, setSearchValue] = useState("");
    const dropdownOptions: IDropdownOption[] = useMemo(() => {
      const filteredLanguages = availableLanguages.filter((lang) =>
        lang.toLowerCase().includes(searchValue.toLowerCase())
      );
      return filteredLanguages.map((lang) => ({
        key: lang,
        text: lang,
      }));
    }, [availableLanguages, searchValue]);

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
          />
        </Label>
      </Section>
    );
  };

const LanguagesConfigurationScreen: React.VFC =
  function LanguagesConfigurationScreen() {
    const { appID } = useParams() as { appID: string };
    const { availableLanguages } = useSystemConfig();

    const appConfigForm = useAppConfigForm({
      appID,
      constructFormState,
      constructConfig,
    });

    // TODO(1380)
    // Add fallback language to supported language
    const onChangePrimaryLanguage = useCallback(
      (primaryLanguage: string) => {
        appConfigForm.setState((state) => {
          return {
            ...state,
            fallbackLanguage: primaryLanguage,
          };
        });
      },
      [appConfigForm]
    );

    return (
      <FormContainer form={appConfigForm} canSave={true}>
        <ScreenContent>
          <ScreenTitle className={cn("col-span-8", "tablet:col-span-full")}>
            <FormattedMessage id="LanguagesConfigurationScreen.title" />
          </ScreenTitle>
          <SelectPrimaryLanguageSection
            className={styles.pageSection}
            availableLanguages={availableLanguages}
            primaryLanguage={appConfigForm.state.fallbackLanguage}
            onChangePrimaryLanguage={onChangePrimaryLanguage}
          />
        </ScreenContent>
      </FormContainer>
    );
  };

export default LanguagesConfigurationScreen;
