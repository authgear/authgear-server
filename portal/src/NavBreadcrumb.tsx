import React, { useContext, useCallback } from "react";
import { useHref, useNavigate, useParams } from "react-router-dom";
import { Breadcrumb, IBreadcrumbItem, IRenderFunction } from "@fluentui/react";
import { Context } from "./intl";
import useNavIsActive from "./hook/useNavIsActive";

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

const breadcrumbStyles = {
  root: {
    margin: "0 -8px",
  },
  itemLink: {
    fontSize: "28px",
    margin: "0px",
  },
  item: {
    fontSize: "28px",
    margin: "0px",
  },
};

const FuncLink: React.VFC<FuncLinkProps> = function FuncLink(
  props: FuncLinkProps
) {
  const { item, renderFunc } = props;
  const href = useHref(item.href!);
  const isActive = useNavIsActive(item.href!);

  const navigate = useNavigate();
  const onLinkClicked = useCallback(
    (ev?: React.MouseEvent<HTMLElement>, _item?: IBreadcrumbItem) => {
      ev?.stopPropagation();
      ev?.preventDefault();
      navigate(href);
    },
    [navigate, href]
  );

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
    onClick: onLinkClicked,
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
const NavBreadcrumb: React.VFC<Props> = function NavBreadcrumb(props: Props) {
  const { className, items } = props;
  const { renderToString } = useContext(Context);
  const { appID } = useParams() as { appID: string };

  const breadcrumbItems: IBreadcrumbItem[] = [];
  for (const item of items) {
    const link = item.to.replace("~/", `/project/${appID}/`);
    breadcrumbItems.push({
      key: link,
      href: link,
      // @ts-expect-error text actually can be React.ReactNode. Their typedef is incorrect.
      text: item.label,
    });
  }

  const label = renderToString("NavBreadcrumb.label");

  return (
    <div className={className}>
      <Breadcrumb
        styles={breadcrumbStyles}
        ariaLabel={label}
        items={breadcrumbItems}
        onRenderItem={onRenderItem}
      />
    </div>
  );
};

export default NavBreadcrumb;
