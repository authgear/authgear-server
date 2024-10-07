import React, { useCallback, useState } from "react";
import ControlledEditor from "@monaco-editor/react";
import { editor } from "monaco-editor";
import { IconButton } from "@fluentui/react";
import { useCopyFeedback } from "./hook/useCopyFeedback";
import cn from "classnames";
import styles from "./CodeBlock.module.css";

const CODE_BLOCK_OPTIONS: editor.IStandaloneEditorConstructionOptions = {
  readOnly: true,
  minimap: { enabled: false },
  wordWrap: "on",
  wrappingIndent: "deepIndent",
  renderLineHighlight: "none",
};

interface CodeBlockCopyButtonProps {
  copyValue: string;
}

const CodeBlockCopyButton: React.VFC<CodeBlockCopyButtonProps> = (props) => {
  const { copyValue } = props;
  // eslint-disable-next-line no-useless-assignment
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

export interface CodeBlockProps {
  className?: string;
  value?: string;
  language?: string;
}

const CodeBlock: React.VFC<CodeBlockProps> = function CodeBlock(props) {
  const { className, value, language } = props;
  const [isCodeBlockMounted, setIsCodeBlockMounted] = useState<boolean>(false);

  const onCodeBlockMounted = useCallback(() => {
    setIsCodeBlockMounted(true);
  }, []);

  return (
    <div className={cn(styles.root, className)}>
      {isCodeBlockMounted ? (
        // The copy button should be visible when the editor is mounted
        // otherwise it will be shown when the editor is still in loading state
        <CodeBlockCopyButton copyValue={value ?? ""} />
      ) : null}
      <ControlledEditor
        height="100%"
        value={value}
        onMount={onCodeBlockMounted}
        language={language}
        options={CODE_BLOCK_OPTIONS}
      />
    </div>
  );
};

export default CodeBlock;
