import React from "react";
import { ControlledEditor, ControlledEditorProps } from "@monaco-editor/react";

export interface CodeEditorProps extends ControlledEditorProps {
  className?: string;
}

const CodeEditor: React.FC<CodeEditorProps> = function CodeEditor(props) {
  const { className, ...rest } = props;

  return (
    <div className={className}>
      <ControlledEditor height="100%" {...rest} />
    </div>
  );
};

export default CodeEditor;
