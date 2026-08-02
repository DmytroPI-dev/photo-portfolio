// The admin UI intentionally talks to the Go API through this small client
// rather than through a framework adapter. Keeping HTTP concerns here makes
// pages ordinary Chakra/React components and keeps future API changes local.
export class GalleryApiError extends Error {
  constructor(message, status, body) {
    super(message);
    this.name = "GalleryApiError";
    this.status = status;
    this.body = body;
    this.code = body?.error?.code;
  }
}

const collectionPayload = (data, includeSlug) => ({
  ...(includeSlug ? { slug: data.slug.trim() } : {}),
  title: data.title.trim(),
  description: data.description.trim(),
  coverPhotoId: data.coverPhotoId.trim(),
  order: Number(data.order),
  ...(includeSlug ? {} : { version: data.version }),
});

export const GalleryApi = (apiBaseUrl, getAccessToken) => {
  const request = async (path, options = {}) => {
    const accessToken = await getAccessToken();
    const response = await fetch(`${apiBaseUrl}${path}`, {
      method: options.method || "GET",
      headers: {
        ...(accessToken ? { Authorization: `Bearer ${accessToken}` } : {}),
        ...(options.body ? { "Content-Type": "application/json" } : {}),
      },
      body: options.body ? JSON.stringify(options.body) : undefined,
    });

    const text = await response.text();
    let body;
    try {
      body = text ? JSON.parse(text) : undefined;
    } catch {
      body = undefined;
    }

    if (!response.ok) {
      throw new GalleryApiError(body?.error?.message || text || "The gallery API request failed.", response.status, body);
    }
    return body;
  };

  return {
    async listCollections() {
      const response = await request("/admin/collections");
      return response.items;
    },

    getCollection: (id) => request(`/admin/collections/${encodeURIComponent(id)}`),

    createCollection: (data) =>
      request("/admin/collections", {
        method: "POST",
        body: collectionPayload(data, true),
      }),

    async updateCollection(id, data) {
      // Use the version captured with the edit form so a concurrent save is
      // rejected by the server instead of silently overwriting newer changes.
      return request(`/admin/collections/${encodeURIComponent(id)}`, {
        method: "PATCH",
        body: collectionPayload(data, false),
      });
    },

    publishCollection: (id, version) =>
      request(`/admin/collections/${encodeURIComponent(id)}/publish`, {
        method: "POST",
        body: { version },
      }),

    archiveCollection: (id, version) =>
      request(`/admin/collections/${encodeURIComponent(id)}/archive`, {
        method: "POST",
        body: { version },
      }),

    restoreCollection: (id, version) =>
      request(`/admin/collections/${encodeURIComponent(id)}/restore`, {
        method: "POST",
        body: { version },
      }),

    deleteCollection: (id, version, confirmationSlug) =>
      request(`/admin/collections/${encodeURIComponent(id)}`, {
        method: "DELETE",
        body: { version, confirmationSlug },
      }),

    async listPhotos() {
      const response = await request("/admin/photos");
      return response.items;
    },

    getPhoto: (id) => request(`/admin/photos/${encodeURIComponent(id)}`),
  };
};
