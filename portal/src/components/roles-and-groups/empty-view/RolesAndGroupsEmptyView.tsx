import React, { MouseEventHandler } from "react";
import cn from "classnames";
import { Text } from "@fluentui/react";
import styles from "./RolesAndGroupsEmptyView.module.css";
import PrimaryButton from "../../../PrimaryButton";

function CreateButton(props: {
  className?: string;
  href?: string;
  onClick?: MouseEventHandler<HTMLButtonElement | HTMLAnchorElement>;
  text: React.ReactNode;
}) {
  const { className, href, onClick, text } = props;
  return (
    <PrimaryButton
      href={href}
      onClick={onClick}
      className={className}
      text={text}
      iconProps={{ iconName: "Add" }}
    />
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
      <div className={styles.icon}>{icon}</div>
      <Text className={styles.title}>{title}</Text>
      <Text className={styles.description}>{description}</Text>
      <div className={styles.button}>{button}</div>
    </div>
  );
};

export const RolesAndGroupsEmptyView = Object.assign(RolesAndGroupsEmptyView_, {
  CreateButton,
});
