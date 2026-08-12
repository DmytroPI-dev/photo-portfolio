import {
  AddIcon,
  ArrowBackIcon,
  ArrowDownIcon,
  ArrowUpIcon,
  DragHandleIcon,
  EditIcon,
} from "@chakra-ui/icons";
import {
  Box,
  Button,
  Flex,
  FormControl,
  FormLabel,
  Heading,
  HStack,
  IconButton,
  Select,
  Spinner,
  Stack,
  Table,
  Tbody,
  Td,
  Text,
  Th,
  Thead,
  Tr,
  useToast,
} from "@chakra-ui/react";
import { useCallback, useEffect, useState } from "react";
import { Link as RouterLink, useNavigate, useParams } from "react-router-dom";
import { EmptyState, ErrorNotice, LoadingState, Page, StatusBadge } from "../../components/ConsolePrimitives";
import { GalleryApiError } from "../../galleryDataProvider";
import {
  imageDetails,
  PhotoDropzone,
  PhotoForm,
  PhotoPreview,
  ProcessingStatus,
  titleFromFilename,
  usePhotoCollections,
} from "./PhotoParts";

const groupPhotosByCollection = (photos) => photos.reduce((groups, photo) => {
  const group = groups.get(photo.collectionId) || [];
  group.push(photo);
  groups.set(photo.collectionId, group);
  return groups;
}, new Map());

export const PhotosPage = ({ api }) => {
  const [photos, setPhotos] = useState([]);
  const [error, setError] = useState(null);
  const [loading, setLoading] = useState(true);
  const [pendingCollection, setPendingCollection] = useState(null);
  const [draggedID, setDraggedID] = useState(null);
  const navigate = useNavigate();

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      setPhotos(await api.listPhotos());
    } catch (reason) {
      setError(reason);
    } finally {
      setLoading(false);
    }
  }, [api]);
  useEffect(() => { load(); }, [load]);

  const reorder = async (collectionId, collectionPhotos, fromID, toID) => {
    const fromIndex = collectionPhotos.findIndex((photo) => photo.id === fromID);
    const toIndex = collectionPhotos.findIndex((photo) => photo.id === toID);
    if (fromIndex < 0 || toIndex < 0 || fromIndex === toIndex) return;
    const nextOrder = [...collectionPhotos];
    nextOrder.splice(toIndex, 0, nextOrder.splice(fromIndex, 1)[0]);
    setPendingCollection(collectionId);
    setError(null);
    try {
      const updated = await api.reorderPhotos(collectionId, nextOrder);
      const byID = new Map(updated.map((photo) => [photo.id, photo]));
      setPhotos((current) => current.map((photo) => byID.get(photo.id) || photo));
    } catch (reason) {
      setError(reason);
    } finally {
      setPendingCollection(null);
      setDraggedID(null);
    }
  };

  const grouped = groupPhotosByCollection(photos);
  return (
    <Page>
      <Flex justify="space-between" align={{ base: "flex-start", sm: "center" }} gap={4} mb={7} direction={{ base: "column", sm: "row" }}>
        <Box>
          <Heading size="lg">Photos</Heading>
          <Text mt={2} color="whiteAlpha.700">Arrange works within each collection and manage publication metadata.</Text>
        </Box>
        <Button leftIcon={<AddIcon />} colorScheme="yellow" as={RouterLink} to="/photos/new">New photo</Button>
      </Flex>
      <ErrorNotice error={error} />
      {loading ? <LoadingState /> : null}
      {!loading && !error && photos.length === 0 ? <EmptyState>No photos have been created yet.</EmptyState> : null}
      {!loading && !error ? (
        <Stack spacing={8}>
          {[...grouped.entries()].map(([collectionId, collectionPhotos]) => (
            <Box key={collectionId} borderWidth="1px" borderColor="whiteAlpha.200" borderRadius="md" overflowX="auto">
              <Flex px={5} py={4} align="center" justify="space-between" borderBottomWidth="1px" borderColor="whiteAlpha.200">
                <Heading size="sm">{collectionId}</Heading>
                {pendingCollection === collectionId ? <Spinner size="sm" color="gold.300" /> : null}
              </Flex>
              <Table variant="striped" colorScheme="whiteAlpha">
                <Thead><Tr bg="gray.300"><Th aria-label="Reorder" /><Th>Preview</Th><Th>Order</Th><Th>Title</Th><Th>Status</Th><Th>Featured</Th><Th aria-label="Actions" /></Tr></Thead>
                <Tbody>
                  {collectionPhotos.sort((left, right) => left.order - right.order).map((photo, index) => (
                    <Tr
                      key={photo.id}
                      draggable={!pendingCollection}
                      onDragStart={() => setDraggedID(photo.id)}
                      onDragOver={(event) => event.preventDefault()}
                      onDrop={() => reorder(collectionId, collectionPhotos, draggedID, photo.id)}
                      _hover={{ bg: "whiteAlpha.50" }}
                    >
                      <Td><IconButton aria-label={`Drag ${photo.title}`} icon={<DragHandleIcon />} size="sm" variant="solid" colorScheme="green" cursor="grab" /></Td>
                      <Td><PhotoPreview photo={photo} api={api} size="48px" /></Td>
                      <Td>{photo.order}</Td>
                      <Td fontWeight="semibold">{photo.title}</Td>
                      <Td><StatusBadge status={photo.status} /><ProcessingStatus status={photo.processingStatus} /></Td>
                      <Td>{photo.featured ? "Yes" : "No"}</Td>
                      <Td textAlign="right">
                        <HStack justify="flex-end">
                          <IconButton aria-label={`Move ${photo.title} up`} icon={<ArrowUpIcon />} size="sm" variant="solid" colorScheme="green" isDisabled={index === 0 || pendingCollection === collectionId} onClick={() => reorder(collectionId, collectionPhotos, photo.id, collectionPhotos[index - 1].id)} />
                          <IconButton aria-label={`Move ${photo.title} down`} icon={<ArrowDownIcon />} size="sm" variant="solid" colorScheme="yellow" isDisabled={index === collectionPhotos.length - 1 || pendingCollection === collectionId} onClick={() => reorder(collectionId, collectionPhotos, photo.id, collectionPhotos[index + 1].id)} />
                          <IconButton aria-label={`Edit ${photo.title}`} icon={<EditIcon />} size="sm" variant="solid" colorScheme="teal" onClick={() => navigate(`/photos/${photo.id}`)} />
                        </HStack>
                      </Td>
                    </Tr>
                  ))}
                </Tbody>
              </Table>
            </Box>
          ))}
        </Stack>
      ) : null}
    </Page>
  );
};

