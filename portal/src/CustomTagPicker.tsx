import React, { useCallback, useMemo, useRef, useState } from "react";
import cn from "classnames";
import { Text } from "@radix-ui/themes";
import { Cross2Icon } from "@radix-ui/react-icons";
import { FormField } from "./components/v2/FormField/FormField";
import styles from "./CustomTagPicker.module.css";

export interface Tag {
  key: string | number;
  name: string;
}

export interface CustomTagPickerProps {
  className?: string;
  label?: React.ReactNode;
  disabled?: boolean;
  selectedItems?: Tag[];
  onChange?: (items?: Tag[]) => void;
  onResolveSuggestions: (filterText: string, tagList?: Tag[]) => Tag[];
  /** When provided, the raw input text is added on Enter / blur (free-text tags). */
  onAdd?: (item: string) => void;
  inputProps?: React.InputHTMLAttributes<HTMLInputElement>;
}

const CustomTagPicker: React.VFC<CustomTagPickerProps> =
  function CustomTagPicker(props: CustomTagPickerProps) {
    const {
      className,
      label,
      disabled,
      selectedItems = [],
      onChange,
      onResolveSuggestions,
      onAdd,
      inputProps,
    } = props;

    const inputRef = useRef<HTMLInputElement | null>(null);
    const [filterText, setFilterText] = useState("");
    const [isFocused, setIsFocused] = useState(false);
    const [activeIndex, setActiveIndex] = useState(0);

    const suggestions = useMemo(() => {
      if (filterText.trim() === "") {
        return [];
      }
      const selectedKeys = new Set(selectedItems.map((t) => t.key));
      return onResolveSuggestions(filterText, selectedItems).filter(
        (t) => !selectedKeys.has(t.key)
      );
    }, [filterText, onResolveSuggestions, selectedItems]);

    const isSuggestionsOpen = isFocused && suggestions.length > 0;

    const addTag = useCallback(
      (tag: Tag) => {
        onChange?.([...selectedItems, tag]);
        setFilterText("");
        setActiveIndex(0);
      },
      [onChange, selectedItems]
    );

    const removeTag = useCallback(
      (key: Tag["key"]) => {
        onChange?.(selectedItems.filter((t) => t.key !== key));
      },
      [onChange, selectedItems]
    );

    const onInputChange = useCallback(
      (e: React.ChangeEvent<HTMLInputElement>) => {
        setFilterText(e.currentTarget.value);
        setActiveIndex(0);
      },
      []
    );

    const onInputKeyDown = useCallback(
      (e: React.KeyboardEvent<HTMLInputElement>) => {
        switch (e.key) {
          case "Enter": {
            e.preventDefault();
            if (suggestions.length > 0) {
              addTag(
                suggestions[Math.min(activeIndex, suggestions.length - 1)]
              );
            } else if (onAdd != null && filterText !== "") {
              onAdd(filterText);
              setFilterText("");
              setActiveIndex(0);
            }
            break;
          }
          case "Backspace": {
            if (filterText === "" && selectedItems.length > 0) {
              removeTag(selectedItems[selectedItems.length - 1].key);
            }
            break;
          }
          case "ArrowDown": {
            if (suggestions.length > 0) {
              e.preventDefault();
              setActiveIndex((i) => Math.min(i + 1, suggestions.length - 1));
            }
            break;
          }
          case "ArrowUp": {
            if (suggestions.length > 0) {
              e.preventDefault();
              setActiveIndex((i) => Math.max(i - 1, 0));
            }
            break;
          }
          case "Escape": {
            setFilterText("");
            setActiveIndex(0);
            break;
          }
          default:
            break;
        }
      },
      [
        activeIndex,
        addTag,
        filterText,
        onAdd,
        removeTag,
        selectedItems,
        suggestions,
      ]
    );

    const onInputFocus = useCallback(() => {
      setIsFocused(true);
    }, []);

    const onInputBlur = useCallback(() => {
      setIsFocused(false);
      if (onAdd != null && filterText !== "") {
        onAdd(filterText);
      }
      setFilterText("");
      setActiveIndex(0);
    }, [filterText, onAdd]);

    const onFieldMouseDown = useCallback((e: React.MouseEvent) => {
      if (e.target !== inputRef.current) {
        e.preventDefault();
        inputRef.current?.focus();
      }
    }, []);

    const field = (
      <div className={cn(styles.root, className)}>
        <div
          className={cn(
            styles.field,
            disabled === true && styles.fieldDisabled
          )}
          onMouseDown={onFieldMouseDown}
        >
          {selectedItems.map((tag) => (
            <span key={tag.key} className={styles.tag}>
              <Text size="2">{tag.name}</Text>
              {disabled === true ? null : (
                <button
                  type="button"
                  className={styles.tagRemove}
                  aria-label={`Remove ${tag.name}`}
                  onClick={() => removeTag(tag.key)}
                >
                  <Cross2Icon width={12} height={12} />
                </button>
              )}
            </span>
          ))}
          <input
            {...inputProps}
            ref={inputRef}
            className={styles.input}
            type="text"
            disabled={disabled}
            value={filterText}
            onChange={onInputChange}
            onKeyDown={onInputKeyDown}
            onFocus={onInputFocus}
            onBlur={onInputBlur}
          />
        </div>
        {isSuggestionsOpen ? (
          <div className={styles.suggestions} role="listbox">
            {suggestions.map((tag, index) => (
              <div
                key={tag.key}
                role="option"
                aria-selected={index === activeIndex}
                className={styles.suggestion}
                data-active={index === activeIndex ? true : undefined}
                // Prevent the input blur so the click handler still fires.
                onMouseDown={(e) => e.preventDefault()}
                onMouseEnter={() => setActiveIndex(index)}
                onClick={() => addTag(tag)}
              >
                <Text size="2">{tag.name}</Text>
              </div>
            ))}
          </div>
        ) : null}
      </div>
    );

    if (label != null) {
      return (
        <FormField size="2" labelSize="2" labelSpace="1" label={label}>
          {field}
        </FormField>
      );
    }
    return field;
  };

export default CustomTagPicker;
