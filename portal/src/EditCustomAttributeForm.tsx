import React, { useMemo, useContext, useCallback, useState } from "react";
import {
  Dropdown,
  DetailsList,
  DetailsRow,
  Label,
  IconButton,
  Icon,
  SelectionMode,
  Text,
  IDropdownOption,
  IDragDropEvents,
  IRenderFunction,
  IDetailsRowProps,
} from "@fluentui/react";
import { Context, FormattedMessage } from "./intl";
import Widget from "./Widget";
import FormTextField from "./FormTextField";
import PrimaryButton from "./PrimaryButton";
import { parseJSONPointer, jsonPointerToString } from "./util/jsonpointer";
import { checkNumberInput, checkIntegerInput } from "./util/input";
import {
  customAttributeTypes,
  CustomAttributeType,
  isCustomAttributeType,
} from "./types";
import { useSystemConfig } from "./context/SystemConfigContext";
import FormErrorMessageText from "./FormErrorMessageText";
import styles from "./EditCustomAttributeForm.module.css";
import { makeValidationErrorCustomMessageIDRule } from "./error/parse";

const REMOVE_BUTTON_ICON_PROPS = {
  iconName: "Blocked12",
};

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

interface CustomAttributeTypeNumberOptionProps {
  parentJSONPointer: string;
  draft: CustomAttributeDraft;
  onChangeDraft: (draft: CustomAttributeDraft) => void;
  checkFunction: (value: string) => boolean;
}

function CustomAttributeTypeNumberOption(
  props: CustomAttributeTypeNumberOptionProps
) {
  const { parentJSONPointer, draft, onChangeDraft, checkFunction } = props;
  const { renderToString } = useContext(Context);

  const onChangeMinimum = useCallback(
    (_e: React.FormEvent<unknown>, newValue?: string) => {
      if (newValue == null) {
        return;
      }
      if (!checkFunction(newValue)) {
        return;
      }
      onChangeDraft({
        ...draft,
        minimum: newValue,
      });
    },
    [draft, onChangeDraft, checkFunction]
  );

  const onChangeMaximum = useCallback(
    (_e: React.FormEvent<unknown>, newValue?: string) => {
      if (newValue == null) {
        return;
      }
      if (!checkFunction(newValue)) {
        return;
      }
      onChangeDraft({
        ...draft,
        maximum: newValue,
      });
    },
    [draft, onChangeDraft, checkFunction]
  );

  return (
    <div className={styles.numberOptionContainer}>
      <Label className={styles.numberOptionLabel}>
        <FormattedMessage id="EditCustomAttributeForm.label.options" />
      </Label>
      <FormTextField
        parentJSONPointer={parentJSONPointer}
        fieldName="minimum"
        className={styles.numberOptionMin}
        prefix={renderToString("EditCustomAttributeForm.label.min")}
        value={draft.minimum}
        onChange={onChangeMinimum}
      />
      <FormTextField
        parentJSONPointer={parentJSONPointer}
        fieldName="maximum"
        className={styles.numberOptionMax}
        prefix={renderToString("EditCustomAttributeForm.label.max")}
        value={draft.maximum}
        onChange={onChangeMaximum}
      />
    </div>
  );
}

interface CustomAttributeTypeEnumOptionProps {
  parentJSONPointer: string;
  draft: CustomAttributeDraft;
  onChangeDraft: (draft: CustomAttributeDraft) => void;
}

