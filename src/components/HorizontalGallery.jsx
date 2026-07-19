import { useEffect, useState, useRef } from "react";
import { Box,useColorModeValue } from "@chakra-ui/react";
import HorizonTiles from "./HorizonTiles";

export default function HorizontalGallery() {
  const [isMouseInside, setIsMouseInside] = useState(false);
  const customScrollRef = useHorizontalScroll(); // Use the hook
  
  const handleMouseEnter = () => {
    setIsMouseInside(true);
  };
  
  const handleMouseLeave = () => {
    setIsMouseInside(false);
  };

  function useHorizontalScroll() {
    const elRef = useRef();
    useEffect(() => {
      const el = elRef.current;
      if (el) {
        const onWheel = e => {
          if (e.deltaY === 0) return;
          e.preventDefault();
          el.scrollTo({
            left: el.scrollLeft + e.deltaY,
            behavior: "smooth"
          });
        };
        el.addEventListener("wheel", onWheel);
        return () => el.removeEventListener("wheel", onWheel);
      }
    }, []);
    return elRef;
  }

  
  useEffect(() => {
    const customScrollDiv = customScrollRef.current;
  
    // Function to handle the scroll behavior change
    const handleScrollBehavior = (isVisible, isMouseInside) => {
      if (isVisible && !isMouseInside) {
        customScrollDiv.style.overflowX = "scroll";
        customScrollDiv.style.overflowY = "hidden";
      } else {
        customScrollDiv.style.overflowX = "hidden";
        customScrollDiv.style.overflowY = "scroll";
      }
    };
  
    // Intersection Observer setup
    const observer = new IntersectionObserver(
      (entries) => {
        const isVisible = entries[0].isIntersecting;
        handleScrollBehavior(isVisible);
      },
      { root: null, threshold: 0.2 } // Adjust threshold as needed
    );
  
    // Start observing the customScrollDiv
    observer.observe(customScrollDiv);
  
    // Cleanup: stop observing when the component unmounts
    return () => {
      observer.unobserve(customScrollDiv);
    };
  }, [isMouseInside, customScrollRef]);
  

  return (
    <Box
      ref={customScrollRef}
      onMouseEnter={handleMouseEnter}
      onMouseLeave={handleMouseLeave}
      style={{ cursor: isMouseInside ? "col-resize" : "auto" }}
      width={"100%"}
      height={"90vh"}
      bg={useColorModeValue("gray.100", "gray.900")}
    >
      <HorizonTiles/>
	    </Box>
  );
}

