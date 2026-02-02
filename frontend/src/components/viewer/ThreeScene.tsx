import { Suspense, useMemo } from 'react'
import { Canvas } from '@react-three/fiber'
import { GizmoHelper, GizmoViewport, Environment } from '@react-three/drei'
import type { Part } from '../../types'
import { PartMesh } from './PartMesh'
import { CameraControls } from './CameraControls'

const defaultCamera = { x: 4, y: 4, z: 8 }

type Props = {
  parts: Part[]
  decompositionLevel: number
  selectedPartId: string | null
  hoveredPartId: string | null
  onSelectPart: (id: string) => void
  onHoverPart: (id: string | null) => void
  viewState?: {
    cameraPosition: { x: number; y: number; z: number }
    cameraTarget: { x: number; y: number; z: number }
    zoom: number
  }
  onViewStateChange: (state: {
    cameraPosition: { x: number; y: number; z: number }
    cameraTarget: { x: number; y: number; z: number }
    zoom: number
  }) => void
}

export const ThreeScene = ({
  parts,
  decompositionLevel,
  selectedPartId,
  hoveredPartId,
  onSelectPart,
  onHoverPart,
  viewState,
  onViewStateChange,
}: Props) => {
  const cameraPosition = useMemo<[number, number, number]>(
    () =>
      viewState?.cameraPosition
        ? [viewState.cameraPosition.x, viewState.cameraPosition.y, viewState.cameraPosition.z]
        : [defaultCamera.x, defaultCamera.y, defaultCamera.z],
    [viewState],
  )

  return (
    <Canvas camera={{ position: cameraPosition, fov: 45 }}>
      <ambientLight intensity={0.4} />
      <directionalLight position={[5, 8, 6]} intensity={0.9} />
      <directionalLight position={[-5, -2, -3]} intensity={0.4} />
      <Suspense fallback={null}>
        {parts.map((part) => (
          <PartMesh
            key={part.id}
            part={part}
            decompositionLevel={decompositionLevel}
            isSelected={selectedPartId === part.id}
            isHovered={hoveredPartId === part.id}
            onSelect={onSelectPart}
            onHover={onHoverPart}
          />
        ))}
        <Environment preset="studio" />
      </Suspense>
      <CameraControls viewState={viewState} onChange={onViewStateChange} />
      <GizmoHelper alignment="bottom-right" margin={[80, 80]}>
        <GizmoViewport axisColors={['#ff6b6b', '#4ecdc4', '#1a535c']} labelColor="white" />
      </GizmoHelper>
    </Canvas>
  )
}
