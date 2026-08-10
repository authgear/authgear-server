import React, { useContext, useCallback, useMemo, useId } from "react";
import cn from "classnames";
import { Select, Text, TextField as RadixTextField } from "@radix-ui/themes";
import { PlusIcon } from "@radix-ui/react-icons";
import { Context, FormattedMessage } from "./intl";
import { TextField } from "./components/v2/TextField/TextField";
import { PrimaryButton } from "./components/v2/Button/PrimaryButton/PrimaryButton";
import {
  IconButton,
  IconButtonIcon,
} from "./components/v2/IconButton/IconButton";
import { parseJSONPointer, jsonPointerToString } from "./util/jsonpointer";
import { checkNumberInput, checkIntegerInput } from "./util/input";
import {
  customAttributeTypes,
  CustomAttributeType,
  isCustomAttributeType,
} from "./types";
import styles from "./EditCustomAttributeForm.module.css";
import { makeValidationErrorCustomMessageIDRule } from "./error/parse";

export interface CustomAttributeDraft {
  pointer: string;
  type: CustomAttributeType | "";
  minimum: string;
  maximum: string;
  enum: string[];
}

export interface EditCustomAttributeFormProps {
  className?: string;
  index: number;
  draft: CustomAttributeDraft;
  mode: "new" | "edit";
  onChangeDraft: (draft: CustomAttributeDraft) => void;
}

// ---------------------------------------------------------------------------
// Number option (min / max)
// ---------------------------------------------------------------------------

interface NumberOptionProps {
  parentJSONPointer: string;
  draft: CustomAttributeDraft;
  onChangeDraft: (draft: CustomAttributeDraft) => void;
  checkFunction: (value: string) => boolean;
}

function NumberOption({
  parentJSONPointer,
  draft,
  onChangeDraft,
  checkFunction,
}: NumberOptionProps) {
  const onChangeMinimum = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const newValue = e.target.value;
      if (!checkFunction(newValue)) {
        return;
      }
      onChangeDraft({ ...draft, minimum: newValue });
    },
    [draft, onChangeDraft, checkFunction]
  );

  const onChangeMaximum = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const newValue = e.target.value;
      if (!checkFunction(newValue)) {
        return;
      }
      onChangeDraft({ ...draft, maximum: newValue });
    },
    [draft, onChangeDraft, checkFunction]
  );

  return (
    <div className={styles.tableRow}>
      <div className={styles.numberRangeRow}>
        <div className={styles.fieldBlock}>
          <Text as="p" size="2" className={styles.fieldLabel}>
            <FormattedMessage id="EditCustomAttributeForm.label.min" />
          </Text>
          <TextField
            size="2"
            type="text"
            value={draft.minimum}
            onChange={onChangeMinimum}
            parentJSONPointer={parentJSONPointer}
            fieldName="minimum"
          />
        </div>
        <div className={styles.fieldBlock}>
          <Text as="p" size="2" className={styles.fieldLabel}>
            <FormattedMessage id="EditCustomAttributeForm.label.max" />
          </Text>
          <TextField
            size="2"
            type="text"
            value={draft.maximum}
            onChange={onChangeMaximum}
            parentJSONPointer={parentJSONPointer}
            fieldName="maximum"
          />
        </div>
      </div>
    </div>
  );
}

// ---------------------------------------------------------------------------
// Enum option
// ---------------------------------------------------------------------------

interface EnumOptionProps {
  parentJSONPointer: string;
  draft: CustomAttributeDraft;
  onChangeDraft: (draft: CustomAttributeDraft) => void;
}

function EnumOption({ draft, onChangeDraft }: EnumOptionProps) {
  const addInputId = useId();
  const [addValue, setAddValue] = React.useState("");

  const onChangeAddValue = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      setAddValue(e.target.value);
    },
    []
  );

  const onClickAdd = useCallback(
    (e: React.MouseEvent<unknown>) => {
      e.preventDefault();
      e.stopPropagation();
      if (addValue.trim() === "") return;
      onChangeDraft({ ...draft, enum: [...draft.enum, addValue.trim()] });
      setAddValue("");
    },
    [draft, onChangeDraft, addValue]
  );

  const onKeyDown = useCallback(
    (e: React.KeyboardEvent<HTMLInputElement>) => {
      if (e.key === "Enter") {
        e.preventDefault();
        if (addValue.trim() === "") return;
        onChangeDraft({ ...draft, enum: [...draft.enum, addValue.trim()] });
        setAddValue("");
      }
    },
    [draft, onChangeDraft, addValue]
  );

  const makeOnRemove = useCallback(
    (index: number) => (e: React.MouseEvent<HTMLButtonElement>) => {
      e.preventDefault();
      e.stopPropagation();
      onChangeDraft({
        ...draft,
        enum: draft.enum.filter((_, i) => i !== index),
      });
    },
    [draft, onChangeDraft]
  );

  return (
    <div className={styles.fieldBlock}>
      <Text as="p" size="2" className={styles.fieldLabel}>
        <FormattedMessage id="EditCustomAttributeForm.label.options" />
      </Text>
      <div className={styles.enumContainer}>
        {/* Existing enum values */}
        {draft.enum.length > 0 ? (
          <div className={styles.enumList}>
            {draft.enum.map((value, i) => (
              <div key={i} className={styles.enumItem}>
                <Text
                  as="p"
                  size="2"
                  weight="medium"
                  className={styles.enumValue}
                >
                  {value}
                </Text>
                <IconButton
                  variant="destroy"
                  size="1"
                  icon={IconButtonIcon.Trash}
                  onClick={makeOnRemove(i)}
                />
              </div>
            ))}
          </div>
        ) : null}

        {/* Add new value */}
        <div className={styles.enumAddRow}>
          <RadixTextField.Root
            id={addInputId}
            size="2"
            variant="surface"
            value={addValue}
            onChange={onChangeAddValue}
            onKeyDown={onKeyDown}
          />
          <div className={styles.enumAddActions}>
            <PrimaryButton
              size="2"
              disabled={addValue.trim() === ""}
              onClick={onClickAdd}
              text={
                <span className={styles.addButtonContent}>
                  <PlusIcon />
                  <FormattedMessage id="add" />
                </span>
              }
            />
          </div>
        </div>
      </div>
    </div>
  );
}

