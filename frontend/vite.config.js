import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

const backendPort = process.env.BACKEND_PORT || "8080";

export default defineConfig({
  root: __dirname,
  plugins: [react()],
  build: {
    outDir: "../web",
    emptyOutDir: true,
  },
  server: {
    // Keep this default in sync with the backend PORT in token.env(.example).
    proxy: {
      "/api": `http://localhost:${backendPort}`,
    },
  },
});
