import { useEffect, useRef, useState } from "react";
import { completeSignIn } from "./authProvider";

export const AuthCallback = () => {
  const hasStarted = useRef(false);
  const [error, setError] = useState("");

  useEffect(() => {
    // Development Strict Mode re-runs effects. The authorization code is
    // single-use, so process it once even while React performs that check.
    if (hasStarted.current) {
      return;
    }

    hasStarted.current = true;
    completeSignIn()
      .then(() => window.location.replace("/collections"))
      .catch((reason) => setError(reason instanceof Error ? reason.message : "Unable to complete sign-in."));
  }, []);

  if (error) {
    return <main className="auth-status">Sign-in could not be completed: {error}</main>;
  }

  return <main className="auth-status">Completing secure sign-in...</main>;
};
