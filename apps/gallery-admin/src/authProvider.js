import { UserManager, WebStorageStateStore } from "oidc-client-ts";

const authority = import.meta.env.VITE_COGNITO_AUTHORITY;
const clientId = import.meta.env.VITE_COGNITO_CLIENT_ID;
const cognitoDomain = import.meta.env.VITE_COGNITO_DOMAIN;
const scope = import.meta.env.VITE_COGNITO_SCOPES;

export const isAuthConfigured = Boolean(authority && clientId && cognitoDomain && scope);

// The SPA stores short-lived tokens in sessionStorage so closing the browser
// ends the local session. No client secret is present: Cognito protects this
// authorization-code flow with PKCE.
const userManager = isAuthConfigured
  ? new UserManager({
      authority,
      client_id: clientId,
      redirect_uri: `${window.location.origin}/auth/callback`,
      post_logout_redirect_uri: `${window.location.origin}/`,
      response_type: "code",
      scope,
      userStore: new WebStorageStateStore({ store: window.sessionStorage }),
    })
  : null;

const requireUserManager = () => {
  if (!userManager) {
    throw new Error("Cognito configuration is missing. Add the VITE_COGNITO_* values to .env.local.");
  }
  return userManager;
};

// Cognito's hosted-login page owns credential entry. Keeping this redirect in
// one function prevents the SPA from ever handling a password or client secret.
export const startSignIn = () =>
  requireUserManager().signinRedirect({
    redirect_uri: `${window.location.origin}/auth/callback`,
    response_type: "code",
    scope,
  });

export const signOut = async () => {
  const manager = requireUserManager();
  const user = await manager.getUser();
  await manager.removeUser();

  // Cognito's hosted /logout endpoint is not a standard OIDC end-session
  // endpoint. It needs the registered client ID and logout URI explicitly.
  // Avoid redirecting there when no Cognito session exists, for example after a
  // locally expired access token.
  if (!user?.id_token) {
    window.location.assign("/");
    return;
  }

  const logoutUrl = new URL("/logout", cognitoDomain);
  logoutUrl.searchParams.set("client_id", clientId);
  logoutUrl.searchParams.set("logout_uri", `${window.location.origin}/`);
  window.location.assign(logoutUrl.toString());
};

// The callback route completes PKCE before protected console routes inspect
// the session. Keeping it here gives the app one owner for token lifecycle.
export const completeSignIn = () => requireUserManager().signinRedirectCallback();

export const getAuthenticatedUser = async () => {
  const user = await requireUserManager().getUser();
  return user && !user.expired ? user : null;
};

export const getAccessToken = async () => {
  const user = await getAuthenticatedUser();
  return user?.access_token || null;
};
