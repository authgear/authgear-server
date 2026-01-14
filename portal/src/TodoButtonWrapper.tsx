import React, { useCallback, useContext, useMemo, useState } from "react";
import cn from "classnames";
import { Dialog, DialogFooter } from "@fluentui/react";
import { Context, FormattedMessage } from "./intl";
import PrimaryButton from "./PrimaryButton";

import styles from "./TodoButtonWrapper.module.css";

interface TodoButtonWrapperProps {
  children: React.ReactNode;
  className?: string;
}

const TodoButtonWrapper: React.VFC<TodoButtonWrapperProps> =
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
            <PrimaryButton
              onClick={onTodoDialogDismissed}
              text={<FormattedMessage id="confirm" />}
            />
          </DialogFooter>
        </Dialog>
        <div style={{ pointerEvents: "none" }}>{children}</div>
      </div>
    );
  };

export default TodoButtonWrapper;
