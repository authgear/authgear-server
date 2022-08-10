// This script generates ./src/codesplit.ts
// This script is NOT automatically run in postinstall phase
// because our Dockerfile only copies package.json and package-lock.json
// Fortunately this script rarely needs rerun because we rarely add new
// dependencies.
import { readFile, open } from "fs/promises";

async function writeImport(dep) {
  await filehandle.write("import(");
  await filehandle.write(JSON.stringify(dep));
  // We include whitespaces here so that the generated file is formatted
  // according to Prettier's taste.
  // So npm run fmt WILL NOT format that file again.
  await filehandle.write(").finally(() => {});\n");
}

const packageJSON = JSON.parse(
  await readFile("./package.json", { encoding: "utf8" })
);

const deps = [];
for (const key of Object.keys(packageJSON["dependencies"])) {
  deps.push(key);
}
deps.sort();

const productionOnlyDeps = [
  "@apollo/client",
  "@fluentui/react",
  "@fluentui/react-hooks",
  "@oursky/react-messageformat",
  "react-dom",
  "react-helmet-async",
  "react-router-dom",
];

const filehandle = await open("./src/codesplit.ts", "w");
for (const dep of deps) {
  if (!productionOnlyDeps.includes(dep)) {
    await writeImport(dep);
  }
}

await filehandle.write("\n");
await filehandle.write('if (process.env.NODE_ENV === "production") {\n');
for (const dep of productionOnlyDeps) {
  // Indentation for if-else block
  await filehandle.write("  ");
  await writeImport(dep);
}
await filehandle.write("}\n");

await filehandle.close();