export const PhotoCreatePage = ({ api }) => {
  const { collections, collectionError } = usePhotoCollections(api);
  const [collectionId, setCollectionId] = useState("");
  const [pending, setPending] = useState(false);
  const [uploadedPhotos, setUploadedPhotos] = useState([]);
  const [error, setError] = useState(null);

  const upload = async (files) => {
    if (!collectionId) {
      setError(new Error("Choose a collection before adding images."));
      return;
    }
    setPending(true);
    setError(null);
    try {
      // Finish each image before starting the next so automatic collection
      // ordering remains deterministic and partial batch results stay visible.
      for (const file of files) {
        const details = await imageDetails(file);
        const uploadDetails = await api.uploadOriginal(file);
        const title = titleFromFilename(file.name) || "Untitled photo";
        const photo = await api.createPhoto({
          uploadId: uploadDetails.photoId,
          originalKey: uploadDetails.originalKey,
          title,
          description: "",
          collectionId,
          width: details.width,
          height: details.height,
          year: "",
          location: "",
          featured: false,
          order: 0,
          altText: title,
          tags: [],
        });
        setUploadedPhotos((current) => [photo, ...current]);
      }
    } catch (reason) {
      setError(reason);
    } finally {
      setPending(false);
    }
  };

  return (
    <Page>
      <HStack spacing={3} mb={7}>
        <IconButton as={RouterLink} to="/photos" aria-label="Back to photos" icon={<ArrowBackIcon />} variant="solid" colorScheme="yellow" />
        <Heading size="lg">Add photos</Heading>
      </HStack>
      <ErrorNotice error={error || collectionError} />
      <Stack spacing={5} maxW="780px">
        <FormControl isRequired>
          <FormLabel>Collection</FormLabel>
          <Select value={collectionId} onChange={(event) => setCollectionId(event.target.value)} isDisabled={pending}>
            <option value="" disabled style={{ color: "gray" }}>Choose a collection</option>
            {collections.filter((collection) => collection.status !== "archived").map((collection) => (
              <option key={collection.id} value={collection.id} style={{ color: "white", backgroundColor: "#1f2937" }}>{collection.title}</option>
            ))}
          </Select>
        </FormControl>
        <PhotoDropzone onSelect={upload} pending={pending} />
        {uploadedPhotos.length > 0 ? (
          <Stack spacing={3} pt={2}>
            {uploadedPhotos.map((photo) => (
              <HStack key={photo.id} spacing={3}>
                <PhotoPreview photo={photo} api={api} size="56px" />
                <Box><Text fontWeight="semibold">{photo.title}</Text><HStack spacing={2} mt={1}><StatusBadge status={photo.status} /><ProcessingStatus status={photo.processingStatus} /></HStack></Box>
              </HStack>
            ))}
          </Stack>
        ) : null}
      </Stack>
    </Page>
  );
};

