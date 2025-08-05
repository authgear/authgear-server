import React from "react";

interface EditOAuthClientFormQuickStartContentProps {
  className?: string;
}

export const EditOAuthClientFormQuickStartContent: React.VFC<EditOAuthClientFormQuickStartContentProps> =
  function EditOAuthClientFormQuickStartContent(props) {
    const { className } = props;
    return <div className={className}></div>;
  };
