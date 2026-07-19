// Import Swiper React components
import { Swiper, SwiperSlide } from "swiper/react";
import { Box, useColorModeValue } from "@chakra-ui/react";
import { featuredPhotos } from "../data/photos";
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

export default function Carousel() {
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
        {featuredPhotos.map((photo) => (
          <SwiperSlide key={photo.id}>
            <img src={photo.src} alt={photo.title} />
          </SwiperSlide>
        ))}
      </Swiper>
    </Box>
  );
}
