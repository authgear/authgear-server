import React, { useContext } from "react";
import cn from "classnames";
import { Text } from "@radix-ui/themes";
import { ChevronRightIcon } from "@radix-ui/react-icons";
import { useParams } from "react-router-dom";
import { Context } from "./intl";
import Link from "./Link";
import useNavIsActive from "./hook/useNavIsActive";
import styles from "./NavBreadcrumb.module.css";

export interface BreadcrumbItem {
  to: string;
  label: React.ReactNode;
}

export interface Props {
  className?: string;
  items: BreadcrumbItem[];
}

function NavBreadcrumbItem({
  to,
  label,
  isLast,
}: {
  to: string;
  label: React.ReactNode;
  isLast: boolean;
}): React.ReactElement {
  const isActive = useNavIsActive(to);

  // The active item (or the last one) is the current page: render it as a
  // plain heading. Earlier items are links back up the hierarchy.
  if (isActive || isLast) {
    return (
      <Text as="span" size="7" weight="bold" className={styles.current}>
        {label}
      </Text>
    );
  }

  return (
    <Link to={to} className={styles.link}>
      <Text as="span" size="7" weight="bold">
        {label}
      </Text>
    </Link>
  );
}

// NavBreadcrumb renders a react-router-aware breadcrumb: earlier items are
// links, and the current (active/last) item is the page title.
const NavBreadcrumb: React.VFC<Props> = function NavBreadcrumb(props: Props) {
  const { className, items } = props;
  const { renderToString } = useContext(Context);
  const { appID } = useParams() as { appID: string };

  const label = renderToString("NavBreadcrumb.label");

  return (
    <nav aria-label={label} className={cn(styles.root, className)}>
      {items.map((item, index) => {
        const to = item.to.replace("~/", `/project/${appID}/`);
        const isLast = index === items.length - 1;
        return (
          <React.Fragment key={to}>
            {index > 0 ? (
              <ChevronRightIcon className={styles.separator} />
            ) : null}
            <NavBreadcrumbItem to={to} label={item.label} isLast={isLast} />
          </React.Fragment>
        );
      })}
    </nav>
  );
};

export default NavBreadcrumb;
