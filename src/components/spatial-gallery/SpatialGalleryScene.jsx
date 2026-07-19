import {
  Html,
  MeshReflectorMaterial,
} from "@react-three/drei";
import { Suspense, useMemo } from "react";
import { getPhotoById } from "../../data/photos";
import { placements } from "../../data/placements";
import CameraRig from "./CameraRig";
import PhotoPanel from "./PhotoPanel";

export default function SpatialGalleryScene({ photos, selectedPhotoId, onSelect }) {
  const selectedPlacement = useMemo(
    () => placements.find((placement) => placement.photoId === selectedPhotoId),
    [selectedPhotoId]
  );

  return (
    <>
      <color attach="background" args={["#191920"]} />
      <fog attach="fog" args={["#191920", 0, 15]} />
      <hemisphereLight intensity={0.45} groundColor="#070707" />
      <ambientLight intensity={0.62} />
      <spotLight
        position={[0, 5.2, 4.5]}
        angle={0.55}
        penumbra={0.9}
        intensity={7.2}
        castShadow
      />
      <pointLight position={[-3, 2.6, 3.5]} intensity={1.4} color="#f4dfc2" />
      <pointLight position={[3, 2.4, 1]} intensity={1.2} color="#d7e7ff" />

      <group position={[0, -0.5, 0]}>
        <Suspense
          fallback={
            <Html center style={{ color: "white", whiteSpace: "nowrap" }}>
              Loading gallery
            </Html>
          }
        >
          {photos.map((photo) => {
            const placement = placements.find(
              (item) => item.photoId === photo.id
            );
            if (!placement) return null;

            return (
              <PhotoPanel
                key={photo.id}
                photo={photo}
                placement={placement}
                isSelected={selectedPhotoId === photo.id}
                isDimmed={
                  Boolean(selectedPhotoId) && selectedPhotoId !== photo.id
                }
                onSelect={onSelect}
              />
            );
          })}
        </Suspense>

        <mesh rotation={[-Math.PI / 2, 0, 0]}>
          <planeGeometry args={[50, 50]} />
          <MeshReflectorMaterial
            blur={[300, 100]}
            resolution={2048}
            mixBlur={1}
            mixStrength={80}
            roughness={1}
            depthScale={1.2}
            minDepthThreshold={0.4}
            maxDepthThreshold={1.4}
            color="#050505"
            metalness={0.5}
          />
        </mesh>
      </group>

      <CameraRig
        selectedPlacement={
          selectedPlacement && getPhotoById(selectedPlacement.photoId)
            ? selectedPlacement
            : null
        }
      />
    </>
  );
}
