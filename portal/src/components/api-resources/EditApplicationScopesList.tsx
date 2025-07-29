import React from "react";

export interface EditApplicationScopesListItem {
  scope: string;
  isAssigned: boolean;
}

interface EditApplicationScopesListProps {
  className?: string;
  scopes: EditApplicationScopesListItem[];
}

export const EditApplicationScopesList: React.VFC<EditApplicationScopesListProps> =
  function EditApplicationScopesList({}: EditApplicationScopesListProps) {
    // TODO
    return <></>;
  };
