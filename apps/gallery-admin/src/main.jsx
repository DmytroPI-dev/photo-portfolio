import React, { useEffect, useMemo, useState } from "react";
import { createRoot } from "react-dom/client";
import { ChakraProvider, CSSReset, extendTheme } from "@chakra-ui/react";
import { BrowserRouter, Navigate, Route, Routes, useLocation } from "react-router-dom";
import { AdminConsole } from "./AdminConsole";
import { AuthCallback } from "./AuthCallback";
import { CognitoLogin } from "./CognitoLogin";
import { getAccessToken, getAuthenticatedUser, isAuthConfigured } from "./authProvider";
import { GalleryApi } from "./galleryDataProvider";
import "./styles.css";

// Development requests use Vite's same-origin /api proxy. A deployed console
// receives the API hostname through VITE_GALLERY_API_URL instead.
const apiUrl = import.meta.env.VITE_GALLERY_API_URL || "/api";

// This deliberately shares Chakra with the public site, but keeps the internal
// console dense, neutral, and legible during long metadata-editing sessions.
const theme = extendTheme({
  config: { initialColorMode: "dark", useSystemColorMode: false },
  colors: { gold: { 300: "#d3ad78", 400: "#b88e56", 500: "#9c713f" } },
  styles: { global: { body: { bg: "gray.950", color: "whiteAlpha.900" } } },
  fonts: { body: "system-ui, sans-serif", heading: "system-ui, sans-serif" },
});

const SessionGate = () => {
  const location = useLocation();
  const [user, setUser] = useState(undefined);
  const api = useMemo(() => GalleryApi(apiUrl, getAccessToken), []);

  useEffect(() => {
    let alive = true;
    getAuthenticatedUser().then((currentUser) => alive && setUser(currentUser));
    return () => { alive = false; };
  }, []);

  if (user === undefined) {
    return <main className="auth-status">Checking secure session...</main>;
  }
  if (!user) {
    return <Navigate to="/login" replace state={{ from: location.pathname }} />;
  }
  return <AdminConsole api={api} user={user} />;
};

const App = () => {
  if (!isAuthConfigured) {
    return <main className="auth-status">Cognito configuration is required before this administrator console can start.</main>;
  }

  return (
    <ChakraProvider theme={theme}>
      <CSSReset />
      <BrowserRouter>
        <Routes>
          <Route path="/auth/callback" element={<AuthCallback />} />
          <Route path="/login" element={<CognitoLogin />} />
          <Route path="*" element={<SessionGate />} />
        </Routes>
      </BrowserRouter>
    </ChakraProvider>
  );
};

const rootElement = document.getElementById("root");
if (!rootElement) {
  throw new Error("Gallery administration could not find its root element.");
}

// Vite may re-evaluate this module during HMR. Store the actual React root on
// window, not on a React-internal DOM property, so development never calls
// createRoot twice for the same container.
const root = window.__galleryAdminRoot || (window.__galleryAdminRoot = createRoot(rootElement));

root.render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
);
