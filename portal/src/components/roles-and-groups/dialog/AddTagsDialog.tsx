import React, {
  ReactElement,
  useCallback,
  useEffect,
  useId,
  useLayoutEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import { Cross2Icon } from "@radix-ui/react-icons";
import { Dialog, Flex, Text } from "@radix-ui/themes";
import { FormattedMessage } from "../../../intl";
import { PrimaryButton } from "../../v2/Button/PrimaryButton/PrimaryButton";
import { SecondaryButton } from "../../v2/Button/SecondaryButton/SecondaryButton";
import { FormField } from "../../v2/FormField/FormField";
import { useDebounced } from "../../../hook/useDebounced";
import styles from "./AddTagsDialog.module.css";

export interface AddTagsDialogTag {
  key: string;
  name: string;
}

interface AddTagsDialogProps {
  isHidden: boolean;
  isLoading: boolean;

  title: string;
  tagPickerLabel: string;
  onResolveSuggestions: (
    filter: string,
    selectedTags?: AddTagsDialogTag[]
  ) => AddTagsDialogTag[] | Promise<AddTagsDialogTag[]>;
  onSubmit?: (tags: AddTagsDialogTag[]) => void;
  onDismiss: () => void;
  onDismissed?: () => void;
}

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
  const inputId = useId();
  const inputRef = useRef<HTMLInputElement | null>(null);
  const containerRef = useRef<HTMLDivElement | null>(null);
  const wasOpenRef = useRef(!isHidden);

  const [tags, setTags] = useState<AddTagsDialogTag[]>([]);
  const [searchKeyword, setSearchKeyword] = useState("");
  const [suggestions, setSuggestions] = useState<AddTagsDialogTag[]>([]);
  const [suggestionsOpen, setSuggestionsOpen] = useState(false);
  const [suggestionsLoading, setSuggestionsLoading] = useState(false);
  const [suggestionPos, setSuggestionPos] = useState<{
    top: number;
    left: number;
    width: number;
  }>({ top: 0, left: 0, width: 0 });

  const [debouncedSearchKeyword] = useDebounced(searchKeyword, 200);
  const hasSearchKeyword = debouncedSearchKeyword.trim() !== "";

  // Recalculate the suggestion dropdown position whenever it opens.
  useLayoutEffect(() => {
    if (!suggestionsOpen || !containerRef.current) {
      return;
    }
    const rect = containerRef.current.getBoundingClientRect();
    setSuggestionPos({
      top: rect.bottom + 4,
      left: rect.left,
      width: rect.width,
    });
  }, [suggestionsOpen]);

  const resetState = useCallback(() => {
    setTags([]);
    setSearchKeyword("");
    setSuggestions([]);
    setSuggestionsOpen(false);
    setSuggestionsLoading(false);
  }, []);

  useEffect(() => {
    if (wasOpenRef.current && isHidden) {
      resetState();
      propsOnDismissed?.();
    }
    wasOpenRef.current = !isHidden;
  }, [isHidden, propsOnDismissed, resetState]);

  const suggestionsActive = !isHidden && hasSearchKeyword;

  // Adjust the dropdown state during render whenever the suggestion query
  // changes: open it in a loading state when a query becomes active, and
  // clear it when the query becomes inactive. The effect below only
  // performs the asynchronous fetch.
  const fetchKey = useMemo(
    () => ({
      suggestionsActive,
      debouncedSearchKeyword,
      tags,
      onResolveSuggestions,
    }),
    [suggestionsActive, debouncedSearchKeyword, tags, onResolveSuggestions]
  );
  const [prevFetchKey, setPrevFetchKey] = useState(fetchKey);
  if (prevFetchKey !== fetchKey) {
    setPrevFetchKey(fetchKey);
    if (suggestionsActive) {
      setSuggestionsLoading(true);
      setSuggestionsOpen(true);
    } else {
      setSuggestions([]);
      setSuggestionsLoading(false);
      if (!hasSearchKeyword) {
        setSuggestionsOpen(false);
      }
    }
  }

  useEffect(() => {
    if (isHidden || !hasSearchKeyword) {
      return;
    }

    let cancelled = false;
    Promise.resolve(onResolveSuggestions(debouncedSearchKeyword, tags)).then(
      (result) => {
        if (!cancelled) {
          setSuggestions(result);
          setSuggestionsLoading(false);
        }
      },
      () => {
        if (!cancelled) {
          setSuggestions([]);
          setSuggestionsLoading(false);
        }
      }
    );

    return () => {
      cancelled = true;
    };
  }, [
    debouncedSearchKeyword,
    hasSearchKeyword,
    isHidden,
    onResolveSuggestions,
    tags,
  ]);

  const onDialogDismiss = useCallback(() => {
    if (isLoading || isHidden) {
      return;
    }
    onDismiss();
  }, [isHidden, isLoading, onDismiss]);

  const onOpenChange = useCallback(
    (open: boolean) => {
      if (!open) {
        onDialogDismiss();
      }
    },
    [onDialogDismiss]
  );

  const onClickAdd = useCallback(() => {
    onSubmit?.(tags);
  }, [onSubmit, tags]);

  const onRemoveTag = useCallback((key: string) => {
    setTags((prev) => prev.filter((tag) => tag.key !== key));
  }, []);

  const onClearTags = useCallback(() => {
    setTags([]);
    setSearchKeyword("");
    inputRef.current?.focus();
  }, []);

  const onSelectSuggestion = useCallback((tag: AddTagsDialogTag) => {
    setTags((prev) => {
      if (prev.some((item) => item.key === tag.key)) {
        return prev;
      }
      return [...prev, tag];
    });
    setSearchKeyword("");
    setSuggestions([]);
    setSuggestionsOpen(false);
    inputRef.current?.focus();
  }, []);

  const onSearchChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const nextValue = e.target.value;
      setSearchKeyword(nextValue);
      if (nextValue.trim() === "") {
        setSuggestions([]);
        setSuggestionsOpen(false);
        setSuggestionsLoading(false);
      } else {
        setSuggestionsLoading(true);
      }
    },
    []
  );

  const onInputKeyDown = useCallback(
    (e: React.KeyboardEvent<HTMLInputElement>) => {
      if (e.key === "Backspace" && searchKeyword === "" && tags.length > 0) {
        e.preventDefault();
        setTags((prev) => prev.slice(0, -1));
        return;
      }
      if (e.key === "Escape") {
        setSuggestionsOpen(false);
        return;
      }
      if (e.key === "Enter" && suggestions.length > 0) {
        e.preventDefault();
        onSelectSuggestion(suggestions[0]);
      }
    },
    [onSelectSuggestion, searchKeyword, suggestions, tags.length]
  );

  const canSubmit = tags.length > 0 && !isLoading;

  const suggestionItems = useMemo(() => {
    if (suggestionsLoading) {
      return (
        <Text as="p" size="2" color="gray" className={styles.suggestionEmpty}>
          <FormattedMessage id="loading" />
        </Text>
      );
    }
    if (suggestions.length === 0) {
      return (
        <Text as="p" size="2" color="gray" className={styles.suggestionEmpty}>
          <FormattedMessage id="AddTagsDialog.no-results" />
        </Text>
      );
    }
    return suggestions.map((suggestion) => (
      <button
        key={suggestion.key}
        type="button"
        className={styles.suggestionItem}
        onMouseDown={(e) => {
          // Prevent the input from losing focus before the click fires.
          e.preventDefault();
        }}
        onClick={() => {
          onSelectSuggestion(suggestion);
        }}
      >
        <Text size="2">{suggestion.name}</Text>
      </button>
    ));
  }, [onSelectSuggestion, suggestions, suggestionsLoading]);

  return (
    <Dialog.Root open={!isHidden} onOpenChange={onOpenChange}>
      <Dialog.Content maxWidth="420px" size="3">
        <Dialog.Title>{title}</Dialog.Title>
        <Dialog.Description className={styles.srOnly}>
          {tagPickerLabel}
        </Dialog.Description>
        <div className={styles.content}>
          <FormField size="2" label={tagPickerLabel} htmlFor={inputId}>
            {/* Wrapper used to measure dropdown position */}
            <div ref={containerRef} className={styles.pickerWrapper}>
              <div className={styles.pickerInputRow}>
                <div className={styles.chipsAndInput}>
                  {tags.map((tag) => (
                    <span key={tag.key} className={styles.chip}>
                      <Text size="1" weight="medium">
                        {tag.name}
                      </Text>
                      <button
                        type="button"
                        className={styles.chipRemove}
                        aria-label={tag.name}
                        disabled={isLoading}
                        onClick={() => {
                          onRemoveTag(tag.key);
                        }}
                      >
                        <Cross2Icon width="0.75rem" height="0.75rem" />
                      </button>
                    </span>
                  ))}
                  <input
                    id={inputId}
                    ref={inputRef}
                    className={styles.searchInput}
                    type="text"
                    value={searchKeyword}
                    disabled={isLoading}
                    autoFocus={true}
                    autoComplete="off"
                    onChange={onSearchChange}
                    onKeyDown={onInputKeyDown}
                  />
                </div>
                {tags.length > 0 ? (
                  <button
                    type="button"
                    className={styles.clearButton}
                    disabled={isLoading}
                    aria-label="Clear"
                    onClick={onClearTags}
                  >
                    <Cross2Icon width="0.875rem" height="0.875rem" />
                  </button>
                ) : null}
              </div>
              {/*
               * position:fixed escapes Dialog's overflow:auto clipping and sits
               * above the Dialog overlay without needing createPortal.
               * onMouseDown:preventDefault on the container prevents the input
               * from losing focus when the user clicks a suggestion.
               */}
              {suggestionsOpen ? (
                <div
                  className={styles.suggestions}
                  style={{
                    top: suggestionPos.top,
                    left: suggestionPos.left,
                    width: suggestionPos.width,
                  }}
                  onMouseDown={(e) => e.preventDefault()}
                  role="listbox"
                >
                  {suggestionItems}
                </div>
              ) : null}
            </div>
          </FormField>
        </div>
        <Flex gap="3" mt="4" justify="end">
          <SecondaryButton
            size="2"
            disabled={isLoading}
            onClick={onDialogDismiss}
            text={<FormattedMessage id="cancel" />}
          />
          <PrimaryButton
            size="2"
            disabled={!canSubmit}
            loading={isLoading}
            onClick={onClickAdd}
            text={<FormattedMessage id="add" />}
          />
        </Flex>
      </Dialog.Content>
    </Dialog.Root>
  );
}

export default AddTagsDialog;
