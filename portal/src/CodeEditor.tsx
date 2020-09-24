import React, { useCallback } from "react";
import { ControlledEditor } from "@monaco-editor/react";

export interface CodeEditorOnChange {
  (e: unknown, value: string | undefined): void;
}

export interface Props {
  className?: string;
  language: string;
  value: string;
  onChange: CodeEditorOnChange;
}

const CodeEditor: React.FC<Props> = function CodeEditor(props) {
  const { className, language, value, onChange } = props;

  const callback = useCallback(
    (e: unknown, value: string | undefined) => {
      onChange(e, value);
      return undefined;
    },
    [onChange]
  );

  return (
    <div className={className}>
      <ControlledEditor
        height="100%"
        value={value}
        language={language}
        onChange={callback}
      />
    </div>
  );
};

export default CodeEditor;