// ---------------------------------------------------------------------------
// Main form
// ---------------------------------------------------------------------------

const EditCustomAttributeForm: React.VFC<EditCustomAttributeFormProps> =
  function EditCustomAttributeForm(props: EditCustomAttributeFormProps) {
    const { className, draft, index, onChangeDraft, mode } = props;
    const { renderToString } = useContext(Context);

    const parentJSONPointer = useMemo(() => {
      return "/user_profile/custom_attributes/attributes/" + String(index);
    }, [index]);

    const fieldName = useMemo(() => {
      if (draft.pointer === "") {
        return "";
      }
      return parseJSONPointer(draft.pointer)[0];
    }, [draft]);

    const onChangeFieldName = useCallback(
      (e: React.ChangeEvent<HTMLInputElement>) => {
        const newValue = e.target.value;
        onChangeDraft({
          ...draft,
          pointer: jsonPointerToString([newValue]),
        });
      },
      [draft, onChangeDraft]
    );

    const onChangeType = useCallback(
      (value: string) => {
        if (!isCustomAttributeType(value)) {
          return;
        }
        onChangeDraft({ ...draft, type: value });
      },
      [draft, onChangeDraft]
    );

    return (
      <div className={cn(styles.tableWrapper, className)}>
        <div className={styles.topFields}>
          <div className={styles.fieldBlock}>
            <Text as="p" size="2" className={styles.fieldLabel}>
              <FormattedMessage id="EditCustomAttributeForm.label.attribute-name" />
            </Text>
            <TextField
              size="2"
              type="text"
              required={true}
              value={fieldName}
              onChange={onChangeFieldName}
              parentJSONPointer={parentJSONPointer}
              fieldName="pointer"
              errorRules={[
                makeValidationErrorCustomMessageIDRule(
                  "not",
                  /\/pointer$/,
                  "EditCustomAttributeForm.error.not"
                ),
                makeValidationErrorCustomMessageIDRule(
                  "duplicated",
                  /\/pointer$/,
                  "EditCustomAttributeForm.error.duplicated-attribute-name"
                ),
              ]}
            />
            <Text
              as="p"
              size="1"
              color="gray"
              className={styles.fieldDescription}
            >
              <FormattedMessage id="EditCustomAttributeForm.description.attribute-name" />
            </Text>
          </div>
          <div className={styles.fieldBlock}>
            <Text as="p" size="2" className={styles.fieldLabel}>
              <FormattedMessage id="EditCustomAttributeForm.label.type" />
            </Text>
            <Select.Root
              value={draft.type}
              onValueChange={onChangeType}
              disabled={mode === "edit"}
            >
              <Select.Trigger variant="surface" />
              <Select.Content>
                {customAttributeTypes.map((key) => (
                  <Select.Item key={key} value={key}>
                    {renderToString("custom-attribute-type." + key)}
                  </Select.Item>
                ))}
              </Select.Content>
            </Select.Root>
          </div>
        </div>

        {draft.type === "number" || draft.type === "integer" ? (
          <div className={styles.table}>
            {draft.type === "number" ? (
              <NumberOption
                parentJSONPointer={parentJSONPointer}
                draft={draft}
                onChangeDraft={onChangeDraft}
                checkFunction={checkNumberInput}
              />
            ) : null}
            {draft.type === "integer" ? (
              <NumberOption
                parentJSONPointer={parentJSONPointer}
                draft={draft}
                onChangeDraft={onChangeDraft}
                checkFunction={checkIntegerInput}
              />
            ) : null}
          </div>
        ) : null}

        {draft.type === "enum" ? (
          <EnumOption
            parentJSONPointer={parentJSONPointer}
            draft={draft}
            onChangeDraft={onChangeDraft}
          />
        ) : null}
      </div>
    );
  };

export default EditCustomAttributeForm;
