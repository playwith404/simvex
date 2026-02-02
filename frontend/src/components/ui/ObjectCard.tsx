import { Link } from 'react-router-dom'
import type { ObjectModel } from '../../types'

export const ObjectCard = ({ item }: { item: ObjectModel }) => {
  return (
    <Link to={`/viewer/${item.id}`} className="object-card">
      <div className="object-card__image">
        <img src={item.thumbnail} alt={item.name} loading="lazy" />
      </div>
      <div className="object-card__body">
        <h3>{item.name}</h3>
        <p>{item.description}</p>
      </div>
    </Link>
  )
}
