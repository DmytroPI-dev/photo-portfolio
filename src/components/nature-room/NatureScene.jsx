import { useFrame, useThree } from "@react-three/fiber";
import { useRef } from "react";
import * as THREE from "three";
import NaturePanel from "./NaturePanel";

export default function NatureScene({
  photo,
  direction,
  phase,
  isExpanded,
  onToggleExpanded,
}) {
  const { camera, pointer } = useThree();
  const lightRig = useRef();

  useFrame((state, delta) => {
    // Keep the camera nearly fixed. The photo itself responds to pointer
    // movement, so the camera should only give a tiny parallax hint.
    camera.position.x = THREE.MathUtils.damp(
      camera.position.x,
      pointer.x * 0.16,
      2,
      delta
    );
    camera.position.y = THREE.MathUtils.damp(
      camera.position.y,
      0.2 + pointer.y * 0.08,
      2,
      delta
    );
    camera.lookAt(pointer.x * 0.1, 0.18 + pointer.y * 0.06, 0);

    if (lightRig.current) {
      // A small animated light drift prevents the foreground frame from feeling
      // pasted over the blurred background.
      lightRig.current.position.x = Math.sin(state.clock.elapsedTime * 0.35) * 0.18;
      lightRig.current.position.y = Math.cos(state.clock.elapsedTime * 0.28) * 0.08;
    }
  });

  return (
    <>
      <ambientLight intensity={0.36} color="#f7ead6" />
      <hemisphereLight intensity={0.42} color="#f8e7c7" groundColor="#090604" />

      <group ref={lightRig}>
        <spotLight
          position={[-2.3, 3.8, 3.2]}
          angle={0.52}
          penumbra={0.9}
          intensity={2.7}
          distance={8}
          color="#f2c48b"
        />
        <pointLight position={[2.2, 1.8, 2.8]} intensity={0.4} color="#cdd8ff" />
      </group>

      <NaturePanel
        key={photo?.id}
        photo={photo}
        direction={direction}
        phase={phase}
        isExpanded={isExpanded}
        onToggleExpanded={onToggleExpanded}
      />
    </>
  );
}
