import { SpotLight, Text, useScroll } from "@react-three/drei";
import { useFrame, useLoader } from "@react-three/fiber";
import { useEffect, useMemo, useRef } from "react";
import * as THREE from "three";

const MAX_ART_HEIGHT = 3.0;
const MAX_ART_WIDTH = 3.4;
const PLAQUE_FONT = "/fonts/sacramento.woff";

// Position is local to each artwork group. Because each group is part of the
// scrolling rail, the light source moves with its artwork while the actual
// SpotLight target remains near the scene center. That is what creates the
// "turning light" effect as a piece enters/leaves the viewport.
const LIGHT_POSITION = [2.34, 2.5, 1.2];

// Keep offset math stable when Drei wraps infinite scroll around 0/1.
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

  // The visible light source should feel like a soft lamp/candle, not a hard
  // yellow mesh. A tiny canvas-generated radial gradient gives the sprite a
  // feathered edge and avoids adding another image asset to the project.
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

  // Preserve the source image aspect ratio and fit it into a predictable visual
  // envelope so wide and tall works can share the same horizontal rail.
  const artHeight = Math.min(MAX_ART_HEIGHT, MAX_ART_WIDTH / aspect);
  const artWidth = artHeight * aspect;
  const plaqueY = -artHeight / 2 - 0.58;

  useEffect(() => {
    // Vite gives us image URLs; Three still needs texture color/quality hints.
    // sRGB keeps colors from looking washed out, and anisotropy helps angled
    // images stay sharper on desktop GPUs.
    texture.colorSpace = THREE.SRGBColorSpace;
    texture.anisotropy = 8;
    texture.needsUpdate = true;
  }, [texture]);

  useEffect(() => () => glowTexture.dispose(), [glowTexture]);

  useFrame((state, delta) => {
    if (!group.current) return;

    // `trackIndex` may be negative or greater than photos.length because the
    // scene renders previous/current/next rail copies for circular scroll.
    // Comparing it to the normalized current index tells each duplicate whether
    // it is the one currently near the center.
    const currentIndex = normalizeOffset(scroll.offset) * Math.max(totalPhotos, 1);
    const distanceFromCenter = trackIndex - currentIndex;
    const targetRotationY = THREE.MathUtils.clamp(distanceFromCenter * -0.025, -0.08, 0.08);

    // Damping gives the subtle gallery motion without a spring library. The
    // farther a work is from center, the more it tilts away.
    group.current.rotation.y = THREE.MathUtils.damp(
      group.current.rotation.y,
      targetRotationY,
      4,
      delta
    );

    // Focus is a soft 0..1 weight used by several small effects: centered works
    // become a touch larger/brighter, while side works recede.
    const focus = Math.max(0, 1 - Math.abs(distanceFromCenter) * 0.38);

    // Layered sine waves make the lamp intensity feel organic. This is purely
    // visual; it does not drive layout or scroll state.
    const flicker =
      0.92 +
      Math.sin(state.clock.elapsedTime * 2.3 + index * 1.7) * 0.055 +
      Math.sin(state.clock.elapsedTime * 7.1 + index) * 0.025;
    const targetScale = 0.98 + focus * 0.025;
    group.current.scale.setScalar(
      THREE.MathUtils.damp(group.current.scale.x, targetScale, 5, delta)
    );

    if (artwork.current) {
      // The artwork uses a normal mapped material, but a tiny emissive lift
      // prevents dark images from disappearing completely in the black room.
      artwork.current.material.emissiveIntensity = THREE.MathUtils.damp(
        artwork.current.material.emissiveIntensity,
        0.02 + focus * 0.045,
        3,
        delta
      );
    }

    if (plaque.current) {
      // The plaque background reacts to focus; the actual text below uses a
      // local Sacramento font so it does not depend on a remote Google request
      // during the 3D scene load.
      plaque.current.material.color.lerp(
        new THREE.Color(focus > 0.7 ? "#f4eee3" : "#bdb7ac"),
        1 - Math.exp(-delta * 4)
      );
    }

    if (light.current) {
      // No castShadow here: when circular scroll duplicates the rail, per-light
      // shadow maps exceed MAX_TEXTURE_IMAGE_UNITS on some Mac Chrome/GPU
      // combinations. The volumetric cone remains, but shadow textures do not.
      light.current.intensity = THREE.MathUtils.damp(
        light.current.intensity,
        (5.15 + focus * 0.95) * flicker,
        6,
        delta
      );
    }

    if (glow.current) {
      // Match the visible sprite glow to the same flicker curve as the actual
      // light, keeping the source soft and candle-like.
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
        angle={0.94}
        attenuation={1.35}
        anglePower={4.2}
        intensity={5.7}
        distance={8.8}
        opacity={1.2}
        color="#daad7b"
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
      >
        <boxGeometry args={[1.18, 0.28, 0.035]} />
        <meshStandardMaterial color="#f1d328" roughness={0.86} metalness={0} />
      </mesh>

      <Text
        position={[0, plaqueY + 0.01, wallZ + 0.13]}
        font={PLAQUE_FONT}
        fontSize={0.16}
        color="#3a3028"
        anchorX="center"
        anchorY="middle"
        maxWidth={1.02}
        textAlign="center"
      >
        {photo.title}
      </Text>
    </group>
  );
}
