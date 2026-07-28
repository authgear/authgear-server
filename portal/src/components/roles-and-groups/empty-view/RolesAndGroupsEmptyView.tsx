import React, { MouseEventHandler } from "react";
import cn from "classnames";
import { Text } from "@radix-ui/themes";
import { PlusIcon } from "@radix-ui/react-icons";
import styles from "./RolesAndGroupsEmptyView.module.css";
import { PrimaryButton } from "../../v2/Button/PrimaryButton/PrimaryButton";

function CreateButton(props: {
  className?: string;
  href?: string;
  onClick?: MouseEventHandler<HTMLButtonElement | HTMLAnchorElement>;
  text: React.ReactNode;
}) {
  const { className, onClick, text } = props;
  return (
    <span className={className}>
      <PrimaryButton
        size="2"
        onClick={onClick}
        text={
          <span className="inline-flex items-center gap-2">
            <PlusIcon width="1rem" height="1rem" />
            {text}
          </span>
        }
      />
    </span>
  );
}

const RolesAndGroupsEmptyView_: React.VFC<{
  className?: string;
  icon: React.ReactNode;
  title: React.ReactNode;
  description: React.ReactNode;
  button: React.ReactNode;
}> = function RolesAndGroupsEmptyView_({
  className,
  icon,
  title,
  description,
  button,
}) {
  return (
    <div className={cn(className, styles.container)}>
      <div className={styles.content}>
        <div className={styles.icon}>{icon}</div>
        <Text as="p" size="3" weight="bold" className={styles.title}>
          {title}
        </Text>
        <Text as="p" size="2" className={styles.description}>
          {description}
        </Text>
      </div>
      <div className={styles.button}>{button}</div>
    </div>
  );
};

export const RolesAndGroupsEmptyView = Object.assign(RolesAndGroupsEmptyView_, {
  CreateButton,
});
