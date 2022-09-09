import React from "react";
import ControlledEditor, { EditorProps } from "@monaco-editor/react";

export interface CodeEditorProps extends EditorProps {
  className?: string;
}

const CodeEditor: React.VFC<CodeEditorProps> = function CodeEditor(props) {
  const { className, ...rest } = props;

  return (
    <div className={className}>
      <ControlledEditor height="100%" {...rest} />
    </div>
  );
};

export default CodeEditor;
