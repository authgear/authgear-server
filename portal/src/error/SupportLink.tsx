import React, { createContext, useContext } from "react";
import ExternalLink from "../ExternalLink";
import { makeHelpDeskURL } from "../util/helpDesk";

// The error parser renders this link, so it must not import the API clients.
// A provider fills in the project and viewer; without one the help desk form
// just asks for them.
const SupportLinkHrefContext = createContext<string>(
  makeHelpDeskURL({ purpose: "other" })
);

export const SupportLinkHrefProvider = SupportLinkHrefContext.Provider;

const SupportLink: React.VFC<{ children?: React.ReactNode }> =
  function SupportLink({ children }) {
    const href = useContext(SupportLinkHrefContext);
    return <ExternalLink href={href}>{children}</ExternalLink>;
  };

export default SupportLink;
