import { stringify } from "csv";
import { readFile, writeFile, mkdir } from "fs/promises";
import { parseArgs } from "node:util";
import { join, dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import { cwd } from "node:process";

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

const {
  values: { "output-file": outputFile },
} = parseArgs({
  options: {
    "output-file": {
      short: "f",
      type: "string",
      default: "output/v2-translations.csv",
    },
  },
});

const supportedLanguages = ["en", "zh-HK", "zh-TW"];

async function main() {
  const messagesByLocale = {};
  for (const locale of supportedLanguages) {
    const filePath = join(
      __dirname,
      "../..",
      `resources/authgear/templates/${locale}/translation.json`
    );
    console.info(`Reading data from ${filePath}`);
    const jsonStr = String(await readFile(filePath));
    const messages = JSON.parse(jsonStr);
    const v2Messages = Object.entries(messages).reduce(
      (msgs, [key, message]) => {
        if (!key.startsWith("v2-")) {
          return msgs;
        }
        msgs[key] = message;
        return msgs;
      },
      {}
    );

    messagesByLocale[locale] = v2Messages;
  }

  // messagesByLocale is object[locale][key]
  // we want to transform the object into array[key][locale]
  const outData = [];
  for (const key of Object.keys(messagesByLocale[supportedLanguages[0]])) {
    const row = [key];
    for (const locale of supportedLanguages) {
      row.push(messagesByLocale[locale][key]);
    }
    outData.push(row);
  }

  stringify(
    outData,
    {
      header: true,
      columns: ["Key", ...supportedLanguages],
    },
    async (err, output) => {
      if (err) {
        console.log(err);
        process.exit(1);
      }
      const absOutputFile = resolve(cwd(), outputFile);
      const outputDir = dirname(absOutputFile);
      await mkdir(`${outputDir}`, { recursive: true });
      await writeFile(outputFile, output);
    }
  );
}

main().catch((e) => console.error("Failed to export v2 translations", e));
