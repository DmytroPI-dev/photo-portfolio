import { SpotLight, useScroll } from "@react-three/drei";
import { useFrame, useLoader } from "@react-three/fiber";
import { useEffect, useMemo, useRef } from "react";
import * as THREE from "three";

const MAX_ART_HEIGHT = 3.0;
const MAX_ART_WIDTH = 3.4;
const LIGHT_POSITION = [-1.28, 2.34, 0.88];
const normalizeOffset = (offset) => ((offset % 1) + 1) % 1;

export default function DrawingArtwork({
  photo,
  index,
  trackIndex = index,
  spacing,
  wallZ,
  totalPhotos,
}) {
  const scroll = useScroll();
  const group = useRef();
  const artwork = useRef();
  const plaque = useRef();
  const light = useRef();
  const glow = useRef();
  const texture = useLoader(THREE.TextureLoader, photo.src);
  const glowTexture = useMemo(() => {
    const canvas = document.createElement("canvas");
    canvas.width = 128;
    canvas.height = 128;

    const context = canvas.getContext("2d");
    const gradient = context.createRadialGradient(64, 64, 2, 64, 64, 62);
    gradient.addColorStop(0, "rgba(255, 208, 144, 0.95)");
    gradient.addColorStop(0.28, "rgba(255, 181, 104, 0.42)");
    gradient.addColorStop(1, "rgba(255, 181, 104, 0)");

    context.fillStyle = gradient;
    context.fillRect(0, 0, 128, 128);

    const glowMap = new THREE.CanvasTexture(canvas);
    glowMap.colorSpace = THREE.SRGBColorSpace;
    return glowMap;
  }, []);
  const aspect = photo.width && photo.height ? photo.width / photo.height : 0.75;
  const artHeight = Math.min(MAX_ART_HEIGHT, MAX_ART_WIDTH / aspect);
  const artWidth = artHeight * aspect;
  const plaqueY = -artHeight / 2 - 0.58;

  useEffect(() => {
    texture.colorSpace = THREE.SRGBColorSpace;
    texture.anisotropy = 8;
    texture.needsUpdate = true;
  }, [texture]);

  useEffect(() => () => glowTexture.dispose(), [glowTexture]);

  useFrame((state, delta) => {
    if (!group.current) return;

    const currentIndex = normalizeOffset(scroll.offset) * Math.max(totalPhotos, 1);
    const distanceFromCenter = trackIndex - currentIndex;
    const targetRotationY = THREE.MathUtils.clamp(distanceFromCenter * -0.025, -0.08, 0.08);

    group.current.rotation.y = THREE.MathUtils.damp(
      group.current.rotation.y,
      targetRotationY,
      4,
      delta
    );

    const focus = Math.max(0, 1 - Math.abs(distanceFromCenter) * 0.38);
    const flicker =
      0.92 +
      Math.sin(state.clock.elapsedTime * 2.3 + index * 1.7) * 0.055 +
      Math.sin(state.clock.elapsedTime * 7.1 + index) * 0.025;
    const targetScale = 0.98 + focus * 0.025;
    group.current.scale.setScalar(
      THREE.MathUtils.damp(group.current.scale.x, targetScale, 5, delta)
    );

    if (artwork.current) {
      artwork.current.material.emissiveIntensity = THREE.MathUtils.damp(
        artwork.current.material.emissiveIntensity,
        0.02 + focus * 0.045,
        3,
        delta
      );
    }

    if (plaque.current) {
      plaque.current.material.color.lerp(
        new THREE.Color(focus > 0.7 ? "#f4eee3" : "#bdb7ac"),
        1 - Math.exp(-delta * 4)
      );
    }

    if (light.current) {
      light.current.intensity = THREE.MathUtils.damp(
        light.current.intensity,
        (5.15 + focus * 0.95) * flicker,
        6,
        delta
      );
    }

    if (glow.current) {
      glow.current.scale.setScalar(0.86 + flicker * 0.08);
      glow.current.material.opacity = 0.2 + focus * 0.14;
    }
  });

  return (
    <group ref={group} position={[trackIndex * spacing, 0, 0]}>
      <SpotLight
        ref={light}
        position={LIGHT_POSITION}
        penumbra={1}
        angle={0.64}
        attenuation={1.35}
        anglePower={4.2}
        intensity={5.7}
        distance={8.8}
        opacity={0.72}
        color="#daad7b"
        castShadow
      />

      <sprite ref={glow} position={LIGHT_POSITION} scale={[0.34, 0.34, 1]}>
        <spriteMaterial
          map={glowTexture}
          color="#ffd1a0"
          transparent
          opacity={0.28}
          blending={THREE.AdditiveBlending}
          depthWrite={false}
          toneMapped={false}
        />
      </sprite>

      <mesh
        ref={artwork}
        position={[0, 0.42, wallZ + 0.08]}
        castShadow
        receiveShadow
      >
        <boxGeometry args={[artWidth, artHeight, 0.07]} />
        <meshStandardMaterial
          map={texture}
          roughness={0.46}
          metalness={0}
        />
      </mesh>

      <mesh
        ref={plaque}
        position={[0, plaqueY, wallZ + 0.1]}
        castShadow
      >
        <boxGeometry args={[1.18, 0.28, 0.035]} />
        <meshStandardMaterial color="#efe8dc" roughness={0.86} metalness={0} />
      </mesh>

    </group>
  );
}
