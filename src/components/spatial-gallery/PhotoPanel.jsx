import { Image as DreiImage, Text, useCursor } from "@react-three/drei";
import { useFrame } from "@react-three/fiber";
import { useRef, useState } from "react";
import * as THREE from "three";

const GOLDEN_RATIO = 1.61803398875;

export default function PhotoPanel({
  photo,
  placement,
  isSelected,
  isDimmed,
  onSelect,
}) {
  const ref = useRef();
  const image = useRef();
  const frame = useRef();
  const [hovered, setHovered] = useState(false);
  const [motionSeed] = useState(() => Math.random() * 10000);

  useCursor(hovered);

  useFrame((_, delta) => {
    if (!ref.current) return;
    const targetScale = isSelected ? 1.08 : hovered ? 1.04 : 1;
    ref.current.scale.x = THREE.MathUtils.damp(
      ref.current.scale.x,
      targetScale,
      7,
      delta
    );
    ref.current.scale.y = THREE.MathUtils.damp(
      ref.current.scale.y,
      targetScale,
      7,
      delta
    );

    if (frame.current) {
      frame.current.material.color.lerp(
        new THREE.Color(hovered || isSelected ? "#f2b263" : "#f6f0e8"),
        1 - Math.exp(-delta * 8)
      );
    }

    if (image.current) {
      image.current.material.zoom =
        2 + Math.sin(motionSeed + _.clock.elapsedTime / 3) / 2;
      image.current.material.grayscale = THREE.MathUtils.damp(
        image.current.material.grayscale,
        isDimmed ? 0.35 : 0,
        6,
        delta
      );
    }
  });

  return (
    <group
      ref={ref}
      position={placement.position}
      rotation={placement.rotation}
      onClick={(event) => {
        event.stopPropagation();
        onSelect(photo.id);
      }}
      onPointerOver={(event) => {
        event.stopPropagation();
        setHovered(true);
      }}
      onPointerOut={() => setHovered(false)}
    >
      <mesh
        name={photo.id}
        scale={[1, GOLDEN_RATIO, 0.05]}
        position={[0, GOLDEN_RATIO / 2, 0]}
        castShadow
        receiveShadow
      >
        <boxGeometry />
        <meshStandardMaterial
          color="#101010"
          metalness={0.28}
          roughness={0.42}
        />
        <mesh
          ref={frame}
          raycast={() => null}
          scale={[0.9, 0.93, 0.9]}
          position={[0, 0, 0.2]}
        >
          <boxGeometry />
          <meshBasicMaterial color="#f6f0e8" toneMapped={false} fog={false} />
        </mesh>
        <DreiImage
          ref={image}
          raycast={() => null}
          url={photo.src}
          scale={[0.72, 0.72, 1]}
          position={[0, 0, 0.7]}
          toneMapped={false}
        />
      </mesh>
      <Text
        position={[0.55, GOLDEN_RATIO, 0]}
        fontSize={0.025}
        color={isDimmed ? "#777777" : "#f4eadc"}
        anchorX="left"
        anchorY="top"
        maxWidth={0.1}
      >
        {photo.title}
      </Text>
    </group>
  );
}
