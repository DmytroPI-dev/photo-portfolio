// Import Swiper React components
import { Swiper, SwiperSlide } from "swiper/react";
import { Box, useColorModeValue } from "@chakra-ui/react";
import { proxy, useSnapshot } from "valtio";
import "swiper/css/effect-coverflow";
import "swiper/css/pagination";
import "../assets/css/swiperStyles.css";
// Import Swiper styles
import "swiper/css";
// import required modules
import {
  EffectCoverflow,
  Pagination,
  Zoom,
  Navigation,
  Autoplay,
} from "swiper/modules";

const state = proxy({
  clicked: null,
  urls: [
    1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 12, 13, 1, 5, 3, 2, 7, 8, 2, 4, 9, 6,
  ].map((u) => `/images/${u}.jpg`),
});

export default function Carousel() {
  const { urls } = useSnapshot(state); // Get the URLs from your state object
  return (
    <Box mt={4} mx="auto" bg={useColorModeValue("gray.100", "gray.900")}>
      {/* Add margin-top and control the width */}
      <Swiper
        effect={"coverflow"}
        autoHeight={true}
        autoplay={{
          delay: 2500,
          disableOnInteraction: false,
        }}
        pagination={{
          clickable: true,
        }}
        grabCursor={true}
        centeredSlides={true}
        slidesPerView={"3"}
        loop={true}
        coverflowEffect={{
          rotate: 50,
          stretch: 0,
          depth: 100,
          modifier: 1,
          slideShadows: true,
        }}
        modules={[EffectCoverflow, Zoom, Navigation, Autoplay, Pagination]}
        className="mySwiper"
      >
        {urls.map((url, index) => (
          <SwiperSlide key={index}>
            <img src={url} alt="" /> {/* Use the dynamic URL */}
          </SwiperSlide>
        ))}
      </Swiper>
    </Box>
  );
}
