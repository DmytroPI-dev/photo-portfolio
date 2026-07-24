import { Float, Text } from "@react-three/drei";
import { useFrame, useThree } from "@react-three/fiber";
import { useMemo, useRef } from "react";
import * as THREE from "three";

const DISPLAY_FONT = "/fonts/sacramento.woff";

// These fragments suggest a frame that has come apart without needing a GLB
// asset or expensive transmission materials. Keeping the scene procedural also
// means the 404 stays small and reliable on the same lower-end GPUs targeted by
// the other gallery rooms.
const FRAME_FRAGMENTS = [
  { position: [-1.7, 0.9, 0.1], rotation: [0.1, 0.22, -0.2], size: [1.55, 0.15, 0.16] },
  { position: [1.72, 0.84, -0.1], rotation: [-0.08, -0.18, 0.3], size: [1.4, 0.15, 0.16] },
  { position: [-1.8, -0.8, 0], rotation: [-0.12, 0.12, 0.24], size: [1.35, 0.15, 0.16] },
  { position: [1.65, -0.92, 0.12], rotation: [0.1, -0.24, -0.15], size: [1.42, 0.15, 0.16] },
  { position: [-2.1, 0.02, -0.12], rotation: [0.15, 0.2, 0.08], size: [0.15, 1.3, 0.16] },
  { position: [2.08, -0.04, -0.05], rotation: [-0.16, -0.18, -0.12], size: [0.15, 1.42, 0.16] },
];

const SHARDS = [
  { position: [-2.4, 1.5, -0.4], rotation: [0.3, 0.6, 0.2], scale: 0.34, speed: 1.8 },
  { position: [2.45, 1.15, -0.65], rotation: [-0.4, 0.2, -0.5], scale: 0.28, speed: 2.1 },
  { position: [-2.28, -1.45, -0.3], rotation: [0.5, -0.4, 0.3], scale: 0.25, speed: 1.5 },
  { position: [2.32, -1.5, -0.2], rotation: [-0.3, 0.5, 0.1], scale: 0.3, speed: 1.9 },
  { position: [0.1, 1.75, -0.8], rotation: [0.2, -0.35, 0.5], scale: 0.2, speed: 2.3 },
];

function FrameFragment({ position, rotation, size, index }) {
  return (
    <Float speed={1.1 + index * 0.08} rotationIntensity={0.16} floatIntensity={0.22}>
      <mesh position={position} rotation={rotation}>
        <boxGeometry args={size} />
        <meshStandardMaterial
          color="#1c1b19"
          emissive="#16120d"
          emissiveIntensity={0.38}
          metalness={0.7}
          roughness={0.3}
        />
      </mesh>
    </Float>
  );
}

function GlassShard({ position, rotation, scale, speed }) {
  return (
    <Float speed={speed} rotationIntensity={0.38} floatIntensity={0.32}>
      <mesh position={position} rotation={rotation} scale={scale}>
        <tetrahedronGeometry args={[1, 0]} />
        <meshStandardMaterial
          color="#6f8190"
          emissive="#283746"
          emissiveIntensity={0.7}
          metalness={0.82}
          roughness={0.18}
          transparent
          opacity={0.76}
        />
      </mesh>
    </Float>
  );
}

export default function NotFoundScene() {
  const scene = useRef();
  const { pointer } = useThree();
  const warmLight = useMemo(() => new THREE.Color("#d6ae77"), []);

  useFrame((state, delta) => {
    if (!scene.current) return;

    // Pointer movement gives the otherwise static error page a quiet museum-
    // display parallax. Damping avoids an abrupt response on trackpads.
    scene.current.rotation.y = THREE.MathUtils.damp(
      scene.current.rotation.y,
      pointer.x * 0.16,
      3.4,
      delta
    );
    scene.current.rotation.x = THREE.MathUtils.damp(
      scene.current.rotation.x,
      pointer.y * -0.08,
      3.4,
      delta
    );
    scene.current.position.y = Math.sin(state.clock.elapsedTime * 0.45) * 0.05;
  });

  return (
    <>
      <ambientLight intensity={0.28} />
      <hemisphereLight args={["#526776", "#070606", 0.5]} />
      <spotLight
        position={[-2.8, 3.4, 4]}
        angle={0.62}
        penumbra={1}
        intensity={4.2}
        color={warmLight}
      />
      <pointLight position={[2.4, -0.8, 2.8]} intensity={1.4} color="#6a9bc0" />

      <group ref={scene}>
        {FRAME_FRAGMENTS.map((fragment, index) => (
          <FrameFragment key={index} {...fragment} index={index} />
        ))}
        {SHARDS.map((shard, index) => (
          <GlassShard key={index} {...shard} />
        ))}

        <Text
          font={DISPLAY_FONT}
          position={[0, 0.08, 0]}
          fontSize={2.65}
          color="#f1ebe0"
          anchorX="center"
          anchorY="middle"
          outlineWidth={0.012}
          outlineColor="#161311"
        >
          404
        </Text>
        <Text
          font={DISPLAY_FONT}
          position={[0, -0.85, 0.02]}
          fontSize={0.38}
          color="#cbbda8"
          anchorX="center"
          anchorY="middle"
        >
          missing frame
        </Text>
      </group>
    </>
  );
}
