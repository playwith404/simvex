import { useMemo, useEffect } from 'react'
import { useGLTF } from '@react-three/drei'
import type { Group, Mesh, MeshStandardMaterial, Plane } from 'three'
import type { Part, ViewerMode } from '../../types'

type Props = {
  part: Part
  decompositionLevel: number
  isSelected: boolean
  isHovered: boolean
  onSelect: (partId: string) => void
  onHover: (partId: string | null) => void
  mode: ViewerMode
  onMeasurePoint?: (point: { x: number; y: number; z: number }) => void
  clippingPlanes: Plane[]
}

const highlightColor = '#f8c86a'
const measureHighlight = '#ff6b6b'

export const PartMesh = ({
  part,
  decompositionLevel,
  isSelected,
  isHovered,
  onSelect,
  onHover,
  mode,
  onMeasurePoint,
  clippingPlanes,
}: Props) => {
  const gltf = useGLTF(part.modelPath) as { scene: Group }

  const scene = useMemo(() => gltf.scene.clone(true), [gltf.scene])

  useEffect(() => {
    scene.traverse((child) => {
      const mesh = child as Mesh
      if (!mesh.isMesh) return
      const material = mesh.material as MeshStandardMaterial
      if (!material.userData.baseColor) {
        material.userData.baseColor = material.color.clone()
        material.userData.baseEmissive = material.emissive.clone()
      }
      const active = isSelected || isHovered
      if (active) {
        const color = mode === 'measure' && isSelected ? measureHighlight : highlightColor
        material.color.set(color)
        material.emissive.set(color)
        material.emissiveIntensity = isSelected ? 0.6 : 0.3
      } else {
        material.color.copy(material.userData.baseColor)
        material.emissive.copy(material.userData.baseEmissive)
        material.emissiveIntensity = 0
      }
      material.clippingPlanes = clippingPlanes
      material.clipShadows = true
      material.needsUpdate = true
    })
  }, [scene, isSelected, isHovered, clippingPlanes, mode])

  const offsetX = part.decomposeDirX * part.decomposeDistance * decompositionLevel
  const offsetY = part.decomposeDirY * part.decomposeDistance * decompositionLevel
  const offsetZ = part.decomposeDirZ * part.decomposeDistance * decompositionLevel

  return (
    <group
      position={[part.localPosX + offsetX, part.localPosY + offsetY, part.localPosZ + offsetZ]}
      onPointerOver={(e) => {
        e.stopPropagation()
        onHover(part.id)
      }}
      onPointerOut={(e) => {
        e.stopPropagation()
        onHover(null)
      }}
      onPointerDown={(e) => {
        e.stopPropagation()
        if (mode === 'measure' && onMeasurePoint) {
          onMeasurePoint({ x: e.point.x, y: e.point.y, z: e.point.z })
        } else {
          onSelect(part.id)
        }
      }}
    >
      <primitive object={scene} />
    </group>
  )
}
