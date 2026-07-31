import React, { useCallback, useContext, useState } from "react";
import cn from "classnames";
import {
  ChevronDownIcon,
  Cross2Icon,
  MagnifyingGlassIcon,
} from "@radix-ui/react-icons";
import {
  Popover,
  Spinner,
  Text,
  TextField as RadixTextField,
} from "@radix-ui/themes";
import { Context as MessageContext } from "../../intl";
import styles from "./UsersFilterDropdown.module.css";

export interface UsersFilterDropdownOption {
  key: string;
  text: string;
}

interface UsersFilterDropdownProps<T extends UsersFilterDropdownOption> {
  className?: string;
  placeholder: string;
  isLoadingOptions: boolean;
  options: T[];
  searchValue: string;
  onSearchValueChange: (value: string) => void;
  selectedItem: T | null;
  onChange: (option: T) => void;
  onClear: () => void;
}

export function UsersFilterDropdown<T extends UsersFilterDropdownOption>({
  className,
  placeholder,
  isLoadingOptions,
  options,
  searchValue,
  onSearchValueChange,
  selectedItem,
  onChange,
  onClear,
}: UsersFilterDropdownProps<T>): React.ReactElement {
  const { renderToString } = useContext(MessageContext);
  const [open, setOpen] = useState(false);

  const onSearchChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      onSearchValueChange(e.currentTarget.value);
    },
    [onSearchValueChange]
  );

  const onClearClick = useCallback(
    (e: React.MouseEvent<HTMLSpanElement>) => {
      e.preventDefault();
      e.stopPropagation();
      onClear();
    },
    [onClear]
  );

  return (
    <div className={cn(styles.root, className)}>
      <Popover.Root open={open} onOpenChange={setOpen}>
        <Popover.Trigger>
          <button type="button" className={styles.trigger}>
            <span
              className={cn(
                styles.triggerLabel,
                selectedItem == null && styles.triggerPlaceholder
              )}
            >
              {selectedItem?.text ?? placeholder}
            </span>
            {selectedItem != null ? (
              <span
                role="button"
                tabIndex={0}
                className={styles.triggerAction}
                aria-label={renderToString("clear")}
                onClick={onClearClick}
                onKeyDown={(e) => {
                  if (e.key === "Enter" || e.key === " ") {
                    e.preventDefault();
                    e.stopPropagation();
                    onClear();
                  }
                }}
              >
                <Cross2Icon />
              </span>
            ) : (
              <ChevronDownIcon className={styles.triggerIcon} />
            )}
          </button>
        </Popover.Trigger>
        <Popover.Content
          className={styles.content}
          sideOffset={4}
          align="start"
        >
          <RadixTextField.Root
            size="2"
            type="search"
            value={searchValue}
            placeholder={renderToString("search")}
            onChange={onSearchChange}
          >
            <RadixTextField.Slot side="left">
              <MagnifyingGlassIcon />
            </RadixTextField.Slot>
          </RadixTextField.Root>
          <div className={styles.options}>
            {isLoadingOptions ? (
              <div className={styles.feedback}>
                <Spinner size="1" />
              </div>
            ) : options.length === 0 ? (
              <Text size="2" color="gray" className={styles.feedback}>
                {renderToString("SearchableDropdown.empty")}
              </Text>
            ) : (
              options.map((option) => (
                <button
                  key={option.key}
                  type="button"
                  className={cn(
                    styles.option,
                    selectedItem?.key === option.key && styles.optionSelected
                  )}
                  onClick={() => {
                    onChange(option);
                    setOpen(false);
                  }}
                >
                  <Text size="2" className={styles.optionText}>
                    {option.text}
                  </Text>
                </button>
              ))
            )}
          </div>
        </Popover.Content>
      </Popover.Root>
    </div>
  );
}
