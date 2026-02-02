import { useEffect, useState } from 'react'
import { fetchObjects } from '../services/api'
import type { ObjectModel } from '../types'
import { ObjectCard } from '../components/ui/ObjectCard'

export const ObjectList = () => {
  const [objects, setObjects] = useState<ObjectModel[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    fetchObjects()
      .then(setObjects)
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false))
  }, [])

  return (
    <section className="object-list">
      <div className="object-list__header">
        <h2>학습할 오브젝트 선택</h2>
        <p>기계 구조를 3D로 탐색하고 학습 플로우를 시작하세요.</p>
      </div>
      {loading && <p>오브젝트를 불러오는 중...</p>}
      {error && <p className="error">{error}</p>}
      <div className="object-list__grid">
        {objects.map((item) => (
          <ObjectCard key={item.id} item={item} />
        ))}
      </div>
    </section>
  )
}
