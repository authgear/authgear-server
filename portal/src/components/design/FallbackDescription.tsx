import React, { useContext } from "react";
import { Text } from "@fluentui/react";
import {
  Context as MFContext,
  FormattedMessage,
} from "@oursky/react-messageformat";
import cn from "classnames";

import { LanguageTag } from "../../util/resource";

interface FallbackDescriptionProps {
  fallbackLanguage: LanguageTag;
}
const FallbackDescription: React.VFC<FallbackDescriptionProps> =
  function FallbackDescription(props) {
    const { fallbackLanguage } = props;
    const { renderToString } = useContext(MFContext);
    return (
      <Text
        className={cn("text-neutral-secondary")}
        variant="small"
        block={true}
      >
        <FormattedMessage
          id="DesignScreen.configuration.fallback"
          values={{
            fallbackLanguage: renderToString(`Locales.${fallbackLanguage}`),
          }}
        />
      </Text>
    );
  };

export default FallbackDescription;
