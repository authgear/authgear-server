import { createContext } from "react";
import { Environment } from "relay-runtime";
import { makeEnvironment } from "./graphql/adminapi/relay";

const AppContext = createContext<Environment>(makeEnvironment(""));
export default AppContext;
