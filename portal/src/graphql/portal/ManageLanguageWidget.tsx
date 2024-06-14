import React, { useCallback, useContext, useMemo } from "react";
import cn from "classnames";
import {
  Label,
  Dropdown,
  IDropdownOption,
  Text,
  IRenderFunction,
} from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";

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

    const templateLocaleOptions: IDropdownOption[] = useMemo(() => {
      const options = [];

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
          key: locale,
          text: localeDisplay,
          data: {
            isFallbackLanguage: fallbackLanguage === locale,
          },
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
          <Text
            styles={(_, theme) => ({
              root: option?.disabled
                ? {
                    fontStyle: "italic",
                    color: theme.semanticColors.disabledText,
                  }
                : undefined,
            })}
          >
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
        <div className={cn(className, styles.root)}>
          <div className={styles.container}>
            {showLabel ? (
              <Label className={styles.titleLabel}>
                <FormattedMessage id="ManageLanguageWidget.title" />
              </Label>
            ) : null}
            <div className={styles.control}>
              <Dropdown
                id="language-widget"
                className={styles.dropdown}
                options={templateLocaleOptions}
                onChange={onChangeTemplateLocale}
                selectedKey={selectedLanguage}
                onRenderTitle={onRenderTitle}
                onRenderOption={onRenderOption}
              />
            </div>
          </div>
        </div>
      </>
    );
  };

export default ManageLanguageWidget;
