import React from "react";
import {
  ChakraProvider,
  CSSReset,
  Box,
} from "@chakra-ui/react";
import theme from "./theme";
import Footer from "./components/Footer";
import Navbar from "./Navbar";
import Hero from "./components/Hero";
import HorizontalGallery from "./components/HorizontalGallery";
import Carousel from "./components/Carousel";
import Testimonials from "./components/Testimonials";
import Contact from "./components/Contacts";

export default function App() {
  return (
    <ChakraProvider theme={theme}>
      <CSSReset />
      <Box minH="100vh">
        <Navbar />
        <Box ml={{ base: 0, md: 60 }} p="4">
          <Hero />
          <HorizontalGallery />
          <Carousel />
          <Testimonials />
          <Contact />
        </Box>
        <Footer />
      </Box>
    </ChakraProvider>
  );
}
