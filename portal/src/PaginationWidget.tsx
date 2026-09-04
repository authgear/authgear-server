import React, { useMemo, useCallback, useContext } from "react";
import cn from "classnames";
import { Button, IconButton } from "@radix-ui/themes";
import {
  ChevronLeftIcon,
  ChevronRightIcon,
  DoubleArrowLeftIcon,
  DoubleArrowRightIcon,
} from "@radix-ui/react-icons";
import { Context } from "./intl";
import { getPaginationRenderData } from "./util/pagination";
import styles from "./PaginationWidget.module.css";

export interface PaginationProps {
  offset: number;
  pageSize: number;
  totalCount?: number;
  onChangeOffset?: (offset: number) => void;
}

export interface PaginationWidgetProps extends PaginationProps {
  className?: string;
}

const PaginationWidget: React.VFC<PaginationWidgetProps> =
  function PaginationWidget(props: PaginationWidgetProps) {
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
          type="button"
          variant="ghost"
          color="gray"
          size="1"
          className={styles.navButton}
          title={labelFirst}
          aria-label={labelFirst}
          disabled={!firstPageButtonEnabled}
          onClick={onClickFirst}
        >
          <DoubleArrowLeftIcon />
        </IconButton>
        <IconButton
          type="button"
          variant="ghost"
          color="gray"
          size="1"
          className={styles.navButton}
          title={labelPrev}
          aria-label={labelPrev}
          disabled={!prevPageButtonEnabled}
          onClick={onClickPrev}
        >
          <ChevronLeftIcon />
        </IconButton>
        <div className={styles.pages}>
          {offsets.map((offset) => {
            const page = offset / pageSize + 1;
            const label = renderToString("PaginationWidget.Page", {
              PAGE: page,
            });
            const isCurrent = currentOffset === offset;
            return (
              <Button
                key={offset}
                type="button"
                variant={isCurrent ? "soft" : "ghost"}
                color={isCurrent ? undefined : "gray"}
                size="1"
                className={styles.pageButton}
                title={label}
                aria-label={label}
                aria-current={isCurrent ? "page" : undefined}
                onClick={(e: React.MouseEvent<HTMLElement>) => {
                  e.preventDefault();
                  e.stopPropagation();
                  if (!isCurrent) {
                    onChangeOffset?.(offset);
                  }
                }}
              >
                {String(page)}
              </Button>
            );
          })}
        </div>
        <IconButton
          type="button"
          variant="ghost"
          color="gray"
          size="1"
          className={styles.navButton}
          title={labelNext}
          aria-label={labelNext}
          disabled={!nextPageButtonEnabled}
          onClick={onClickNext}
        >
          <ChevronRightIcon />
        </IconButton>
        <IconButton
          type="button"
          variant="ghost"
          color="gray"
          size="1"
          className={styles.navButton}
          title={labelLast}
          aria-label={labelLast}
          disabled={!lastPageButtonEnabled}
          onClick={onClickLast}
        >
          <DoubleArrowRightIcon />
        </IconButton>
      </div>
    );
  };

export default PaginationWidget;
