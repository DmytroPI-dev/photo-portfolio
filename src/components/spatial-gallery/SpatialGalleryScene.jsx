import {
  Html,
  MeshReflectorMaterial,
} from "@react-three/drei";
import { useFrame } from "@react-three/fiber";
import { Suspense, useMemo, useRef } from "react";
import * as THREE from "three";
import { getPhotoById } from "../../data/photos";
import CameraRig from "./CameraRig";
import PhotoPanel from "./PhotoPanel";

const FLOOR_HEIGHT = 6.7;

function ElevatorShell() {
  // The first elevator pass had literal glass rails. They helped describe the
  // idea, but in screenshots they read as hard gray borders around the art and
  // the reflector. Keep this component as a named placeholder for future subtle
  // glass effects, but render nothing for now.
  return null;
}

function GalleryFloor({ floor, activeFloorIndex, onSelect, selectedPhotoId }) {
  const distanceFromActive = Math.abs(floor.index - activeFloorIndex);
  const floorFocus = distanceFromActive === 0 ? 1 : 0.28;

  return (
    <group position={[0, -floor.index * FLOOR_HEIGHT - 0.5, 0]}>
      <mesh position={[0, 1.2, -3.45]}>
        <planeGeometry args={[44, 18]} />
        <meshStandardMaterial
          color={floor.index === activeFloorIndex ? "#030304" : "#010101"}
          roughness={1}
          metalness={0}
        />
      </mesh>

      {floor.photos.map((photo) => {
        const placement = floor.placements.find(
          (item) => item.photoId === photo.id
        );
        if (!placement) return null;

        return (
          <PhotoPanel
            key={`${floor.id}-${photo.id}`}
            photo={photo}
            placement={placement}
            floorAccent={floor.accent}
            floorFocus={floorFocus}
            isSelected={selectedPhotoId === photo.id}
            isDimmed={
              distanceFromActive > 0 ||
              (Boolean(selectedPhotoId) && selectedPhotoId !== photo.id)
            }
            onSelect={onSelect}
          />
        );
      })}
    </group>
  );
}

export default function SpatialGalleryScene({
  floors,
  activeFloorIndex,
  selectedPhotoId,
  onSelect,
}) {
  const movingFloors = useRef();
  const activeFloor = floors[activeFloorIndex] ?? floors[0];
  const selectedPlacement = useMemo(
    () =>
      activeFloor?.placements.find(
        (placement) => placement.photoId === selectedPhotoId
      ),
    [activeFloor, selectedPhotoId]
  );
  const visibleFloors = useMemo(
    () =>
      floors.filter(
        (floor) => Math.abs(floor.index - activeFloorIndex) <= 1
      ),
    [activeFloorIndex, floors]
  );

  useFrame((_, delta) => {
    if (!movingFloors.current) return;

    // Each floor is authored below the previous one. Moving the parent upward by
    // activeFloorIndex * FLOOR_HEIGHT centers the selected level in the camera.
    const targetY = activeFloorIndex * FLOOR_HEIGHT;
    movingFloors.current.position.y = THREE.MathUtils.damp(
      movingFloors.current.position.y,
      targetY,
      3.2,
      delta
    );
  });

  return (
    <>
      <color attach="background" args={["#191920"]} />
      <fog attach="fog" args={["#191920", 2, 17]} />
      <hemisphereLight intensity={0.36} groundColor="#060606" />
      <ambientLight intensity={0.48} />
      <spotLight
        position={[0, 5.2, 4.5]}
        angle={0.58}
        penumbra={0.9}
        intensity={6.8}
        castShadow
      />
      <pointLight position={[-3, 2.6, 3.5]} intensity={1.4} color="#f4dfc2" />
      <pointLight position={[3, 2.4, 1]} intensity={1.2} color="#d7e7ff" />

      <ElevatorShell />

      <group ref={movingFloors}>
        <Suspense
          fallback={
            <Html center style={{ color: "white", whiteSpace: "nowrap" }}>
              Loading gallery
            </Html>
          }
        >
          {visibleFloors.map((floor) => (
            <GalleryFloor
              key={floor.id}
              floor={floor}
              activeFloorIndex={activeFloorIndex}
              selectedPhotoId={selectedPhotoId}
              onSelect={onSelect}
            />
          ))}
        </Suspense>
      </group>

      <mesh position={[0, -1.02, 0]} rotation={[-Math.PI / 2, 0, 0]}>
        <planeGeometry args={[50, 50]} />
        <MeshReflectorMaterial
          blur={[300, 100]}
          resolution={1024}
          mixBlur={1}
          mixStrength={42}
          roughness={1}
          depthScale={0.7}
          minDepthThreshold={0.35}
          maxDepthThreshold={1.25}
          color="#050505"
          metalness={0.42}
        />
      </mesh>

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
