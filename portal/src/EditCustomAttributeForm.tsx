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
import {
  NumberRangeFields,
  StringConstraintsFields,
  ValidationSection,
} from "./StringConstraintsFields";
import {
  StringConstraintsDraft,
  draftFromConstraints,
  patternErrorKey,
} from "./util/stringConstraints";
import { TranslationValuesForm } from "./hook/useTranslationValuesForm";

export interface CustomAttributeDraft {
  pointer: string;
  type: CustomAttributeType | "";
  minimum: string;
  maximum: string;
  enum: string[];
  // Only for type string.
  constraints: StringConstraintsDraft;
}

export const EMPTY_CONSTRAINTS: StringConstraintsDraft =
  draftFromConstraints(undefined);

export interface EditCustomAttributeFormProps {
  className?: string;
  index: number;
  draft: CustomAttributeDraft;
  mode: "new" | "edit";
  onChangeDraft: (draft: CustomAttributeDraft) => void;
  patternMessage: string;
  onPatternMessageChange: (message: string) => void;
}

// Keeps the pattern's message with the attribute: it follows a renamed
// attribute, and goes when the pattern or the string type does.
export function syncPatternMessage(
  prev: CustomAttributeDraft,
  next: CustomAttributeDraft,
  translations: TranslationValuesForm
): void {
  const oldKey = patternErrorKey("custom", prev.pointer);
  const message = translations.getValue(oldKey);
  if (message === "") {
    return;
  }
  if (next.type !== "string" || next.constraints.pattern === "") {
    translations.setValue(oldKey, "");
    return;
  }
  const newKey = patternErrorKey("custom", next.pointer);
  if (newKey !== oldKey) {
    translations.setValue(oldKey, "");
    translations.setValue(newKey, message);
  }
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
    <section className={styles.panel}>
      <Text as="p" size="3" weight="bold">
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
    </section>
  );
}

// ---------------------------------------------------------------------------
// Main form
// ---------------------------------------------------------------------------

const EditCustomAttributeForm: React.VFC<EditCustomAttributeFormProps> =
  function EditCustomAttributeForm(props: EditCustomAttributeFormProps) {
    const {
      className,
      draft,
      index,
      onChangeDraft,
      mode,
      patternMessage,
      onPatternMessageChange,
    } = props;
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

    // Boolean attributes are not offered when creating, but an existing one keeps its type shown.
    const selectableTypes = useMemo(
      () =>
        customAttributeTypes.filter(
          (key) => key !== "boolean" || draft.type === "boolean"
        ),
      [draft.type]
    );

    const onChangeType = useCallback(
      (value: string) => {
        if (!isCustomAttributeType(value)) {
          return;
        }
        // Only a string carries constraints; the schema refuses them on
        // other types.
        onChangeDraft({
          ...draft,
          type: value,
          constraints:
            value === "string" ? draft.constraints : EMPTY_CONSTRAINTS,
        });
      },
      [draft, onChangeDraft]
    );

    return (
      <div className={cn(styles.tableWrapper, className)}>
        <section className={styles.panel}>
          <Text as="p" size="3" weight="bold">
            <FormattedMessage id="EditCustomAttributeForm.section.details" />
          </Text>
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
                {selectableTypes.map((key) => (
                  <Select.Item key={key} value={key}>
                    {renderToString("custom-attribute-type." + key)}
                  </Select.Item>
                ))}
              </Select.Content>
            </Select.Root>
          </div>
        </section>

        {draft.type === "number" || draft.type === "integer" ? (
          <ValidationSection>
            <NumberRangeFields
              minimum={draft.minimum}
              maximum={draft.maximum}
              onChange={(minimum, maximum) =>
                onChangeDraft({ ...draft, minimum, maximum })
              }
              checkInput={
                draft.type === "integer" ? checkIntegerInput : checkNumberInput
              }
              parentJSONPointer={parentJSONPointer}
            />
          </ValidationSection>
        ) : null}

        {draft.type === "string" ? (
          <ValidationSection>
            <StringConstraintsFields
              draft={draft.constraints}
              onChange={(constraints) =>
                onChangeDraft({ ...draft, constraints })
              }
              patternMessage={patternMessage}
              onPatternMessageChange={onPatternMessageChange}
              parentJSONPointer={parentJSONPointer}
            />
          </ValidationSection>
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
