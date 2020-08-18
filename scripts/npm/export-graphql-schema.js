const {spawnSync} = require("child_process");
const {
    getIntrospectionQuery,
    buildClientSchema,
    printSchema,
} = require("graphql/utilities");

const { stdout, stderr } = spawnSync("go", ["run", "../graphqlschema/main.go"], {
    input: getIntrospectionQuery(),
})
if (stderr.length > 0) {
    console.error(stderr.toString());
    process.exit(1);
}

const result = JSON.parse(stdout.toString());
if (result.errors) {
    console.error(result.errors);
    process.exit(1);
}

const schema = buildClientSchema(result.data);
console.log(printSchema(schema));