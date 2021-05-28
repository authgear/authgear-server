import React, { useCallback, useContext, useMemo, useState } from "react";
import cn from "classnames";
import { Dialog, DialogFooter, PrimaryButton } from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";

import styles from "./TodoButtonWrapper.module.scss";

interface TodoButtonWrapperProps {
  children: React.ReactNode;
  className?: string;
}

const TodoButtonWrapper: React.FC<TodoButtonWrapperProps> =
  function TodoButtonWrapper(props: TodoButtonWrapperProps) {
    const { children, className } = props;
    const { renderToString } = useContext(Context);

    const [todoDialogVisible, setTodoDialogVisible] = useState(false);

    const onTodoButtonClicked = useCallback(() => {
      setTodoDialogVisible(true);
    }, []);

    const onTodoDialogDismissed = useCallback(() => {
      setTodoDialogVisible(false);
    }, []);

    const todoDialogContentProps = useMemo(() => {
      return {
        title: <FormattedMessage id="TodoButtonWrapper.dialog-title" />,
        subText: renderToString("TodoButtonWrapper.dialog-message"),
      };
    }, [renderToString]);

    return (
      <div className={cn(className, styles.root)} onClick={onTodoButtonClicked}>
        <Dialog
          hidden={!todoDialogVisible}
          dialogContentProps={todoDialogContentProps}
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
