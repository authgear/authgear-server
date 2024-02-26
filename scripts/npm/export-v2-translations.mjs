import { stringify } from "csv";
import { readFile } from "fs/promises";
import { join, dirname } from "node:path";
import { fileURLToPath } from "node:url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

const supportedLanguages = ["en", "zh-HK", "zh-TW"];

async function main() {
  const messagesByLocale = {};
  for (const locale of supportedLanguages) {
    const filePath = join(
      __dirname,
      "../..",
      `resources/authgear/templates/${locale}/translation.json`,
    );
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
      {},
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
      process.stdout.write(output);
    },
  );
}

main().catch((e) => console.error("Failed to export v2 translations", e));
