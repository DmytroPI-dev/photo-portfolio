import { useScroll } from "@react-three/drei";
import { useFrame, useThree } from "@react-three/fiber";
import { useRef } from "react";
import * as THREE from "three";
import DrawingArtwork from "./DrawingArtwork";

const ART_SPACING = 4.7;
const WALL_Z = -0.1;
const WALL_HEIGHT = 7.5;
const WALL_PADDING = 12;
const normalizeOffset = (offset) => ((offset % 1) + 1) % 1;

export default function DrawingRoomScene({ photos, onActivePhotoChange }) {
  const scroll = useScroll();
  const { camera, pointer } = useThree();
  const artGroup = useRef();
  const lastActivePhotoId = useRef(null);

  const cycleLength = Math.max(photos.length, 1) * ART_SPACING;
  useFrame((state, delta) => {
    const progress = normalizeOffset(scroll.offset);
    const targetX = -progress * cycleLength;

    if (artGroup.current) {
      if (Math.abs(artGroup.current.position.x - targetX) > cycleLength / 2) {
        artGroup.current.position.x = targetX;
      } else {
        artGroup.current.position.x = THREE.MathUtils.damp(
          artGroup.current.position.x,
          targetX,
          7,
          delta
        );
      }
    }

    camera.position.x = THREE.MathUtils.damp(
      camera.position.x,
      pointer.x * 0.18,
      2.2,
      delta
    );
    camera.position.y = THREE.MathUtils.damp(
      camera.position.y,
      0.05 + pointer.y * 0.12,
      2.2,
      delta
    );
    camera.lookAt(pointer.x * 0.28, 0.16 + pointer.y * 0.08, -0.08);

    const nextIndex = photos.length
      ? Math.round(progress * photos.length) % photos.length
      : 0;
    const nextPhotoId = photos[nextIndex]?.id;

    if (nextPhotoId && nextPhotoId !== lastActivePhotoId.current) {
      lastActivePhotoId.current = nextPhotoId;
      onActivePhotoChange(nextPhotoId);
    }
  });

  return (
    <>
      <color attach="background" args={["#030303"]} />
      <fog attach="fog" args={["#050505", 9, 16]} />

      <ambientLight intensity={0.36} color="#ffffff" />

      <group ref={artGroup}>
        {[-1, 0, 1].map((cycleOffset) =>
          photos.map((photo, index) => (
            <DrawingArtwork
              key={`${cycleOffset}-${photo.id}`}
              photo={photo}
              index={index}
              trackIndex={index + cycleOffset * photos.length}
              spacing={ART_SPACING}
              wallZ={WALL_Z}
              totalPhotos={photos.length}
            />
          ))
        )}
      </group>

      <mesh position={[0, 0, WALL_Z - 0.06]} receiveShadow>
        <planeGeometry args={[cycleLength + WALL_PADDING, WALL_HEIGHT]} />
        <meshStandardMaterial color="#050505" roughness={1} metalness={0} />
      </mesh>
    </>
  );
}
