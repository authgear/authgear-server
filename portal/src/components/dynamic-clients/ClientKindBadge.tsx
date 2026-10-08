import React from "react";
import { FormattedMessage } from "../../intl";
import { Badge } from "../v2/Badge/Badge";

export interface ClientKindBadgeProps {
  firstParty: boolean;
}

// One look for first-party vs third-party across the DCR tab: the comparison
// card, the token list, the create-token dialog and the client list.
export function ClientKindBadge({
  firstParty,
}: ClientKindBadgeProps): React.ReactElement {
  return firstParty ? (
    <Badge
      size="1"
      variant="warning"
      text={<FormattedMessage id="ClientKindBadge.first-party" />}
    />
  ) : (
    <Badge
      size="1"
      variant="neutral"
      text={<FormattedMessage id="ClientKindBadge.third-party" />}
    />
  );
}
