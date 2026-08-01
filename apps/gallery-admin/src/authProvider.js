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

export const authProvider = {
  // React-admin may call login with form/redirect parameters of its own. Keep
  // the Cognito request explicit so only the registered PKCE callback and
  // approved scopes are ever sent to the authorization endpoint.
  login: () =>
    requireUserManager().signinRedirect({
      redirect_uri: `${window.location.origin}/auth/callback`,
      response_type: "code",
      scope,
    }),

  async logout() {
    const manager = requireUserManager();
    const user = await manager.getUser();
    await manager.removeUser();

    // Cognito's managed /logout endpoint isn't a standard OIDC end-session
    // endpoint. It requires its app client ID and an allow-listed `logout_uri`,
    // whereas oidc-client-ts sends an OIDC `id_token_hint`. Build Cognito's
    // documented URL directly, after clearing the app's local token session.
    // React-admin also calls logout when an anonymous visitor is redirected to
    // /login; only contact Cognito when there was an actual OIDC session.
    if (!user?.id_token) {
      return;
    }

    const logoutUrl = new URL("/logout", cognitoDomain);
    logoutUrl.searchParams.set("client_id", clientId);
    logoutUrl.searchParams.set("logout_uri", `${window.location.origin}/`);
    window.location.assign(logoutUrl.toString());
  },

  async checkAuth() {
    const user = await requireUserManager().getUser();
    if (!user || user.expired) {
      return Promise.reject();
    }
  },

  async checkError(error) {
    if (error?.status === 401 || error?.status === 403) {
      await requireUserManager().removeUser();
      return Promise.reject();
    }
  },

  async getIdentity() {
    const user = await requireUserManager().getUser();
    if (!user || user.expired) {
      return Promise.reject();
    }

    return {
      id: user.profile.sub,
      fullName: user.profile.name || user.profile.email || user.profile.username,
    };
  },

  async getPermissions() {
    const user = await requireUserManager().getUser();
    return user?.scope?.split(" ") || [];
  },
};

// The callback route completes the PKCE exchange before React-admin checks the
// session. Keeping it here gives the app one authority for login, logout, and
// future bearer tokens supplied to protected /admin API requests.
export const completeSignIn = () => requireUserManager().signinRedirectCallback();

export const getAccessToken = async () => {
  const user = await requireUserManager().getUser();
  return user && !user.expired ? user.access_token : null;
};
