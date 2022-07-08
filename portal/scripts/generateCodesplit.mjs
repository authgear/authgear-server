import { readFile, open } from "fs/promises";

const packageJSON = JSON.parse(
  await readFile("./package.json", { encoding: "utf8" })
);

const deps = [];
for (const key of Object.keys(packageJSON["dependencies"])) {
  deps.push(key);
}
deps.sort();

const filehandle = await open("./src/codesplit.ts", "w");
for (const dep of deps) {
  // Special case: normalize.css is referenced by index.html.
  if (dep !== "normalize.css") {
    await filehandle.write("import(");
    await filehandle.write(JSON.stringify(dep));
    // We include whitespaces here so that the generated file is formatted
    // according to Prettier's taste.
    // So npm run fmt WILL NOT format that file again.
    await filehandle.write(").finally(() => {});\n");
  }
}
await filehandle.close();
