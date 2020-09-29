import React, { useCallback, useContext, useState } from "react";
import cn from "classnames";
import { Dialog, DialogFooter, PrimaryButton } from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";

import styles from "./TodoButtonWrapper.module.scss";

interface TodoButtonWrapperProps {
  children: React.ReactNode;
  className?: string;
}

const TodoButtonWrapper: React.FC<TodoButtonWrapperProps> = function TodoButtonWrapper(
  props: TodoButtonWrapperProps
) {
  const { children, className } = props;
  const { renderToString } = useContext(Context);

  const [todoDialogVisible, setTodoDialogVisible] = useState(false);

  const onTodoButtonClicked = useCallback(() => {
    setTodoDialogVisible(true);
  }, []);

  const onTodoDialogDismissed = useCallback(() => {
    setTodoDialogVisible(false);
  }, []);

  return (
    <div className={cn(className, styles.root)} onClick={onTodoButtonClicked}>
      <Dialog
        hidden={!todoDialogVisible}
        title={<FormattedMessage id="TodoButtonWrapper.dialog-title" />}
        subText={renderToString("TodoButtonWrapper.dialog-message")}
        onDismiss={onTodoDialogDismissed}
      >
        <DialogFooter>
          <PrimaryButton onClick={onTodoDialogDismissed}>
            <FormattedMessage id="confirm" />
          </PrimaryButton>
        </DialogFooter>
      </Dialog>
      <div style={{ pointerEvents: "none" }}>{children}</div>
    </div>
  );
};

export default TodoButtonWrapper;
