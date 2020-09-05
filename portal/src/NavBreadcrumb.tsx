import React, { useContext } from "react";
import { useHref } from "react-router-dom";
import cn from "classnames";
import { Breadcrumb, IBreadcrumbItem, IRenderFunction } from "@fluentui/react";
import { Context } from "@oursky/react-messageformat";
import useNavIsActive from "./hook/useNavIsActive";
import styles from "./NavBreadcrumb.module.scss";

export interface BreadcrumbItem {
  to: string;
  label: React.ReactNode;
}

export interface Props {
  className?: string;
  items: BreadcrumbItem[];
}

interface FuncLinkProps {
  item: IBreadcrumbItem;
  renderFunc: IRenderFunction<IBreadcrumbItem>;
}

const FuncLink: React.FC<FuncLinkProps> = function FuncLink(
  props: FuncLinkProps
) {
  const { item, renderFunc } = props;
  const href = useHref(item.href!);
  const isActive = useNavIsActive(item.href!);

  if (isActive) {
    return renderFunc({
      ...item,
      href: undefined,
      as: "h1",
    });
  }

  return renderFunc({
    ...item,
    href,
  });
};

function onRenderItem(
  item?: IBreadcrumbItem,
  renderFunc?: IRenderFunction<IBreadcrumbItem>
) {
  return (
    <FuncLink
      // @ts-expect-error it is never null
      item={item}
      // @ts-expect-error it is never null
      renderFunc={renderFunc}
    />
  );
}

// NavBreadcrumb is an integration between Breadcrumb and react-router-dom.
// The biggest trick here is to provide onRenderItem, which accept IBreadcrumbItem and the original renderItem function.
// And then we render a function component, which allows us to use hooks.
const NavBreadcrumb: React.FC<Props> = function NavBreadcrumb(props: Props) {
  const { className, items } = props;
  const { renderToString } = useContext(Context);

  const breadcrumbItems: IBreadcrumbItem[] = [];
  for (const item of items) {
    breadcrumbItems.push({
      key: item.to,
      href: item.to,
      // @ts-expect-error text actually can be React.ReactNode. Their typedef is incorrect.
      text: item.label,
    });
  }

  const label = renderToString("NavBreadcrumb.label");

  return (
    <Breadcrumb
      ariaLabel={label}
      className={cn(className, styles.root)}
      items={breadcrumbItems}
      onRenderItem={onRenderItem}
    />
  );
};

export default NavBreadcrumb;
