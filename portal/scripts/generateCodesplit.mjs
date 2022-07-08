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
  await filehandle.write("import(");
  await filehandle.write(JSON.stringify(dep));
  await filehandle.write(").then(()=>{});\n");
}
await filehandle.close();
