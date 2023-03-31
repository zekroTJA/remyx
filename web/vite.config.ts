import SVGInjectPlugin from "vite-plugin-svgr-component";
import { defineConfig } from "vite";
import solidPlugin from "vite-plugin-solid";

export default defineConfig({
  plugins: [solidPlugin(), SVGInjectPlugin()],
  server: {
    port: 3000,
  },
  build: {
    target: "esnext",
  },
});
