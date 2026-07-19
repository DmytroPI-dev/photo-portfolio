import { Routes, Route } from "react-router-dom";
import {
    Hero,
    Carousel,
    HorizontalGallery,
    Testimonials,
    ErrorPage,
    Contacts
} from "../components";


export const AllRoutes = () => {
    return (
        <main>
            <Routes>
                <Route path="/" element={<Hero />}></Route>
                <Route path="/carousel" element={<Carousel />}></Route>
                <Route path="/gallery" element={<HorizontalGallery />}></Route>
                <Route path="/testimonials" element={<Testimonials />}></Route>
                <Route path="/contacts" element={<Contacts />}></Route>
                <Route path="*" element={<ErrorPage />}></Route>

            </Routes>
        </main>
    );
};
