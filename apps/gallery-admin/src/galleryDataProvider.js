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

const photoPayload = (data, includeUpload) => ({
  ...(includeUpload
    ? { uploadId: data.uploadId, originalKey: data.originalKey }
    : { version: data.version }),
  title: data.title.trim(),
  description: data.description.trim(),
  collectionId: data.collectionId,
  width: Number(data.width),
  height: Number(data.height),
  year: data.year.trim(),
  location: data.location.trim(),
  featured: Boolean(data.featured),
  order: Number(data.order),
  altText: data.altText.trim(),
  tags: data.tags,
});

// The placeholders belong to the public portfolio rather than the separate
// admin bundle. Keeping this origin configurable lets the same metadata show
// a real preview locally and after the sites move to independent hostnames.
const previewOrigin = (import.meta.env.VITE_GALLERY_PREVIEW_ORIGIN || "https://photo-gallery.i-dmytro.org").replace(/\/$/, "");

export const photoPreviewURL = (src) => {
  if (!src || /^https?:\/\//i.test(src)) {
    return src;
  }
  return `${previewOrigin}${src.startsWith("/") ? src : `/${src}`}`;
};

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

    getPhotoPreview: (id) =>
      request(`/admin/photos/${encodeURIComponent(id)}/preview`),

    async uploadOriginal(file) {
      const contentType = file.type || imageContentType(file.name);
      const upload = await request("/admin/uploads", {
        method: "POST",
        body: {
          filename: file.name,
          contentType,
          size: file.size,
        },
      });
      const body = new FormData();
      Object.entries(upload.uploadFields).forEach(([name, value]) => body.append(name, value));
      body.append("file", file);
      const response = await fetch(upload.uploadUrl, { method: "POST", body });
      if (!response.ok) {
        throw new Error("The image could not be uploaded to private storage.");
      }
      return upload;
    },

    createPhoto: (data) =>
      request("/admin/photos", {
        method: "POST",
        body: photoPayload(data, true),
      }),

    updatePhoto: (id, data) =>
      request(`/admin/photos/${encodeURIComponent(id)}`, {
        method: "PATCH",
        body: photoPayload(data, false),
      }),

    publishPhoto: (id, version) =>
      request(`/admin/photos/${encodeURIComponent(id)}/publish`, {
        method: "POST",
        body: { version },
      }),

    archivePhoto: (id, version) =>
      request(`/admin/photos/${encodeURIComponent(id)}/archive`, {
        method: "POST",
        body: { version },
      }),

    restorePhoto: (id, version) =>
      request(`/admin/photos/${encodeURIComponent(id)}/restore`, {
        method: "POST",
        body: { version },
      }),

    async reorderPhotos(collectionId, photos) {
      const response = await request("/admin/photos/reorder", {
        method: "POST",
        body: { collectionId, photos: photos.map(({ id, version }) => ({ id, version })) },
      });
      return response.items;
    },
  };
};

const imageContentType = (filename) => {
  const extension = filename.split(".").pop()?.toLowerCase();
  return (
    {
      jpg: "image/jpeg",
      jpeg: "image/jpeg",
      png: "image/png",
      webp: "image/webp",
    }[extension] || "application/octet-stream"
  );
};
