import { Image as DreiImage, Text, useCursor } from "@react-three/drei";
import { useFrame } from "@react-three/fiber";
import { useMemo, useRef, useState } from "react";
import * as THREE from "three";

const GOLDEN_RATIO = 1.61803398875;

export default function PhotoPanel({
  photo,
  placement,
  floorAccent = "#f2b263",
  floorFocus = 1,
  isSelected,
  isDimmed,
  onSelect,
}) {
  const ref = useRef();
  const image = useRef();
  const frame = useRef();
  const [hovered, setHovered] = useState(false);
  const [motionSeed] = useState(() => Math.random() * 10000);
  const baseScale = placement.scale ?? 1;
  const accentColor = useMemo(() => new THREE.Color(floorAccent), [floorAccent]);
  const defaultFrameColor = useMemo(() => new THREE.Color("#f6f0e8"), []);

  useCursor(hovered);

  useFrame((_, delta) => {
    if (!ref.current) return;

    // The floor itself moves vertically, while each panel has its own smaller
    // hover/selection breathing motion. Multiplying by placement.scale lets the
    // same component support hero-sized center images and smaller side images.
    const interactionScale = isSelected ? 1.08 : hovered ? 1.04 : 1;
    const targetScale = baseScale * interactionScale;
    ref.current.scale.setScalar(
      THREE.MathUtils.damp(ref.current.scale.x, targetScale, 7, delta)
    );

    if (frame.current) {
      const targetColor = hovered || isSelected ? accentColor : defaultFrameColor;
      frame.current.material.color.lerp(targetColor, 1 - Math.exp(-delta * 8));
      frame.current.material.opacity = THREE.MathUtils.damp(
        frame.current.material.opacity,
        0.42 + floorFocus * 0.58,
        6,
        delta
      );
    }

    if (image.current) {
      image.current.material.transparent = true;
      image.current.material.zoom =
        2 + Math.sin(motionSeed + _.clock.elapsedTime / 3) / 2;
      image.current.material.grayscale = THREE.MathUtils.damp(
        image.current.material.grayscale,
        isDimmed ? 0.35 : 0,
        6,
        delta
      );
      image.current.material.opacity = THREE.MathUtils.damp(
        image.current.material.opacity,
        0.36 + floorFocus * 0.64,
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
          <meshBasicMaterial
            color="#f6f0e8"
            toneMapped={false}
            fog={false}
            transparent
            opacity={1}
          />
        </mesh>
        <DreiImage
          ref={image}
          raycast={() => null}
          url={photo.src}
          scale={[0.72, 0.72, 1]}
          position={[0, 0, 0.7]}
          toneMapped={false}
          transparent
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
