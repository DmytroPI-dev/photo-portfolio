import { createContext, useContext, useEffect, useMemo, useState } from "react";
import { collections as fallbackCollections } from "./collections";
import { photos as fallbackPhotos } from "./photos";

const GalleryDataContext = createContext(null);

// Vite replaces this value when it builds the public site. Leaving it blank is
// intentional: local development stays self-contained and renders the bundled
// placeholder collection data until an API URL is explicitly configured.
const publicApiBaseUrl = (import.meta.env.VITE_GALLERY_API_URL ?? "")
  .trim()
  .replace(/\/+$/, "");

const sortByOrder = (left, right) => left.order - right.order;

const fallbackGallery = {
  collections: [...fallbackCollections].sort(sortByOrder),
  photos: [...fallbackPhotos].sort(sortByOrder),
};

async function readJSON(path, signal) {
  const response = await fetch(`${publicApiBaseUrl}${path}`, { signal });
  if (!response.ok) {
    throw new Error(`gallery API request failed with ${response.status}`);
  }
  return response.json();
}

async function loadPublicGallery(signal) {
  const collectionList = await readJSON("/collections", signal);
  if (!Array.isArray(collectionList.items)) {
    throw new Error("gallery API returned an invalid collection list");
  }

  // Collection detail is the public API's intentionally compact read model: it
  // contains both collection metadata and its ordered published photos. Fetch
  // every published room before changing UI state, so Home never mixes a fresh
  // collection list with stale placeholder photos from another source.
  const details = await Promise.all(
    collectionList.items.map((collection) =>
      readJSON(`/collections/${encodeURIComponent(collection.slug)}`, signal)
    )
  );

  const collections = [];
  const photos = [];
  details.forEach((detail, index) => {
    const { photos: collectionPhotos, ...collection } = detail;
    if (!Array.isArray(collectionPhotos)) {
      throw new Error("gallery API returned an invalid collection detail");
    }

    collections.push({ ...collectionList.items[index], ...collection });
    photos.push(...collectionPhotos);
  });

  return {
    collections: collections.sort(sortByOrder),
    photos: photos.sort(sortByOrder),
  };
}

// The public gallery has no authentication dependency. It renders bundled
// artwork immediately, upgrades to the published API snapshot when configured,
// and deliberately keeps the fallback if the API is unavailable.
export function GalleryDataProvider({ children }) {
  const [gallery, setGallery] = useState(fallbackGallery);
  const [isLoading, setIsLoading] = useState(Boolean(publicApiBaseUrl));
  const [source, setSource] = useState("fallback");

  useEffect(() => {
    if (!publicApiBaseUrl) {
      return undefined;
    }

    const controller = new AbortController();

    loadPublicGallery(controller.signal)
      .then((remoteGallery) => {
        setGallery(remoteGallery);
        setSource("api");
      })
      .catch((error) => {
        // Abort is expected when React Strict Mode replays effects or the user
        // navigates away. Other failures intentionally retain the fallback:
        // a temporary API outage should not turn a visual portfolio into a
        // blank page.
        if (error.name !== "AbortError") {
          setGallery(fallbackGallery);
          setSource("fallback");
        }
      })
      .finally(() => {
        if (!controller.signal.aborted) {
          setIsLoading(false);
        }
      });

    return () => controller.abort();
  }, []);

  const value = useMemo(
    () => ({ ...gallery, isLoading, source }),
    [gallery, isLoading, source]
  );

  return (
    <GalleryDataContext.Provider value={value}>
      {children}
    </GalleryDataContext.Provider>
  );
}

export function useGalleryData() {
  const gallery = useContext(GalleryDataContext);
  if (!gallery) {
    throw new Error("useGalleryData must be used inside GalleryDataProvider");
  }
  return gallery;
}
