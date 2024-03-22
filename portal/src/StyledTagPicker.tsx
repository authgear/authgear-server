import React, { useCallback, useRef, useState } from "react";
import { ITag, ITagPickerProps, IconButton, TagPicker } from "@fluentui/react";
import styles from "./StyledTagPicker.module.css";
import { fixTagPickerStyles } from "./bugs";
import cn from "classnames";

/*
Expected usage:
  const [searchKeyword, setSearchKeyword] = useState("");
  const [tags, setTags] = useState<ITag[]>();
  const { refetch } = useQuery()
  const onChangeTags = useCallback((tags?: ITag[]) => {
    setTags(tags);
  }, []);
  const onResolveSuggestions = useCallback(async (): Promise<ITag[]> => {
    const result = await refetch({
      searchKeyword,
    });

    if (result.data.roles?.edges == null) {
      return [];
    }
    return result.data.roles.edges.map((edge) => {
      return {
        key: edge?.node?.key ?? "",
        name: edge?.node?.name ?? "",
      };
    });
  }, [refetch, searchKeyword]);

  const onInputChange = useCallback((value: string): string => {
    setSearchKeyword(value);
    return value;
  }, []);

  return <StyledTagPicker
    value={searchKeyword}
    onInputChange={onInputChange}
    selectedItems={tags}
    onChange={onChangeTags}
    onResolveSuggestions={onResolveSuggestions}
    onClearTags={onClearTags}
  />


 */

interface StyledPickerProps extends ITagPickerProps {
  value: string;
  onClearTags: () => void;
  autoFocus?: boolean;
}

const StyledTagPicker: React.VFC<StyledPickerProps> = function StyledTagPicker(
  props: StyledPickerProps
) {
  const {
    selectedItems,
    onChange,
    className,
    value,
    onClearTags: onClearInput,
    autoFocus,
    ...rest
  } = props;
  const tagPickerRef = useRef<HTMLDivElement | null>(null);

  const _onChange = useCallback(
    (tags?: ITag[]) => {
      const tagsByKey = tags?.map((tag) => tag.key) ?? [];
      const filteredTags = tags?.filter(
        (tag, index) => tagsByKey.indexOf(tag.key) === index
      );
      onChange?.(filteredTags);
    },
    [onChange]
  );
  const [calloutWidth, setCalloutWidth] = useState(0);
  const onCalloutLayoutMounted = useCallback(() => {
    setCalloutWidth(tagPickerRef.current?.offsetWidth ?? 0);
  }, []);

  return (
    // NOTE: directly add ref to TagPicker doesn't work
    <div ref={tagPickerRef} className={styles.tagPickerContainer}>
      <TagPicker
        {...rest}
        styles={{
          ...fixTagPickerStyles,
          itemsWrapper: {
            // padding of input field
            maxWidth: "calc(100% - 32px)",
          },
        }}
        className={cn(styles.tagPicker, className)}
        inputProps={{
          value,
          className: styles.pickerInput,
          autoFocus: autoFocus,
        }}
        pickerCalloutProps={{
          target: tagPickerRef,
          calloutWidth: calloutWidth,
          onLayerMounted: onCalloutLayoutMounted,
        }}
        pickerSuggestionsProps={{
          className: "min-w-0",
        }}
        onChange={_onChange}
        selectedItems={selectedItems}
      />
      {(selectedItems?.length ?? 0) > 0 ? (
        <IconButton
          onClick={onClearInput}
          iconProps={{ iconName: "Clear" }}
          className={styles.pickerIconButton}
        />
      ) : null}
    </div>
  );
};

export default StyledTagPicker;
