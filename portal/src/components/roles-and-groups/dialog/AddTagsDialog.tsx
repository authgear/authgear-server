import {
  Dialog,
  DialogFooter,
  IDialogContentProps,
  IModalProps,
  ITag,
  Label,
} from "@fluentui/react";
import { FormattedMessage } from "@oursky/react-messageformat";
import React, { ReactElement, useCallback, useMemo, useState } from "react";
import PrimaryButton from "../../../PrimaryButton";
import DefaultButton from "../../../DefaultButton";
import StyledTagPicker from "../../../StyledTagPicker";
import styles from "./AddTagsDialog.module.css";

interface AddTagsDialogProps {
  isHidden: boolean;
  isLoading: boolean;

  title: string;
  tagPickerLabel: string;
  onResolveSuggestions: (
    filter: string,
    selectedTags?: ITag[]
  ) => ITag[] | Promise<ITag[]>;
  onSubmit?: (tags: ITag[]) => void;
  onDismiss: () => void;
  onDismissed?: () => void;
}

const dialogStyles = { main: { minHeight: 0 } };

function AddTagsDialog({
  isHidden,
  isLoading,
  onSubmit,
  onResolveSuggestions,
  onDismiss,
  onDismissed: propsOnDismissed,
  title,
  tagPickerLabel,
}: AddTagsDialogProps): ReactElement {
  const [tags, setTags] = useState<ITag[]>([]);

  const onChangeTags = useCallback((tags?: ITag[]) => {
    if (tags === undefined) {
      setTags([]);
    } else {
      setTags(tags);
    }
  }, []);
  const [searchKeyword, setSearchKeyword] = useState<string>("");
  const onSearchInputChange = useCallback((value: string): string => {
    setSearchKeyword(value);
    return value;
  }, []);
  const onClearTags = useCallback(() => setTags([]), []);

  const onDialogDismiss = useCallback(() => {
    if (isLoading || isHidden) {
      return;
    }
    onDismiss();
  }, [isHidden, isLoading, onDismiss]);

  const modalProps = useMemo((): IModalProps => {
    return {
      onDismissed: () => {
        // Reset states on dismiss
        setTags([]);

        propsOnDismissed?.();
      },
    };
  }, [propsOnDismissed]);

  const onClick = useCallback(() => {
    onSubmit?.(tags);
  }, [onSubmit, tags]);

  const dialogContentProps: IDialogContentProps = useMemo(() => {
    return {
      title: title,
    };
  }, [title]);

  const onEmptyResolveSuggestions = useCallback(
    async (selectedTags?: ITag[]) => onResolveSuggestions("", selectedTags),
    [onResolveSuggestions]
  );

  return (
    <>
      <Dialog
        hidden={isHidden}
        onDismiss={onDialogDismiss}
        modalProps={modalProps}
        dialogContentProps={dialogContentProps}
        styles={dialogStyles}
        maxWidth="560px"
      >
        <div className={styles.content}>
          <div className={styles.field}>
            <Label>{tagPickerLabel}</Label>
            <StyledTagPicker
              autoFocus={true}
              value={searchKeyword}
              onInputChange={onSearchInputChange}
              selectedItems={tags}
              onChange={onChangeTags}
              onResolveSuggestions={onResolveSuggestions}
              onEmptyResolveSuggestions={onEmptyResolveSuggestions}
              onClearTags={onClearTags}
            />
          </div>
        </div>
        <DialogFooter>
          <PrimaryButton
            disabled={isLoading}
            onClick={onClick}
            text={<FormattedMessage id="add" />}
          />
          <DefaultButton
            onClick={onDialogDismiss}
            disabled={isLoading}
            text={<FormattedMessage id="cancel" />}
          />
        </DialogFooter>
      </Dialog>
    </>
  );
}

export default AddTagsDialog;
