import React, { Fragment } from "react";
import { Label, Text } from "@fluentui/react";
import CodeEditor from "../../CodeEditor";
import cn from "classnames";
import styles from "./EditTemplatesWidget.module.scss";

export interface EditTemplatesWidgetItem {
  key: string;
  title: React.ReactNode;
  language: "html" | "plaintext" | "json" | "css";
  value: string;
  onChange: (value: string | undefined, e: unknown) => void;
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

const EditTemplatesWidget: React.FC<EditTemplatesWidgetProps> =
  function EditTemplatesWidget(props: EditTemplatesWidgetProps) {
    const { className, sections } = props;

    return (
      <div className={cn(styles.form, className)}>
        {sections.map((section) => {
          return (
            <Fragment key={section.key}>
              <Label className={styles.boldLabel}>{section.title}</Label>
              {section.items.map((item) => {
                return (
                  <Fragment key={item.key}>
                    <Text className={styles.label} block={true}>
                      {item.title}
                    </Text>
                    <CodeEditor
                      className={styles.codeEditor}
                      language={item.language}
                      value={item.value}
                      onChange={item.onChange}
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
