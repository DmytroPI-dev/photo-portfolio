import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig({
  plugins: [react()],
  server: {
    port: 5174,
    strictPort: true,
    // Development requests stay same-origin, so the admin can be exercised
    // against the deployed read API before its localhost:5174 CORS allowance
    // is applied. Production uses VITE_GALLERY_API_URL directly.
    proxy: {
      "/api": {
        target: "https://m02dauw9h9.execute-api.eu-central-1.amazonaws.com",
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api/, ""),
      },
    },
  },
});
