import { HttpError } from "react-admin";

const unsupported = (operation) => () =>
  Promise.reject(new Error(`${operation} is not available until protected admin mutations are deployed.`));

// React-admin remains independent of the Go route shapes. This adapter is the
// only place that understands today's public read API and tomorrow's protected
// admin endpoints, so backend routes can remain small and domain-oriented.
export const GalleryDataProvider = (apiBaseUrl) => {
  const request = async (path) => {
    const response = await fetch(`${apiBaseUrl}${path}`);
    if (!response.ok) {
      throw new HttpError(await response.text(), response.status);
    }
    return response.json();
  };

  const listPhotos = async () => {
    const { items: collections } = await request("/collections");
    const details = await Promise.all(collections.map((collection) => request(`/collections/${collection.slug}`)));
    return details.flatMap((collection) => collection.photos);
  };

  return {
    async getList(resource, params) {
      const records = resource === "collections" ? (await request("/collections")).items : await listPhotos();
      const { page, perPage } = params.pagination;
      const start = (page - 1) * perPage;

      return {
        data: records.slice(start, start + perPage),
        total: records.length,
      };
    },

    async getOne(resource, params) {
      if (resource === "collections") {
        return { data: await request(`/collections/${params.id}`) };
      }
      return { data: await request(`/photos/${params.id}`) };
    },

    async getMany(resource, params) {
      const records = await Promise.all(
        params.ids.map((id) => request(resource === "collections" ? `/collections/${id}` : `/photos/${id}`)),
      );
      return { data: records };
    },

    getManyReference: unsupported("getManyReference"),
    create: unsupported("create"),
    update: unsupported("update"),
    updateMany: unsupported("updateMany"),
    delete: unsupported("delete"),
    deleteMany: unsupported("deleteMany"),
  };
};
