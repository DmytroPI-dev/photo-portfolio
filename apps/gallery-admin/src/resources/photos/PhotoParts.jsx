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
      if (!readOnly) onSubmit({ ...form, tags: form.tagsText.split(",") });
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
