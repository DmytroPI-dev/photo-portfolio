import React from "react";
import { createRoot } from "react-dom/client";
import { Admin, Resource, defaultTheme } from "react-admin";
import { GalleryDataProvider } from "./galleryDataProvider";
import { CollectionList, CollectionShow } from "./resources/collections";
import { PhotoList, PhotoShow } from "./resources/photos";
import "./styles.css";

// In local development Vite proxies /api to the deployed read API. A deployed
// console receives its explicit API origin through VITE_GALLERY_API_URL.
const apiUrl = import.meta.env.VITE_GALLERY_API_URL || "/api";

// This theme intentionally stays operational and restrained. The public
// portfolio owns the cinematic visual language; the administrator needs a
// quiet surface for checking metadata and, later, making careful edits.
const theme = {
  ...defaultTheme,
  palette: {
    ...defaultTheme.palette,
    mode: "dark",
    primary: { main: "#c7a36b" },
    background: { default: "#151515", paper: "#202020" },
  },
};

createRoot(document.getElementById("root")).render(
  <React.StrictMode>
    <Admin dataProvider={GalleryDataProvider(apiUrl)} theme={theme} title="Gallery Administration">
      <Resource name="collections" list={CollectionList} show={CollectionShow} />
      <Resource name="photos" list={PhotoList} show={PhotoShow} />
    </Admin>
  </React.StrictMode>,
);
