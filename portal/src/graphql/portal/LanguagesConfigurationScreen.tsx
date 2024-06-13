import React, { PropsWithChildren, useMemo, useState } from "react";
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
}
const SelectPrimaryLanguageSection: React.VFC<SelectPrimaryLanguageWidgetProps> =
  function SelectPrimaryLanguageSection(props) {
    const { className, availableLanguages } = props;

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
    return (
      <FormContainer form={appConfigForm} canSave={true}>
        <ScreenContent>
          <ScreenTitle className={cn("col-span-8", "tablet:col-span-full")}>
            <FormattedMessage id="LanguagesConfigurationScreen.title" />
          </ScreenTitle>
          <SelectPrimaryLanguageSection
            className={styles.pageSection}
            availableLanguages={availableLanguages}
          />
        </ScreenContent>
      </FormContainer>
    );
  };

export default LanguagesConfigurationScreen;
