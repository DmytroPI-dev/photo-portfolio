import { Text, useScroll } from "@react-three/drei";
import { useFrame, useLoader } from "@react-three/fiber";
import { useEffect, useMemo, useRef } from "react";
import * as THREE from "three";

const MAX_PANEL_WIDTH = 3.35;
const MAX_PANEL_HEIGHT = 2.65;
const normalizeOffset = (offset) => ((offset % 1) + 1) % 1;

export default function NaturePanel({
  photo,
  index,
  trackIndex = index,
  spacing,
  totalPhotos,
  wallZ,
}) {
  const scroll = useScroll();
  const group = useRef();
  const image = useRef();
  const glass = useRef();
  const texture = useLoader(THREE.TextureLoader, photo.src);
  const aspect = photo.width && photo.height ? photo.width / photo.height : 0.75;

  // Fit both vertical placeholders and future landscape photos into a wider
  // nature-gallery envelope. When real landscape images arrive, they will expand
  // horizontally without changing this component.
  const panelWidth = Math.min(MAX_PANEL_WIDTH, MAX_PANEL_HEIGHT * aspect);
  const panelHeight = panelWidth / aspect;
  const labelY = -panelHeight / 2 - 0.28;

  const baseRotationY = useMemo(() => {
    if (index % 3 === 1) return Math.PI / 18;
    if (index % 3 === 2) return -Math.PI / 18;
    return 0;
  }, [index]);

  useEffect(() => {
    texture.colorSpace = THREE.SRGBColorSpace;
    texture.anisotropy = 8;
    texture.needsUpdate = true;
  }, [texture]);

  useFrame((_, delta) => {
    if (!group.current) return;

    const currentIndex = normalizeOffset(scroll.offset) * Math.max(totalPhotos, 1);
    const distanceFromCenter = trackIndex - currentIndex;
    const focus = Math.max(0, 1 - Math.abs(distanceFromCenter) * 0.42);

    group.current.rotation.y = THREE.MathUtils.damp(
      group.current.rotation.y,
      baseRotationY + THREE.MathUtils.clamp(distanceFromCenter * -0.035, -0.16, 0.16),
      4,
      delta
    );
    group.current.position.z = THREE.MathUtils.damp(
      group.current.position.z,
      focus * 0.16,
      3,
      delta
    );
    group.current.scale.setScalar(
      THREE.MathUtils.damp(group.current.scale.x, 0.96 + focus * 0.055, 4, delta)
    );

    if (image.current) {
      image.current.material.emissiveIntensity = THREE.MathUtils.damp(
        image.current.material.emissiveIntensity,
        0.02 + focus * 0.06,
        4,
        delta
      );
    }

    if (glass.current) {
      glass.current.material.opacity = THREE.MathUtils.damp(
        glass.current.material.opacity,
        0.12 + focus * 0.11,
        4,
        delta
      );
    }
  });

  return (
    <group ref={group} position={[trackIndex * spacing, 0.48, 0]}>
      <mesh position={[0, 0, wallZ + 0.02]}>
        <boxGeometry args={[panelWidth + 0.28, panelHeight + 0.28, 0.06]} />
        <meshStandardMaterial color="#0a0f0b" roughness={0.72} metalness={0.12} />
      </mesh>

      <mesh ref={image} position={[0, 0, wallZ + 0.08]}>
        <boxGeometry args={[panelWidth, panelHeight, 0.035]} />
        <meshStandardMaterial
          map={texture}
          roughness={0.52}
          metalness={0}
          emissive="#dceee0"
          emissiveIntensity={0.04}
        />
      </mesh>

      <mesh ref={glass} position={[0, 0, wallZ + 0.11]}>
        <planeGeometry args={[panelWidth, panelHeight]} />
        <meshBasicMaterial
          color="#dcefe2"
          transparent
          opacity={0.16}
          blending={THREE.MultiplyBlending}
          depthWrite={false}
        />
      </mesh>

      <Text
        position={[0, labelY, wallZ + 0.13]}
        fontSize={0.08}
        color="#c7d7c4"
        anchorX="center"
        anchorY="top"
        maxWidth={panelWidth}
        textAlign="center"
      >
        {photo.title}
      </Text>
    </group>
  );
}
