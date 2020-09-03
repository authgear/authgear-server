import React, { useMemo, useCallback, useContext } from "react";
import cn from "classnames";
import {
  DefaultButton,
  IconButton,
  IIconProps,
  IButtonStyles,
} from "@fluentui/react";
import { Context } from "@oursky/react-messageformat";
import { getPaginationRenderData } from "./util/pagination";
import styles from "./PaginationWidget.module.scss";

export interface Props {
  className?: string;
  offset: number;
  pageSize: number;
  totalCount?: number;
  onChangeOffset?: (offset: number) => void;
}

const iconFirst: IIconProps = {
  iconName: "DoubleChevronLeft8",
};

const iconPrev: IIconProps = {
  iconName: "ChevronLeftSmall",
};

const iconNext: IIconProps = {
  iconName: "ChevronRightSmall",
};

const iconLast: IIconProps = {
  iconName: "DoubleChevronRight8",
};

const iconButtonStyles: IButtonStyles = {
  root: {
    width: "24px",
    height: "24px",
  },
  rootDisabled: {
    backgroundColor: "transparent",
  },
};

const pageButtonStyles: IButtonStyles = {
  root: {
    border: "none",
    minWidth: "0px",
    padding: "6px",
    fontWeight: "900",
  },
  rootDisabled: {
    backgroundColor: "transparent",
  },
};

const PaginationWidget: React.FC<Props> = function PaginationWidget(
  props: Props
) {
  const { className, offset, pageSize, totalCount, onChangeOffset } = props;

  const { renderToString } = useContext(Context);

  const {
    currentOffset,
    offsets,
    firstPageButtonEnabled,
    prevPageButtonEnabled,
    nextPageButtonEnabled,
    lastPageButtonEnabled,
    maxOffset,
  } = useMemo(() => {
    return getPaginationRenderData({
      offset,
      pageSize,
      totalCount,
    });
  }, [offset, pageSize, totalCount]);

  const onClickFirst = useCallback(
    (e: React.MouseEvent<HTMLElement>) => {
      e.preventDefault();
      e.stopPropagation();
      onChangeOffset?.(0);
    },
    [onChangeOffset]
  );

  const onClickPrev = useCallback(
    (e: React.MouseEvent<HTMLElement>) => {
      e.preventDefault();
      e.stopPropagation();
      onChangeOffset?.(currentOffset - pageSize);
    },
    [currentOffset, pageSize, onChangeOffset]
  );

  const onClickNext = useCallback(
    (e: React.MouseEvent<HTMLElement>) => {
      e.preventDefault();
      e.stopPropagation();
      onChangeOffset?.(currentOffset + pageSize);
    },
    [currentOffset, pageSize, onChangeOffset]
  );

  const onClickLast = useCallback(
    (e: React.MouseEvent<HTMLElement>) => {
      e.preventDefault();
      e.stopPropagation();
      if (maxOffset != null) {
        onChangeOffset?.(maxOffset);
      }
    },
    [maxOffset, onChangeOffset]
  );

  const labelFirst = renderToString("PaginationWidget.First");
  const labelPrev = renderToString("PaginationWidget.Prev");
  const labelNext = renderToString("PaginationWidget.Next");
  const labelLast = renderToString("PaginationWidget.Last");

  return (
    <div className={cn(className, styles.root)}>
      <IconButton
        title={labelFirst}
        ariaLabel={labelFirst}
        styles={iconButtonStyles}
        iconProps={iconFirst}
        disabled={!firstPageButtonEnabled}
        onClick={onClickFirst}
      />
      <IconButton
        title={labelPrev}
        ariaLabel={labelPrev}
        styles={iconButtonStyles}
        iconProps={iconPrev}
        disabled={!prevPageButtonEnabled}
        onClick={onClickPrev}
      />
      <div className={styles.pages}>
        {offsets.map((offset) => {
          const page = offset / pageSize + 1;
          const label = renderToString("PaginationWidget.Page", {
            PAGE: page,
          });
          return (
            <DefaultButton
              key={offset}
              title={label}
              ariaLabel={label}
              styles={pageButtonStyles}
              disabled={currentOffset === offset}
              onClick={(e: React.MouseEvent<HTMLElement>) => {
                e.preventDefault();
                e.stopPropagation();
                onChangeOffset?.(offset);
              }}
            >
              {page}
            </DefaultButton>
          );
        })}
      </div>
      <IconButton
        title={labelNext}
        ariaLabel={labelNext}
        styles={iconButtonStyles}
        iconProps={iconNext}
        disabled={!nextPageButtonEnabled}
        onClick={onClickNext}
      />
      <IconButton
        title={labelLast}
        ariaLabel={labelLast}
        styles={iconButtonStyles}
        iconProps={iconLast}
        disabled={!lastPageButtonEnabled}
        onClick={onClickLast}
      />
    </div>
  );
};

export default PaginationWidget;
