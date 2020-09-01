module.exports = {
  client: {
    service: {
      name: "portal",
      localSchemaFile: "./src/graphql/portal/schema.graphql",
    },
    includes: ["./src/graphql/portal/**/*.{js,jsx,ts,tsx}"],
  },
};
