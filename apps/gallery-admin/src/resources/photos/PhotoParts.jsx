import {
  Box,
  Badge,
  Button,
  FormControl,
  FormLabel,
  HStack,
  Image,
  Input,
  Select,
  Stack,
  Switch,
  Text,
  Textarea,
} from "@chakra-ui/react";
import { useEffect, useRef, useState } from "react";
import { Link as RouterLink } from "react-router-dom";
import { photoPreviewURL } from "../../galleryDataProvider";

export const ProcessingStatus = ({ status }) =>
  status && status !== "not_required" ? (
    <Badge colorScheme={status === "ready" ? "green" : "yellow"}>
      {status}
    </Badge>
  ) : null;

export const PhotoPreview = ({ photo, api, size = "72px" }) => {
  const [privatePreviewURL, setPrivatePreviewURL] = useState("");
  const publicPreviewURL = photo.previewURL || photoPreviewURL(photo.src);

  useEffect(() => {
    let active = true;
    if (publicPreviewURL || !api || !photo.id || !photo.originalKey) {
      setPrivatePreviewURL("");
      return () => { active = false; };
    }
    api.getPhotoPreview(photo.id)
      .then((response) => active && setPrivatePreviewURL(response.url))
      .catch(() => active && setPrivatePreviewURL(""));
    return () => { active = false; };
  }, [api, photo.id, photo.originalKey, publicPreviewURL]);

  const src = publicPreviewURL || privatePreviewURL;
  return src ? (
    <Image src={src} alt={photo.altText || photo.title || "Photo preview"} boxSize={size} objectFit="cover" borderRadius="sm" bg="whiteAlpha.100" />
  ) : <Box boxSize={size} borderRadius="sm" bg="whiteAlpha.100" />;
};

export const imageDetails = (file) => new Promise((resolve, reject) => {
  const objectURL = URL.createObjectURL(file);
  const image = new window.Image();
  image.onload = () => {
    URL.revokeObjectURL(objectURL);
    resolve({ width: image.naturalWidth, height: image.naturalHeight });
  };
  image.onerror = () => {
    URL.revokeObjectURL(objectURL);
    reject(new Error("The selected file is not a readable image."));
  };
  image.src = objectURL;
});

export const titleFromFilename = (filename) => filename
  .replace(/\.[^.]+$/, "")
  .replace(/[-_]+/g, " ")
  .replace(/\s+/g, " ")
  .trim();

const googlePhotosClientID = import.meta.env.VITE_GOOGLE_PHOTOS_CLIENT_ID;
const googlePhotosScope = "https://www.googleapis.com/auth/photospicker.mediaitems.readonly";
const googlePhotosAPI = "https://photospicker.googleapis.com/v1";
const maxGoogleImportWidth = 4096;
const maxGoogleImportBytes = 25 * 1024 * 1024;
let googleIdentityPromise;

// Google Identity Services is loaded only when an administrator explicitly
// imports selected Photos items. That keeps the normal drag-and-drop workflow
// dependency-free and avoids requesting the Picker scope at console startup.
const loadGoogleIdentityServices = () => {
  if (window.google?.accounts?.oauth2) return Promise.resolve(window.google);
  if (googleIdentityPromise) return googleIdentityPromise;
  googleIdentityPromise = new Promise((resolve, reject) => {
    const script = document.createElement("script");
    script.src = "https://accounts.google.com/gsi/client";
    script.async = true;
    script.onload = () => window.google?.accounts?.oauth2 ? resolve(window.google) : reject(new Error("Google sign-in could not be initialized."));
    script.onerror = () => reject(new Error("Google sign-in could not be loaded."));
    document.head.appendChild(script);
  });
  return googleIdentityPromise;
};

const requestGooglePhotosToken = () => {
  const google = window.google;
  if (!google?.accounts?.oauth2) {
    return Promise.reject(new Error("Google sign-in is still loading. Try again in a moment."));
  }
  return new Promise((resolve, reject) => {
    const client = google.accounts.oauth2.initTokenClient({
      client_id: googlePhotosClientID,
      scope: googlePhotosScope,
      callback: (response) => response.error || !response.access_token
        ? reject(new Error(response.error_description || "Google Photos access was not granted."))
        : resolve(response.access_token),
      error_callback: () => reject(new Error("Google Photos sign-in was cancelled.")),
    });
    client.requestAccessToken();
  });
};

