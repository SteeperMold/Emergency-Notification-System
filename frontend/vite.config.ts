import tailwindcss from "@tailwindcss/vite";
import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";
import svgr from "vite-plugin-svgr";

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    react(),
    tailwindcss(),
    svgr(),
  ],
  server: {
    port: 3000,
    watch: {
      usePolling: true,
    },
    host: true,
    strictPort: true,
  },
  resolve: {
    alias: {
      "src": "/src",
    },
  },
});
