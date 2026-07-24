import { Text, useCursor } from "@react-three/drei";
import { useFrame, useLoader, useThree } from "@react-three/fiber";
import { useEffect, useMemo, useRef, useState } from "react";
import * as THREE from "three";

const MAX_PHOTO_WIDTH = 2.45;
const MAX_PHOTO_HEIGHT = 3.15;
const ENTER_OFFSET_X = 4.2;
const MAT_BORDER = 0.22;
// Keep the shadow strength in one place. Its material starts at this same
// value, so it remains stable once the animation loop begins.
const SHADOW_OPACITY = 0.9;
const EXPANDED_SHADOW_OPACITY = 1;
const SHADOW_COLOR = "#000000";

function createShadowTexture() {
  const canvas = document.createElement("canvas");
  canvas.width = 512;
  canvas.height = 512;

  const context = canvas.getContext("2d");

  // This texture acts like a CSS box-shadow, but lives inside WebGL so it can
  // move, rotate, and fade with the framed photo. The hard center is hidden
  // behind the actual frame; the useful part is the large blurred edge.
  context.shadowColor = "rgba(0, 0, 0, 1)";;
  context.shadowBlur = 92;
  context.shadowOffsetX = 30;
  context.shadowOffsetY = 42;
  context.fillStyle = "rgba(0, 0, 0, 0.7)";
  context.fillRect(116, 96, 280, 320);

  const texture = new THREE.CanvasTexture(canvas);
  texture.needsUpdate = true;
  return texture;
}

function getFittedSize(photo) {
  const aspect = photo?.width && photo?.height ? photo.width / photo.height : 0.75;
  const photoWidth = Math.min(MAX_PHOTO_WIDTH, MAX_PHOTO_HEIGHT * aspect);
  const photoHeight = photoWidth / aspect;

  return {
    photoWidth,
    photoHeight,
    frameWidth: photoWidth + MAT_BORDER * 2,
    frameHeight: photoHeight + MAT_BORDER * 2,
  };
}

