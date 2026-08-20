import React from "react";
import ReactDOM from "react-dom/client";
import "./index.css";
import App from "./App";
import { GalleryDataProvider } from "./data/galleryData";

const rootElement = document.getElementById("root");

if (!rootElement) {
  throw new Error("Photo portfolio could not find its root element.");
}

// Vite may re-evaluate this module during HMR. Keep the actual React root in
// a stable application-owned location rather than a React-internal DOM field.
const root = window.__photoPortfolioRoot || (window.__photoPortfolioRoot = ReactDOM.createRoot(rootElement));

root.render(
  <React.StrictMode>
    <GalleryDataProvider>
      <App />
    </GalleryDataProvider>
  </React.StrictMode>
);