function CustomAttributeTypeEnumOption(
  props: CustomAttributeTypeEnumOptionProps
) {
  const { parentJSONPointer, draft, onChangeDraft } = props;
  const { enum: items } = draft;
  const { renderToString } = useContext(Context);
  const { themes } = useSystemConfig();
  const { destructive } = themes;
  const [dndIndex, setDNDIndex] = useState<number | undefined>(undefined);
  const [value, setValue] = useState("");
  const onChangeValue = useCallback(
    (_e: React.FormEvent<unknown>, newValue?: string) => {
      if (newValue != null) {
        setValue(newValue);
      }
    },
    []
  );
  const onClickAdd = useCallback(
    (e: React.MouseEvent<unknown>) => {
      e.preventDefault();
      e.stopPropagation();
      onChangeDraft({
        ...draft,
        enum: [...draft.enum, value],
      });
      setValue("");
    },
    [draft, onChangeDraft, value]
  );

  const onRenderValue = useCallback(
    (item?: string, index?: number) => {
      if (item == null || index == null) {
        return null;
      }
      return (
        <div className={styles.itemCell}>
          <Text className={styles.enumValue} block={true}>
            {item}
          </Text>
          <FormErrorMessageText
            block={true}
            parentJSONPointer={parentJSONPointer + "/enum"}
            fieldName={String(index)}
          />
        </div>
      );
    },
    [parentJSONPointer]
  );
  const onRenderRemoveButton = useCallback(
    (_item?: string, index?: number) => {
      if (index == null) {
        return null;
      }
      const onClick = (e: React.MouseEvent<unknown>) => {
        e.preventDefault();
        e.stopPropagation();
        onChangeDraft({
          ...draft,
          enum: draft.enum.filter((_, i) => i !== index),
        });
      };
      return (
        <IconButton
          theme={destructive}
          iconProps={REMOVE_BUTTON_ICON_PROPS}
          title={renderToString("remove")}
          ariaLabel={renderToString("remove")}
          onClick={onClick}
          disabled={items.length <= 1}
        />
      );
    },
    [draft, items, onChangeDraft, renderToString, destructive]
  );
  const onRenderReorderHandle = useCallback(() => {
    return (
      <div className={styles.reorderHandle}>
        <Icon iconName="GlobalNavButton" />
      </div>
    );
  }, []);
  const columns = useMemo(() => {
    return [
      {
        key: "value",
        name: "",
        minWidth: 0,
        onRender: onRenderValue,
      },
      {
        key: "remove",
        name: "",
        minWidth: 24,
        maxWidth: 24,
        onRender: onRenderRemoveButton,
      },
      {
        key: "reorder",
        name: "",
        minWidth: 24,
        maxWidth: 24,
        onRender: onRenderReorderHandle,
      },
    ];
  }, [onRenderValue, onRenderRemoveButton, onRenderReorderHandle]);

  const reorder = useCallback(
    (index: number, item: string) => {
      const itemsWithoutIndex = [
        ...items.slice(0, index),
        ...items.slice(index + 1),
      ];
      const insertIndex = items.indexOf(item);
      if (insertIndex >= 0) {
        itemsWithoutIndex.splice(insertIndex, 0, items[index]);
        onChangeDraft({
          ...draft,
          enum: itemsWithoutIndex,
        });
      }
    },
    [items, draft, onChangeDraft]
  );

  const dragDropEvents: IDragDropEvents = useMemo(() => {
    return {
      canDrop: () => true,
      canDrag: () => true,
      onDragEnter: () => styles.onDragEnter,
      onDragLeave: () => {},
      onDragStart: (_item?: string, index?: number) => {
        if (index != null) {
          setDNDIndex(index);
        }
      },
      onDragEnd: (_item?: string) => {
        setDNDIndex(undefined);
      },
      onDrop: (item?: string) => {
        if (dndIndex != null && item != null) {
          reorder(dndIndex, item);
        }
      },
    };
  }, [reorder, dndIndex]);

  const onRenderRow: IRenderFunction<IDetailsRowProps> = useCallback(
    (props?: IDetailsRowProps) => {
      if (props == null) {
        return null;
      }
      let className = "";
      const { itemIndex } = props;
      if (dndIndex != null) {
        if (itemIndex < dndIndex) {
          className = styles.before;
        } else if (itemIndex > dndIndex) {
          className = styles.after;
        }
      }
      return <DetailsRow {...props} className={className} />;
    },
    [dndIndex]
  );

  return (
    <div className={styles.enumOptionContainer}>
      <Label className={styles.enumOptionLabel}>
        <FormattedMessage id="EditCustomAttributeForm.label.options" />
      </Label>
      <FormTextField
        parentJSONPointer={parentJSONPointer}
        fieldName="enum"
        className={styles.enumOptionTextField}
        value={value}
        onChange={onChangeValue}
      />
      <PrimaryButton
        className={styles.enumOptionAddButton}
        onClick={onClickAdd}
        text={renderToString("add")}
        disabled={value === ""}
      />
      <div className={styles.enumOptionList}>
        <DetailsList
          columns={columns}
          items={items}
          selectionMode={SelectionMode.none}
          onRenderRow={onRenderRow}
          isHeaderVisible={false}
          dragDropEvents={dragDropEvents}
        />
      </div>
    </div>
  );
}

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
      (_e: React.FormEvent<unknown>, newValue?: string) => {
        if (newValue == null) {
          return;
        }
        onChangeDraft({
          ...draft,
          pointer: jsonPointerToString([newValue]),
        });
      },
      [draft, onChangeDraft]
    );

    const typeOptions: IDropdownOption[] = useMemo(() => {
      return customAttributeTypes.map((key) => {
        return {
          key,
          text: renderToString("custom-attribute-type." + key),
        };
      });
    }, [renderToString]);

    const onChangeType = useCallback(
      (_e: React.FormEvent<unknown>, option?: IDropdownOption) => {
        if (option == null) {
          return;
        }
        if (!isCustomAttributeType(option.key)) {
          return;
        }
        onChangeDraft({
          ...draft,
          type: option.key,
        });
      },
      [draft, onChangeDraft]
    );

    return (
      <Widget className={className}>
        <FormTextField
          parentJSONPointer={parentJSONPointer}
          fieldName="pointer"
          required={true}
          value={fieldName}
          onChange={onChangeFieldName}
          label={renderToString("EditCustomAttributeForm.label.attribute-name")}
          description={renderToString(
            "EditCustomAttributeForm.description.attribute-name"
          )}
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
        <Dropdown
          selectedKey={draft.type}
          options={typeOptions}
          label={renderToString("EditCustomAttributeForm.label.type")}
          onChange={onChangeType}
          disabled={mode === "edit"}
        />
        {draft.type === "number" ? (
          <CustomAttributeTypeNumberOption
            parentJSONPointer={parentJSONPointer}
            draft={draft}
            onChangeDraft={onChangeDraft}
            checkFunction={checkNumberInput}
          />
        ) : null}
        {draft.type === "integer" ? (
          <CustomAttributeTypeNumberOption
            parentJSONPointer={parentJSONPointer}
            draft={draft}
            onChangeDraft={onChangeDraft}
            checkFunction={checkIntegerInput}
          />
        ) : null}
        {draft.type === "enum" ? (
          <CustomAttributeTypeEnumOption
            parentJSONPointer={parentJSONPointer}
            draft={draft}
            onChangeDraft={onChangeDraft}
          />
        ) : null}
      </Widget>
    );
  };

export default EditCustomAttributeForm;
