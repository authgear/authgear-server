import React, { useCallback, useRef } from "react";
import { ITag, ITagPickerProps, TagPicker } from "@fluentui/react";
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
  />


 */

interface StyledPickerProps extends ITagPickerProps {
  value: string;
}

const StyledTagPicker: React.VFC<StyledPickerProps> = function StyledTagPicker(
  props: StyledPickerProps
) {
  const { onChange, className, value, ...rest } = props;
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

  return (
    // NOTE: directly add ref to TagPicker doesn't work
    <div ref={tagPickerRef}>
      <TagPicker
        {...rest}
        styles={fixTagPickerStyles}
        className={cn(styles.tagPicker, className)}
        inputProps={{
          value,
        }}
        pickerSuggestionsProps={{
          suggestionsClassName: styles.pickerSuggestions,
        }}
        pickerCalloutProps={{
          className: styles.pickerCallout,
          target: tagPickerRef,
          calloutMaxHeight: 152,
          calloutMaxWidth: 510,
        }}
        onChange={_onChange}
      />
    </div>
  );
};

export default StyledTagPicker;
