import { useEffect, useRef, useState } from "react";
import { startSignIn } from "./authProvider";

// This app delegates credential entry entirely to Cognito, so the local login
// route immediately begins the hosted authorization-code-with-PKCE flow.
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
    startSignIn().catch((reason) => {
      setError(reason instanceof Error ? reason.message : "Unable to start secure sign-in.");
    });
  }, []);

  if (error) {
    return <main className="auth-status">Sign-in could not be started: {error}</main>;
  }

  return <main className="auth-status">Redirecting to secure sign-in...</main>;
};
