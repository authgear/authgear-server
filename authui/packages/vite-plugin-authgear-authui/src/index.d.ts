import { Plugin } from "vite";

interface ViteAuthgearAuthUIOptions {
  input: Record<string, string>;
}

declare const viteAuthgearAuthUI: (
  options: ViteAuthgearAuthUIOptions
) => Plugin[];

export { viteAuthgearAuthUI };
