import { Helmet } from 'react-helmet'
import { useTranslation } from 'react-i18next'
import { Link } from 'react-router-dom'
import { Hero } from 'src/components/hero'

export default function Home() {
  const { t } = useTranslation('translation')
  return (
    <>
      <Helmet>
        <title>{t('title')}</title>
      </Helmet>
      <div className="flex h-screen flex-col items-center justify-center">
        <Link to="/maps">Maps</Link>
      </div>
    </>
  )
}
