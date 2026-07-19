import { useFrame, useThree } from "@react-three/fiber";
import { useMemo } from "react";
import * as THREE from "three";

const GOLDEN_RATIO = 1.61803398875;
const galleryOffset = new THREE.Vector3(0, -0.5, 0);
const overviewPosition = new THREE.Vector3(0, 0, 5.5);

export default function CameraRig({ selectedPlacement }) {
  const { camera, pointer } = useThree();
  const dummy = useMemo(() => new THREE.Object3D(), []);

  useFrame((_, delta) => {
    const selectedRotation = selectedPlacement
      ? new THREE.Euler(...selectedPlacement.rotation)
      : null;
    const selectedOffset =
      selectedPlacement && new THREE.Vector3(0, GOLDEN_RATIO / 2, 1.85)
        .applyEuler(selectedRotation)
        .add(new THREE.Vector3(...selectedPlacement.position))
        .add(galleryOffset);

    const targetPosition = selectedPlacement
      ? selectedOffset
      : overviewPosition.clone().add(new THREE.Vector3(pointer.x * 0.08, pointer.y * 0.05, 0));

    const targetQuaternion = selectedPlacement
      ? new THREE.Quaternion().setFromEuler(selectedRotation)
      : dummy.quaternion.identity().clone();

    camera.position.lerp(targetPosition, 1 - Math.exp(-delta * 2.5));
    camera.quaternion.slerp(targetQuaternion, 1 - Math.exp(-delta * 2.5));
  });

  return null;
}
