import React from "react";
import { Separator as RadixSeparator } from "@radix-ui/themes";
import cn from "classnames";

interface SeparatorProps {
  className?: string;
}
const Separator: React.VFC<SeparatorProps> = function Separator(props) {
  const { className } = props;
  return (
    <RadixSeparator size="4" className={cn("w-full", "my-6", className)} />
  );
};

export default Separator;
