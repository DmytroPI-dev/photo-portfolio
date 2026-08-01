import { useEffect, useRef, useState } from "react";
import { authProvider } from "./authProvider";

// React-admin normally renders a local username/password form at /login. This
// application delegates credential entry entirely to Cognito, so reaching that
// route immediately starts the hosted-login authorization-code flow instead.
export const CognitoLogin = () => {
  const hasStarted = useRef(false);
  const [error, setError] = useState("");

  useEffect(() => {
    // React Strict Mode runs effects twice in development. The guard prevents a
    // second PKCE state record and a competing Cognito redirect.
    if (hasStarted.current) {
      return;
    }

    hasStarted.current = true;
    authProvider.login().catch((reason) => {
      setError(reason instanceof Error ? reason.message : "Unable to start secure sign-in.");
    });
  }, []);

  if (error) {
    return <main className="auth-status">Sign-in could not be started: {error}</main>;
  }

  return <main className="auth-status">Redirecting to secure sign-in...</main>;
};
