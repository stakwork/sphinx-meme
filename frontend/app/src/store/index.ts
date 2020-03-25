import React from 'react'
import { ErrStore } from './errs'

const ctx = React.createContext({
  errStore: new ErrStore(),
})

export const useStores = () => React.useContext(ctx)