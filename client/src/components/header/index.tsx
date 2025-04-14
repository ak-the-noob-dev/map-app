import React, { ReactNode } from 'react'
import { LanguageSelector } from '../language-selector'
import { Button } from '../ui/button'
import { useTranslation } from 'react-i18next'
import { Github } from 'lucide-react'

interface IProps {
  leftNode?: ReactNode
}
export function Header(props: IProps) {
  const { t } = useTranslation()

  return (
    <div className="fixed left-0 top-0 flex w-full items-center justify-between border bg-slate-50 bg-opacity-70 px-4 py-4 md:px-12"></div>
  )
}
