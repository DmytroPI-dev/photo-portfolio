import { Box } from "@chakra-ui/react";
import { Navigate, Route, Routes } from "react-router-dom";
import { ConsoleHeader } from "./components/ConsoleHeader";
import {
  CollectionCreatePage,
  CollectionEditPage,
  CollectionsPage,
} from "./resources/collections/CollectionPages";
import {
  PhotoCreatePage,
  PhotoDetailPage,
  PhotosPage,
} from "./resources/photos/PhotoPages";

// Feature modules own their screens and local interaction state. This file is
// deliberately limited to the protected shell and the resource route map.
export const AdminConsole = ({ api, user }) => (
  <Box minH="100vh" bg="gray.950" color="whiteAlpha.900">
    <ConsoleHeader user={user} />
    <Routes>
      <Route path="/" element={<Navigate to="/collections" replace />} />
      <Route path="/collections" element={<CollectionsPage api={api} />} />
      <Route
        path="/collections/new"
        element={<CollectionCreatePage api={api} />}
      />
      <Route
        path="/collections/:id"
        element={<CollectionEditPage api={api} />}
      />
      <Route path="/photos" element={<PhotosPage api={api} />} />
      <Route path="/photos/new" element={<PhotoCreatePage api={api} />} />
      <Route path="/photos/:id" element={<PhotoDetailPage api={api} />} />
      <Route path="*" element={<Navigate to="/collections" replace />} />
    </Routes>
  </Box>
);
