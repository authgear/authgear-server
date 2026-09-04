import React, {
  useCallback,
  useMemo,
  useState,
  ReactElement,
  ReactNode,
} from "react";
import cn from "classnames";
import { Checkbox } from "@radix-ui/themes";
import { DragHandleDots2Icon } from "@radix-ui/react-icons";
import styles from "./PriorityList.module.css";

export interface PriorityListItem {
  key: string;
  checked: boolean;
  disabled: boolean;
  content: ReactNode;
}

export interface PriorityListProps {
  className?: string;
  items: PriorityListItem[];
  checkedColumnLabel: string;
  keyColumnLabel: string;
  onChangeChecked: (key: string, checked: boolean) => void;
  onMove: (fromIndex: number, toIndex: number) => void;
}

interface LocalCheckboxProps {
  item: PriorityListItem;
  onChangeChecked: (key: string, checked: boolean) => void;
}

function LocalCheckbox(props: LocalCheckboxProps): ReactElement {
  const { item, onChangeChecked } = props;

  const onCheckedChange = useCallback(
    (checked: boolean | "indeterminate") => {
      if (checked === "indeterminate") {
        return;
      }
      onChangeChecked(item.key, checked);
    },
    [item.key, onChangeChecked]
  );

  return (
    <Checkbox
      checked={item.checked}
      onCheckedChange={onCheckedChange}
      disabled={Boolean(item.disabled && !item.checked)}
    />
  );
}

// A transparent 1x1 GIF used to suppress the browser's default drag
// preview snapshot, which would otherwise follow the cursor and block
// the view. The drop position indicators are enough visual feedback.
let emptyDragImage: HTMLImageElement | undefined;
function getEmptyDragImage(): HTMLImageElement {
  if (emptyDragImage == null) {
    emptyDragImage = new Image(1, 1);
    emptyDragImage.src =
      "data:image/gif;base64,R0lGODlhAQABAIAAAAAAAP///yH5BAEAAAAALAAAAAABAAEAAAIBRAA7";
  }
  return emptyDragImage;
}

function PriorityList(props: PriorityListProps): ReactElement {
  const {
    className,
    items,
    checkedColumnLabel,
    keyColumnLabel,
    onChangeChecked,
    onMove,
  } = props;

  const [dndIndex, setDNDIndex] = useState<number | undefined>(undefined);
  // The row currently hovered during a drag; the insertion indicator is
  // drawn on this row, on the side where the dragged row would land.
  const [dndOverIndex, setDNDOverIndex] = useState<number | undefined>(
    undefined
  );

  const endDrag = useCallback(() => {
    setDNDIndex(undefined);
    setDNDOverIndex(undefined);
  }, []);

  // Only activated items are orderable; disabled or unchecked items are
  // pinned at the bottom and cannot be dragged nor be a drop target.
  // Indices used for drag state are display positions; each entry keeps
  // its index in the original items array for onMove.
  const displayItems = useMemo(() => {
    const withIndex = items.map((item, index) => ({
      item,
      index,
      pinned: item.disabled || !item.checked,
    }));
    return [
      ...withIndex.filter((entry) => !entry.pinned),
      ...withIndex.filter((entry) => entry.pinned),
    ];
  }, [items]);

  return (
    <div className={cn(styles.table, className)}>
      <div className={styles.headerRow}>
        <div className={styles.headerReorderSpacer} />
        <div className={styles.headerCellChecked}>{checkedColumnLabel}</div>
        <div className={styles.headerCellKey}>{keyColumnLabel}</div>
      </div>
      {displayItems.map(({ item, index: itemIndex, pinned }, displayIndex) => (
        <div
          key={item.key}
          className={styles.row}
          draggable={!pinned}
          onDragStart={(e) => {
            e.dataTransfer.effectAllowed = "move";
            e.dataTransfer.setDragImage(getEmptyDragImage(), 0, 0);
            setDNDIndex(displayIndex);
          }}
          onDragEnd={endDrag}
          onDragOver={(e) => {
            if (pinned) {
              // Not a drop target; clear any stale indicator.
              if (dndOverIndex != null) {
                setDNDOverIndex(undefined);
              }
              return;
            }
            e.preventDefault();
            if (dndOverIndex !== displayIndex) {
              setDNDOverIndex(displayIndex);
            }
          }}
          onDrop={() => {
            if (!pinned && dndIndex != null && dndIndex !== displayIndex) {
              onMove(displayItems[dndIndex].index, itemIndex);
            }
            endDrag();
          }}
          data-dnd-dragging={dndIndex === displayIndex ? true : undefined}
          data-dnd-insert-above={
            dndIndex != null &&
            dndOverIndex === displayIndex &&
            displayIndex < dndIndex
              ? true
              : undefined
          }
          data-dnd-insert-below={
            dndIndex != null &&
            dndOverIndex === displayIndex &&
            displayIndex > dndIndex
              ? true
              : undefined
          }
        >
          <div className={styles.reorderHandle}>
            {!pinned ? <DragHandleDots2Icon /> : null}
          </div>
          <div className={styles.cellChecked}>
            <LocalCheckbox item={item} onChangeChecked={onChangeChecked} />
          </div>
          <div className={styles.cellKey}>{item.content}</div>
        </div>
      ))}
    </div>
  );
}

export default PriorityList;
