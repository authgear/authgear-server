import React, { Fragment, useCallback } from "react";
import { Label, Text } from "@fluentui/react";
import CodeEditor from "../../CodeEditor";
import cn from "classnames";
import styles from "./EditTemplatesWidget.module.css";
import TextField from "../../TextField";

export interface TextFieldWidgetIteProps {
  className?: string;
  value: string;
  onChange: (value: string | undefined, e: unknown) => void;
  readOnly?: boolean;
}

// TextFieldWidgetItem is a wrapper of TextField
// The positional arguments order of onChange functions are different between
// TextField and CodeEditor, so we need to wrap the TextField
const TextFieldWidgetItem: React.VFC<TextFieldWidgetIteProps> =
  function TextFieldWidgetItem(props) {
    const { className, value, onChange: onChangeProps, readOnly } = props;

    const onChange = useCallback(
      (
        event: React.FormEvent<HTMLInputElement | HTMLTextAreaElement>,
        newValue?: string | undefined
      ) => onChangeProps(newValue, event),
      [onChangeProps]
    );

    return (
      <TextField
        className={className}
        value={value}
        onChange={onChange}
        readOnly={readOnly}
      />
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
  sections: EditTemplatesWidgetSection[];
}

const EditTemplatesWidget: React.VFC<EditTemplatesWidgetProps> =
  function EditTemplatesWidget(props: EditTemplatesWidgetProps) {
    const { className, sections } = props;

    return (
      <div className={cn(styles.form, className)}>
        {sections.map((section) => {
          return (
            <Fragment key={section.key}>
              {section.title != null ? (
                <Label className={styles.boldLabel}>{section.title}</Label>
              ) : null}
              {section.items.map((item) => {
                return item.editor === "code" ? (
                  <Fragment key={item.key}>
                    <Text className={styles.label} block={true}>
                      {item.title}
                    </Text>
                    <CodeEditor
                      className={styles.codeEditor}
                      language={item.language}
                      value={item.value}
                      onChange={item.onChange}
                      options={{ readOnly: item.readOnly }}
                    />
                  </Fragment>
                ) : (
                  <Fragment key={item.key}>
                    <Text className={styles.label} block={true}>
                      {item.title}
                    </Text>
                    <TextFieldWidgetItem
                      className={styles.textField}
                      value={item.value}
                      onChange={item.onChange}
                      readOnly={item.readOnly}
                    />
                  </Fragment>
                );
              })}
            </Fragment>
          );
        })}
      </div>
    );
  };

export default EditTemplatesWidget;
