import React from "react";
import ReactDOM from "react-dom/client";
import "./index.css";
import App from "./App";

const rootElement = document.getElementById("root");

if (!rootElement) {
  throw new Error("Photo portfolio could not find its root element.");
}

// Vite can evaluate this entry more than once during development. Check if
// React has already created a root on this element (_reactRootContainer).
// Only create a new root if one doesn't already exist.
let root = rootElement._reactRootContainer;

if (!root) {
  root = ReactDOM.createRoot(rootElement);
}

root.render(
  <React.StrictMode>
    <App />
  </React.StrictMode>
);
