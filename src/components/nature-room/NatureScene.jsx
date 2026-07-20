import { MeshReflectorMaterial, useScroll } from "@react-three/drei";
import { useFrame, useThree } from "@react-three/fiber";
import { useRef } from "react";
import * as THREE from "three";
import NaturePanel from "./NaturePanel";

const PANEL_SPACING = 5.2;
const WALL_Z = -0.28;
const WALL_HEIGHT = 8.4;
const normalizeOffset = (offset) => ((offset % 1) + 1) % 1;

export default function NatureScene({ photos, onActivePhotoChange }) {
  const scroll = useScroll();
  const { camera, pointer } = useThree();
  const rail = useRef();
  const lastActivePhotoId = useRef(null);
  const cycleLength = Math.max(photos.length, 1) * PANEL_SPACING;

  useFrame((_, delta) => {
    const progress = normalizeOffset(scroll.offset);
    const targetX = -progress * cycleLength;

    if (rail.current) {
      // The rail is rendered three times. When infinite scroll wraps, snapping
      // over the numeric jump keeps the visual loop continuous.
      if (Math.abs(rail.current.position.x - targetX) > cycleLength / 2) {
        rail.current.position.x = targetX;
      } else {
        rail.current.position.x = THREE.MathUtils.damp(
          rail.current.position.x,
          targetX,
          6,
          delta
        );
      }
    }

    // Nature gets a slower, calmer camera than Drawings. It responds to pointer
    // movement like the viewer leaning in an open corridor, not navigating a
    // strict 3D world.
    camera.position.x = THREE.MathUtils.damp(
      camera.position.x,
      pointer.x * 0.22,
      1.8,
      delta
    );
    camera.position.y = THREE.MathUtils.damp(
      camera.position.y,
      0.35 + pointer.y * 0.12,
      1.8,
      delta
    );
    camera.lookAt(pointer.x * 0.34, 0.34 + pointer.y * 0.1, -0.12);

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
      <color attach="background" args={["#030604"]} />
      <fog attach="fog" args={["#030604", 8, 18]} />

      <ambientLight intensity={0.44} color="#dceee0" />
      <hemisphereLight intensity={0.72} color="#d8f1dd" groundColor="#020302" />
      <directionalLight position={[-4, 4.5, 3.8]} intensity={1.35} color="#d8f0d2" />
      <pointLight position={[3.5, 2.1, 3.4]} intensity={1.1} color="#d7e7ff" />

      <group ref={rail}>
        {[-1, 0, 1].map((cycleOffset) =>
          photos.map((photo, index) => (
            <NaturePanel
              key={`${cycleOffset}-${photo.id}`}
              photo={photo}
              index={index}
              trackIndex={index + cycleOffset * photos.length}
              spacing={PANEL_SPACING}
              totalPhotos={photos.length}
              wallZ={WALL_Z}
            />
          ))
        )}
      </group>

      {/* Oversize wall/floor surfaces keep their edges outside the camera view.
          Small planes created distracting borders in the Home scene. */}
      <mesh position={[0, 0.45, WALL_Z - 0.1]}>
        <planeGeometry args={[cycleLength + 24, WALL_HEIGHT]} />
        <meshStandardMaterial color="#04100b" roughness={1} metalness={0} />
      </mesh>

      <mesh position={[0, -1.08, 1.2]} rotation={[-Math.PI / 2, 0, 0]}>
        <planeGeometry args={[cycleLength + 28, 24]} />
        <MeshReflectorMaterial
          blur={[520, 180]}
          resolution={1024}
          mixBlur={1}
          mixStrength={26}
          roughness={1}
          depthScale={0.45}
          minDepthThreshold={0.25}
          maxDepthThreshold={1.15}
          color="#020504"
          metalness={0.28}
        />
      </mesh>
    </>
  );
}
