import React from "react";
import { DetailsList, IDetailsListProps, IColumn } from "@fluentui/react";
import { Context } from "@oursky/react-messageformat";

import OrderButtons from "./OrderButtons";

interface DetailsListWithOrderingProps extends IDetailsListProps {
  onSwapClicked: (index1: number, index2: number) => void;
  columns: IColumn[];
  onRenderItemColumnClassName?: (
    item?: any,
    index?: number,
    column?: IColumn
  ) => string;
  onRenderItemColumn: (
    item?: any,
    index?: number,
    column?: IColumn
  ) => React.ReactNode;
  renderAriaLabel: (index?: number) => string;
  orderColumnMinWidth?: number;
  orderColumnMaxWidth?: number;
}

const DetailsListWithOrdering: React.FC<DetailsListWithOrderingProps> =
  function DetailsListWithOrdering(props: DetailsListWithOrderingProps) {
    const { renderToString } = React.useContext(Context);
    const onRenderItemColumn = React.useCallback(
      (item?: any, index?: number, column?: IColumn) => {
        const className = props.onRenderItemColumnClassName?.(
          item,
          index,
          column
        );

        const component =
          column?.key === "order" ? (
            <OrderButtons
              index={index}
              itemCount={props.items.length}
              onSwapClicked={props.onSwapClicked}
              renderAriaLabel={props.renderAriaLabel}
            />
          ) : (
            props.onRenderItemColumn(item, index, column)
          );
        return <span className={className}>{component}</span>;
      },
      [props]
    );

    const columns: IColumn[] = [
      ...props.columns,
      {
        key: "order",
        fieldName: "order",
        name: renderToString("DetailsListWithOrdering.order"),
        minWidth: props.orderColumnMinWidth ?? 100,
        maxWidth: props.orderColumnMaxWidth ?? 100,
      },
    ];

    return (
      <DetailsList
        {...props}
        columns={columns}
        onRenderItemColumn={onRenderItemColumn}
      />
    );
  };

export default DetailsListWithOrdering;
