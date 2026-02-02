import { Suspense, useEffect, useMemo, useRef, type RefObject } from 'react'
import { Canvas, useThree } from '@react-three/fiber'
import { GizmoHelper, GizmoViewport, Environment } from '@react-three/drei'
import { Box3, Vector3, Group } from 'three'
import type { Part } from '../../types'
import type { PdfImageData } from '../../types/pdf'
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
  onCanvasReady?: (canvas: HTMLCanvasElement) => void
  onCaptureReady?: (capture: () => Promise<PdfImageData | null>) => void
}

const CaptureBridge = ({
  onCaptureReady,
  partsGroupRef,
}: {
  onCaptureReady?: (capture: () => Promise<PdfImageData | null>) => void
  partsGroupRef: RefObject<Group>
}) => {
  const { gl, scene, camera } = useThree()
  const controls = useThree((state) => state.controls) as any

  useEffect(() => {
    if (!onCaptureReady) return

    onCaptureReady(async () => {
      const group = partsGroupRef.current
      if (!group) return null

      const box = new Box3().setFromObject(group)
      if (box.isEmpty()) {
        return {
          dataUrl: gl.domElement.toDataURL('image/png'),
          width: gl.domElement.width,
          height: gl.domElement.height,
        }
      }

      const center = box.getCenter(new Vector3())
      const size = box.getSize(new Vector3())
      const margin = 1.2

      const prevPosition = camera.position.clone()
      const prevZoom = camera.zoom
      const prevNear = camera.near
      const prevFar = camera.far
      const prevTarget = controls?.target?.clone()

      const fov = (camera.fov * Math.PI) / 180
      const aspect = camera.aspect || gl.domElement.width / gl.domElement.height
      const hFov = 2 * Math.atan(Math.tan(fov / 2) * aspect)
      const fitWidthDistance = (size.x / 2) / Math.tan(hFov / 2)
      const fitHeightDistance = (size.y / 2) / Math.tan(fov / 2)
      const distance = Math.max(fitWidthDistance, fitHeightDistance) * margin

      const target = prevTarget ?? new Vector3(0, 0, 0)
      const direction = new Vector3().subVectors(camera.position, target)
      if (direction.lengthSq() === 0) {
        direction.set(0, 0, 1)
      } else {
        direction.normalize()
      }

      camera.position.copy(center).add(direction.multiplyScalar(distance))
      camera.zoom = 1
      camera.near = distance / 100
      camera.far = distance * 100
      camera.updateProjectionMatrix()

      if (controls?.target) {
        controls.target.copy(center)
        controls.update?.()
      }

      gl.render(scene, camera)
      const dataUrl = gl.domElement.toDataURL('image/png')

      camera.position.copy(prevPosition)
      camera.zoom = prevZoom
      camera.near = prevNear
      camera.far = prevFar
      camera.updateProjectionMatrix()

      if (controls?.target && prevTarget) {
        controls.target.copy(prevTarget)
        controls.update?.()
      }

      gl.render(scene, camera)

      return {
        dataUrl,
        width: gl.domElement.width,
        height: gl.domElement.height,
      }
    })
  }, [onCaptureReady, partsGroupRef, gl, scene, camera, controls])

  return null
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
  onCanvasReady,
  onCaptureReady,
}: Props) => {
  const cameraPosition = useMemo<[number, number, number]>(
    () =>
      viewState?.cameraPosition
        ? [viewState.cameraPosition.x, viewState.cameraPosition.y, viewState.cameraPosition.z]
        : [defaultCamera.x, defaultCamera.y, defaultCamera.z],
    [viewState],
  )
  const partsGroupRef = useRef<Group>(null)

  return (
    <Canvas
      camera={{ position: cameraPosition, fov: 45 }}
      gl={{ preserveDrawingBuffer: true, antialias: true }}
      onCreated={({ gl }) => {
        onCanvasReady?.(gl.domElement)
      }}
    >
      <ambientLight intensity={0.4} />
      <directionalLight position={[5, 8, 6]} intensity={0.9} />
      <directionalLight position={[-5, -2, -3]} intensity={0.4} />
      <Suspense fallback={null}>
        <group ref={partsGroupRef}>
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
        </group>
        <Environment preset="studio" />
      </Suspense>
      <CaptureBridge onCaptureReady={onCaptureReady} partsGroupRef={partsGroupRef} />
      <CameraControls viewState={viewState} onChange={onViewStateChange} />
      <GizmoHelper alignment="bottom-right" margin={[80, 80]}>
        <GizmoViewport axisColors={['#ff6b6b', '#4ecdc4', '#1a535c']} labelColor="white" />
      </GizmoHelper>
    </Canvas>
  )
}
