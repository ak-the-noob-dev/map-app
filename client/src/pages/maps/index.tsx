import MapView from 'src/components/map/view'
import { Helmet } from 'react-helmet'

export default function Maps() {
  return (
    <div className="relative h-screen w-full">
      <Helmet>
        <title>Maps</title>
      </Helmet>
      <MapView />
    </div>
  )
}
