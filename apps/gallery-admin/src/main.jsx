import React from "react";
import { createRoot } from "react-dom/client";
import { Admin, Resource, defaultTheme } from "react-admin";
import { BrowserRouter, Route, Routes } from "react-router-dom";
import { AuthCallback } from "./AuthCallback";
import { CognitoLogin } from "./CognitoLogin";
import { authProvider, getAccessToken, isAuthConfigured } from "./authProvider";
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

const AdminConsole = () => (
  <Admin
    authProvider={authProvider}
    dataProvider={GalleryDataProvider(apiUrl, getAccessToken)}
    loginPage={CognitoLogin}
    requireAuth
    theme={theme}
    title="Gallery Administration"
  >
    <Resource name="collections" list={CollectionList} show={CollectionShow} />
    <Resource name="photos" list={PhotoList} show={PhotoShow} />
  </Admin>
);

// The code callback cannot live inside <Admin>: requireAuth would send it to
// /login before oidc-client-ts could exchange Cognito's code for tokens.
// Routing it outside React-admin leaves admin resources protected while the
// short callback remains publicly reachable and purpose-specific.
const App = () => {
  if (!isAuthConfigured) {
    return <main className="auth-status">Cognito configuration is required before this administrator console can start.</main>;
  }

  return (
    <BrowserRouter>
      <Routes>
        <Route path="/auth/callback" element={<AuthCallback />} />
        <Route path="*" element={<AdminConsole />} />
      </Routes>
    </BrowserRouter>
  );
};

const rootElement = document.getElementById("root");

if (!rootElement) {
  throw new Error("Gallery administration could not find its root element.");
}

// Vite can re-evaluate this entry module during development and may replace
// the root DOM node. Check if React has already created a root on this element
// (React stores it as _reactRootContainer). Only create a new root if one
// doesn't already exist, preventing the "already passed to createRoot" warning.
let root = rootElement._reactRootContainer;

if (!root) {
  root = createRoot(rootElement);
}

root.render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
);