const googleRequest = async (token, path, options = {}) => {
  const response = await fetch(`${googlePhotosAPI}${path}`, {
    ...options,
    headers: {
      Authorization: `Bearer ${token}`,
      ...(options.body ? { "Content-Type": "application/json" } : {}),
      ...options.headers,
    },
  });
  const body = await response.json().catch(() => ({}));
  if (!response.ok) throw new Error(body.error?.message || "Google Photos could not complete the requested action.");
  return body;
};

const durationMilliseconds = (duration, fallback) => {
  const seconds = Number.parseFloat(String(duration || ""));
  return Number.isFinite(seconds) && seconds >= 0 ? seconds * 1000 : fallback;
};

const wait = (milliseconds) => new Promise((resolve) => window.setTimeout(resolve, milliseconds));

const selectedGoogleItems = async (token, sessionID, initialSession) => {
  let session = initialSession;
  let remaining = durationMilliseconds(session.pollingConfig?.timeoutIn, 5 * 60 * 1000);
  while (!session.mediaItemsSet) {
    const interval = Math.max(durationMilliseconds(session.pollingConfig?.pollInterval, 2000), 500);
    if (remaining <= 0) throw new Error("Google Photos selection timed out before any images were chosen.");
    await wait(Math.min(interval, remaining));
    remaining -= interval;
    session = await googleRequest(token, `/sessions/${encodeURIComponent(sessionID)}`);
  }

  const items = [];
  let pageToken = "";
  do {
    const query = new URLSearchParams({ sessionId: sessionID, pageSize: "50" });
    if (pageToken) query.set("pageToken", pageToken);
    const page = await googleRequest(token, `/mediaItems?${query.toString()}`);
    items.push(...(page.mediaItems || []));
    pageToken = page.nextPageToken || "";
  } while (pageToken);
  return items;
};

const filenameForGoogleItem = (item, contentType) => {
  const provided = item.mediaFile?.filename?.trim();
  if (provided) return provided;
  const extension = { "image/jpeg": "jpg", "image/png": "png", "image/webp": "webp" }[contentType] || "jpg";
  return `google-photo-${item.id || Date.now()}.${extension}`;
};

