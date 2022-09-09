import React, { useCallback, useState } from "react";
import ControlledEditor, { EditorProps } from "@monaco-editor/react";
import { IconButton } from "@fluentui/react";
import { useCopyFeedback } from "./hook/useCopyFeedback";
import cn from "classnames";
import styles from "./CodeEditor.module.css";

interface CodeEditorCopyButtonProps {
  copyValue: string;
}

const CodeEditorCopyButton: React.VFC<CodeEditorCopyButtonProps> = (props) => {
  const { copyValue } = props;
  const { copyButtonProps, Feedback } = useCopyFeedback({
    textToCopy: copyValue,
  });

  return (
    <div className={styles.copyButton}>
      <IconButton {...copyButtonProps} />
      <Feedback />
    </div>
  );
};

export interface CodeEditorProps extends EditorProps {
  className?: string;
  isCopyButtonVisible?: boolean;
  copyValue?: string;
}

const CodeEditor: React.VFC<CodeEditorProps> = function CodeEditor(props) {
  const { className, isCopyButtonVisible, copyValue, ...rest } = props;
  const [isCodeEditorMounted, setIsCodeEditorMounted] =
    useState<boolean>(false);

  const onCodeEditorMounted = useCallback(() => {
    setIsCodeEditorMounted(true);
  }, []);

  return (
    <div className={cn(styles.root, className)}>
      {isCopyButtonVisible && isCodeEditorMounted ? (
        // The copy button should be visible when the editor is mounted
        // otherwise it will be shown when the editor is still in loading state
        <CodeEditorCopyButton copyValue={copyValue ?? ""} />
      ) : null}
      <ControlledEditor height="100%" onMount={onCodeEditorMounted} {...rest} />
    </div>
  );
};

export default CodeEditor;
