import React, { useContext } from "react";
import { Text } from "@radix-ui/themes";
import { Context as MFContext, FormattedMessage } from "../../intl";

import { LanguageTag } from "../../util/resource";

interface FallbackDescriptionProps {
  fallbackLanguage: LanguageTag;
}
const FallbackDescription: React.VFC<FallbackDescriptionProps> =
  function FallbackDescription(props) {
    const { fallbackLanguage } = props;
    const { renderToString } = useContext(MFContext);
    return (
      <Text as="p" size="1" color="gray">
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
