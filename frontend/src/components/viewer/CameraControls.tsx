import { useEffect, useRef } from 'react'
import { OrbitControls } from '@react-three/drei'
import { useThree } from '@react-three/fiber'
import { MOUSE } from 'three'

export type ViewState = {
  cameraPosition: { x: number; y: number; z: number }
  cameraTarget: { x: number; y: number; z: number }
  zoom: number
}

type Props = {
  viewState?: ViewState
  onChange: (state: ViewState) => void
}

export const CameraControls = ({ viewState, onChange }: Props) => {
  const controlsRef = useRef<any>(null)
  const { camera } = useThree()

  useEffect(() => {
    if (!viewState || !controlsRef.current) return
    camera.position.set(
      viewState.cameraPosition.x,
      viewState.cameraPosition.y,
      viewState.cameraPosition.z,
    )
    controlsRef.current.target.set(
      viewState.cameraTarget.x,
      viewState.cameraTarget.y,
      viewState.cameraTarget.z,
    )
    camera.zoom = viewState.zoom || 1
    camera.updateProjectionMatrix()
  }, [camera, viewState])

  return (
    <OrbitControls
      ref={controlsRef}
      makeDefault
      enableDamping
      dampingFactor={0.08}
      mouseButtons={{
        LEFT: MOUSE.PAN,
        MIDDLE: MOUSE.PAN,
        RIGHT: MOUSE.ROTATE,
      }}
      onChange={() => {
        if (!controlsRef.current) return
        const target = controlsRef.current.target
        onChange({
          cameraPosition: {
            x: camera.position.x,
            y: camera.position.y,
            z: camera.position.z,
          },
          cameraTarget: { x: target.x, y: target.y, z: target.z },
          zoom: camera.zoom,
        })
      }}
    />
  )
}
