import React, { useCallback, useContext, useMemo } from "react";
import cn from "classnames";
import { Select, Text } from "@radix-ui/themes";
import { Context, FormattedMessage } from "../../intl";

import { LanguageTag } from "../../util/resource";

import styles from "./ManageLanguageWidget.module.css";

interface ManageLanguageWidgetProps {
  className?: string;
  showLabel?: boolean;

  // The supported languages.
  existingLanguages: LanguageTag[];
  supportedLanguages: LanguageTag[];

  // The selected language.
  selectedLanguage: LanguageTag;
  onChangeSelectedLanguage: (newSelectedLanguage: LanguageTag) => void;

  // The fallback language.
  fallbackLanguage: LanguageTag;
}

interface LanguageOption {
  value: LanguageTag;
  text: string;
  isFallbackLanguage: boolean;
  disabled: boolean;
}

function getLanguageLocaleKey(locale: LanguageTag) {
  return `Locales.${locale}`;
}

const ManageLanguageWidget: React.VFC<ManageLanguageWidgetProps> =
  function ManageLanguageWidget(props: ManageLanguageWidgetProps) {
    const {
      className,
      supportedLanguages,
      existingLanguages,
      selectedLanguage,
      onChangeSelectedLanguage,
      fallbackLanguage,
      showLabel = true,
    } = props;

    const { renderToString } = useContext(Context);

    const displayTemplateLocale = useCallback(
      (locale: LanguageTag) => {
        return renderToString(getLanguageLocaleKey(locale));
      },
      [renderToString]
    );

    const templateLocaleOptions: LanguageOption[] = useMemo(() => {
      const options: LanguageOption[] = [];

      const combinedLocales = new Set([
        ...existingLanguages,
        ...supportedLanguages,
      ]);

      for (const locale of combinedLocales) {
        const isNew = !existingLanguages.includes(locale);
        const isRemoved = !supportedLanguages.includes(locale);

        let localeDisplay = displayTemplateLocale(locale);
        if (isRemoved) {
          localeDisplay = renderToString(
            "ManageLanguageWidget.option-removed",
            {
              LANG: localeDisplay,
            }
          );
        }

        options.push({
          value: locale,
          text: localeDisplay,
          isFallbackLanguage: fallbackLanguage === locale,
          disabled: isRemoved || isNew,
        });
      }

      return options;
    }, [
      existingLanguages,
      supportedLanguages,
      displayTemplateLocale,
      fallbackLanguage,
      renderToString,
    ]);

    return (
      <div className={cn(className, styles.root)}>
        <div className={styles.container}>
          {showLabel ? (
            <Text
              as="label"
              size="2"
              weight="medium"
              className={styles.titleLabel}
              htmlFor="language-widget"
            >
              <FormattedMessage id="ManageLanguageWidget.title" />
            </Text>
          ) : null}
          <div className={styles.control}>
            <Select.Root
              value={selectedLanguage}
              onValueChange={onChangeSelectedLanguage}
              size="2"
            >
              <Select.Trigger
                id="language-widget"
                className={styles.dropdown}
              />
              <Select.Content position="popper">
                {templateLocaleOptions.map((option) => (
                  <Select.Item
                    key={option.value}
                    value={option.value}
                    disabled={option.disabled}
                    className={cn(
                      option.disabled ? styles.optionDisabled : null
                    )}
                  >
                    <FormattedMessage
                      id="ManageLanguageWidget.language-label"
                      values={{
                        LANG: option.text,
                        IS_FALLBACK: String(option.isFallbackLanguage),
                      }}
                    />
                  </Select.Item>
                ))}
              </Select.Content>
            </Select.Root>
          </div>
        </div>
      </div>
    );
  };

export default ManageLanguageWidget;
