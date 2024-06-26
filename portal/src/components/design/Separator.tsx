import React from "react";
import cn from "classnames";

interface SeparatorProps {
  className?: string;
}
const Separator: React.VFC<SeparatorProps> = function Separator(props) {
  const { className } = props;
  return <div className={cn("h-px", "my-12", "bg-separator", className)}></div>;
};

export default Separator;
