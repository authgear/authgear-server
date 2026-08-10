import React, { Fragment, useCallback } from "react";
import { Text } from "@radix-ui/themes";
import CodeEditor from "../../CodeEditor";
import cn from "classnames";
import styles from "./EditTemplatesWidget.module.css";
import { TextField as RadixTextField } from "../../components/v2/TextField/TextField";
import { SettingsSectionCard } from "../../components/v2/SettingsSectionCard/SettingsSectionCard";
import { FormattedMessage } from "../../intl";

export interface TextFieldWidgetIteProps {
  className?: string;
  label?: React.ReactNode;
  value: string;
  onChange: (value: string | undefined, e: unknown) => void;
  readOnly?: boolean;
}

// TextFieldWidgetItem adapts Radix TextField's onChange signature to CodeEditor's
// (value, event) so both editor types stay interchangeable at the call site.
const TextFieldWidgetItem: React.VFC<TextFieldWidgetIteProps> =
  function TextFieldWidgetItem(props) {
    const {
      className,
      label,
      value,
      onChange: onChangeProps,
      readOnly,
    } = props;

    const onChange = useCallback(
      (event: React.ChangeEvent<HTMLInputElement>) => {
        onChangeProps(event.currentTarget.value, event);
      },
      [onChangeProps]
    );

    return (
      <div className={className}>
        <RadixTextField
          size="2"
          labelSize="2"
          label={label}
          value={value}
          onChange={onChange}
          readOnly={readOnly}
        />
      </div>
    );
  };

export interface EditTemplatesWidgetItem {
  key: string;
  title: React.ReactNode;
  language: "html" | "plaintext" | "json" | "css" | "yaml";
  editor: "code" | "textfield";
  value: string;
  onChange: (value: string | undefined, e: unknown) => void;
  readOnly?: boolean;
}

export interface EditTemplatesWidgetSection {
  key: string;
  title: React.ReactNode;
  items: EditTemplatesWidgetItem[];
}

export interface EditTemplatesWidgetProps {
  className?: string;
  codeEditorClassname?: string;
  sections: EditTemplatesWidgetSection[];
}

function renderItem(
  item: EditTemplatesWidgetItem,
  codeEditorClassname: string | undefined
): React.ReactElement {
  if (item.editor === "code") {
    return (
      <div key={item.key} className={styles.editorCard}>
        <div className={styles.editorCardHeader}>
          <Text
            as="p"
            size="3"
            weight="medium"
            className={styles.editorCardTitle}
          >
            {item.title}
          </Text>
        </div>
        <CodeEditor
          className={cn(
            styles.codeEditor,
            item.language !== "html" && styles.codeEditorCompact,
            codeEditorClassname
          )}
          language={item.language}
          value={item.value}
          onChange={item.onChange}
          options={{ readOnly: item.readOnly }}
        />
      </div>
    );
  }

  return (
    <SettingsSectionCard
      key={item.key}
      title={<FormattedMessage id="EditTemplatesWidget.mailing-information" />}
    >
      <TextFieldWidgetItem
        label={item.title}
        value={item.value}
        onChange={item.onChange}
        readOnly={item.readOnly}
      />
    </SettingsSectionCard>
  );
}

const EditTemplatesWidget: React.VFC<EditTemplatesWidgetProps> =
  function EditTemplatesWidget(props: EditTemplatesWidgetProps) {
    const { className, codeEditorClassname, sections } = props;

    return (
      <div className={cn(styles.form, className)}>
        {sections.map((section) => {
          const items = section.items.map((item) =>
            renderItem(item, codeEditorClassname)
          );

          // Sections with titles (e.g. New User / Existing User) use a
          // left-label card; channel sections (title stripped) stay flat.
          if (section.title != null) {
            return (
              <SettingsSectionCard
                key={section.key}
                title={section.title}
                contentClassName={styles.sectionContent}
              >
                {items}
              </SettingsSectionCard>
            );
          }

          return <Fragment key={section.key}>{items}</Fragment>;
        })}
      </div>
    );
  };

export default EditTemplatesWidget;