export const PhotoDetailPage = ({ api }) => {
  const { id } = useParams();
  const { collections, collectionError } = usePhotoCollections(api);
  const [photo, setPhoto] = useState(null);
  const [error, setError] = useState(null);
  const [loading, setLoading] = useState(true);
  const [pending, setPending] = useState(false);
  const toast = useToast();
  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      setPhoto(await api.getPhoto(id));
    } catch (reason) {
      setError(reason);
    } finally {
      setLoading(false);
    }
  }, [api, id]);
  useEffect(() => { load(); }, [load]);

  const handleMutationError = async (reason) => {
    if (reason instanceof GalleryApiError && reason.code === "version_conflict") {
      toast({ title: "Photo changed elsewhere", description: "The latest server version has been loaded.", status: "warning", duration: 5000, isClosable: true });
      await load();
    } else {
      setError(reason);
    }
  };
  const save = async (form) => {
    setPending(true);
    setError(null);
    try {
      setPhoto(await api.updatePhoto(id, form));
      toast({ title: "Photo saved", status: "success", duration: 3000, isClosable: true });
    } catch (reason) {
      await handleMutationError(reason);
    } finally {
      setPending(false);
    }
  };
  const transition = async (action) => {
    setPending(true);
    setError(null);
    try {
      const updated = action === "publish"
        ? await api.publishPhoto(id, photo.version)
        : action === "archive"
          ? await api.archivePhoto(id, photo.version)
          : await api.restorePhoto(id, photo.version);
      setPhoto(updated);
      toast({ title: action === "publish" ? "Photo published" : action === "archive" ? "Photo archived" : "Photo restored as draft", status: "success", duration: 3000, isClosable: true });
    } catch (reason) {
      await handleMutationError(reason);
    } finally {
      setPending(false);
    }
  };

  if (loading) return <Page><LoadingState /></Page>;
  if (!photo) return <Page><ErrorNotice error={error || collectionError} /></Page>;
  const formValue = { ...photo, tags: photo.tags || [], tagsText: (photo.tags || []).join(", ") };

  return (
    <Page>
      <Flex justify="space-between" align={{ base: "flex-start", sm: "center" }} gap={4} mb={7} direction={{ base: "column", sm: "row" }}>
        <HStack spacing={3}>
          <IconButton as={RouterLink} to="/photos" aria-label="Back to photos" icon={<ArrowBackIcon />} variant="solid" colorScheme="yellow" />
          <Box>
            <Heading size="lg">Edit photo</Heading>
            <HStack mt={1} spacing={3}><StatusBadge status={photo.status} /><ProcessingStatus status={photo.processingStatus} /><Text color="whiteAlpha.600">Version {photo.version}</Text></HStack>
          </Box>
        </HStack>
        <HStack>
          {photo.status === "draft" ? <Button colorScheme="green" onClick={() => transition("publish")} isLoading={pending} isDisabled={photo.processingStatus && photo.processingStatus !== "ready" && photo.processingStatus !== "not_required"}>Publish</Button> : null}
          {photo.status === "draft" || photo.status === "published" ? <Button colorScheme="orange" onClick={() => transition("archive")} isLoading={pending}>Archive</Button> : null}
          {photo.status === "archived" ? <Button colorScheme="blue" onClick={() => transition("restore")} isLoading={pending}>Restore</Button> : null}
        </HStack>
      </Flex>
      <ErrorNotice error={error || collectionError} />
      <PhotoForm initialValue={formValue} collections={collections} onSubmit={save} pending={pending} submitLabel="Save changes" readOnly={photo.status === "archived"} api={api} />
    </Page>
  );
};
