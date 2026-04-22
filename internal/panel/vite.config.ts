import { defineConfig, loadEnv } from "vite";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";
import path from "path";

// https://vitejs.dev/config/
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, path.resolve(__dirname, "../../"), "");

  const adminHost = env.ADMIN_PANEL_HOST && env.ADMIN_PANEL_HOST !== "0.0.0.0" ? env.ADMIN_PANEL_HOST : "localhost";
  const adminPort = env.ADMIN_PANEL_PORT || "8080";
  const adminTarget = `http://${adminHost}:${adminPort}`;

  return {
    plugins: [react(), tailwindcss()],
    envDir: "../../",
    envPrefix: ["VITE_", "ADMIN_", "PUBLIC_"],
    resolve: {
      alias: {
        "@": path.resolve(__dirname, "./src"),
      },
    },
    server: {
      port: parseInt(env.ADMIN_PANEL_PORT || "8096"),
      watch: {
        ignored: ["**/.env"],
      },
      proxy: {
        "/_api": {
          target: adminTarget,
          changeOrigin: true,
          secure: false,
        },
      },
    },
    build: {
      chunkSizeWarningLimit: 1000,
      rollupOptions: {
        output: {
          manualChunks(id) {
            if (id.includes("node_modules")) {
              return "vendor";
            }
          },
        },
      },
    },
  };
});
