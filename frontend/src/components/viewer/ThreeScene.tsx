import { Suspense, useEffect, useMemo, useRef, type RefObject } from 'react'
import { Canvas, useThree } from '@react-three/fiber'
import { GizmoHelper, GizmoViewport, Environment } from '@react-three/drei'
import {
  Box3,
  Color,
  Group,
  LinearFilter,
  PerspectiveCamera,
  RGBAFormat,
  UnsignedByteType,
  Vector3,
  WebGLRenderTarget,
} from 'three'
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
  onCaptureReady?: (capture: () => PdfImageData | null) => void
}

const CaptureBridge = ({
  onCaptureReady,
  partsGroupRef,
}: {
  onCaptureReady?: (capture: () => PdfImageData | null) => void
  partsGroupRef: RefObject<Group | null>
}) => {
  const { gl, scene, camera } = useThree()
  const controls = useThree((state) => state.controls) as any

  useEffect(() => {
    if (!onCaptureReady) return

    onCaptureReady(() => {
      // Render to an offscreen target so we don't need preserveDrawingBuffer=true on the main canvas.
      // This keeps normal rendering lighter/safer and makes PDF capture deterministic across browsers.
      const renderToDataUrl = (width: number, height: number): string => {
        const prevTarget = gl.getRenderTarget()
        const prevAutoClear = gl.autoClear
        const prevClearAlpha = gl.getClearAlpha()
        const prevClearColor = gl.getClearColor(new Color())

        const target = new WebGLRenderTarget(width, height, {
          minFilter: LinearFilter,
          magFilter: LinearFilter,
          format: RGBAFormat,
          type: UnsignedByteType,
          depthBuffer: true,
          stencilBuffer: false,
        })

        const pixels = new Uint8Array(width * height * 4)

        try {
          gl.setRenderTarget(target)
          gl.autoClear = true
          gl.setClearColor(0x000000, 0)
          gl.clear(true, true, true)
          gl.render(scene, camera)
          gl.readRenderTargetPixels(target, 0, 0, width, height, pixels)
        } finally {
          gl.setRenderTarget(prevTarget)
          gl.autoClear = prevAutoClear
          gl.setClearColor(prevClearColor, prevClearAlpha)
          target.dispose()
        }

        // Flip Y (WebGL origin is bottom-left).
        const canvas = document.createElement('canvas')
        canvas.width = width
        canvas.height = height
        const ctx = canvas.getContext('2d')
        if (!ctx) return ''
        const imageData = ctx.createImageData(width, height)
        for (let y = 0; y < height; y++) {
          const srcRow = (height - 1 - y) * width * 4
          const dstRow = y * width * 4
          imageData.data.set(pixels.subarray(srcRow, srcRow + width * 4), dstRow)
        }
        ctx.putImageData(imageData, 0, 0)
        return canvas.toDataURL('image/png')
      }

      const baseW = gl.domElement.width || 1200
      const baseH = gl.domElement.height || 800
      const capW = Math.min(baseW, 1600)
      const capH = Math.max(1, Math.round((capW * baseH) / baseW))

      const group = partsGroupRef.current
      if (!group) return null

      const box = new Box3().setFromObject(group)
      if (box.isEmpty()) {
        const dataUrl = renderToDataUrl(capW, capH)
        if (!dataUrl) return null
        return { dataUrl, width: capW, height: capH }
      }

      const center = box.getCenter(new Vector3())
      const size = box.getSize(new Vector3())
      const margin = 1.2

      const prevPosition = camera.position.clone()
      const prevZoom = camera.zoom
      const prevNear = camera.near
      const prevFar = camera.far
      const prevTarget = controls?.target?.clone()

      if (!(camera instanceof PerspectiveCamera)) {
        const dataUrl = renderToDataUrl(capW, capH)
        if (!dataUrl) return null
        return { dataUrl, width: capW, height: capH }
      }

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

      const dataUrl = renderToDataUrl(capW, capH)

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

      if (!dataUrl) return null
      return { dataUrl, width: capW, height: capH }
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
  const partsGroupRef = useRef<Group | null>(null)

  return (
    <Canvas
      camera={{ position: cameraPosition, fov: 45 }}
      // Cap DPR to reduce GPU memory pressure (helps prevent Chrome "Aw, Snap" tab crashes on heavy WebGL scenes).
      dpr={[1, 1.5]}
      gl={{ preserveDrawingBuffer: false, antialias: false, powerPreference: 'high-performance' }}
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
