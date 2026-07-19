import {
  ChakraProvider,
  CSSReset,
  Box,
} from "@chakra-ui/react";
import { BrowserRouter, Route, Routes, useLocation } from "react-router-dom";
import theme from "./theme";
import Footer from "./components/Footer";
import Navbar from "./Navbar";
import AboutPage from "./pages/AboutPage";
import CollectionPage from "./pages/CollectionPage";
import ContactPage from "./pages/ContactPage";
import HomePage from "./pages/HomePage";
import NotFoundPage from "./pages/NotFoundPage";

function AppShell() {
  const { pathname } = useLocation();
  const isImmersiveRoute = pathname === "/drawings";

  return (
    <Box minH="100vh" bg="black">
      <Navbar />
      <Box ml={{ base: 0, md: 60 }} pt={{ base: 20, md: 0 }}>
        <Routes>
          <Route path="/" element={<HomePage />} />
          <Route path="/:collectionId" element={<CollectionPage />} />
          <Route path="/about" element={<AboutPage />} />
          <Route path="/contact" element={<ContactPage />} />
          <Route path="*" element={<NotFoundPage />} />
        </Routes>
        {!isImmersiveRoute ? <Footer /> : null}
      </Box>
    </Box>
  );
}

export default function App() {
  return (
    <ChakraProvider theme={theme}>
      <CSSReset />
      <BrowserRouter>
        <AppShell />
      </BrowserRouter>
    </ChakraProvider>
  );
}