const downloadGooglePhoto = async (item, token) => {
  if (item.type !== "PHOTO" || !item.mediaFile?.baseUrl) return null;
  // A bounded download is deliberate: the worker still produces responsive
  // derivatives, while this avoids importing an unnecessarily large original
  // from Google Photos into the 25 MB direct-upload contract.
  const response = await fetch(`${item.mediaFile.baseUrl}=w${maxGoogleImportWidth}`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  if (!response.ok) throw new Error("A selected Google Photos image could not be downloaded.");
  const blob = await response.blob();
  const contentType = blob.type || item.mediaFile.mimeType;
  if (!["image/jpeg", "image/png", "image/webp"].includes(contentType)) return null;
  if (blob.size > maxGoogleImportBytes) return null;
  return new File([blob], filenameForGoogleItem(item, contentType), { type: contentType });
};

export const GooglePhotosImport = ({ onSelect, onError, pending, disabled = false }) => {
  const [importing, setImporting] = useState(false);
  const [pickerSession, setPickerSession] = useState(null);
  const [identityReady, setIdentityReady] = useState(Boolean(window.google?.accounts?.oauth2));
  useEffect(() => {
    if (!googlePhotosClientID) return undefined;
    let active = true;
    loadGoogleIdentityServices()
      .then(() => active && setIdentityReady(true))
      .catch((reason) => active && onError(reason));
    return () => { active = false; };
  }, [onError]);
  if (!googlePhotosClientID) return null;

  const preparePicker = async () => {
    setImporting(true);
    onError(null);
    try {
      const token = await requestGooglePhotosToken();
      const session = await googleRequest(token, "/sessions", {
        method: "POST",
        body: JSON.stringify({ pickingConfig: { maxItemCount: "50" } }),
      });
      if (!session.id || !session.pickerUri) throw new Error("Google Photos did not create a picker session.");
      // Authorization and picker browsing each need a browser popup. Keeping
      // them as consecutive clicks avoids consuming one user gesture for two
      // windows, which Chrome correctly blocks as a popup abuse protection.
      setPickerSession({ token, id: session.id, pickerUri: session.pickerUri, session });
    } catch (reason) {
      onError(reason);
    } finally {
      setImporting(false);
    }
  };

  const openPicker = async () => {
    if (!pickerSession) return;
    const pickerWindow = window.open(
      `${pickerSession.pickerUri.replace(/\/$/, "")}/autoclose`,
      "gallery-google-photos-picker",
      "popup,width=1080,height=760",
    );
    if (!pickerWindow) {
      onError(new Error("Allow pop-ups for this admin console to select images from Google Photos."));
      return;
    }

    setImporting(true);
    onError(null);
    try {
      const items = await selectedGoogleItems(pickerSession.token, pickerSession.id, pickerSession.session);
      const files = [];
      for (const item of items) {
        const file = await downloadGooglePhoto(item, pickerSession.token);
        if (file) files.push(file);
      }
      if (files.length === 0) throw new Error("Select at least one JPEG, PNG, or WebP image under 25 MB from Google Photos.");
      await onSelect(files);
    } catch (reason) {
      onError(reason);
    } finally {
      setImporting(false);
      // Session deletion follows successful retrieval or cancellation. It is
      // best effort because expiry/revocation can make the final cleanup call
      // fail after the user has already returned to the admin console.
      googleRequest(pickerSession.token, `/sessions/${encodeURIComponent(pickerSession.id)}`, { method: "DELETE" }).catch(() => {});
      setPickerSession(null);
    }
  };

  return pickerSession
    ? <Button type="button" variant="outline" colorScheme="blue" onClick={openPicker} isDisabled={disabled || pending} isLoading={importing} loadingText="Importing">Open Google Photos</Button>
    : <Button type="button" variant="outline" colorScheme="blue" onClick={preparePicker} isDisabled={disabled || pending || !identityReady} isLoading={importing || !identityReady} loadingText={identityReady ? "Connecting" : "Loading Google"}>Import from Google Photos</Button>;
};

export const PhotoDropzone = ({ onSelect, pending, readOnly = false }) => {
  const inputRef = useRef(null);
  const [dragging, setDragging] = useState(false);
  const selectFiles = (files) => {
    const imageFiles = Array.from(files || []);
    if (imageFiles.length > 0) onSelect(imageFiles);
  };

  return (
    <FormControl isRequired>
      <FormLabel>Image</FormLabel>
      <Box
        borderWidth="1px"
        borderStyle="dashed"
        borderColor={dragging ? "gold.300" : "whiteAlpha.400"}
        borderRadius="md"
        px={6}
        py={10}
        textAlign="center"
        bg={dragging ? "whiteAlpha.100" : "transparent"}
        cursor={readOnly || pending ? "default" : "pointer"}
        onClick={() => !readOnly && !pending && inputRef.current?.click()}
        onDragEnter={(event) => {
          event.preventDefault();
          if (!readOnly && !pending) setDragging(true);
        }}
        onDragOver={(event) => event.preventDefault()}
        onDragLeave={() => setDragging(false)}
        onDrop={(event) => {
          event.preventDefault();
          setDragging(false);
          if (!readOnly && !pending) selectFiles(event.dataTransfer.files);
        }}
      >
        <Stack spacing={3} align="center">
          <Text color="whiteAlpha.800">{pending ? "Uploading images..." : "Drop images here"}</Text>
          <Button
            type="button"
            size="sm"
            variant="outline"
            colorScheme="yellow"
            isDisabled={readOnly || pending}
            onClick={(event) => {
              event.stopPropagation();
              inputRef.current?.click();
            }}
          >
            Choose image
          </Button>
        </Stack>
        <Input ref={inputRef} type="file" accept="image/jpeg,image/png,image/webp" multiple display="none" onChange={(event) => selectFiles(event.target.files)} />
      </Box>
    </FormControl>
  );
};

export const PhotoForm = ({ initialValue, collections, onSubmit, pending, submitLabel, readOnly = false, api }) => {
  const [form, setForm] = useState(initialValue);
  const collectionOptions = collections.filter((collection) => collection.status !== "archived" || collection.id === form.collectionId);

  useEffect(() => setForm(initialValue), [initialValue]);
  const change = (field) => (event) => setForm((current) => ({ ...current, [field]: event.target.value }));

  return (
    <Box as="form" onSubmit={(event) => {
      event.preventDefault();
      if (!readOnly) {
        const tags = form.tagsText.split(",").map((tag) => tag.trim()).filter(Boolean);
        onSubmit({ ...form, tags });
      }
    }}>
      <Stack spacing={5} maxW="780px">
        {form.previewURL || form.src || form.originalKey ? <PhotoPreview photo={form} api={api} size="180px" /> : null}
        <FormControl isRequired><FormLabel>Title</FormLabel><Input value={form.title} onChange={change("title")} isReadOnly={readOnly} autoComplete="off" /></FormControl>
        <FormControl isRequired>
          <FormLabel>Collection</FormLabel>
          <Select value={form.collectionId} onChange={change("collectionId")} isDisabled={readOnly}>
            <option value="" disabled style={{ color: "gray" }}>Choose a collection</option>
            {collectionOptions.map((collection) => (
              <option key={collection.id} value={collection.id} disabled={collection.status === "archived"} style={{ color: collection.status === "archived" ? "gray" : "#f7fafc", backgroundColor: collection.status === "archived" ? "inherit" : "gray" }}>
                {collection.title} ({collection.status})
              </option>
            ))}
          </Select>
        </FormControl>
        <FormControl><FormLabel>Description</FormLabel><Textarea value={form.description} onChange={change("description")} isReadOnly={readOnly} rows={4} resize="vertical" /></FormControl>
        <FormControl><FormLabel>Alt text</FormLabel><Input value={form.altText} onChange={change("altText")} isReadOnly={readOnly} autoComplete="off" /></FormControl>
        <FormControl><FormLabel>Tags</FormLabel><Input value={form.tagsText} onChange={change("tagsText")} isReadOnly={readOnly} autoComplete="off" /></FormControl>
        <HStack align="start" spacing={4} flexWrap="wrap">
          <FormControl display="flex" alignItems="center" minW="160px" pt={8}>
            <FormLabel mb="0">Featured</FormLabel>
            <Switch isChecked={form.featured} onChange={(event) => setForm((current) => ({ ...current, featured: event.target.checked }))} isDisabled={readOnly} />
          </FormControl>
        </HStack>
        <HStack align="start" spacing={4} flexWrap="wrap">
          <FormControl minW="160px" flex="1"><FormLabel>Year</FormLabel><Input value={form.year} onChange={change("year")} isReadOnly={readOnly} /></FormControl>
          <FormControl minW="160px" flex="1"><FormLabel>Location</FormLabel><Input value={form.location} onChange={change("location")} isReadOnly={readOnly} /></FormControl>
        </HStack>
        {!readOnly ? (
          <HStack pt={2}>
            <Button type="submit" colorScheme="yellow" isLoading={pending} loadingText="Saving">{submitLabel}</Button>
            <Button as={RouterLink} to="/photos" variant="solid" colorScheme="green">Cancel</Button>
          </HStack>
        ) : null}
      </Stack>
    </Box>
  );
};

export const usePhotoCollections = (api) => {
  const [collections, setCollections] = useState([]);
  const [error, setError] = useState(null);
  useEffect(() => {
    let active = true;
    api.listCollections().then((items) => active && setCollections(items)).catch((reason) => active && setError(reason));
    return () => { active = false; };
  }, [api]);
  return { collections, collectionError: error };
};
