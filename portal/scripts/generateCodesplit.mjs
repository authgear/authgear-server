// This script generates ./src/codesplit.ts
// This script is NOT automatically run in postinstall phase
// because our Dockerfile only copies package.json and package-lock.json
// Fortunately this script rarely needs rerun because we rarely add new
// dependencies.
import { readFile, open } from "fs/promises";

const defaultOnly = ["deep-equal"];

let i = 0;
async function writeImport(dep) {
  // We include whitespaces here so that the generated file is formatted
  // according to Prettier's taste.
  // So npm run fmt WILL NOT format that file again.
  //
  i++;
  const name = "_" + String(i);
  if (defaultOnly.includes(dep)) {
    await filehandle.write(`import ${name} from `);
  } else {
    await filehandle.write(`import * as ${name} from `);
  }
  await filehandle.write(JSON.stringify(dep));
  await filehandle.write(";\nconsole.log(");
  await filehandle.write(name);
  await filehandle.write(");\n");
}

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
  await writeImport(dep);
}
await filehandle.close();
