const {spawnSync} = require("child_process");
const {
    getIntrospectionQuery,
    buildClientSchema,
    printSchema,
} = require("graphql/utilities");

const {stdout, stderr} = spawnSync("go", ["run", "../graphqlschema/main.go", process.argv[2] || "unset"], {
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

const schema = buildClientSchema(sortValue(result.data));
console.log(printSchema(schema));

function compare(selector) {
    return (a, b) => {
        if (selector(a) < selector(b)) {
            return -1;
        } else if (selector(a) > selector(b)) {
            return 1;
        } else {
            return 0;
        }
    };
}

function sortValue(obj) {
    if (Array.isArray(obj)) {
        return obj
            .map(elem => sortValue(elem))
            .sort(compare(value => value.name));
    } else if (obj && typeof obj === "object") {
        const entries = Object.entries(obj);
        for (const entry of entries) {
            entry[1] = sortValue(entry[1]);
        }
        entries.sort(compare(entry => entry[0]));
        return Object.fromEntries(entries);
    } else {
        return obj;
    }
}