module.exports = {
  client: {
    service: {
      name: "adminapi",
      localSchemaFile: "./src/graphql/adminapi/schema.graphql",
    },
    includes: ["./src/graphql/adminapi/**/*.{js,jsx,ts,tsx}"],
  },
};
