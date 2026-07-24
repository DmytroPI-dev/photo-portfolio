import { extendTheme } from "@chakra-ui/react";

const theme = extendTheme({
  // The portfolio is a dark gallery experience. Disabling system colour-mode
  // detection prevents an operating-system preference from changing its mood.
  config: {
    initialColorMode: "dark",
    useSystemColorMode: false,
  },
});

export default theme;