export default function NaturePanel({
  photo,
  direction,
  phase,
  isExpanded,
  onToggleExpanded,
}) {
  const { pointer } = useThree();
  const group = useRef();
  const frame = useRef();
  const mat = useRef();
  const image = useRef();
  const shadow = useRef();
  const title = useRef();
  const [hovered, setHovered] = useState(false);
  const texture = useLoader(THREE.TextureLoader, photo.src);
  const shadowTexture = useMemo(createShadowTexture, []);
  const { photoWidth, photoHeight, frameWidth, frameHeight } = useMemo(
    () => getFittedSize(photo),
    [photo]
  );
  const initialTransform = useMemo(
    () => {
      // This is intentionally mount-only. The panel is keyed by `photo.id`, so a
      // new image gets a fresh starting transform, while later phase changes do
      // not snap the same group back to a prop-defined position.
      if (phase !== "entering") {
        return {
          position: [0, -0.04, 0],
          rotation: [0, 0, 0],
          scale: 1,
        };
      }

      return {
        position: [direction * ENTER_OFFSET_X, -0.05, -0.35],
        rotation: [0.08, direction * -0.28, direction * -0.1],
        scale: 0.84,
      };
    },
    []
  );

  useCursor(hovered);

  useEffect(() => {
    texture.colorSpace = THREE.SRGBColorSpace;
    texture.anisotropy = 8;
    texture.needsUpdate = true;
  }, [texture]);

  useEffect(() => () => shadowTexture.dispose(), [shadowTexture]);

  useFrame((state, delta) => {
    if (!group.current) return;

    const leaving = phase === "leaving";
    const expandedScale = isExpanded ? 1.44 : 1;
    const floatY = Math.sin(state.clock.elapsedTime * 0.9) * 0.055;
    const waveZ = Math.sin(state.clock.elapsedTime * 0.8 + pointer.x * 1.4) * 0.018;

    const targetX = leaving ? -direction * ENTER_OFFSET_X : 0;
    const targetY = (isExpanded ? 0.02 : -0.04) + floatY;
    const targetZ = isExpanded ? 1.02 : 0;
    const targetScale = (leaving ? 0.84 : expandedScale) * (hovered ? 1.025 : 1);
    const targetRotationY = leaving
      ? direction * 0.34
      : pointer.x * (isExpanded ? 0.055 : 0.13);
    const targetRotationX = pointer.y * (isExpanded ? -0.035 : -0.08);
    const targetRotationZ = leaving
      ? direction * -0.14
      : waveZ + pointer.x * (isExpanded ? -0.01 : -0.025);

    group.current.position.x = THREE.MathUtils.damp(
      group.current.position.x,
      targetX,
      3.8,
      delta
    );
    group.current.position.y = THREE.MathUtils.damp(
      group.current.position.y,
      targetY,
      3.2,
      delta
    );
    group.current.position.z = THREE.MathUtils.damp(
      group.current.position.z,
      targetZ,
      3.2,
      delta
    );
    group.current.scale.setScalar(
      THREE.MathUtils.damp(group.current.scale.x, targetScale, 3.4, delta)
    );
    group.current.rotation.x = THREE.MathUtils.damp(
      group.current.rotation.x,
      targetRotationX,
      3.2,
      delta
    );
    group.current.rotation.y = THREE.MathUtils.damp(
      group.current.rotation.y,
      targetRotationY,
      3.2,
      delta
    );
    group.current.rotation.z = THREE.MathUtils.damp(
      group.current.rotation.z,
      targetRotationZ,
      3.2,
      delta
    );

    const targetOpacity = leaving ? 0 : 1;
    [frame, mat, image, title].forEach((item) => {
      if (!item.current?.material) return;
      item.current.material.opacity = THREE.MathUtils.damp(
        item.current.material.opacity,
        targetOpacity,
        4.4,
        delta
      );
    });

    if (shadow.current) {
      shadow.current.material.opacity = THREE.MathUtils.damp(
        shadow.current.material.opacity,
        leaving ? 0 : isExpanded ? EXPANDED_SHADOW_OPACITY : SHADOW_OPACITY,
        4.4,
        delta
      );
    }

    if (frame.current) {
      frame.current.material.color.lerp(
        new THREE.Color(hovered || isExpanded ? "#d8c28f" : "#11100e"),
        1 - Math.exp(-delta * 4)
      );
    }
  });

  return (
    <group
      ref={group}
      position={initialTransform.position}
      rotation={initialTransform.rotation}
      scale={initialTransform.scale}
      onClick={(event) => {
        event.stopPropagation();
        onToggleExpanded();
      }}
      onPointerOver={(event) => {
        event.stopPropagation();
        setHovered(true);
      }}
      onPointerOut={() => setHovered(false)}
    >
      {/*
        The soft shadow is a transparent WebGL plane rather than a CSS
        box-shadow. `renderOrder` and disabled depth testing make it paint
        first; the frame then covers its rectangular center, leaving only the
        blurred perimeter visible behind the artwork.
      */}
      <mesh
        ref={shadow}
        position={[0.24, -0.32, -0.18]}
        renderOrder={-1}
      >
        <planeGeometry args={[frameWidth + 1.65, frameHeight + 1.75]} />
        <meshBasicMaterial
          map={shadowTexture}
          color={SHADOW_COLOR}
          transparent
          opacity={SHADOW_OPACITY}
          depthTest={false}
          depthWrite={false}
          toneMapped={false}
        />
      </mesh>

      <mesh ref={frame} position={[0, 0, -0.05]}>
        <boxGeometry args={[frameWidth + 0.18, frameHeight + 0.18, 0.14]} />
        <meshStandardMaterial
          color="#11100e"
          roughness={0.42}
          metalness={0.54}
          transparent
          opacity={1}
        />
      </mesh>

      <mesh ref={mat} position={[0, 0, 0.035]}>
        <boxGeometry args={[frameWidth, frameHeight, 0.055]} />
        <meshStandardMaterial
          color="#3e4740"
          roughness={0.92}
          metalness={0}
          transparent
          opacity={1}
        />
      </mesh>

      <mesh ref={image} position={[0, 0.1, 0.05]}>
        <boxGeometry args={[photoWidth, photoHeight, 0.035]} />
        <meshBasicMaterial
          map={texture}
          transparent
          opacity={1}
          toneMapped={false}
        />
      </mesh>

      <Text
        ref={title}
        position={[0, -frameHeight / 2 - 0.26, 0.08]}
        fontSize={0.105}
        color="#f5ebd9"
        anchorX="center"
        anchorY="top"
        maxWidth={frameWidth}
        textAlign="center"
        material-transparent
        material-opacity={1}
      >
        {photo.title}
      </Text>
    </group>
  );
}
