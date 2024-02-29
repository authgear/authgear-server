import { parse } from "csv";
import { readFile, writeFile } from "fs/promises";
import { readFileSync } from "fs";
import { parseArgs } from "node:util";
import { fileURLToPath } from "node:url";
import { join, dirname } from "node:path";

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

const supportedLanguages = ["en", "zh-HK", "zh-TW"];
const columns = ["Key", ...supportedLanguages];

async function main() {
  const updatedMessagesByLocale = {};
  supportedLanguages.forEach((lang) => {
    updatedMessagesByLocale[lang] = {};
  });
  const content = readFileSync(process.stdin.fd);
  const records = parse(content, {
    columns: () => columns,
    relaxColumnCount: true,
  });
  for await (const record of records) {
    for (const lang of supportedLanguages) {
      updatedMessagesByLocale[lang][record["Key"]] = record[lang];
    }
  }

  for (const locale of supportedLanguages) {
    const filePath = join(
      __dirname,
      "../..",
      `resources/authgear/templates/${locale}/translation.json`
    );
    console.info(`Reading data from ${filePath}`);
    const jsonStr = String(await readFile(filePath));
    const messages = JSON.parse(jsonStr);
    const allMessages = Object.entries(messages).reduce(
      (msgs, [key, message]) => {
        msgs[key] = message;
        return msgs;
      },
      {}
    );

    const newMessages = {
      ...allMessages,
      ...updatedMessagesByLocale[locale],
    };
    console.info(`Writing data to ${filePath}`);
    await writeFile(filePath, JSON.stringify(newMessages, null, 2));
  }
}

main().catch((e) => console.error("Failed to import v2 translations", e));
